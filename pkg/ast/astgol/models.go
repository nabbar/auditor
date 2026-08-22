/*
 * MIT License
 *
 * Copyright (c) 2026 Nicolas JUHEL
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

// Package astgol provides Go-specific AST parsing capabilities for the auditor tool.
//
// This package implements specialized Abstract Syntax Tree (AST) parsing functionality
// specifically designed for Go programming language files. It handles the identification
// and parsing of Go module files (go.mod) and Go source code files (.go). The package
// provides a structured approach to analyzing Go language constructs through AST-based
// parsing techniques.
//
// Key features include:
//   - Language identification for Go module files (go.mod)
//   - AST creation for Go source code files
//   - Pattern matching for Go file extensions and naming conventions
//   - Integration with the general AST framework for consistent language handling
package astgol

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	audids "github.com/nabbar/auditor/pkg/data/id"
	auddbm "github.com/nabbar/auditor/pkg/data/manager"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	auduxi "github.com/nabbar/auditor/pkg/uxi"
	semtps "github.com/nabbar/golib/semaphore/types"
	astgop "golang.org/x/tools/go/packages"
)

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

	// c is a cache for source code content to avoid repeated file reads
	c map[string][]byte
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

// ParseModule parses the Go module information from the go.mod file.
//
// This function reads the go.mod file and extracts the module name, returning a Module object.
// It also stores the module in the collection for later retrieval. The ParseModule function
// serves as the entry point for extracting module-level information from Go module files
// (go.mod). It processes the file to identify the module name and creates appropriate data
// structures for storage in the auditor's collection system.
//
// Key features:
//   - Reads the go.mod file content from disk using getFileSrc method, which caches file contents for performance
//   - Parses the module name using string manipulation techniques with byte-level operations to extract the module identifier
//   - Creates a Module object with appropriate naming and file path information for tracking and management
//   - Stores the module in the manager's collection for later retrieval via ModAdd, ensuring persistent storage
//   - Implements proper error handling for missing or malformed module declarations to prevent silent failures
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during parsing operations.
//
// Returns:
//   - audmod.Module: The parsed Module object containing module information, or nil if parsing fails.
//   - error: Any error encountered during parsing operations, including file read errors or missing module declarations.
func (o *mdl) ParseModule(ctx context.Context) (audmod.Module, error) {
	var (
		e error
		c []byte
		l []byte
		f [][]byte
		m audmod.Module
		i audids.ID

		of = o.getFileName()
		od = o.getDirWork()
		op = o.getFilePath()
	)

	if c, e = o.getFileSrc(op); e != nil {
		return nil, e
	}

	o.u.Info("[Path: %s] Preparing vendor folder", od)

	commands := [][]string{
		{"mod", "tidy"},
		{"mod", "download", "all"},
		{"mod", "vendor"},
	}

	for _, args := range commands {
		cmd := exec.CommandContext(ctx, "go", args...) // #nosec

		cmd.Dir = od // changing working directory to go.mod root folder

		// catching stdout to add information
		if out, err := cmd.CombinedOutput(); err != nil {
			o.u.ErrorStack(fmt.Sprintf("go %s failed: %s", strings.Join(args, " "), string(out)), err)
			return nil, err
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}

	o.u.Info("[Path: %s] Successfully vendoring folder", od)

	for _, l = range bytes.Split(c, []byte("\n")) {
		l = bytes.TrimSpace(l)

		if bytes.HasPrefix(l, []byte("module ")) {
			f = bytes.Fields(l)

			if len(f) < 2 {
				o.u.Info("[File: %s] invalid module match in file: '%s'", op, l)
				continue
			}

			if ctx.Err() != nil {
				return nil, ctx.Err()
			}

			m = o.a.ModNew(string(f[1]), op)

			if m == nil {
				return nil, errors.New("module cannot be registered")
			}

			i = m.GetID()
			o.a.ModAdd(m)
			o.setMod(i)

			o.u.Info("[Path: %s] added new module: '%s'", op, m.GetMod())
			return m, e
		}
	}

	return nil, fmt.Errorf("module not found in go.mod file %s", of)
}

// ParsePackage parses all Go packages within the module directory.
//
// This function uses the golang.org/x/tools/go/packages package to load and parse
// all packages in the directory, then processes each package's files to extract
// functions, types, and other declarations. The ParsePackage function serves as the
// core mechanism for processing all packages within a Go module directory. It leverages
// the golang.org/x/tools/go/packages library to load and parse package information,
// then iterates through each package's files to extract detailed AST information.
//
// Key features:
//   - Uses astgop.Load with appropriate configuration modes (NeedName, NeedFiles, NeedSyntax, NeedTypes)
//     for comprehensive package analysis, ensuring all required information is collected
//   - Processes all packages in the module directory using the configured directory path to ensure complete coverage
//   - Handles package-level parsing for each file within packages, maintaining granular control over individual files
//   - Implements error handling for package loading failures and empty package collections to prevent silent failures
//   - Uses the parseFile function to process individual source files with proper error propagation for detailed analysis
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during package processing.
//
// Returns:
//   - error: Any error encountered during package processing, including package loading or parsing errors.
//     This includes cases where packages cannot be loaded from the directory or where individual file parsing fails.
func (o *mdl) ParsePackage(ctx context.Context) error {
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
		ptn []string
	)

	defer func() {
		if bar != nil {
			bar.DeferMain()
		}
	}()

	// Validate that module has been initialized before attempting package parsing
	if o.getModId() == 0 {
		return errors.New("module initialized, call ParseModule before ParsePackage")
	}

	// Retrieve the parsed module information for validation and access
	if mod = o.getMod(); mod == nil || mod.GetID() == 0 {
		return errors.New("module initialized, call ParseModule before ParsePackage")
	}

	// Generate package list based on the module name and directory path
	if ptn, err = o.getPkgList(mod.GetMod(), od, false); err != nil {
		return err
	}

	// Load all packages in the specified directory with required configurations for AST analysis
	if alp, err = astgop.Load(cfg, ptn...); err != nil {
		o.u.ErrorStack("cannot start AST for Golang", err)
		return err
	}

	// Validate that packages were successfully loaded to prevent empty processing
	if len(alp) < 1 {
		err = errors.New("no package not found in path " + cfg.Dir)
		o.u.ErrorStack("error trigger while parsing package in module", err)
		return err
	}

	// Count all syntax files across all packages for progress tracking and UI updates
	for i := 0; i < len(alp); i++ {
		if len(alp[i].Syntax) < 1 {
			continue
		}

		mpk += len(alp[i].Syntax)
	}

	// Initialize the progress bar with package count for visual feedback during processing
	bar = o.u.NewBar("Module "+mod.GetMod(), mpk)

	o.u.Info("[Path: %s] Found %d Entries for %d Packages", od, mpk, len(alp))

	// Create package references quickly for all loaded packages to ensure proper tracking and management
	for _, alk = range alp {
		if len(alk.Syntax) < 1 {
			continue
		}

		_ = o.a.PkgNewOrMerge(o.getModId(), alk.PkgPath)
	}

	// Process each package to analyze their files and declarations
	for _, alk = range alp {
		if len(alk.Syntax) < 1 {
			continue
		}

		var pkg = o.a.PkgSearch(alk.PkgPath)

		// Validate that the package was found in the collection before processing
		if pkg == nil || pkg.IsEmpty() {
			o.u.Error("no package found in path %s", alk.PkgPath)
			bar.Inc(len(alk.Syntax)) // increase the counter to prevent incomplete bar total
			continue
		}

		// Process each file within the package for detailed AST analysis
		for j, f := range alk.Syntax {
			if ctx.Err() != nil {
				o.u.ErrorStack("context error", ctx.Err())
				return ctx.Err() // use deferMain to stop goroutines
			}

			if bar.Err() != nil {
				o.u.ErrorStack("context error", bar.Err())
				return bar.Err() // use deferMain to stop goroutines
			}

			// Create a new worker for concurrent processing
			if err = bar.NewWorker(); err != nil {
				o.u.ErrorStack("error while create new worker", err)
				return err // use deferMain to stop goroutines
			}

			// Real file path extraction for logging and error reporting
			fn := alk.GoFiles[j]

			// Launch concurrent processing of the file using goroutines for parallel execution
			go func(x context.Context, i audids.ID, n string, a *ast.File) {
				defer bar.DeferWorker()

				o.u.Info("[File %s] Parsing file", fn)
				if e := o.parseFile(x, i, n, a); e != nil {
					o.u.ErrorStack("parsing file error: "+n, e)
					nok.Store(true)
					return
				}
			}(bar, pkg.GetID(), fn, f)
		}
	}

	// Wait for all concurrent processing to complete before proceeding
	if err = bar.WaitAll(); err != nil {
		o.u.ErrorStack("error while waiting worker", err)
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
		o.u.Error("error trigger on Parsing Module")
		return err
	}

	return nil
}

// UpdatePackage updates package information by analyzing vendor dependencies.
//
// This function orchestrates the process of updating package information for vendor dependencies.
// It collects local packages, maps vendor dependencies, and processes vendor packages concurrently.
// The UpdatePackage function serves as a mechanism for maintaining updated information about
// vendor dependencies in the Go module's package collection. It ensures that package data reflects
// the latest state of vendor dependencies by analyzing their AST structures.
//
// Key features:
//   - Collects all local packages belonging to the current module to identify dependencies
//   - Maps vendor package IDs to slices of vendor entry IDs for efficient dependency tracking
//   - Processes vendor packages concurrently using semaphore-based parallelism for performance
//   - Maintains thread safety and proper resource management throughout the update process
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during update operations.
//
// Returns:
//   - error: Any error encountered during package update processing, including dependency collection or analysis errors.
func (o *mdl) UpdatePackage(ctx context.Context) error {
	var (
		nbr int
		lpi []audids.ID
		ptn map[audids.ID][]audids.ID

		mid = o.getModId()
	)

	if mid == 0 {
		return errors.New("module initialized, call ParseModule before ParsePackage")
	}

	// 1. Collecting local packages
	if lpi = o.collectModulePackages(mid); len(lpi) < 1 {
		return nil
	}

	// 2. Mapping vendor dependencies: Vendor Pkg ID -> []Vendor Entry IDs
	if ptn, nbr = o.collectVendorDependencies(lpi); len(ptn) < 1 || nbr == 0 {
		return nil
	}

	// 3. Processing vendor packages concurrently
	return o.getVdrPkg(ctx, ptn, nbr) // placeholder waiting finish implementation
}
