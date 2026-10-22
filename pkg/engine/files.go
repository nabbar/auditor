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

// Package engine orchestrates the multi-pass structural analysis, asynchronous
// dependency resolution, linguistic LLM reasoning, and final report generation
// for automated source code auditing.
package engine

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	audast "github.com/nabbar/auditor/pkg/ast"
	astgen "github.com/nabbar/auditor/pkg/ast/generic"
	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	audtps "github.com/nabbar/auditor/pkg/data/types"
	librun "github.com/nabbar/golib/runner"
)

// file.go contains function about scanning files, order entries, ...

// ScanFiles performs an initial AST-driven traversal of the repository to discover and parse source code files.
//
// This function executes a comprehensive filesystem walk starting from the configured repository path,
// identifying source code files across multiple supported programming languages. The process follows
// a two-phase approach:
//
//  1. File Discovery: Utilizes language-specific glob patterns from the AST package to efficiently filter
//     potential source code files based on filename patterns.
//  2. Language Identification and AST Parsing: For each discovered file, performs precise language identification
//     and constructs an Abstract Syntax Tree (AST) for further analysis.
//
// Key design considerations:
//
// - Each file is assigned to exactly one language to ensure unambiguous categorization.
// - Pattern-based filtering provides efficient initial screening of files before expensive parsing operations.
// - Language-specific identification ensures accurate categorization even in ambiguous cases.
// - Concurrent processing with semaphore coordination helps manage resource limits while maximizing throughput.
// - Comprehensive error handling maintains system stability and provides meaningful feedback for debugging.
//
// The function returns a generic error if any processing step fails, allowing the caller to determine whether
// the operation completed successfully or encountered issues that require attention.
//
// Example usage:
//
//	err := engine.ScanFiles()
//	if err != nil {
//	    // Handle error appropriately
//	}
func (o *eng) ScanFiles() error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("auditor/engine/ScanFile", r)
		}
	}()

	o.u.Info("Walking filesystem topology starting from root: %s", o.o.RepoPath)

	var (
		err error
		lst []audast.Lang
		mch map[audast.Lang][]string

		ret = errors.New("failed to scanning file for AST")
	)

	if mch, err = o.getLangFiles(); err != nil {
		o.u.ErrorStack("failed to walk repository directory: %w", err)
		return ret
	}

	lst = slices.Collect(maps.Keys(mch))

	for _, lng := range lst {
		if e := o.addColAst(lng, mch[lng]); e != nil {
			return e
		}
	}

	for i := 1; i < 5; i++ {
		if e := o.astStep(i); e != nil {
			o.u.ErrorStack(fmt.Sprintf("error trigger on ast step %d", i), e)
		}
	}

	// clean empty modules + update vendor status
	o.d.ModWalk(func(mod audmod.Module) bool {
		if mod == nil || mod.IsEmpty() {
			return true
		}

		if mod.IsVendor() {
			mod.SetStatusVendor()
			o.d.ModAdd(mod)
		}

		return true
	})

	// clean empty package + update vendor status
	o.d.PkgWalk(func(pkg audpkg.Package) bool {
		if pkg == nil || pkg.IsEmpty() {
			return true
		}

		if pkg.IsVendor() {
			pkg.SetStatusVendor()
			o.d.PkgAdd(pkg)
		}

		return true
	})

	// clean empty entry + update vendor status
	o.d.EntWalk(func(ent audent.Entry) bool {
		if ent == nil || ent.IsEmpty() {
			return true
		}

		if c, _ := ent.GetType(); c == audtps.EntryNone || c == audtps.EntryDepend {
			o.u.Info("Cleaning useless entry: %s (type: %s | id: %d)", ent.GetFullPath(), ent.GetTypeString(), ent.GetID())
			o.d.EntDel(ent.GetID())
			return true
		}

		if ent.IsVendor() {
			return true
		}

		var msg []string
		for _, d := range ent.GetDepend() {
			msg = append(msg, fmt.Sprintf("'%s' (ID: %d | Type: %s | PkgId: %d)", d.GetFullPath(), d.GetID(), d.GetTypeString(), d.GetPkgId()))
		}

		o.u.Info("Entry Stored: '%s' (ID: %d | Type: %s | Pkg: %d) - Dep: \n%s", ent.GetFullPath(), ent.GetID(), ent.GetTypeString(), ent.GetPkgId(), strings.Join(msg, "\n"))

		return true
	})

	var lid = o.d.EntPkgIds(false)

	for _, i := range o.d.EntDepPkgIds(false) {
		if !slices.Contains(lid, i) {
			lid = append(lid, i)
		}
	}

	// clean empty entry + update vendor status
	o.d.EntWalk(func(ent audent.Entry) bool {
		if ent == nil || ent.IsEmpty() || !ent.IsVendor() {
			return true
		}

		if pid := ent.GetPkgId(); !slices.Contains(lid, pid) {
			o.u.Info("Cleaning useless entry: %s (type: %s | id: %d)", ent.GetFullPath(), ent.GetTypeString(), ent.GetID())
			o.d.EntDel(ent.GetID())
		}

		return true
	})

	lid = make([]audids.ID, 0, o.d.EntLen())

	// clean empty entry + update vendor status
	o.d.EntWalk(func(ent audent.Entry) bool {
		if ent == nil || ent.IsEmpty() || ent.IsVendor() {
			return true
		}

		lid = append(lid, ent.GetID())

		for _, i := range ent.GetDependIds() {
			if !slices.Contains(lid, i) {
				lid = append(lid, i)
			}
		}

		if ids := ent.GetDependIds(); len(ent.GetDepend()) != len(ids) {
			o.u.Warning("Entry stored '%s' (id: %d) has unknown dependencies: %v", ent.GetFullPath(), ent.GetID(), ids)
		}

		return true
	})

	for _, i := range o.d.EntKeys() {
		ent := o.d.EntGet(i)
		if ent == nil || ent.IsEmpty() || !ent.IsVendor() {
			continue
		}

		if !slices.Contains(lid, i) {
			o.u.Info("Cleaning useless Entry: '%s' (id: %d), IsVendor: %v, Status: %s", ent.GetFullPath(), ent.GetID(), ent.IsVendor(), ent.GetStatus().String())
			o.d.EntDel(i)
		} else {
			o.d.UnSetStatusOutdated(ent.GetID())
		}
	}

	lid = o.d.EntPkgIds(true)

	for _, i := range o.d.PkgKeys() {
		if !slices.Contains(lid, i) {
			pkg := o.d.PkgGet(i)
			if pkg != nil && !pkg.IsEmpty() && pkg.IsVendor() {
				o.u.Info("Cleaning useless Package: '%s' (id: %d), IsVendor: %v, Status: %s", pkg.GetFullPath(), pkg.GetID(), pkg.IsVendor(), pkg.GetStatus().String())
				o.d.PkgDel(i)
			}
		}
	}

	lid = o.d.PkgModIds(true)

	for _, i := range o.d.ModKeys() {
		if !slices.Contains(lid, i) {
			mod := o.d.ModGet(i)
			if mod != nil && !mod.IsEmpty() && mod.IsVendor() {
				o.u.Info("Cleaning useless module: %s (id: %d, file: %s), IsVendor: %v, Status: %s", mod.GetMod(), mod.GetID(), mod.GetFile(), mod.IsVendor(), mod.GetStatus().String())
				o.d.ModDel(i)
			}
		}
	}

	o.u.Info("File scanning completed successfully")
	return nil
}

