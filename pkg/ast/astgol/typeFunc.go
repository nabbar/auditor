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
	audtps "github.com/nabbar/auditor/pkg/data/types"
)

// setFunc processes a function declaration (*ast.FuncDecl) from a Go source file
// and creates or merges an entry in the auditor's data model.
//
// This function handles the initial creation of function entries during AST parsing.
// It extracts the source code for the function, creates a new entry or merges with
// an existing one, and registers it in the package's entry collection.
//
// The function:
//  1. Validates the FuncDecl node and package identifier.
//  2. Skips anonymous functions and blank identifiers (_).
//  3. Retrieves the source code for the function declaration (including doc comments).
//  4. Creates or merges an entry with EntryFunction as the code type.
//  5. Sets the entry status and registers it in the package collection.
//
// Note: This function only creates the basic function entry. Parameter types, return
// types, and dependencies are populated later by updFunc during the update phase.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this function.
//   - fs: The file path containing the function declaration.
//   - afs: The AST File structure containing the function.
//   - node: The AST Node representing the function declaration, expected to be *ast.FuncDecl.
func (o *mdl) setFunc(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, node ast.Node) {
	var (
		err error
		cnt []byte
		ent audent.Entry

		ok bool
		tf *ast.FuncDecl
	)

	// Type-assert the node to *ast.FuncDecl. Validate that the function declaration
	// is not nil, has a non-nil name, and has a non-nil type signature.
	if tf, ok = node.(*ast.FuncDecl); !ok || tf == nil || tf.Name == nil || tf.Type == nil {
		err = errors.New("invalid AST function declaration")
		o.u.ErrorStack("[setFunc] trigger error", err)
		return
	}

	// Ignore anonymous functions and blank identifiers (_).
	// Anonymous functions have no name and cannot be tracked as entries.
	// Blank identifiers are not valid function names.
	if len(tf.Name.Name) < 1 || tf.Name.Name == "_" {
		return
	}

	// Validate that a non-zero package identifier is provided.
	if pkg == 0 {
		err = errors.New("invalid nil package")
		o.u.ErrorStack("[setFunc] trigger error", err)
		return
	}

	o.u.Info("[setFunc][File %s][Function %s] Setting entry function", fs, tf.Name.Name)

	// Retrieve the raw source code for the FuncDecl node from the file.
	// This includes the function signature, body, and any associated documentation comments.
	if cnt, err = o.getSrc(ctx, fs, afs, tf); err != nil {
		err = fmt.Errorf("cannot retrieve content of file %s: %v", fs, err)
		o.u.ErrorStack("[setFunc] trigger error on retrieve source code", err)
		return
	}

	// Create a new entry or merge with an existing one for this function.
	// The type string is empty at this stage; parameter/return types are populated by updFunc.
	if ent = o.a.EntNewOrMerge(pkg, tf.Name.Name, "", audtps.EntryFunction, Lang, cnt); ent == nil || ent.IsEmpty() {
		err = fmt.Errorf("invalid function %s in file %s", tf.Name.Name, fs)
		o.u.ErrorStack("[setFunc] trigger error on creating/updating entry", err)
		return
	}

	// Mark the entry as successfully parsed.
	ent.SetStatusParsed()

	// Configure vendor/outdated status. Vendor entries are marked as outdated
	// so they can be cleaned up by the engine; non-vendor entries are kept current.
	if ent.IsVendor() {
		ent.SetStatusOutdated()
	} else {
		ent.UnSetStatusOutdated()
	}

	// Add the entry to the package collection and log the result.
	o.a.EntAdd(ent)
	o.u.Info("[setFunc][File %s][Function %s] Entry added '%s' (id: %d) / Package '%s' (id %d) / Module '%s' (id: %d)", fs, tf.Name.Name, ent.GetFullPath(), ent.GetID(), ent.GetPkg().GetFullPath(), ent.GetPkgId(), ent.GetPkg().GetMod().GetMod(), ent.GetPkg().GetModId())
}

