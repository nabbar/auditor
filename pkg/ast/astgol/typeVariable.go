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

// setVariable processes a variable declaration (*ast.ValueSpec) from a Go source file
// and creates or merges an entry in the auditor's data model.
//
// This function handles the parsing of variable declarations found in Go source code.
// It extracts the source code for the variable, determines its type (explicit, inherited,
// or inferred from value), creates a new entry or merges with an existing one, and
// registers it in the package's entry collection.
//
// The function processes each named variable in the ValueSpec individually. Blank
// identifiers (_) and unnamed variables are skipped. For each valid variable name:
//  1. The source code content is retrieved via getSrc using the ValueSpec node position.
//  2. The type is determined by first checking for an explicit type declaration,
//     then falling back to inherited types from the GenDecl block, and finally
//     inferring from the assigned value expression.
//  3. An entry is created or merged with EntNewOrMerge using EntryVariable as the code type.
//  4. The entry status is set to Parsed, and vendor/outdated flags are configured.
//  5. The entry is added to the package collection via EntAdd.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this variable, used for context association.
//   - fs: The file path containing the variable declaration, used for logging and error reporting.
//   - afs: The AST File structure containing the variable, providing access to source code content.
//   - node: The AST Node representing the variable declaration, expected to be *ast.ValueSpec.
func (o *mdl) setVariable(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, node ast.Node) {
	var (
		err error
		cnt []byte
		ent audent.Entry
		tps string

		ok bool
		vs *ast.ValueSpec
	)

	// Type-assert the node to *ast.ValueSpec. If the node is not a ValueSpec,
	// or if it is nil, or if it contains no names, the declaration is invalid
	// and processing is aborted with an error.
	vs, ok = node.(*ast.ValueSpec)

	if !ok || vs == nil || len(vs.Names) < 1 {
		err = errors.New("invalid AST type spec for variable")
		o.u.ErrorStack("[setVariable] trigger error", err)
		return
	}

	// Validate that a non-zero package identifier is provided.
	if pkg == 0 {
		err = errors.New("invalid nil package")
		o.u.ErrorStack("[setVariable] trigger error", err)
		return
	}

	// Retrieve the raw source code for the ValueSpec node from the file.
	cnt, err = o.getSrc(ctx, fs, afs, vs)
	if err != nil {
		o.u.ErrorStack("[setVariable] trigger error on retrieve source code", err)
		return
	}

	// Iterate over all names declared in this ValueSpec.
	for i, itm := range vs.Names {
		// Skip nil entries, empty names, and blank identifiers (_).
		if itm == nil || len(itm.Name) < 1 || itm.Name == "_" {
			continue
		}

		o.u.Info("[setVariable][File %s][Variable %s] Setting entry variable", fs, itm.Name)

		// Attempt to retrieve the explicit or inherited type.
		_, tps, ok = o.parseValueSpecType(ctx, afs, vs)
		if !ok {
			var vals = vs.Values

			// If there are no explicit values on this ValueSpec, try to inherit
			// values from a previous ValueSpec in the same GenDecl block.
			if len(vals) < 1 {
				_, vals = o.getInheritedValueSpec(afs, vs)
			}

			// Implicit type inferred from the corresponding value expression.
			if len(vals) > i && vals[i] != nil {
				_, tps, ok = o.parseExprTypeFromExpr(ctx, vals[i])
			} else if len(vals) == 1 && vals[0] != nil {
				_, tps, ok = o.parseExprTypeFromExpr(ctx, vals[0])
			}
		}

		// If the type could not be determined, log an error and skip this variable.
		if !ok {
			err = fmt.Errorf("cannot identify type entry for %s", itm.Name)
			o.u.ErrorStack("[setVariable] trigger error on identify type/code", err)
			continue
		}

		// Create a new entry or merge with an existing one for this variable.
		if ent = o.a.EntNewOrMerge(pkg, itm.Name, tps, audtps.EntryVariable, Lang, cnt); ent == nil || ent.IsEmpty() {
			err = fmt.Errorf("invalid variable entry %s", itm.Name)
			o.u.ErrorStack("[setVariable] trigger error on creating/updating entry", err)
			continue
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
		o.u.Info("[setVariable][File %s][Variable %s] Entry added '%s' (id: %d) / Package '%s' (id %d) / Module '%s' (id: %d)", fs, itm.Name, ent.GetFullPath(), ent.GetID(), ent.GetPkg().GetFullPath(), ent.GetPkgId(), ent.GetPkg().GetMod().GetMod(), ent.GetPkg().GetModId())
	}
}

// updVariable updates an existing variable entry (*ast.ValueSpec) in the auditor's data model.
//
// This function handles the updating of variable declarations that have already been
// created during a previous parsing pass. It retrieves the existing entry, updates its
// type information, collects any type dependencies, and re-registers the entry.
//
// The update process follows these steps for each named variable:
//  1. Validate the ValueSpec node and package identifier.
//  2. Determine the type (explicit, inherited, or inferred from value).
//  3. Find the existing entry by package ID and variable name.
//  4. Update the entry's type if it has changed.
//  5. Collect type dependencies (external packages referenced by the variable's type).
//  6. Re-add the entry to the package collection.
//
// Unlike setVariable, updVariable does not retrieve source code content because the
// entry already exists. Instead, it focuses on type validation and dependency tracking.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this variable.
//   - fs: The file path containing the variable declaration.
//   - afs: The AST File structure containing the variable.
//   - node: The AST Node representing the variable declaration, expected to be *ast.ValueSpec.
func (o *mdl) updVariable(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, node ast.Node) {
	var (
		err error
		dep map[string]bool
		ent audent.Entry
		tps string

		ok bool
		vs *ast.ValueSpec
		pr ast.Expr
	)

	// Type-assert the node to *ast.ValueSpec.
	vs, ok = node.(*ast.ValueSpec)

	if !ok || vs == nil || len(vs.Names) < 1 {
		err = errors.New("invalid AST type spec for variable")
		o.u.ErrorStack("[updVariable] trigger error", err)
		return
	}

	// Validate that a non-zero package identifier is provided.
	if pkg == 0 {
		err = errors.New("invalid nil package")
		o.u.ErrorStack("[updVariable] trigger error", err)
		return
	}

	// Iterate over all names declared in this ValueSpec.
	for i, itm := range vs.Names {
		// Skip nil entries, empty names, and blank identifiers (_).
		if itm == nil || len(itm.Name) < 1 || itm.Name == "_" {
			continue
		}

		pr = nil
		dep = make(map[string]bool)

		o.u.Info("[updVariable][File %s][Variable %s] Updating entry variable", fs, itm.Name)

		// Attempt to retrieve the explicit or inherited type.
		_, tps, ok = o.parseValueSpecType(ctx, afs, vs)
		if !ok {
			var vals = vs.Values

			if len(vals) < 1 {
				_, vals = o.getInheritedValueSpec(afs, vs)
			}

			// Implicit type inferred from the corresponding value expression.
			// The parent expression (pr) is captured for dependency tracking.
			if len(vals) > i && vals[i] != nil {
				pr = vals[i]
				_, tps, ok = o.parseExprTypeFromExpr(ctx, vals[i])
			} else if len(vals) == 1 && vals[0] != nil {
				pr = vals[0]
				_, tps, ok = o.parseExprTypeFromExpr(ctx, vals[0])
			}
		} else {
			// If the type is explicitly declared (vs.Type != nil), use it as the
			// parent expression for dependency tracking.
			// Otherwise, inherit values from a previous ValueSpec in the block.
			if vs.Type != nil {
				pr = vs.Type
			} else {
				_, vals := o.getInheritedValueSpec(afs, vs)
				if len(vals) > i && vals[i] != nil {
					pr = vals[i]
				} else if len(vals) == 1 && vals[0] != nil {
					pr = vals[0]
				}
			}
		}

		// If the type could not be determined, log an error and skip this variable.
		if !ok {
			err = fmt.Errorf("cannot identify type entry for %s", itm.Name)
			o.u.ErrorStack("[updVariable] trigger error on checking type/code", err)
			continue
		}

		// Find the existing entry by package ID and variable name.
		if ent, err = o.findEntry(pkg, itm.Name); err != nil {
			o.u.ErrorStack("[updVariable] trigger error on find variable entry", err)
			continue
		} else if ent == nil || ent.IsEmpty() {
			err = fmt.Errorf("cannot find variable entry %s for package %d", itm.Name, pkg)
			o.u.ErrorStack("[updVariable] trigger error on find variable entry", err)
			continue
		}

		// Check and update the entry's type.
		if c, s := ent.GetType(); c != audtps.EntryVariable || s != tps {
			ent.SetType(audtps.EntryVariable, tps)
		}

		// Collect type dependencies from the parent expression (pr).
		for _, et := range o.getTypeDepend(ctx, fs, afs, pkg, pr) {
			if et != nil && !et.IsEmpty() {
				id := et.GetID()
				fn := et.GetFullPath()
				vd := et.IsVendor()

				if id < 1 {
					o.u.ErrorStackData(fmt.Sprintf("[updVariable][File %s] trigger error on getting dependencies", fs), et, errors.New("empty ID for entry"))
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

		// Re-add the updated entry to the package collection and log the result.
		o.a.EntAdd(ent)
		o.u.Info("[updVariable][File %s][Variable %s] Entry updated '%s' (id: %d) / Package '%s' (id %d) / Module '%s' (id: %d)", fs, itm.Name, ent.GetFullPath(), ent.GetID(), ent.GetPkg().GetFullPath(), ent.GetPkgId(), ent.GetPkg().GetMod().GetMod(), ent.GetPkg().GetModId())
	}
}
