//
// MIT License
//
// Copyright (c) 2026 Nicolas JUHEL
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//

// Package astgol implements Go module AST parsing.
//
// It provides a step-based pipeline for reading go.mod, parsing local and vendor
// packages, and performing global dependency resolution across all cached ASTs.
// The four public steps are Step1, Step2, Step3, and Step4.
package astgol

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	audids "github.com/nabbar/auditor/pkg/data/id"
	auddbm "github.com/nabbar/auditor/pkg/data/manager"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audtps "github.com/nabbar/auditor/pkg/data/types"
	auduxi "github.com/nabbar/auditor/pkg/uxi"
	semtps "github.com/nabbar/golib/semaphore/types"
	astgop "golang.org/x/tools/go/packages"
)

// astCache holds the AST and package ID for a single parsed source file.
type astCache struct {
	p audids.ID
	a *ast.File
}

// mdl represents a Go module AST parser implementation.
//
// This struct holds the state for parsing Go modules and packages, including file paths,
// module information, and cached source code content. The mdl struct serves as the core
// data structure for managing the AST parsing process for Go language files within the
// auditor tool. It maintains all necessary state information required for processing
// Go modules and their constituent packages.
//
// The mdl struct implements the astgen.AST interface and provides a complete implementation
// for Go language AST parsing operations. It manages the lifecycle of AST processing for Go modules,
// including module parsing, package analysis, and vendor package handling.
type mdl struct {
	// x is a read-write mutex protecting concurrent access to all fields in this struct
	x sync.RWMutex

	// u is the UXI Manager instance used for logging and user interaction
	u auduxi.Manager

	// a is the AST manager for handling collections of AST elements and managing workflow
	a *auddbm.Linker

	// d is the directory path where Go module files are located
	d string

	// f is the filename of the go.mod file being parsed
	f string

	// m is a function that retrieves the module information once it's been parsed
	m func() audmod.Module

	// l is the list of module paths discovered in go.mod (dependencies of this module)
	l []string

	// c is a cache for source code content to avoid repeated file reads
	c map[string][]byte

	// s is a cache for AST content to avoid repeated AST parsing, organized as Filepath => astCache{pkg id, *ast.File}
	s map[string]astCache

	// h is a handler map to analyze dedicated type
	h map[audtps.CodeType]map[bool]func(context.Context, audids.ID, string, *ast.File, ast.Node)
}

// Close cleans up the AST parser state by clearing all cached data and references.
//
// This method implements the io.Closer interface and ensures proper resource cleanup.
// The Close function serves as the primary mechanism for cleaning up resources when
// the AST parser is no longer needed. It clears all cached data and references to prevent
// memory leaks and ensure proper state management. This method is essential for maintaining
// clean state after AST processing operations are completed.
//
// Key features:
//   - Clears the source code cache (c) to free up memory and prevent memory leaks
//   - Resets the module retrieval function (m) to nil to prevent dangling references and maintain data integrity
//   - Clears filename (f) and directory (d) fields to reset state and ensure clean restart conditions
//   - Implements io.Closer interface compliance for proper resource management and lifecycle control
//
// Returns:
//   - error: Any error encountered during cleanup operations. While this implementation
//     currently returns nil, it maintains compatibility with the io.Closer interface.
func (o *mdl) Close() error {
	o.x.Lock()
	defer o.x.Unlock()

	o.c = nil
	o.m = nil
	o.f = ""
	o.d = ""

	return nil
}

// GetFile returns the filename of the go.mod file being parsed.
//
// This method provides access to the name of the Go module file that this AST parser is currently handling.
// It retrieves the filename from the internal state, ensuring thread-safe access through the read mutex.
//
// Returns:
//   - string: The name of the go.mod file being processed.
func (o *mdl) GetFile() string {
	return o.getFileName()
}

// GetDirectory returns the directory path where Go module files are located.
//
// This method provides access to the working directory path for the Go module files.
// It retrieves the directory from the internal state, ensuring thread-safe access through the read mutex.
//
// Returns:
//   - string: The directory path where the go.mod file is located.
func (o *mdl) GetDirectory() string {
	return o.getDirWork()
}

// GetFullPath returns the full path to the go.mod file being parsed.
//
// This method constructs and returns the complete file path for the go.mod file by joining
// the directory path with the filename. It ensures thread-safe access through the read mutex.
//
// Returns:
//   - string: The full path to the go.mod file.
func (o *mdl) GetFullPath() string {
	return o.getFilePath()
}