func (o *eng) getLangFiles() (map[audast.Lang][]string, error) {
	var (
		err error
		ptn = audast.Pattern()
		mch = make(map[audast.Lang][]string)
	)

	// Walk the filesystem starting from RepoPath
	// This recursive traversal discovers all files in the repository directory
	// and its subdirectories, building a map of files grouped by language type
	err = filepath.Walk(o.o.RepoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories - only process regular files
		// Directories are skipped to avoid processing directory entries as source code
		if info.IsDir() {
			return nil
		}

		// Skip private files
		if strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}

		// Skip private files
		for _, d := range strings.Split(filepath.Dir(path[len(o.o.RepoPath):]), string(filepath.Separator)) {
			if len(d) < 1 {
				continue
			}

			d = strings.TrimPrefix(d, string(filepath.Separator))

			if strings.HasPrefix(d, ".") {
				return nil
			}
		}

		//		o.u.Info("file scanned %s", path)

		// Pattern matching phase: Identify which language patterns match the current file
		// This approach uses AST package patterns for efficient initial filtering
		for l, g := range ptn {
			for _, s := range g {
				var (
					e error
					c bool
					p = path
				)

				if !strings.Contains(s, "/") {
					p = filepath.Base(path)
				}

				// Handle negation patterns (starting with "^")
				// These patterns indicate files that should NOT match the pattern
				if strings.HasPrefix(s, "^") {
					c, e = filepath.Match(p, s[1:])
					c = !c
				} else {
					c, e = filepath.Match(p, s)
				}

				// If pattern matching succeeds and file matches, assign to language bucket
				// The return statement ensures each file is processed only once for a single language
				if e == nil && c {
					if len(mch[l]) < 1 {
						mch[l] = make([]string, 0)
					}

					o.u.Info("file match %s for language %s", path, l.String())
					mch[l] = append(mch[l], path)
					return nil
				}
			}
		}

		return nil
	})

	return mch, err
}

