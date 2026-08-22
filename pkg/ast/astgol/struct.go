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

package astgol

import (
	"context"
	"fmt"
	"go/ast"

	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audprm "github.com/nabbar/auditor/pkg/data/params"
	audtps "github.com/nabbar/auditor/pkg/data/types"
)

// parseStructType analyzes the fields of a Go struct to extract inputs and dependencies.
//
// This function processes Go struct type declarations by examining their field definitions
// and extracting parameter information along with associated dependencies. It handles both
// named fields and embedded types, creating comprehensive entry representations for structs
// that can be stored in the auditor's data collection system.
//
// Key features:
//   - Processes struct fields to extract input parameters and their types
//   - Identifies and tracks dependencies from field types through getTypeDepend analysis
//   - Creates or merges entry objects with appropriate metadata for storage
//   - Handles source code content retrieval for documentation purposes
//   - Manages dependency tracking to prevent duplicate entries in the dependency graph
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this struct, used for context association.
//   - fs: The file path containing this struct declaration, used for logging and error reporting.
//   - afs: The AST File node representing the complete file structure, providing context for position calculations.
//   - atc: The AST Node for the struct type specification, specifying the exact location within the file.
//   - name: The name of the struct type being processed.
//   - st: The AST StructType node containing field definitions to analyze.
//   - tcd: The categorized code type for the struct (EntryStruct).
//   - tst: The string representation of the struct type for documentation purposes.
//
// Returns:
//   - error: Any error encountered during struct analysis, including invalid AST structures,
//     parameter extraction issues, or problems with Entry creation. This ensures that individual
//     struct processing errors do not prevent continued analysis of other types in the same file.
func (o *mdl) parseStructType(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, atc ast.Node, name string, st *ast.StructType, tcd audtps.CodeType, tst string) error {
	// Validate that the struct type and its fields are not nil
	if st == nil || st.Fields == nil {
		return nil
	}

	var (
		err error
		dep = make(map[string]bool)
		cnt []byte
		ent audent.Entry
	)

	o.u.Info("[File %s][Struct %s] Parsing entry struct", fs, name)

	// Retrieve source code content for documentation purposes
	if cnt, err = o.getSrc(ctx, fs, afs, atc); err != nil {
		return fmt.Errorf("cannot retrieve content of file %s: %v", fs, err)
	}

	// Create or merge an Entry object with appropriate struct information for storage
	ent = o.a.EntNewOrMerge(pkg, name, tst, tcd, Lang, cnt)

	// Validate that the Entry was successfully created and is not empty
	if ent == nil || ent.IsEmpty() {
		return fmt.Errorf("invalid function %s in file %s", name, fs)
	}

	// Process each field in the struct to extract parameter information
	for _, f := range st.Fields.List {
		pr, dp, er := o.getParams(ctx, pkg, fs, afs, name, f)

		if er != nil {
			return er
		}

		// Process dependencies from field types
		for _, et := range dp {
			if et == nil {
				continue
			}

			id := et.GetID()
			fn := et.GetFullPath()

			// Skip duplicate dependencies to prevent redundant tracking
			if dep[fn] {
				o.u.Info("[File %s][Struct %s] Skip existing dependency for entry params", fs, name)
				continue
			}

			dep[fn] = true

			o.u.Info("[File %s][Struct %s] Add dependency %s", fs, name, et.GetFullPath())
			ent.AddDepend(id)
		}

		// Process input parameters for the struct fields
		for _, prm := range pr {
			o.u.Info("[File %s][Struct %s] Add Input Params type %s", fs, name, prm.String())
		}

		// Add input parameters to the entry for comprehensive tracking
		ent.AddInputs(pr...)
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	o.u.Info("[File %s][Struct %s] Update entry struct", fs, name)
	o.a.EntAdd(ent)
	return nil
}

// parseInterfaceType analyzes the methods and embedded types of a Go interface to extract I/O and dependencies.
//
// This function processes Go interface type declarations by examining their method definitions
// and embedded types. It handles both explicitly declared named methods and embedded interfaces,
// extracting parameter information along with associated dependencies for comprehensive interface analysis.
//
// Key features:
//   - Processes interface methods to extract input/output parameters and their types
//   - Identifies and tracks dependencies from method signatures through getTypeDepend analysis
//   - Handles embedded interfaces by analyzing their type specifications
//   - Creates or merges entry objects with appropriate metadata for storage
//   - Manages dependency tracking to prevent duplicate entries in the dependency graph
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this interface, used for context association.
//   - fs: The file path containing this interface declaration, used for logging and error reporting.
//   - afs: The AST File node representing the complete file structure, providing context for position calculations.
//   - atc: The AST Node for the interface type specification, specifying the exact location within the file.
//   - name: The name of the interface type being processed.
//   - it: The AST InterfaceType node containing method definitions to analyze.
//   - tcd: The categorized code type for the interface (EntryInterface).
//   - tst: The string representation of the interface type for documentation purposes.
//
// Returns:
//   - error: Any error encountered during interface analysis, including invalid AST structures,
//     parameter extraction issues, or problems with Entry creation. This ensures that individual
//     interface processing errors do not prevent continued analysis of other types in the same file.
func (o *mdl) parseInterfaceType(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, atc ast.Node, name string, it *ast.InterfaceType, tcd audtps.CodeType, tst string) error {
	// Validate that the interface type and its methods are not nil
	if it == nil || it.Methods == nil {
		return nil
	}

	var (
		err error
		dep = make(map[string]bool)
		cnt []byte
		ent audent.Entry
	)

	o.u.Info("[File %s][Interface %s] Parsing entry interface", fs, name)

	// Retrieve source code content for documentation purposes
	if cnt, err = o.getSrc(ctx, fs, afs, atc); err != nil {
		return fmt.Errorf("cannot retrieve content of file %s: %v", fs, err)
	}

	// Create or merge an Entry object with appropriate interface information for storage
	ent = o.a.EntNewOrMerge(pkg, name, tst, tcd, Lang, cnt)

	// Validate that the Entry was successfully created and is not empty
	if ent == nil || ent.IsEmpty() {
		return fmt.Errorf("invalid function %s in file %s", name, fs)
	}

	// Process each method in the interface to extract parameter information
	for _, m := range it.Methods.List {

		// Explicit declared named method (ex: MyFunc(context.Context) error)
		if len(m.Names) > 0 {
			ft, ok := m.Type.(*ast.FuncType)

			if !ok {
				continue
			}

			mn := m.Names[0].Name // Method Name
			mp := audprm.New(mn, audtps.EntryFunction)

			// Process input parameters for the method
			if ft.Params != nil && len(ft.Params.List) > 0 {
				for _, p := range ft.Params.List {
					pr, dp, er := o.getParams(ctx, pkg, fs, afs, mn, p)

					if er != nil {
						return er
					}

					// Process dependencies from input parameters
					for _, et := range dp {
						if et == nil {
							continue
						}

						id := et.GetID()
						fn := et.GetFullPath()

						// Skip duplicate dependencies to prevent redundant tracking
						if dep[fn] {
							o.u.Info("[File %s][Interface %s][Method %s] Skip existing dependency for input params", fs, name, mn)
							continue
						}

						dep[fn] = true

						o.u.Info("[File %s][Interface %s][Method %s] Add dependency %s", fs, name, et.GetFullPath(), mn)
						ent.AddDepend(id)
					}

					// Process input parameters for the method
					for _, prm := range pr {
						o.u.Info("[File %s][Interface %s][Method %s] Add Input Params type %s", fs, name, prm.String(), mn)
					}

					mp.SetInput(pr...)
				}
			}

			// Process output parameters for the method
			if ft.Results != nil {
				for _, p := range ft.Results.List {
					pr, dp, er := o.getParams(ctx, pkg, fs, afs, mn, p)

					if er != nil {
						return er
					}

					// Process dependencies from output parameters
					for _, et := range dp {
						if et == nil {
							continue
						}

						id := et.GetID()
						fn := et.GetFullPath()

						// Skip duplicate dependencies to prevent redundant tracking
						if dep[fn] {
							o.u.Info("[File %s][Interface %s][Method %s] Skip existing dependency for output result", fs, name, mn)
							continue
						}

						dep[fn] = true

						o.u.Info("[File %s][Interface %s][Method %s] Add dependency %s", fs, name, et.GetFullPath(), mn)
						ent.AddDepend(id)
					}

					// Process output parameters for the method
					for _, prm := range pr {
						o.u.Info("[File %s][Interface %s][Method %s] Add Output Result type %s", fs, name, prm.String(), mn)
					}

					mp.SetOutput(pr...)
				}
			}

			o.u.Info("[File %s][Interface %s] Add Method %s as entry params for interface", fs, name, mn)
			ent.AddInputs(mp)
			// end for named method
			continue
		}

		// Embedded Interface (ex: myPackage.MyInterface)
		// root dependencies ignored (cf resolveRootIdent)
		if let := o.getTypeDepend(ctx, fs, afs, pkg, m.Type); len(let) > 0 {
			for _, et := range let {
				if et == nil || et.IsEmpty() {
					continue
				}

				id := et.GetID()
				fn := et.GetFullPath()

				// Skip duplicate dependencies to prevent redundant tracking
				if !dep[fn] {
					dep[fn] = true
					o.u.Info("[File %s][Type %s] Add embedded interface dependency %s", fs, name, fn)
					ent.AddDepend(id)
				}
			}

			// end for dependencies for embedded interface
		}

		// Process embedded interfaces by extracting their type information
		_, mn, _ := o.parseExprCode(ctx, m.Type)
		mp := audprm.New(mn, audtps.EntryInterface)
		o.u.Info("[File %s][Interface %s] Add embedded interface %s as entry params for interface", fs, name, mn)
		ent.AddInputs(mp)

		// end for embedded interfaces
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	o.u.Info("[File %s][Struct %s] Update entry struct", fs, name)
	o.a.EntAdd(ent)
	return nil
}