// Step1 initializes the module by reading go.mod, parsing the module name,
// and registering it in the AST manager. It also discovers and registers
// vendor module dependencies listed in go.mod.
func (o *mdl) Step1(ctx context.Context) error {
	var (
		e error
		c []byte
		n string
		l []string
		m audmod.Module
		i audids.ID

		of = o.getFileName()
		op = o.getFilePath()
	)

	if e = o.initVendor(ctx); e != nil {
		return e
	}

	if c, e = o.getFileSrc(op); e != nil {
		return e
	}

	n, l = o.parseGoMod(c)
	if len(n) == 0 {
		return fmt.Errorf("module not found in go.mod file %s", of)
	}

	m = o.a.ModNew(n, op)

	if m == nil {
		return errors.New("module cannot be registered")
	}

	i = m.GetID()

	m.SetStatusParsed()
	m.UnSetStatusOutdated()

	o.a.ModAdd(m)
	o.setMod(i)

	o.u.Info("[Path: %s] added new module: '%s'", op, m.GetMod())

	n = o.getMod().GetMod()
	o.l = make([]string, 0, len(l))

	for _, p := range l {
		if len(p) < 1 {
			continue
		}

		if strings.HasPrefix(p, n) {
			continue
		}

		m = o.a.ModNewOrMerge(p, "")
		o.u.Info("[Path: %s] new or get vendor module: '%s'", op, m.GetMod())

		m.SetStatusParsed()
		if m.IsVendor() {
			m.UnSetStatusOutdated()
		}

		o.l = append(o.l, m.GetMod())
	}

	return nil
}

// Step2 parses vendor packages within the module.
// Step1 must be called successfully before invoking this method.
func (o *mdl) Step2(ctx context.Context) error {
	var mod = o.getMod()

	// Validate that module has been initialized before attempting package parsing
	if mod == nil || mod.IsEmpty() {
		return errors.New("module not initialized, call Step1 before other Step")
	}

	return o.parsePackage(ctx, true)
}

// Step3 parses local (non-vendor) packages within the module.
// Step1 & Step2 must be called successfully before invoking this method.
func (o *mdl) Step3(ctx context.Context) error {
	return o.parsePackage(ctx, false)
}

// Step4 performs global dependency resolution across all cached ASTs
// by walking every declaration in each cached file.
func (o *mdl) Step4(ctx context.Context) error {
	var lst = o.listAST()

	if len(lst) < 1 {
		return nil
	}

	o.u.Info("[Step 4] Starting global dependency resolution across cached ASTs")

	for filePath, ach := range lst {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		for _, adc := range ach.a.Decls {
			switch i := adc.(type) {
			case *ast.FuncDecl:
				o.processNode(ctx, ach.p, filePath, ach.a, i, true)
			case *ast.GenDecl:
				o.processNode(ctx, ach.p, filePath, ach.a, i, true)
			}
		}
	}

	o.u.Info("[Step 4] Successfully completed global dependency resolution")
	return nil
}

