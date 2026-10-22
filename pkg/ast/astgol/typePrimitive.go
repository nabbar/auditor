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

// setPrimitive processes a primitive type declaration (*ast.TypeSpec) from a Go source file
// and creates or merges an entry in the auditor's data model.
//
// Primitive types are type definitions based on built-in Go types (e.g., type MyInt int,
// type MyString string). The function validates that the parsed type is EntryPrimitive,
// extracts the source code, creates or merges an entry, and registers it.
//
// The function:
//  1. Validates the TypeSpec node and package identifier.
//  2. Retrieves the source code for the type declaration.
//  3. Uses parseExprCode to determine the type classification and string representation.
//  4. Verifies that the parsed type is EntryPrimitive.
//  5. Creates or merges an entry with EntryPrimitive as the code type.
//  6. Sets the entry status and registers it in the package collection.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this primitive type.
//   - fs: The file path containing the type declaration.
//   - afs: The AST File structure containing the type.
//   - node: The AST Node representing the type declaration, expected to be *ast.TypeSpec.
func (o *mdl) setPrimitive(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, node ast.Node) {
	var (
		err error
		cnt []byte
		ent audent.Entry
		tpc audtps.CodeType
		tps string

		ok bool
		ts *ast.TypeSpec
	)

	// Type-assert the node to *ast.TypeSpec. Validate that the type spec is not nil,
	// has a non-nil name, and the name is not empty.
	ts, ok = node.(*ast.TypeSpec)
	if !ok || ts == nil || ts.Name == nil || len(ts.Name.Name) < 1 {
		err = errors.New("invalid AST type spec for primitive")
		o.u.ErrorStack("[setPrimitive] trigger error", err)
		return
	}

	// Validate that a non-zero package identifier is provided.
	if pkg == 0 {
		err = errors.New("invalid nil package")
		o.u.ErrorStack("[setPrimitive] trigger error", err)
		return
	}

	o.u.Info("[setPrimitive][File %s][Primitive %s] Setting entry primitive", fs, ts.Name.Name)

	// Retrieve the raw source code for the TypeSpec node from the file.
	cnt, err = o.getSrc(ctx, fs, afs, ts)
	if err != nil {
		o.u.ErrorStack("[setPrimitive] trigger error on retrieve source code", err)
		return
	}

	// Use parseExprCode to determine the type classification and string representation
	// of the type expression (ts.Type).
	tpc, tps, ok = o.parseExprCode(ctx, ts.Type)

	if !ok {
		err = fmt.Errorf("cannot identify type entry for %s", ts.Name.Name)
		o.u.ErrorStack("[setPrimitive] trigger error on retrieve type/code", err)
		return
	}

	// Verify that the parsed type is EntryPrimitive.
	if tpc != audtps.EntryPrimitive {
		err = fmt.Errorf("invalid entry primitive for %s: %s - %s", ts.Name.Name, tpc, tps)
		o.u.ErrorStack("[setPrimitive] trigger error on checking type/code", err)
		return
	}

	// Create a new entry or merge with an existing one for this primitive type.
	if ent = o.a.EntNewOrMerge(pkg, ts.Name.Name, tps, audtps.EntryPrimitive, Lang, cnt); ent == nil || ent.IsEmpty() {
		err = fmt.Errorf("invalid primitive entry %s", ts.Name.Name)
		o.u.ErrorStack("[setPrimitive] trigger error on creating/updating entry", err)
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
	o.u.Info("[setPrimitive][File %s][Primitive %s] Entry added '%s' (id: %d) / Package '%s' (id %d) / Module '%s' (id: %d)", fs, ts.Name.Name, ent.GetFullPath(), ent.GetID(), ent.GetPkg().GetFullPath(), ent.GetPkgId(), ent.GetPkg().GetMod().GetMod(), ent.GetPkg().GetModId())
}

// updPrimitive updates an existing primitive type entry (*ast.TypeSpec) in the auditor's data model.
//
// This function handles updating primitive type declarations that have already been
// created by setPrimitive. It retrieves the existing entry, verifies the type is still
// EntryPrimitive, updates type information if changed, collects type dependencies,
// and re-registers the entry.
//
// The update process:
//  1. Validates the TypeSpec node and package identifier.
//  2. Finds the existing entry by package ID and type name.
//  3. Uses parseExprCode to verify the type is still EntryPrimitive.
//  4. Updates the entry's type if it has changed.
//  5. Collects type dependencies from the type expression.
//  6. Re-adds the entry to the package collection.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this primitive type.
//   - fs: The file path containing the type declaration.
//   - afs: The AST File structure containing the type.
//   - node: The AST Node representing the type declaration, expected to be *ast.TypeSpec.
func (o *mdl) updPrimitive(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, node ast.Node) {
	var (
		err error
		dep = make(map[string]bool)
		ent audent.Entry
		tpc audtps.CodeType
		tps string

		ok bool
		ts *ast.TypeSpec
	)

	// Type-assert the node to *ast.TypeSpec. Validate that the type spec is not nil,
	// has a non-nil name, and the name is not empty.
	ts, ok = node.(*ast.TypeSpec)
	if !ok || ts == nil || ts.Name == nil || len(ts.Name.Name) < 1 {
		err = errors.New("invalid AST type spec for primitive")
		o.u.ErrorStack("[updPrimitive] trigger error", err)
		return
	}

	// Validate that a non-zero package identifier is provided.
	if pkg == 0 {
		err = errors.New("invalid nil package")
		o.u.ErrorStack("[updPrimitive] trigger error", err)
		return
	}

	// Find the existing entry by package ID and type name.
	if ent, err = o.findEntry(pkg, ts.Name.Name); err != nil {
		o.u.ErrorStack("[updPrimitive] trigger error on finding entry", err)
		return
	} else if ent == nil || ent.IsEmpty() {
		err = fmt.Errorf("cannot find primitive entry %s for package %d", ts.Name.Name, pkg)
		o.u.ErrorStack("[updPrimitive] trigger error on finding entry", err)
		return
	}

	o.u.Info("[updPrimitive][File %s][Primitive %s] Updating entry primitive", fs, ts.Name.Name)

	// Use parseExprCode to determine the type classification and string representation
	// of the type expression (ts.Type).
	tpc, tps, ok = o.parseExprCode(ctx, ts.Type)

	if !ok {
		err = fmt.Errorf("cannot identify type entry for %s", ts.Name.Name)
		o.u.ErrorStack("[updPrimitive] trigger error on retrieve type/code", err)
		return
	}

	// Verify that the parsed type is EntryPrimitive.
	if tpc != audtps.EntryPrimitive {
		err = fmt.Errorf("invalid entry primitive for %s: %s - %s", ts.Name.Name, tpc, tps)
		o.u.ErrorStack("[updPrimitive] trigger error on checking type/code", err)
		return
	}

	// Check and update the entry's type. If the code type or type string
	// has changed, update the entry with the new type information.
	if c, s := ent.GetType(); c != audtps.EntryPrimitive || s != tps {
		ent.SetType(audtps.EntryPrimitive, tps)
	}

	// Collect type dependencies from the type expression (ts.Type).
	// Dependencies are tracked using a deduplication map to avoid duplicates.
	for _, et := range o.getTypeDepend(ctx, fs, afs, pkg, ts.Type) {
		if et != nil && !et.IsEmpty() {
			id := et.GetID()
			fn := et.GetFullPath()
			vd := et.IsVendor()

			// Skip entries with invalid IDs (less than 1).
			if id < 1 {
				o.u.ErrorStackData(fmt.Sprintf("[updPrimitive][File %s] trigger error on getting dependencies", fs), et, errors.New("empty ID for entry"))
				continue
			}

			// Skip duplicate dependencies (same full path already processed).
			if dep[fn] {
				continue
			}

			dep[fn] = true

			// If the dependency is from a vendor package, update its status
			// to indicate it is being used (not outdated).
			if vd {
				o.a.UnSetStatusOutdated(id)
			}

			// Add the dependency to the primitive type entry.
			ent.AddDepend(id)
		}
	}

	// Re-add the updated entry to the package collection and log the result.
	o.a.EntAdd(ent)
	o.u.Info("[updPrimitive][File %s][Primitive %s] Entry updated '%s' (id: %d) / Package '%s' (id %d) / Module '%s' (id: %d)", fs, ts.Name.Name, ent.GetFullPath(), ent.GetID(), ent.GetPkg().GetFullPath(), ent.GetPkgId(), ent.GetPkg().GetMod().GetMod(), ent.GetPkg().GetModId())
}
