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

	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audprm "github.com/nabbar/auditor/pkg/data/params"
	audtps "github.com/nabbar/auditor/pkg/data/types"
)

// setInterface processes an interface type declaration (*ast.TypeSpec with *ast.InterfaceType)
// from a Go source file and creates or merges an entry in the auditor's data model.
//
// This function handles the initial creation of interface entries during AST parsing.
// It validates that the type is indeed an interface, extracts the source code,
// creates a new entry or merges with an existing one, and registers it.
//
// The function:
//  1. Validates the TypeSpec node and package identifier.
//  2. Type-asserts that ts.Type is an *ast.InterfaceType.
//  3. Retrieves the source code for the interface declaration.
//  4. Creates or merges an entry with EntryInterface as the code type.
//  5. Sets the entry status and registers it in the package collection.
//
// Note: Interface methods are populated later by updInterface during the update phase.
// This function only creates the basic interface entry.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this interface.
//   - fs: The file path containing the interface declaration.
//   - afs: The AST File structure containing the interface.
//   - node: The AST Node representing the type declaration, expected to be *ast.TypeSpec
//     with ts.Type being *ast.InterfaceType.
func (o *mdl) setInterface(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, node ast.Node) {
	var (
		err error
		cnt []byte
		ent audent.Entry

		ok bool
		ts *ast.TypeSpec
		it *ast.InterfaceType
	)

	// Type-assert the node to *ast.TypeSpec. Validate that the type spec is not nil,
	// has a non-nil name, and the name is not empty.
	if ts, ok = node.(*ast.TypeSpec); !ok || ts == nil || ts.Name == nil || len(ts.Name.Name) < 1 {
		err = errors.New("invalid AST type spec declaration for interface")
		o.u.ErrorStack("[setInterface] trigger error", err)
		return
	}

	// Type-assert that ts.Type is an *ast.InterfaceType. If the type is not an interface,
	// this function cannot handle it and returns an error.
	if it, ok = ts.Type.(*ast.InterfaceType); !ok || it == nil {
		err = fmt.Errorf("invalid interface type for %s", ts.Name.Name)
		o.u.ErrorStack("[setInterface] trigger error", err)
		return
	}

	// Validate that a non-zero package identifier is provided.
	if pkg == 0 {
		err = errors.New("invalid nil package")
		o.u.ErrorStack("[setInterface] trigger error", err)
		return
	}

	o.u.Info("[setInterface][File %s][Interface %s] Setting entry interface", fs, ts.Name.Name)

	// Retrieve the raw source code for the TypeSpec node from the file.
	// This includes the interface declaration text and any associated documentation comments.
	cnt, err = o.getSrc(ctx, fs, afs, ts)
	if err != nil {
		err = fmt.Errorf("cannot retrieve content of file %s: %v", fs, err)
		o.u.ErrorStack("[setInterface] trigger error on retrieve source code", err)
		return
	}

	// Create a new entry or merge with an existing one for this interface.
	// The type string is empty; interface methods are populated by updInterface.
	if ent = o.a.EntNewOrMerge(pkg, ts.Name.Name, "", audtps.EntryInterface, Lang, cnt); ent == nil || ent.IsEmpty() {
		err = fmt.Errorf("invalid interface %s in file %s", ts.Name.Name, fs)
		o.u.ErrorStack("[setInterface] trigger error on creating/updating entry", err)
		return
	}

	// Mark the entry as successfully parsed.
	ent.SetStatusParsed()

	// Configure vendor/outdated status.
	if ent.IsVendor() {
		ent.SetStatusOutdated()
	} else {
		ent.UnSetStatusOutdated()
	}

	// Add the entry to the package collection and log the result.
	o.a.EntAdd(ent)
	o.u.Info("[setInterface][File %s][Interface %s] Entry added '%s' (id: %d) / Package '%s' (id %d) / Module '%s' (id: %d)", fs, ts.Name.Name, ent.GetFullPath(), ent.GetID(), ent.GetPkg().GetFullPath(), ent.GetPkgId(), ent.GetPkg().GetMod().GetMod(), ent.GetPkg().GetModId())
}

