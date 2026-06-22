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
	"context"
	"errors"
	"fmt"
	"go/ast"
	"slices"
	"sync/atomic"
	"time"

	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	semtps "github.com/nabbar/golib/semaphore/types"
	astgop "golang.org/x/tools/go/packages"
)

// collectModulePackages returns all package IDs belonging to the current module ID.
//
// This function walks through all packages in the collection and filters those that belong
// to the specified module ID. It ensures that only packages associated with the current module
// are included in the returned slice, which is essential for dependency analysis and vendor package processing.
//
// Key features:
//   - Iterates through all packages using PkgWalk for comprehensive coverage
//   - Filters packages based on their module ID to ensure proper scope isolation
//   - Prevents duplicate package IDs by checking against existing entries in the result slice
//   - Returns a clean slice of package identifiers for further processing
//
// Parameters:
//   - modID: The identifier of the module for which to collect package IDs.
//
// Returns:
//   - []audids.ID: A slice containing all package IDs belonging to the specified module.
func (o *mdl) collectModulePackages(modID audids.ID) []audids.ID {
	lpk := make([]audids.ID, 0)

	o.a.PkgWalk(func(pkg audpkg.Package) bool {
		if pkg.GetModId() != modID {
			return true
		}

		pid := pkg.GetID()
		if !slices.Contains(lpk, pid) {
			lpk = append(lpk, pid)
		}

		return true
	})

	return lpk
}

// collectVendorDependencies maps vendor Package IDs to slices of vendor Entry IDs.
//
// This function analyzes all entries in the collection to identify those that depend on vendor packages.
// It creates a mapping from vendor package identifiers to their corresponding entry identifiers,
// which is essential for processing vendor dependencies during AST analysis. The function ensures
// that only relevant vendor dependencies are included based on local package associations.
//
// Key features:
//   - Walks through all entries using EntWalk for comprehensive dependency analysis
//   - Filters entries based on their package ID to ensure they belong to the local module
//   - Identifies vendor dependencies by checking the IsVendor flag on dependency entries
//   - Builds a map structure for efficient lookup and processing of vendor packages
//   - Maintains a count of vendor entries for progress tracking and resource management
//
// Parameters:
//   - localPkgIDs: A slice of package identifiers belonging to the local module.
//
// Returns:
//   - map[audids.ID][]audids.ID: A mapping from vendor package IDs to slices of vendor entry IDs.
//   - int: The total count of vendor entries found for processing.
func (o *mdl) collectVendorDependencies(localPkgIDs []audids.ID) (map[audids.ID][]audids.ID, int) {
	ptn := make(map[audids.ID][]audids.ID)
	nbr := 0

	o.a.EntWalk(func(ent audent.Entry) bool {
		if ent == nil || ent.IsEmpty() {
			return true
		}

		// Skip entries that don't belong to the local module packages
		if !slices.Contains(localPkgIDs, ent.GetPkgId()) {
			return true
		}

		// Process each dependency of the entry
		for _, d := range ent.GetDepend() {
			// Skip dependencies that are nil, empty, or not vendor packages
			if d == nil || d.IsEmpty() || !d.IsVendor() {
				continue
			}

			dp := d.GetPkgId()
			vendorEntID := d.GetID()

			// Skip if the vendor entry ID is invalid
			if vendorEntID == 0 {
				continue
			}

			// Initialize the slice for this package ID if it doesn't exist yet
			if _, exists := ptn[dp]; !exists {
				ptn[dp] = make([]audids.ID, 0)
			}

			// Add the vendor entry ID to the slice if it's not already present
			if !slices.Contains(ptn[dp], vendorEntID) {
				ptn[dp] = append(ptn[dp], vendorEntID)
				nbr++
			}
		}

		return true
	})

	return ptn, nbr
}

