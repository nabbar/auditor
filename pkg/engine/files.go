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
	audmod "github.com/nabbar/auditor/pkg/data/mod"
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

		mch = make(map[audast.Lang][]string)
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

	if e := o.getAstMod(); e != nil {
		return e
	}

	if e := o.getAstPkg(); e != nil {
		return e
	}

	if e := o.updAstPkg(); e != nil {
		return e
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

func (o *eng) getAstMod() error {
	var iOk = new(atomic.Bool)

	for _, lng := range o.c.GetLang() {

		var pgs = o.u.NewBar("Parsing Module with AST "+lng.String(), o.c.LenAst(lng))
		defer pgs.DeferMain()

		for _, f := range o.c.GetAst(lng) {
			if e := pgs.NewWorker(); e != nil {
				iOk.Store(false)
				o.u.ErrorStack(fmt.Sprintf("AST Lang %s - failed to create new worker instance for file: %s", lng.String(), f), e)
				return e
			}

			go func(as astgen.AST) {
				defer func() {
					if r := recover(); r != nil {
						librun.RecoveryCaller("auditor/engine/ScanFile/getAstMod", r)
					}
				}()

				var (
					er error
					md audmod.Module
				)

				defer func() {
					pgs.DeferWorker()
				}()

				// Parse module information from the source file
				// This extracts high-level structural information such as module names
				// and other top-level declarations that define the codebase structure
				md, er = f.ParseModule(pgs)
				if er != nil {
					o.u.ErrorStack(fmt.Sprintf("AST Lang %s - failed to parse module with AST for file: %s", lng.String(), f.GetFile()), er)
					iOk.Store(false)
					return
				}

				o.u.Info("AST Lang %s - Module found: %s (id: %d)", lng.String(), md.GetMod(), md.GetID())
			}(f)
		}

		// Wait for all concurrent processing to complete
		// This ensures that all language processing goroutines finish before proceeding
		if e := pgs.WaitAll(); e != nil {
			o.u.ErrorStack("failed to wait parsing module for lang "+lng.String(), e)
		}

		// force update bar
		for !pgs.Completed() {
			pgs.Inc(1)
			time.Sleep(time.Millisecond)
		}

		// timer to wait ui is updated
		time.Sleep(500 * time.Millisecond)
	}

	if iOk.Load() {
		return errors.New("failed to parse module")
	}

	return nil
}

func (o *eng) getAstPkg() error {
	var iOk = new(atomic.Bool)

	for _, lng := range o.c.GetLang() {

		var pgs = o.u.NewBar("Parsing Packages with AST "+lng.String(), o.c.LenAst(lng))
		defer pgs.DeferMain()

		for _, f := range o.c.GetAst(lng) {
			if e := pgs.NewWorker(); e != nil {
				iOk.Store(false)
				o.u.ErrorStack(fmt.Sprintf("AST Lang %s - failed to create new worker instance for AST Parsing Package in base path: %s", lng.String(), f.GetDirectory()), e)
				return e
			}

			go func() {
				defer func() {
					if r := recover(); r != nil {
						librun.RecoveryCaller("auditor/engine/ScanFile/getAstPkg", r)
					}
				}()

				var er error

				defer func() {
					pgs.DeferWorker()
				}()

				// Parse module information from the source file
				// This extracts high-level structural information such as module names
				// and other top-level declarations that define the codebase structure
				if er = f.ParsePackage(pgs); er != nil {
					o.u.ErrorStack(fmt.Sprintf("AST Lang %s - failed to parse packages for base path: %s", lng.String(), f.GetDirectory()), er)
					iOk.Store(false)
					return
				}

				o.u.Info("AST Lang %s - Package parsed in base path: %s", lng.String(), f.GetDirectory())
			}()
		}

		// Wait for all concurrent processing to complete
		// This ensures that all language processing goroutines finish before proceeding
		if e := pgs.WaitAll(); e != nil {
			o.u.ErrorStack("failed to wait all parsing package for lang "+lng.String(), e)
		}

		// force update bar
		for !pgs.Completed() {
			pgs.Inc(1)
			time.Sleep(time.Millisecond)
		}

		// timer to wait ui is updated
		time.Sleep(500 * time.Millisecond)
	}

	if iOk.Load() {
		return errors.New("failed to parse package")
	}

	return nil
}

func (o *eng) updAstPkg() error {
	var iOk = new(atomic.Bool)

	for _, lng := range o.c.GetLang() {

		var pgs = o.u.NewBar("Update packages with AST "+lng.String(), o.c.LenAst(lng))
		defer pgs.DeferMain()

		for _, f := range o.c.GetAst(lng) {
			if e := pgs.NewWorker(); e != nil {
				iOk.Store(false)
				o.u.ErrorStack(fmt.Sprintf("AST Lang %s - failed to create new worker instance for AST Updating Package in base path: %s", lng.String(), f.GetDirectory()), e)
				return e
			}

			go func() {
				defer func() {
					if r := recover(); r != nil {
						librun.RecoveryCaller("auditor/engine/ScanFile/updAstPkg", r)
					}
				}()

				var er error

				defer func() {
					pgs.DeferWorker()
				}()

				// Parse module information from the source file
				// This extracts high-level structural information such as module names
				// and other top-level declarations that define the codebase structure
				if er = f.UpdatePackage(pgs); er != nil {
					o.u.ErrorStack(fmt.Sprintf("AST Lang %s - failed to update packages for path: %s", lng.String(), f.GetDirectory()), er)
					iOk.Store(false)
					return
				}

				o.u.Info("AST Lang %s - Package Updated for path: %s", lng.String(), f.GetDirectory())
			}()
		}

		// Wait for all concurrent processing to complete
		// This ensures that all language processing goroutines finish before proceeding
		if e := pgs.WaitAll(); e != nil {
			o.u.ErrorStack("failed to wait updating packages for lang "+lng.String(), e)
		}

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