// updFunc updates an existing function entry (*ast.FuncDecl) in the auditor's data model.
//
// This function handles the detailed update of function declarations that have already
// been created by setFunc. It populates parameter types, return types, and dependency
// information by analyzing the function's AST structure.
//
// The update process performs the following steps:
//  1. Validates the FuncDecl node and package identifier.
//  2. Skips anonymous functions and blank identifiers (_).
//  3. Finds the existing entry by package ID and function name.
//  4. Updates the entry's type to EntryFunction.
//  5. Processes input parameters: extracts types, creates parameter objects, and tracks dependencies.
//  6. Processes output (return) parameters: extracts types, creates parameter objects, and tracks dependencies.
//  7. Inspects the function body using ast.Inspect to traverse all AST nodes recursively and identify dependencies:
//     - *ast.SelectorExpr nodes (e.g., the "fmt.Println" selection expression within "fmt.Println()") represent
//     the qualified identifier or method receiver part of a call; resolved via getDepend to locate the
//     external package and function entries.
//     - *ast.CallExpr nodes (e.g., localFunc() or pkg.Func()) represent function call expressions and are
//     resolved via getLocDepend to locate the local function entries within the same package context.
//     When the CallExpr's Fun field is a SelectorExpr, both the SelectorExpr and CallExpr branches
//     will execute, but the deduplication map prevents duplicate dependencies.
//  8. Re-adds the updated entry to the package collection.
//
// Dependencies are tracked using a deduplication map (keyed by full path) to avoid
// adding the same dependency multiple times. Vendor dependencies have their outdated
// status cleared to indicate they are actively used.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this function.
//   - fs: The file path containing the function declaration.
//   - afs: The AST File structure containing the function.
//   - node: The AST Node representing the function declaration, expected to be *ast.FuncDecl.
func (o *mdl) updFunc(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, node ast.Node) {
	var (
		err error
		dep = make(map[string]bool)
		ent audent.Entry

		ok bool
		tf *ast.FuncDecl
	)

	// Type-assert the node to *ast.FuncDecl. Validate that the function declaration
	// is not nil, has a non-nil name, and has a non-nil type signature.
	if tf, ok = node.(*ast.FuncDecl); !ok || tf == nil || tf.Name == nil || tf.Type == nil {
		err = errors.New("invalid AST function declaration")
		o.u.ErrorStack("[updFunc] trigger error", err)
		return
	}

	// Ignore anonymous functions and blank identifiers (_).
	if len(tf.Name.Name) < 1 || tf.Name.Name == "_" {
		return
	}

	// Validate that a non-zero package identifier is provided.
	if pkg == 0 {
		err = errors.New("invalid nil package")
		o.u.ErrorStack("[updFunc] trigger error", err)
		return
	}

	// Find the existing entry by package ID and function name.
	// If the entry does not exist, log an error and return.
	if ent, err = o.findEntry(pkg, tf.Name.Name); err != nil {
		o.u.ErrorStack("[updFunc] trigger error on find entry", err)
		return
	} else if ent == nil || ent.IsEmpty() {
		err = fmt.Errorf("cannot find function entry %s for package %d", tf.Name.Name, pkg)
		o.u.ErrorStack("[updFunc] trigger error on finding entry", err)
		return
	}

	o.u.Info("[updFunc][File %s][Function %s] Updating entry function", fs, tf.Name.Name)

	// Check and update the entry's type. Ensure it is set to EntryFunction with
	// an empty type string (functions do not have a type string like types do).
	if c, s := ent.GetType(); c != audtps.EntryFunction || s != "" {
		ent.SetType(audtps.EntryFunction, "")
	}

	// Process input parameters from the function signature.
	// For each parameter, extract types, create parameter objects, and track dependencies.
	if tf.Type.Params != nil {
		for _, p := range tf.Type.Params.List {
			pr, dp, er := o.getParams(ctx, pkg, fs, afs, tf.Name.Name, p)

			if er != nil {
				o.u.ErrorStack(fmt.Sprintf("[updFunc][File %s][Function %s] trigger error on getting input params", fs, tf.Name.Name), er)
				return
			}

			// Add the extracted parameters as inputs to the function entry.
			ent.AddInputs(pr...)

			// Add dependencies from parameter types (e.g., imported types used as parameters).
			for _, et := range dp {
				if et == nil || et.IsEmpty() {
					continue
				}

				id := et.GetID()
				fn := et.GetFullPath()
				vd := et.IsVendor()

				if id < 1 {
					o.u.ErrorStackData(fmt.Sprintf("[updFunc][File %s][Function %s] trigger error on getting type input", fs, tf.Name.Name), et, errors.New("empty ID for entry"))
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

	// Process output (return) parameters from the function signature.
	// Similar to input parameters, extract types and track dependencies.
	if tf.Type.Results != nil {
		for _, p := range tf.Type.Results.List {
			pr, dp, er := o.getParams(ctx, pkg, fs, afs, tf.Name.Name, p)

			if er != nil {
				o.u.ErrorStack(fmt.Sprintf("[updFunc][File %s][Function %s] trigger error on getting result params", fs, tf.Name.Name), er)
				return
			}

			ent.AddOutputs(pr...)

			for _, et := range dp {
				if et == nil || et.IsEmpty() {
					continue
				}

				id := et.GetID()
				fn := et.GetFullPath()
				vd := et.IsVendor()

				if id < 1 {
					o.u.ErrorStackData(fmt.Sprintf("[updFunc][File %s][Function %s] trigger error on getting type result", fs, tf.Name.Name), et, errors.New("empty ID for entry"))
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

	// Use ast.Inspect to recursively traverse all AST nodes within the function body.
	// Two kinds of nodes are matched to resolve call dependencies:
	//   - *ast.SelectorExpr: e.g., the "fmt.Println" selection expression part of a
	//     call like "fmt.Println()"; resolves external package and function entries
	//     via getDepend.
	//   - *ast.CallExpr: e.g., local function calls like "myFunc()" or call expressions
	//     with an unqualified receiver; resolves local entries within the same package
	//     context via getLocDepend. When CallExpr's Fun is a SelectorExpr, both cases
	//     will execute for the same call, but the deduplication map prevents duplicates.
	// Context cancellation is checked at each visited node for graceful shutdown.
	ast.Inspect(tf.Body, func(n ast.Node) bool {
		if ctx.Err() != nil {
			return false
		}

		var et audent.Entry
		switch expr := n.(type) {
		case *ast.SelectorExpr:
			// External dependency: a qualified identifier like pkg.Func()
			// Resolve the package and function entry via getDepend.
			if expr != nil {
				et = o.getDepend(ctx, fs, afs, expr)
			}
		case *ast.CallExpr:
			// Local dependency: an unqualified function call like myFunc()
			// Resolve the local entry via getLocDepend.
			if expr != nil {
				et = o.getLocDepend(ctx, fs, pkg, expr)
			}
		default:
			// Skip all other node types (variable declarations, loops, conditionals, etc.)
			return true
		}

		if et == nil || et.IsEmpty() {
			return true
		}

		id := et.GetID()
		fn := et.GetFullPath()
		vd := et.IsVendor()

		if id < 1 {
			o.u.ErrorStackData(fmt.Sprintf("[updFunc][File %s][Function %s] trigger error on getting type dependencies", fs, tf.Name.Name), et, errors.New("empty ID for entry"))
			return true
		}

		if dep[fn] {
			return true
		}

		dep[fn] = true

		if vd {
			o.a.UnSetStatusOutdated(id)
		}

		ent.AddDepend(id)
		return true
	})

	// Re-add the updated entry to the package collection and log the result.
	o.a.EntAdd(ent)
	o.u.Info("[updFunc][File %s][Function %s] Entry updated '%s' (id: %d) / Package '%s' (id %d) / Module '%s' (id: %d)", fs, tf.Name.Name, ent.GetFullPath(), ent.GetID(), ent.GetPkg().GetFullPath(), ent.GetPkgId(), ent.GetPkg().GetMod().GetMod(), ent.GetPkg().GetModId())
}
