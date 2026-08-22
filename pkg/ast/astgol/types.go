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
	"fmt"
	"go/ast"
	"strings"

	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audprm "github.com/nabbar/auditor/pkg/data/params"
	audtps "github.com/nabbar/auditor/pkg/data/types"
)

// parseType extracts information from Go type declarations (structs, interfaces, and primitive types).
//
// This function serves as the core mechanism for processing Go type specifications within the AST parsing framework.
// It categorizes different types of Go constructs (structs, interfaces, primitives) and creates appropriate Entry objects
// for storage in the auditor's data collection system. The parseType method is crucial for understanding the structure
// and relationships between different types in Go code.
//
// The implementation provides comprehensive type analysis including:
// - Type name extraction from AST identifiers
// - Classification of types into primitive, custom, struct, or interface categories
// - Source code content handling for accurate documentation capture
// - Creation of Entry objects with appropriate type classifications for storage
//
// Key features:
//   - Extracts type names and categorizes them appropriately based on their structure
//   - Uses parseExprCode to determine the exact type classification and string representation
//   - Handles different type categories (primitive, custom, struct, interface) with proper categorization
//   - Creates Entry objects for storage in the auditor's data collection system with appropriate metadata
//   - Implements proper error handling for invalid type specifications and classifications
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this type declaration, used to associate
//     the type with its package context in the auditor's data model.
//   - fs: The file path containing this type declaration, used for logging and error reporting.
//   - afs: The AST File structure containing this type declaration, providing access to source code content.
//   - asp: The AST Type Specification node for processing, containing all information about
//     the type definition including name and underlying type structure.
//
// Returns:
//   - error: Any error encountered during type processing, including invalid AST structures,
//     classification issues, or problems with Entry creation. This ensures that individual type
//     processing errors do not prevent continued analysis of other types in the same file.
func (o *mdl) parseType(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, asp *ast.TypeSpec) error {
	var (
		ok  bool
		ent audent.Entry
		tst string
		tcd audtps.CodeType
		err error
		cnt []byte
	)

	// Validate the AST Type Specification node to ensure it's not nil and has a valid name
	if asp == nil || asp.Name == nil || len(asp.Name.Name) < 1 || asp.Name.Name == "_" {
		return fmt.Errorf("invalid AST Type Specification")
	}

	// Validate that the package identifier is valid (non-zero)
	if pkg == 0 {
		return fmt.Errorf("invalid nil package")
	}

	o.u.Info("[file %s][Type %s] Parsing Type", fs, asp.Name.Name)

	// Determine the type classification and string representation using recursive parsing
	tcd, tst, ok = o.parseExprCode(ctx, asp.Type)

	// Validate that parsing was successful and produced valid results
	if !ok {
		return fmt.Errorf("invalid AST Type Specification")
	}

	// Ensure that types requiring string representations have valid string values
	if tcd.NeedString() && len(tst) < 1 {
		return fmt.Errorf("invalid AST Type Specification")
	}

	// Dispatch to specialized parsing functions based on the type of AST node
	switch st := asp.Type.(type) {
	case *ast.StructType:
		// Handle struct type specifications with detailed analysis
		return o.parseStructType(ctx, pkg, fs, afs, asp, asp.Name.Name, st, tcd, tst)
	case *ast.InterfaceType:
		// Handle interface type specifications with detailed analysis
		return o.parseInterfaceType(ctx, pkg, fs, afs, asp, asp.Name.Name, st, tcd, tst)
	}

	// Retrieve the source code content for documentation purposes
	if cnt, err = o.getSrc(ctx, fs, afs, asp); cnt == nil {
		return fmt.Errorf("cannot retrieve content of file %s: %v", fs, err)
	}

	// Create or merge an Entry object with appropriate type information for storage
	ent = o.a.EntNewOrMerge(pkg, asp.Name.Name, tst, tcd, Lang, nil)

	// Validate that the Entry was successfully created and is not empty
	if ent == nil || ent.IsEmpty() {
		return fmt.Errorf("invalid entry %s (%s)", asp.Name.Name, strings.Join([]string{tcd.String(), tst}, " "))
	}

	o.u.Info("[File %s][Type %s] Add entry %s (id: %d)", fs, asp.Name.Name, ent.GetTypeString(), ent.GetID())
	o.a.EntAdd(ent)
	return nil
}