// parsePackage parses all source files in a module, either local or vendor.
// When isVendor is true it walks the vendor packages; otherwise it walks the
// local packages. It loads packages via x/tools/go/packages, registers each
// package in the AST manager, and then processes every file concurrently.
func (o *mdl) parsePackage(ctx context.Context, isVendor bool) error {
	var (
		od = o.getDirWork()

		err error
		bar semtps.SemBar
		mpk int
		mod audmod.Module
		nok = atomic.Bool{}

		alp []*astgop.Package
		cfg = &astgop.Config{
			Mode:       astgop.NeedName | astgop.NeedFiles | astgop.NeedSyntax | astgop.NeedTypes,
			Context:    ctx,
			Logf:       o.u.Info,
			Dir:        od,
			Env:        nil,
			BuildFlags: nil,
			Fset:       nil,
			ParseFile:  nil,
			Tests:      false,
			Overlay:    nil,
		}

		alk *astgop.Package
		btl string
		ptn []string
		vdr = "packages"
	)

	defer func() {
		if bar != nil {
			bar.DeferMain()
		}
	}()

	// Retrieve the parsed module information for validation and access
	if mod = o.getMod(); mod == nil || mod.GetID() == 0 {
		return errors.New("module not initialized, call Step1 before other Step")
	}

	if isVendor {
		vdr = "vendor packages"
		btl = "Module " + mod.GetMod() + " (vendor)"
	} else {
		btl = "Module " + mod.GetMod()
	}

	// Generate package list based on the module name and directory path
	if ptn, err = o.getPkgList(mod.GetMod(), od, isVendor); err != nil {
		return err
	}

	if len(ptn) < 1 {
		o.u.Info("[parsePackage][Path: %s] No %s found", od, vdr)
		return nil
	}

	// Load all packages in the specified directory with required configurations for AST analysis
	if alp, err = astgop.Load(cfg, ptn...); err != nil {
		o.u.ErrorStack(fmt.Sprintf("[parsePackage][Path: %s] Cannot start AST for Golang", od), err)
		return err
	}

	// Validate that packages were successfully loaded to prevent empty processing
	if len(alp) < 1 {
		if !isVendor {
			err = errors.New("no " + vdr + " not found in path " + cfg.Dir)
			o.u.ErrorStack(fmt.Sprintf("[parsePackage][Path: %s] error trigger while parsing %s in module", od, vdr), err)
			return err
		}

		o.u.Info("[parsePackage][Path: %s] No AST %s found", od, vdr)
		return nil
	}

	// Count all syntax files across all packages for progress tracking and UI updates
	for i := 0; i < len(alp); i++ {
		if len(alp[i].Syntax) < 1 {
			continue
		}

		mpk += len(alp[i].Syntax)
	}

	// Initialize the progress bar with package count for visual feedback during processing
	bar = o.u.NewBar(btl, mpk)
	o.u.Info("[parsePackage][Path: %s] Found %d Entries for %d %s", od, mpk, len(alp), vdr)

	// Create package references quickly for all loaded packages to ensure proper tracking and management
	for _, alk = range alp {
		if len(alk.Syntax) < 1 {
			continue
		}

		var (
			mid audids.ID
			pth = alk.PkgPath
		)

		if strings.Contains(pth, "/vendor/") {
			pth = pth[strings.Index(pth, "/vendor/")+8:]
		}

		pth, _ = audmod.CleanPath(pth)

		if mid, pth = o.findMod(pth); mid < 1 {
			err = errors.New("module not found for path " + pth)
			o.u.ErrorStack(fmt.Sprintf("[parsePackage][Path: %s] Error on find module for package path %s, name %s", od, alk.PkgPath, alk.Name), err)
			continue
		}

		o.u.Info("[parsePackage][Path: %s] Adding package '%s' for module id %d", od, pth, mid)
		p := o.a.PkgNewOrMerge(mid, pth)

		if p != nil {
			p.SetStatusParsed()
			// Package (vendor or not) are unset outdated
			// engine will clean outdated entries
			// engine will clean package with no entry
			p.UnSetStatusOutdated()

			o.a.PkgAdd(p)
			o.u.Info("[parsePackage][Path: %s] Package '%s' (id: %d) updated for module '%s' (id: %d), vendor: %v", od, p.GetFullPath(), p.GetID(), p.GetMod().GetMod(), p.GetModId(), p.IsVendor())
		}
	}

	// Process each package to analyze their files and declarations
	for _, alk = range alp {
		if len(alk.Syntax) < 1 {
			continue
		}

		var pth = alk.PkgPath

		if strings.Contains(pth, "/vendor/") {
			pth = pth[strings.Index(pth, "/vendor/")+8:]
		}

		pth, _ = audmod.CleanPath(pth)

		var pkg = o.a.PkgSearch(pth)

		// Validate that the package was found in the collection before processing
		if pkg == nil || pkg.IsEmpty() {
			o.u.ErrorStack(fmt.Sprintf("[parsePackage][Path: %s] Error on find %s path %s with name %s", od, vdr, alk.PkgPath, alk.Name), errors.New("package not found for path "+pth))
			bar.Inc(len(alk.Syntax)) // increase the counter to prevent incomplete bar total
			continue
		}

		// Process each file within the package for detailed AST analysis
		for j, f := range alk.Syntax {
			if ctx.Err() != nil {
				return nil // use deferMain to stop goroutines
			}

			if bar.Err() != nil {
				return nil // use deferMain to stop goroutines
			}

			// Create a new worker for concurrent processing
			if err = bar.NewWorker(); err != nil {
				o.u.ErrorStack(fmt.Sprintf("[parsePackage][Path: %s] error while create new worker", od), err)
				return err // use deferMain to stop goroutines
			}

			// Real file path extraction for logging and error reporting
			fn := alk.GoFiles[j]

			// Add AST Syntax to cache
			if !isVendor {
				// add into cache for local package
				// skipping vendoring
				o.addAST(fn, pkg.GetID(), f)
			}

			// Launch concurrent processing of the file using goroutines for parallel execution
			fct := func(x context.Context, i audids.ID, n string, a *ast.File) {
				defer bar.DeferWorker()

				o.u.Info("[parsePackage][File %s] Parsing file", fn)
				if e := o.parseFile(x, i, n, a); e != nil {
					o.u.ErrorStack(fmt.Sprintf("[parsePackage][Path: %s] error while parsing file %s", od, n), e)
					nok.Store(true)
					return
				}
			}

			go fct(bar, pkg.GetID(), fn, f)
		}
	}

	// Wait for all concurrent processing to complete before proceeding
	if err = bar.WaitAll(); err != nil {
		o.u.ErrorStack(fmt.Sprintf("[parsePackage][Path: %s] error while waiting worker", od), err)
		return err
	}

	// Force update the progress bar to ensure completion status is properly displayed
	for !bar.Completed() {
		bar.Inc(1)
		time.Sleep(time.Millisecond)
	}

	// Wait for UI updates to complete before returning control
	time.Sleep(500 * time.Millisecond)

	// Check if any errors occurred during processing and return appropriate error
	if nok.Load() {
		err = errors.New("error trigger on Parsing Module")
		o.u.ErrorStack(fmt.Sprintf("[parsePackage][Path: %s] error while parsing module", od), err)
		return err
	}

	return nil
}