// updInterface updates an existing interface type entry (*ast.TypeSpec with *ast.InterfaceType)
// in the auditor's data model.
//
// This function handles the detailed update of interface declarations that have already
// been created by setInterface. It populates method signatures, input/output parameters,
// and dependency information by analyzing the interface's method list.
//
// The update process performs the following steps:
//  1. Validates the TypeSpec node and package identifier.
//  2. Type-asserts that ts.Type is an *ast.InterfaceType.
//  3. Finds the existing entry by package ID and interface name.
//  4. Updates the entry's type to EntryInterface.
//  5. Iterates over all method entries in the interface's Method list:
//     - For named methods (len(m.Names) > 0): type-asserts the method type to *ast.FuncType;
//     if successful, extracts the first name as the method name, creates a parameter object
//     with EntryFunction code type, processes input and output parameters to create
//     parameter objects and track dependencies from parameter and return types, then adds
//     the method as an input parameter to the interface entry.
//     - For embedded interfaces (len(m.Names) == 0): resolves the embedded type via
//     getTypeDepend, extracts the type's code and string representation, adds it as an
//     input parameter to the interface entry, and registers it as a dependency on the
//     interface entry. Vendor dependencies have their outdated status cleared.
//  6. Re-adds the updated entry to the package collection.
//
// Dependencies are tracked using a deduplication map (keyed by full path) to avoid
// duplicates. Vendor dependencies have their outdated status cleared.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this interface.
//   - fs: The file path containing the interface declaration.
//   - afs: The AST File structure containing the interface.
//   - node: The AST Node representing the type declaration, expected to be *ast.TypeSpec
//     with ts.Type being *ast.InterfaceType.
func (o *mdl) updInterface(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, node ast.Node) {
	var (
		err error
		dep = make(map[string]bool)
		ent audent.Entry

		ok bool
		ts *ast.TypeSpec
		it *ast.InterfaceType
	)

	// Type-assert the node to *ast.TypeSpec. Validate that the type spec is not nil,
	// has a non-nil name, and the name is not empty.
	if ts, ok = node.(*ast.TypeSpec); !ok || ts == nil || ts.Name == nil || len(ts.Name.Name) < 1 {
		err = errors.New("invalid AST type spec declaration for interface")
		o.u.ErrorStack("[updInterface] trigger error", err)
		return
	}

	// Type-assert that ts.Type is an *ast.InterfaceType.
	if it, ok = ts.Type.(*ast.InterfaceType); !ok || it == nil {
		err = fmt.Errorf("invalid interface type for %s", ts.Name.Name)
		o.u.ErrorStack("[updInterface] trigger error", err)
		return
	}

	// Validate that a non-zero package identifier is provided.
	if pkg == 0 {
		err = errors.New("invalid nil package")
		o.u.ErrorStack("[updInterface] trigger error", err)
		return
	}

	// Find the existing entry by package ID and interface name.
	if ent, err = o.findEntry(pkg, ts.Name.Name); err != nil {
		o.u.ErrorStack("[updInterface] trigger error on find entry", err)
		return
	} else if ent == nil || ent.IsEmpty() {
		err = fmt.Errorf("cannot find interface entry %s for package %d", ts.Name.Name, pkg)
		o.u.ErrorStack("[updInterface] trigger error on find entry", err)
		return
	}

	o.u.Info("[updInterface][File %s][Interface %s] Updating entry interface", fs, ts.Name.Name)

	// Check and update the entry's type. Ensure it is set to EntryInterface with
	// an empty type string.
	if c, s := ent.GetType(); c != audtps.EntryInterface || s != "" {
		ent.SetType(audtps.EntryInterface, "")
	}

	// Process all methods declared in the interface.
	if it.Methods != nil {
		for _, m := range it.Methods.List {
			// Named method: extract function type, process parameters, and add to interface entry.
			if len(m.Names) > 0 {
				// Type-assert the method type to *ast.FuncType; if not a function
				// type, skip this method entry.
				ft, k := m.Type.(*ast.FuncType)
				if !k {
					continue
				}

				// Use the first name as the method name.
				mn := m.Names[0].Name
				// Create a new parameter object for this method with EntryFunction code type.
				mp := audprm.New(mn, audtps.EntryFunction)

				// Process input parameters for the interface method.
				if ft.Params != nil {
					for _, p := range ft.Params.List {
						pr, dp, er := o.getParams(ctx, pkg, fs, afs, mn, p)

						if er != nil {
							o.u.ErrorStack(fmt.Sprintf("[updInterface][File %s][Interface %s][Method %s] trigger error on getting input param of method", fs, ts.Name.Name, mn), er)
							return
						}

						mp.SetInput(pr...)

						// Add dependencies from parameter types.
						for _, et := range dp {
							if et == nil || et.IsEmpty() {
								continue
							}

							id := et.GetID()
							fn := et.GetFullPath()
							vd := et.IsVendor()

							if id < 1 {
								o.u.ErrorStackData(fmt.Sprintf("[updInterface][File %s][Interface %s][Method %s] trigger error on getting Method input params", fs, ts.Name.Name, mn), et, errors.New("empty ID for entry"))
								continue
							}

							if dep[fn] {
								continue
							}

							dep[fn] = true

							if vd {
								o.a.UnSetStatusOutdated(id)
							}

							ent.AddDepend(id)
						}
					}
				}

				// Process output (return) parameters for the interface method.
				if ft.Results != nil {
					for _, p := range ft.Results.List {
						pr, dp, er := o.getParams(ctx, pkg, fs, afs, mn, p)

						if er != nil {
							o.u.ErrorStack(fmt.Sprintf("[updInterface][File %s][Interface %s][Method %s] trigger error on getting result params of method", fs, ts.Name.Name, mn), er)
							return
						}

						mp.SetOutput(pr...)

						// Add dependencies from return types.
						for _, et := range dp {
							if et == nil || et.IsEmpty() {
								continue
							}

							id := et.GetID()
							fn := et.GetFullPath()
							vd := et.IsVendor()

							if id < 1 {
								o.u.ErrorStackData(fmt.Sprintf("[updInterface][File %s][Interface %s][Method %s] trigger error on getting Method result params", fs, ts.Name.Name, mn), et, errors.New("empty ID for entry"))
								continue
							}

							if dep[fn] {
								continue
							}

							dep[fn] = true

							if vd {
								o.a.UnSetStatusOutdated(id)
							}

							ent.AddDepend(id)
						}
					}
				}

				// Add the method parameter as an input to the interface entry.
				ent.AddInputs(mp)

			} else {
				// Embedded interface: no method names, just an embedded type.
				// Resolve the embedded type and add it as a dependency.
				for _, et := range o.getTypeDepend(ctx, fs, afs, pkg, m.Type) {
					if et != nil && !et.IsEmpty() {
						cd, st := et.GetType()
						id := et.GetID()
						fn := et.GetFullPath()
						vd := et.IsVendor()

						if id < 1 {
							o.u.ErrorStackData(fmt.Sprintf("[updInterface][File %s][Interface %s] trigger error on getting embedded interface", fs, ts.Name.Name), et, errors.New("empty ID for entry"))
							continue
						}

						// Add the embedded type as an input parameter.
						ent.AddInputs(audprm.New(st, cd))

						if dep[fn] {
							continue
						}

						dep[fn] = true

						if vd {
							o.a.UnSetStatusOutdated(id)
						}

						ent.AddDepend(id)
					}
				}
			}
		}
	}

	// Re-add the updated entry to the package collection and log the result.
	o.a.EntAdd(ent)
	o.u.Info("[updInterface][File %s][Interface %s] Entry updated '%s' (id: %d) / Package '%s' (id %d) / Module '%s' (id: %d)", fs, ts.Name.Name, ent.GetFullPath(), ent.GetID(), ent.GetPkg().GetFullPath(), ent.GetPkgId(), ent.GetPkg().GetMod().GetMod(), ent.GetPkg().GetModId())
}