// parseExprCode recursively parses AST expressions to determine their code type and string representation.
//
// This function serves as the core recursive parser for Go AST expression types. It analyzes various expression
// nodes and determines their classification within the auditor's type system. The parseExprCode function is
// fundamental to understanding Go type structures and provides the foundation for categorizing different
// AST node types into appropriate categories for auditing purposes.
//
// The implementation provides comprehensive handling of different Go expression types:
// - Identifiers (primitive types, custom types)
// - Struct types
// - Interface types
// - Selector expressions (imported types)
// - Star expressions (pointers)
// - Array types (slices and arrays)
// - Map types
//
// Key features:
//   - Recursive parsing of complex expression structures to handle nested types
//   - Comprehensive type classification for Go AST nodes based on their fundamental characteristics
//   - Proper handling of nested types and composite structures through recursive analysis
//   - Accurate string representation generation for type descriptions that can be used in documentation
//   - Error handling for unrecognized or unsupported expression types that cannot be classified
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during parsing.
//   - expr: The AST Expression node to analyze and classify, which may be a simple identifier or complex structure.
//
// Returns:
//   - audtps.CodeType: The categorized code type for the expression, indicating whether it represents
//     a primitive, custom, struct, interface, or other type category.
//   - string: The string representation of the type for documentation purposes, providing a human-readable
//     description of the type structure.
//   - bool: A flag indicating whether the parsing was successful and the expression could be properly classified.
func (o *mdl) parseExprCode(ctx context.Context, expr ast.Expr) (audtps.CodeType, string, bool) {
	// Validate that the expression is not nil
	if expr == nil {
		return audtps.EntryNone, "", false
	}

	// Check for context cancellation to prevent processing when cancelled
	if ctx.Err() != nil {
		return audtps.EntryNone, "", false
	}

	// Analyze different types of AST expressions based on their node type
	switch t := expr.(type) {
	case *ast.Ident:
		// Handle identifier expressions (simple names like "int", "string")
		if primitives[t.Name] {
			return audtps.EntryPrimitive, t.Name, true
		}
		return audtps.EntryCustom, t.Name, true

	case *ast.FuncType:
		// Handle function type expressions (e.g., func() int)
		return audtps.EntryFunction, "func", true

	case *ast.StructType:
		// Handle struct type expressions (e.g., struct{ field int })
		return audtps.EntryStruct, "struct", true

	case *ast.InterfaceType:
		// Handle interface type expressions (e.g., interface{ Method() })
		return audtps.EntryInterface, "interface", true

	case *ast.SelectorExpr:
		// Handles imported types like "time.Time" or "http.Request"
		if ident, ok := t.X.(*ast.Ident); ok {
			return audtps.EntryCustom, ident.Name + "." + t.Sel.Name, true
		}
		return audtps.EntryCustom, t.Sel.Name, true

	case *ast.StarExpr:
		// Handles pointers (e.g., *int). We add "*" to the underlying type.
		tcd, tst, ok := o.parseExprCode(ctx, t.X)
		return tcd, "*" + tst, ok

	case *ast.ArrayType:
		// Handles slices and arrays (e.g., []string).
		tcd, tst, ok := o.parseExprCode(ctx, t.Elt)
		return tcd, "[]" + tst, ok

	case *ast.MapType:
		// Handles maps (e.g., map[string]int).
		_, keyStr, _ := o.parseExprCode(ctx, t.Key)
		tcd, valStr, ok := o.parseExprCode(ctx, t.Value)
		return tcd, "map[" + keyStr + "]" + valStr, ok

	case *ast.Ellipsis:
		// Handles ellipsis (e.g., ...string).
		tcd, tst, ok := o.parseExprCode(ctx, t.Elt)
		return tcd, "..." + tst, ok

	case *ast.ChanType:
		// Handles channel (e.g., chan []struct{}).
		tcd, tst, ok := o.parseExprCode(ctx, t.Value)
		return tcd, "chan " + tst, ok

	case *ast.IndexExpr:
		// Handles generic types with a single argument (e.g., Collection[Module])
		tcd, tstX, okX := o.parseExprCode(ctx, t.X)
		_, txtS, okS := o.parseExprCode(ctx, t.Index)
		return tcd, tstX + "[" + txtS + "]", okX && okS

	case *ast.IndexListExpr:
		// Handles generic types with multiple arguments (e.g., Collection[Module, Package])
		var (
			tstS []string
			okS  = true
		)

		tcd, tstX, okX := o.parseExprCode(ctx, t.X)

		for _, idx := range t.Indices {
			_, s, k := o.parseExprCode(ctx, idx)
			tstS = append(tstS, s)
			okS = okS && k
		}

		return tcd, tstX + "[" + strings.Join(tstS, ", ") + "]", okX && okS

	default:
		// Unrecognized or too complex for standard audit
		return audtps.EntryNone, "unknown", false
	}
}

