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

// setConstant processes a constant declaration (*ast.ValueSpec) from a Go source file
// and creates or merges an entry in the auditor's data model.
//
// This function handles the parsing of constant declarations found in Go source code.
// It extracts the source code for the constant, determines its type (explicit, inherited,
// or inferred from value), creates a new entry or merges with an existing one, and
// registers it in the package's entry collection.
//
// The function processes each named constant in the ValueSpec individually. Blank
// identifiers (_) and unnamed constants are skipped. For each valid constant name:
//  1. The source code content is retrieved via getSrc using the ValueSpec node position.
//  2. The type is determined by first checking for an explicit type declaration,
//     then falling back to inherited types from the GenDecl block, and finally
//     inferring from the assigned value expression.
//  3. An entry is created or merged with EntNewOrMerge using EntryConstant as the code type.
//  4. The entry status is set to Parsed, and vendor/outdated flags are configured.
//  5. The entry is added to the package collection via EntAdd.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this constant, used for context association.
//   - fs: The file path containing the constant declaration, used for logging and error reporting.
//   - afs: The AST File structure containing the constant, providing access to source code content.
//   - node: The AST Node representing the constant declaration, expected to be *ast.ValueSpec.
func (o *mdl) setConstant(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, node ast.Node) {
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
		err = errors.New("invalid AST type spec for constant")
		o.u.ErrorStack("[setConstant] trigger error", err)
		return
	}

	// Validate that a non-zero package identifier is provided. A zero package ID
	// indicates that the package has not been properly registered, which would
	// make it impossible to associate the constant entry with a package.
	if pkg == 0 {
		err = errors.New("invalid nil package")
		o.u.ErrorStack("[setConstant] trigger error", err)
		return
	}

	// Retrieve the raw source code for the ValueSpec node from the file.
	// This includes the constant declaration text and any associated documentation
	// comments, which will be stored with the entry for reference.
	cnt, err = o.getSrc(ctx, fs, afs, vs)
	if err != nil {
		o.u.ErrorStack("[setConstant] trigger error on retrieve source code", err)
		return
	}

	// Iterate over all names declared in this ValueSpec. A single ValueSpec can
	// declare multiple constants (e.g., "const (a, b, c = 1, 2, 3)").
	for i, itm := range vs.Names {
		// Skip nil entries, empty names, and blank identifiers (_).
		// Blank identifiers are not valid constant names and should be ignored.
		if itm == nil || len(itm.Name) < 1 || itm.Name == "_" {
			continue
		}

		o.u.Info("[setConstant][File %s][Constant %s] Setting entry constant", fs, itm.Name)

		// Attempt to retrieve the explicit or inherited type.
		// For const blocks using iota and a shared type, the type may be declared
		// on an earlier ValueSpec in the same block and inherited by subsequent ones.
		_, tps, ok = o.parseValueSpecType(ctx, afs, vs)
		if !ok {
			var vals = vs.Values

			// If there are no explicit values on this ValueSpec, try to inherit
			// values from a previous ValueSpec in the same GenDecl block.
			if len(vals) < 1 {
				_, vals = o.getInheritedValueSpec(afs, vs)
			}

			// Implicit type inferred from the corresponding value expression.
			// If this ValueSpec has multiple values, use the value at index i.
			// If there is only one value, use it for all names (shared type).
			if len(vals) > i && vals[i] != nil {
				_, tps, ok = o.parseExprTypeFromExpr(ctx, vals[i])
			} else if len(vals) == 1 && vals[0] != nil {
				_, tps, ok = o.parseExprTypeFromExpr(ctx, vals[0])
			}
		}

		// If the type could not be determined through any method, log an error
		// and skip this constant. Processing continues with the next constant.
		if !ok {
			err = fmt.Errorf("cannot identify type entry for %s", itm.Name)
			o.u.ErrorStack("[setConstant] trigger error on identify type/code", err)
			continue
		}

		// Create a new entry or merge with an existing one for this constant.
		// EntNewOrMerge handles both creation of new entries and updating of
		// existing entries with the same name in the same package.
		if ent = o.a.EntNewOrMerge(pkg, itm.Name, tps, audtps.EntryConstant, Lang, cnt); ent == nil || ent.IsEmpty() {
			err = fmt.Errorf("invalid constant entry %s", itm.Name)
			o.u.ErrorStack("[setConstant] trigger error on creating/updating entry", err)
			continue
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

		// Add the entry to the package collection and log the result with full
		// identification including entry ID, package path, and module information.
		o.a.EntAdd(ent)
		o.u.Info("[setConstant][File %s][Constant %s] Entry added '%s' (id: %d) / Package '%s' (id %d) / Module '%s' (id: %d)", fs, itm.Name, ent.GetFullPath(), ent.GetID(), ent.GetPkg().GetFullPath(), ent.GetPkgId(), ent.GetPkg().GetMod().GetMod(), ent.GetPkg().GetModId())
	}
}