// getVdrPkg iterates through vendor packages and runs AST analysis using SemBar.
//
// This function orchestrates the concurrent processing of vendor packages by creating
// a semaphore-based bar for managing worker threads. It processes each vendor package
// in parallel to improve performance during vendor dependency analysis. The function
// ensures proper resource management and error handling throughout the vendor package processing.
//
// Key features:
//   - Initializes a progress bar with the number of vendor packages to process
//   - Creates concurrent workers for each vendor package using goroutines
//   - Manages worker lifecycle through semaphore-based concurrency control
//   - Handles errors and tracks failures through atomic flags for proper error reporting
//   - Ensures thread-safe resource cleanup via deferred functions
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - ptn: A mapping from vendor package IDs to slices of vendor entry IDs.
//   - nbr: The total count of vendor entries to process.
//
// Returns:
//   - error: Any error encountered during vendor package processing, including worker creation or processing errors.
func (o *mdl) getVdrPkg(ctx context.Context, ptn map[audids.ID][]audids.ID, nbr int) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Return early if there are no vendor packages to process
	if len(ptn) < 1 || nbr < 1 {
		return nil
	}

	var (
		err error
		bar semtps.SemBar

		nok = new(atomic.Bool)
		odw = o.getDirWork()
		cfg = &astgop.Config{
			Mode:       astgop.NeedName | astgop.NeedFiles | astgop.NeedSyntax | astgop.NeedTypes,
			Context:    ctx,
			Logf:       o.u.Info,
			Dir:        odw,
			Env:        nil,
			BuildFlags: nil,
			Fset:       nil,
			ParseFile:  nil,
			Tests:      false,
			Overlay:    nil,
		}
	)

	defer func() {
		if bar != nil {
			bar.DeferMain()
		}
	}()

	// Initialize the progress bar for vendor package processing
	bar = o.u.NewBar("Parsing all vendor", len(ptn))

	// Process each vendor package concurrently using goroutines
	for pid, eid := range ptn {
		if err = bar.NewWorker(); err != nil {
			o.u.ErrorStack("error creating new worker", err)
			return err
		}

		go func() {
			defer bar.DeferWorker()

			// Process the vendor package with its associated entries
			if e := o.oneVdrPkg(ctx, pid, eid, cfg); e != nil {
				o.u.ErrorStack("error processing vendor package", e)
				nok.Store(true)
			}
		}()
	}

	// Wait for all concurrent workers to complete
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
		err = errors.New("error triggered on parsing vendor packages")
		o.u.Error("error triggered on parsing vendor packages")
		return err
	}

	return nil
}