// getParams extracts parameter information from AST fields and creates parameter objects.
//
// This function processes AST field nodes to extract parameter details including names, types,
// and dependencies. It handles complex type structures by recursively analyzing nested expressions
// and builds comprehensive parameter information for storage in the auditor's data model.
//
// Key features:
//   - Extracts parameter names from AST field identifiers
//   - Determines parameter types using parseExprCode for accurate classification
//   - Handles complex type structures including pointers, arrays, slices, maps, and channels
//   - Recursively processes function parameters and struct fields to build complete parameter information
//   - Creates parameter objects with appropriate input/output classifications
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this parameter, used for context association.
//   - fs: The file path containing this parameter, used for logging and error reporting.
//   - afs: The AST File structure containing this parameter, providing access to source code content.
//   - en: The name of the entry (function or type) that contains these parameters.
//   - p: The AST Field node representing a parameter in a function signature or struct field.
//
// Returns:
//   - []audprm.Params: A slice of parameter objects built from the AST field information.
//   - []audent.Entry: A slice of dependent entry objects that this parameter references.
//   - error: Any error encountered during parameter processing, including invalid AST structures or classification issues.
func (o *mdl) getParams(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, en string, p *ast.Field) ([]audprm.Params, []audent.Entry, error) {
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}

	var (
		ok bool

		nam string
		tcd audtps.CodeType
		tst string
		stp ast.Expr

		sin = make([]audprm.Params, 0)
		sou = make([]audprm.Params, 0)
		prm = make([]audprm.Params, 0)
		dep = make([]audent.Entry, 0)
	)

	// Determine the type classification and string representation for the parameter
	tcd, tst, ok = o.parseExprCode(ctx, p.Type)

	// Extract parameter name if available
	if len(p.Names) > 0 {
		nam = p.Names[0].Name
	}

	// Collect dependencies from type specifications
	for _, et := range o.getTypeDepend(ctx, fs, afs, pkg, p.Type) {
		if et == nil || et.IsEmpty() {
			continue
		}

		dep = append(dep, et)

		if tcd == audtps.EntryNone {
			tcd = audtps.EntryDepend
		}

		if !ok {
			ok = true
		}
	}

	// Validate that parsing was successful and produced valid results
	if !ok {
		s := nam
		if len(s) < 1 {
			s = "unnamed"
		}
		return nil, nil, fmt.Errorf("invalid params '%s' for entry '%s' in file '%s'", s, en, fs)
	}

	// Handle complex type structures that require recursive processing
	if tcd == audtps.EntryFunction || tcd == audtps.EntryStruct {
		stp = p.Type
		for {
			switch u := stp.(type) {
			case *ast.StarExpr:
				stp = u.X
				continue // continue the for loop
			case *ast.ArrayType:
				stp = u.Elt
				continue // continue the for loop
			case *ast.Ellipsis:
				stp = u.Elt
				continue // continue the for loop
			case *ast.ChanType:
				stp = u.Value
				continue // continue the for loop
			default:
				// nothing, call break following
			}

			break
		}

		// Process function types to extract parameter information
		if ft, kft := stp.(*ast.FuncType); kft {
			// Retrieving the anonymous function's input parameters
			if ft.Params != nil {
				for _, fp := range ft.Params.List {
					sPr, sDep, err := o.getParams(ctx, pkg, fs, afs, en, fp)
					if err == nil {
						sin = append(sin, sPr...)
						dep = append(dep, sDep...)
					}
				}
			}
			// Retrieving the output parameters of the anonymous function
			if ft.Results != nil {
				for _, fr := range ft.Results.List {
					sPr, sDep, err := o.getParams(ctx, pkg, fs, afs, en, fr)
					if err == nil {
						sou = append(sou, sPr...)
						dep = append(dep, sDep...)
					}
				}
			}
		} else if st, isStruct := stp.(*ast.StructType); isStruct {
			// Retrieval of anonymous structure fields (stored as Inputs)
			if st.Fields != nil {
				for _, sf := range st.Fields.List {
					sPr, sDep, err := o.getParams(ctx, pkg, fs, afs, en, sf)
					if err == nil {
						sin = append(sin, sPr...)
						dep = append(dep, sDep...)
					}
				}
			}
		}
	}

	// Determine the count of parameters (handling anonymous parameters)
	count := len(p.Names)
	if count == 0 {
		count = 1 // anonymous params
	}

	// Create parameter objects for each parameter in the field
	for i := 0; i < count; i++ {
		pr := audprm.New(tst, tcd)

		if len(nam) > 0 {
			pr.SetName(nam)
		}

		if len(sin) > 0 {
			pr.SetInput(sin...)
		}

		if len(sou) > 0 {
			pr.SetOutput(sou...)
		}

		prm = append(prm, pr)
	}

	return prm, dep, nil
}