// updConstant updates an existing constant entry (*ast.ValueSpec) in the auditor's data model.
//
// This function handles the updating of constant declarations that have already been
// created during a previous parsing pass. It retrieves the existing entry, updates its
// type information, collects any type dependencies, and re-registers the entry.
//
// The update process follows these steps for each named constant:
//  1. Validate the ValueSpec node and package identifier.
//  2. Determine the type (explicit, inherited, or inferred from value).
//  3. Find the existing entry by package ID and constant name.
//  4. Update the entry's type if it has changed.
//  5. Collect type dependencies (external packages referenced by the constant's type).
//  6. Re-add the entry to the package collection.
//
// Unlike setConstant, updConstant does not retrieve source code content because the
// entry already exists. Instead, it focuses on type validation and dependency tracking.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this constant.
//   - fs: The file path containing the constant declaration.
//   - afs: The AST File structure containing the constant.
//   - node: The AST Node representing the constant declaration, expected to be *ast.ValueSpec.
func (o *mdl) updConstant(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, node ast.Node) {
	var (
		err error
		dep map[string]bool
		ent audent.Entry
		tps string

		ok bool
		vs *ast.ValueSpec
		pr ast.Expr
	)

	// Type-assert the node to *ast.ValueSpec. If the node is not a ValueSpec,
	// or if it is nil, or if it contains no names, the declaration is invalid
	// and processing is aborted with an error.
	vs, ok = node.(*ast.ValueSpec)

	if !ok || vs == nil || len(vs.Names) < 1 {
		err = errors.New("invalid AST type spec for constant")
		o.u.ErrorStack("[updConstant] trigger error", err)
		return
	}

	// Validate that a non-zero package identifier is provided.
	if pkg == 0 {
		err = errors.New("invalid nil package")
		o.u.ErrorStack("[updConstant] trigger error", err)
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

		o.u.Info("[updConstant][File %s][Constant %s] Updating entry constant", fs, itm.Name)

		// Attempt to retrieve the explicit or inherited type.
		// For const blocks using iota and a shared type, the type may be declared
		// on an earlier ValueSpec in the same block and inherited by subsequent ones.
		_, tps, ok = o.parseValueSpecType(ctx, afs, vs)
		if !ok {
			var vals = vs.Values

			// If there are no explicit values on this ValueSpec, try to inherit
			// values from a previous ValueSpec in the same GenDecl block.
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

		// If the type could not be determined, log an error and skip this constant.
		if !ok {
			err = fmt.Errorf("cannot identify type entry for %s", itm.Name)
			o.u.ErrorStack("[updConstant] trigger error on checking type/code", err)
			continue
		}

		// Find the existing entry by package ID and constant name.
		// If the entry does not exist, log an error and skip.
		if ent, err = o.findEntry(pkg, itm.Name); err != nil {
			o.u.ErrorStack("[updConstant] trigger error on find constant entry", err)
			continue
		} else if ent == nil || ent.IsEmpty() {
			err = fmt.Errorf("cannot find constant entry %s for package %d", itm.Name, pkg)
			o.u.ErrorStack("[updConstant] trigger error on find constant entry", err)
			continue
		}

		// Check and update the entry's type. If the code type or type string
		// has changed, update the entry with the new type information.
		if c, s := ent.GetType(); c != audtps.EntryConstant || s != tps {
			ent.SetType(audtps.EntryConstant, tps)
		}

		// Collect type dependencies from the parent expression (pr).
		// Dependencies are tracked using a deduplication map to avoid adding
		// the same dependency multiple times.
		for _, et := range o.getTypeDepend(ctx, fs, afs, pkg, pr) {
			if et != nil && !et.IsEmpty() {
				id := et.GetID()
				fn := et.GetFullPath()
				vd := et.IsVendor()

				// Skip entries with invalid IDs (less than 1).
				if id < 1 {
					o.u.ErrorStackData(fmt.Sprintf("[updConstant][File %s] trigger error on getting dependencies", fs), et, errors.New("empty ID for entry"))
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

				// Add the dependency to the constant entry.
				ent.AddDepend(id)
			}
		}

		// Re-add the updated entry to the package collection and log the result.
		o.a.EntAdd(ent)
		o.u.Info("[updConstant][File %s][Constant %s] Entry updated '%s' (id: %d) / Package '%s' (id %d) / Module '%s' (id: %d)", fs, itm.Name, ent.GetFullPath(), ent.GetID(), ent.GetPkg().GetFullPath(), ent.GetPkgId(), ent.GetPkg().GetMod().GetMod(), ent.GetPkg().GetModId())
	}
}
