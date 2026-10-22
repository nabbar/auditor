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

// setStruct processes a struct type declaration (*ast.TypeSpec with *ast.StructType)
// from a Go source file and creates or merges an entry in the auditor's data model.
//
// This function handles the initial creation of struct entries during AST parsing.
// It validates that the type is indeed a struct, extracts the source code,
// creates a new entry or merges with an existing one, and registers it.
//
// The function:
//  1. Validates the TypeSpec node and package identifier.
//  2. Type-asserts that ts.Type is an *ast.StructType.
//  3. Retrieves the source code for the struct declaration.
//  4. Creates or merges an entry with EntryStruct as the code type.
//  5. Sets the entry status and registers it in the package collection.
//
// Note: Struct fields are populated later by updStruct during the update phase.
// This function only creates the basic struct entry.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this struct.
//   - fs: The file path containing the struct declaration.
//   - afs: The AST File structure containing the struct.
//   - node: The AST Node representing the type declaration, expected to be *ast.TypeSpec
//     with ts.Type being *ast.StructType.
func (o *mdl) setStruct(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, node ast.Node) {
	var (
		err error
		cnt []byte
		ent audent.Entry

		ok bool
		ts *ast.TypeSpec
		st *ast.StructType
	)

	// Type-assert the node to *ast.TypeSpec. Validate that the type spec is not nil,
	// has a non-nil name, and the name is not empty.
	if ts, ok = node.(*ast.TypeSpec); !ok || ts == nil || ts.Name == nil || len(ts.Name.Name) < 1 {
		err = errors.New("invalid AST type spec declaration for struct")
		o.u.ErrorStack("[setStruct] trigger error", err)
		return
	}

	// Type-assert that ts.Type is an *ast.StructType.
	if st, ok = ts.Type.(*ast.StructType); !ok || st == nil {
		err = fmt.Errorf("invalid struct type for %s", ts.Name.Name)
		o.u.ErrorStack("[setStruct] trigger error", err)
		return
	}

	// Validate that a non-zero package identifier is provided.
	if pkg == 0 {
		err = errors.New("invalid nil package")
		o.u.ErrorStack("[setStruct] trigger error", err)
		return
	}

	o.u.Info("[setStruct][File %s][Struct %s] Setting entry struct", fs, ts.Name.Name)

	// Retrieve the raw source code for the TypeSpec node from the file.
	if cnt, err = o.getSrc(ctx, fs, afs, ts); err != nil {
		err = fmt.Errorf("cannot retrieve content of file %s: %v", fs, err)
		o.u.ErrorStack("[setStruct] trigger error on retrieve source code", err)
		return
	}

	// Create a new entry or merge with an existing one for this struct.
	// The type string is empty; struct fields are populated by updStruct.
	if ent = o.a.EntNewOrMerge(pkg, ts.Name.Name, "", audtps.EntryStruct, Lang, cnt); ent == nil || ent.IsEmpty() {
		err = fmt.Errorf("invalid struct %s in file %s", ts.Name.Name, fs)
		o.u.ErrorStack("[setStruct] trigger error on creating/updating entry", err)
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
	o.u.Info("[setStruct][File %s][Struct %s] Entry added '%s' (id: %d) / Package '%s' (id %d) / Module '%s' (id: %d)", fs, ts.Name.Name, ent.GetFullPath(), ent.GetID(), ent.GetPkg().GetFullPath(), ent.GetPkgId(), ent.GetPkg().GetMod().GetMod(), ent.GetPkg().GetModId())
}

// updStruct updates an existing struct type entry (*ast.TypeSpec with *ast.StructType)
// in the auditor's data model.
//
// This function handles updating struct declarations that have already been
// created by setStruct. It retrieves the existing entry, verifies the type is still
// EntryStruct, processes struct fields to extract field types and dependencies,
// and re-registers the entry.
//
// The update process:
//  1. Validates the TypeSpec node and package identifier.
//  2. Type-asserts that ts.Type is an *ast.StructType.
//  3. Finds the existing entry by package ID and struct name.
//  4. Updates the entry's type to EntryStruct.
//  5. Iterates over all struct fields:
//     - For each field, uses getParams to extract field types and dependencies.
//     - Adds field parameters as inputs to the struct entry.
//     - Collects type dependencies from field types.
//  6. Re-adds the entry to the package collection.
//
// Dependencies are tracked using a deduplication map (keyed by full path) to avoid
// duplicates. Vendor dependencies have their outdated status cleared.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this struct.
//   - fs: The file path containing the struct declaration.
//   - afs: The AST File structure containing the struct.
//   - node: The AST Node representing the type declaration, expected to be *ast.TypeSpec
//     with ts.Type being *ast.StructType.
func (o *mdl) updStruct(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, node ast.Node) {
	var (
		err error
		dep = make(map[string]bool)
		ent audent.Entry

		ok bool
		ts *ast.TypeSpec
		st *ast.StructType
	)

	// Type-assert the node to *ast.TypeSpec. Validate that the type spec is not nil,
	// has a non-nil name, and the name is not empty.
	if ts, ok = node.(*ast.TypeSpec); !ok || ts == nil || ts.Name == nil || len(ts.Name.Name) < 1 {
		err = errors.New("invalid AST type spec declaration for struct")
		o.u.ErrorStack("[updStruct] trigger error", err)
		return
	}

	// Type-assert that ts.Type is an *ast.StructType.
	if st, ok = ts.Type.(*ast.StructType); !ok || st == nil {
		err = fmt.Errorf("invalid struct type for %s", ts.Name.Name)
		o.u.ErrorStack("[updStruct] trigger error", err)
		return
	}

	// Validate that a non-zero package identifier is provided.
	if pkg == 0 {
		err = errors.New("invalid nil package")
		o.u.ErrorStack("[updStruct] trigger error", err)
		return
	}

	// Find the existing entry by package ID and struct name.
	if ent, err = o.findEntry(pkg, ts.Name.Name); err != nil {
		o.u.ErrorStack("[updStruct] trigger error on find entry", err)
		return
	} else if ent == nil || ent.IsEmpty() {
		err = fmt.Errorf("cannot find struct entry %s for package %d", ts.Name.Name, pkg)
		o.u.ErrorStack("[updStruct] trigger error on find entry", err)
		return
	}

	o.u.Info("[updStruct][File %s][Struct %s] Updating entry struct", fs, ts.Name.Name)

	// Check and update the entry's type. Ensure it is set to EntryStruct with
	// an empty type string.
	if c, s := ent.GetType(); c != audtps.EntryStruct || s != "" {
		ent.SetType(audtps.EntryStruct, "")
	}

	// Process all fields declared in the struct.
	if st.Fields != nil {
		for _, f := range st.Fields.List {
			// Use getParams to extract field types, parameter objects, and dependencies.
			pr, dp, er := o.getParams(ctx, pkg, fs, afs, ts.Name.Name, f)

			if er != nil {
				o.u.ErrorStack(fmt.Sprintf("[updStruct][File %s][Struct %s] trigger error retrieve field", fs, ts.Name.Name), er)
				return
			}

			// Add the extracted field parameters as inputs to the struct entry.
			ent.AddInputs(pr...)

			// Collect type dependencies from field types (e.g., imported types used as field types).
			for _, et := range dp {
				if et == nil || et.IsEmpty() {
					continue
				}

				id := et.GetID()
				fn := et.GetFullPath()
				vd := et.IsVendor()

				// Skip entries with invalid IDs (less than 1).
				if id < 1 {
					o.u.ErrorStackData(fmt.Sprintf("[updStruct][File %s][Struct %s] trigger error on getting field dependencies", fs, ts.Name.Name), et, errors.New("empty ID for entry"))
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

				// Add the dependency to the struct entry.
				ent.AddDepend(id)
			}
		}
	}

	// Re-add the updated entry to the package collection and log the result.
	o.a.EntAdd(ent)
	o.u.Info("[updStruct][File %s][Struct %s] Entry updated '%s' (id: %d) / Package '%s' (id %d) / Module '%s' (id: %d)", fs, ts.Name.Name, ent.GetFullPath(), ent.GetID(), ent.GetPkg().GetFullPath(), ent.GetPkgId(), ent.GetPkg().GetMod().GetMod(), ent.GetPkg().GetModId())
}