func (o *eng) addColAst(lng audast.Lang, lst []string) error {
	// Create a progress bar for tracking file processing within each language group
	// This provides granular feedback to users about the specific files being processed
	var (
		iOk = new(atomic.Bool)
		pgs = o.u.NewBar("Initiate AST "+lng.String(), len(lst))
	)

	defer pgs.DeferMain()

	for _, f := range lst {
		if e := pgs.NewWorker(); e != nil {
			iOk.Store(false)
			o.u.ErrorStack(fmt.Sprintf("AST Lang %s - failed to create new worker instance for file: %s", lng.String(), f), e)
			return e
		}

		go func(pth string) {
			defer func() {
				if r := recover(); r != nil {
					librun.RecoveryCaller("auditor/engine/ScanFile/addColAst", r)
				}
			}()

			var (
				er error
				as astgen.AST
			)

			defer func() {
				pgs.DeferWorker()
			}()

			// Final language identification check for the file
			// This step ensures that files are properly categorized before AST processing
			// Even though pattern matching has already identified the likely language,
			// this additional verification provides robustness against edge cases
			if !lng.Identify(pth) {
				return
			}

			// Create AST parser instance for the specific language and file
			// This leverages language-specific AST implementations from the AST package
			as, er = lng.AST(o.u, o.d, pth)
			if er != nil || as == nil {
				iOk.Store(false)
				o.u.ErrorStack(fmt.Sprintf("AST Lang %s - failed to create AST instance for file: %s", lng.String(), pth), er)
				return
			}

			o.c.AddAst(lng, as)
		}(f)
	}

	// Wait for all concurrent processing to complete
	// This ensures that all language processing goroutines finish before proceeding
	if e := pgs.WaitAll(); e != nil {
		o.u.ErrorStack("failed to wait all filesystem operations for lang "+lng.String(), e)
	}

	// force update bar
	for !pgs.Completed() {
		pgs.Inc(1)
		time.Sleep(time.Millisecond)
	}

	// timer to wait ui is updated
	time.Sleep(500 * time.Millisecond)

	if iOk.Load() {
		return errors.New("failed to scan file for language " + lng.String())
	}

	return nil
}

func (o *eng) astStep(n int) error {
	var iOk = new(atomic.Bool)

	for _, lng := range o.c.GetLang() {

		var pgs = o.u.NewBar(fmt.Sprintf("Launch AST %s Step %d packages", lng.String(), n), o.c.LenAst(lng))
		defer pgs.DeferMain()

		lst := o.c.GetAst(lng)

		for i := range lst {
			if e := pgs.NewWorker(); e != nil {
				iOk.Store(false)
				o.u.ErrorStack(fmt.Sprintf("AST Lang %s - failed to create new worker instance for AST Updating Package in base path: %s", lng.String(), lst[i].GetDirectory()), e)
				return e
			}

			go func(idx int) {
				defer func() {
					if r := recover(); r != nil {
						librun.RecoveryCaller("auditor/engine/ScanFile/updAstPkg", r)
					}
				}()

				var er error

				defer func() {
					pgs.DeferWorker()
				}()

				switch n {
				case 1:
					er = lst[idx].Step1(pgs)
				case 2:
					er = lst[idx].Step2(pgs)
				case 3:
					er = lst[idx].Step3(pgs)
				case 4:
					er = lst[idx].Step4(pgs)
				}

				// Parse module information from the source file
				// This extracts high-level structural information such as module names
				// and other top-level declarations that define the codebase structure
				if er != nil {
					o.u.ErrorStack(fmt.Sprintf("AST Lang %s - failed to update packages for path: %s", lng.String(), lst[idx].GetDirectory()), er)
					iOk.Store(false)
					return
				}

				o.u.Info("AST Lang %s - Package Updated for path: %s", lng.String(), lst[idx].GetDirectory())
			}(i)
		}

		// Wait for all concurrent processing to complete
		// This ensures that all language processing goroutines finish before proceeding
		if e := pgs.WaitAll(); e != nil {
			o.u.ErrorStack("failed to wait updating packages for lang "+lng.String(), e)
		}

		o.c.SetAst(lng, lst)

		// force update bar
		for !pgs.Completed() {
			pgs.Inc(1)
			time.Sleep(time.Millisecond)
		}

		// timer to wait ui is updated
		time.Sleep(500 * time.Millisecond)
	}

	if iOk.Load() {
		return errors.New("failed to update packages")
	}

	return nil
}