// oneVdrPkg analyzes AST files for a given vendor package.
//
// This function processes a single vendor package by loading its AST and analyzing
// the declarations within its source files. It maps entry names to their IDs for quick lookup
// and then processes each declaration concurrently using goroutines. The function ensures
// proper error handling and resource management throughout the vendor package analysis.
//
// Key features:
//   - Loads the vendor package using astgop.Load with appropriate configuration
//   - Maps entry names to entry IDs for efficient lookup during AST processing
//   - Processes each declaration concurrently using goroutines for parallel execution
//   - Handles function and type declarations specifically for vendor package analysis
//   - Manages worker lifecycle through semaphore-based concurrency control
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pid: The identifier of the vendor package to process.
//   - led: A slice of entry identifiers associated with this vendor package.
//   - cfg: Configuration for AST package loading with required modes.
//
// Returns:
//   - error: Any error encountered during vendor package analysis, including loading or processing errors.
func (o *mdl) oneVdrPkg(ctx context.Context, pid audids.ID, led []audids.ID, cfg *astgop.Config) error {
	if ctx == nil || pid < 1 || cfg == nil {
		return errors.New("invalid argument")
	}

	// Return early if there are no entries to process
	if len(led) < 1 {
		return nil
	}

	var (
		err error
		bar semtps.SemBar
		pth string
		alp []*astgop.Package
		alk *astgop.Package

		let = make(map[string]audids.ID)
		nok = new(atomic.Bool)
		pkg = o.a.PkgGet(pid)
	)

	defer func() {
		if bar != nil {
			bar.DeferMain()
		}
	}()

	// Validate that the package exists and is a vendor package
	if pkg == nil || pkg.IsEmpty() || !pkg.IsVendor() {
		return nil
	}

	// Retrieve the full path of the vendor package for AST loading
	if pth = pkg.GetFullPath(); len(pth) < 0 {
		return errors.New("invalid package path")
	}

	// Load the vendor package with the specified configuration
	if alp, err = astgop.Load(cfg, pth); err != nil {
		o.u.ErrorStack("cannot start AST for Golang vendor package: "+pth, err)
		return err
	}

	// Validate that exactly one package was loaded (as expected for vendor packages)
	if len(alp) != 1 {
		err = fmt.Errorf("no package found in path %s", pth)
		o.u.ErrorStack("error trigger while parsing vendor package", err)
		return err
	}

	alk = alp[0]

	// Map entry names to entry IDs for quick lookup during AST processing
	for _, i := range led {
		e := o.a.EntGet(i)
		if e == nil || e.IsEmpty() {
			bar.Inc(1)
			continue
		}
		let[e.GetName()] = i
	}

	// Initialize the progress bar for vendor entry processing within this package
	bar = o.u.NewBar("Parsing Vendor Entry for package "+pth, len(let))

	// Process each syntax file in the vendor package
	for j, f := range alk.Syntax {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		// Skip files with no declarations to avoid unnecessary processing
		if len(f.Decls) < 1 {
			continue
		}

		fn := alk.GoFiles[j]

		// Process each declaration in the file
		for _, d := range f.Decls {
			switch t := d.(type) {
			case *ast.FuncDecl:
				// Handle function declarations specifically for vendor packages
				if t.Name != nil && len(t.Name.Name) > 0 {
					if eid, ok := let[t.Name.Name]; ok && eid > 0 {
						if err = bar.NewWorker(); err != nil {
							o.u.ErrorStack("error creating new worker", err)
							return err
						}

						go func(etID audids.ID, file string, decl *ast.FuncDecl) {
							defer bar.DeferWorker()

							o.u.Info("[File %s] Parsing vendor function: %s", file, decl.Name.Name)
							// Process the vendor function declaration with appropriate context
							if e := o.vendorDecl(ctx, pid, etID, file, f, decl); e != nil {
								o.u.ErrorStack("error parsing vendor decl: "+file, e)
								nok.Store(true)
							}
						}(eid, fn, t)
					}
				}
			case *ast.GenDecl:
				// Handle general declarations including type specifications
				if t != nil && len(t.Specs) > 0 {
					for _, spec := range t.Specs {
						if s, ok := spec.(*ast.TypeSpec); ok && s.Name != nil && len(s.Name.Name) > 0 {
							// Handle type specifications for vendor packages
							if eid, k := let[s.Name.Name]; k && eid > 0 {
								if err = bar.NewWorker(); err != nil {
									o.u.ErrorStack("error creating new worker", err)
									return err
								}

								go func(etID audids.ID, file string, decl *ast.TypeSpec) {
									defer bar.DeferWorker()

									o.u.Info("[File %s] Parsing vendor type: %s", file, decl.Name.Name)
									// Process the vendor type declaration with appropriate context
									if e := o.vendorDecl(ctx, pid, etID, file, f, decl); e != nil {
										o.u.ErrorStack("error parsing vendor decl: "+file, e)
										nok.Store(true)
									}
								}(eid, fn, s)
							}
						}
					}
				}
			}
		}
	}

	// Wait for all concurrent processing to complete
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
		err = errors.New("error triggered on parsing vendor packages")
		o.u.Error("error triggered on parsing vendor packages")
		return err
	}

	return nil
}

// vendorDecl parses specific vendor AST declarations to update entry signatures and parameters.
//
// This function handles the processing of vendor-specific AST declarations by dispatching
// to appropriate parsing functions based on the declaration type. It ensures that only
// relevant vendor declarations are processed while maintaining proper error handling and
// context management throughout the analysis process.
//
// Key features:
//   - Validates input parameters to ensure proper context and identifiers
//   - Retrieves the entry for validation before processing
//   - Dispatches to specific parsing functions based on AST node type (function or type)
//   - Maintains error handling and reporting for vendor declaration processing
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pid: The identifier of the package containing this vendor declaration.
//   - eid: The identifier of the entry associated with this vendor declaration.
//   - file: The path to the source file containing this declaration.
//   - afs: The AST file structure containing this declaration.
//   - node: The AST node representing the vendor declaration to process.
//
// Returns:
//   - error: Any error encountered during vendor declaration processing, including validation or parsing errors.
func (o *mdl) vendorDecl(ctx context.Context, pid, eid audids.ID, file string, afs *ast.File, node ast.Node) error {
	if ctx == nil || node == nil || pid < 1 || eid < 1 {
		return errors.New("invalid parameters")
	}

	var ent = o.a.EntGet(eid)

	if ent == nil || ent.IsEmpty() {
		return errors.New("vendor entry not found")
	}

	switch decl := node.(type) {
	case *ast.FuncDecl:
		// Logic to parse parameters, return types, and receiver from AST,
		// without recursively traversing internal body calls (IsVendor = true).
		return o.parseFunction(ctx, pid, file, afs, decl)

	case *ast.TypeSpec:
		// Logic to parse underlying struct/interface fields and types.
		return o.parseType(ctx, pid, file, afs, decl)

	default:
		return fmt.Errorf("unsupported vendor declaration type: %T", node)
	}
}
