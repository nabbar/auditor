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
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audprm "github.com/nabbar/auditor/pkg/data/params"
	audtps "github.com/nabbar/auditor/pkg/data/types"
)

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
		return audtps.EntryPrimitive, "func", true

	case *ast.StructType:
		// Handle struct type expressions (e.g., struct{ field int })
		return audtps.EntryPrimitive, "struct", true

	case *ast.InterfaceType:
		// Handle interface type expressions (e.g., interface{ Method() })
		return audtps.EntryPrimitive, "interface", true

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
			o.u.InfoData("[getParams][File %s][Entry %s] Skip empty entry while collecting dependencies from type specifications", p, fs, en)
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

		e := fmt.Errorf("invalid params '%s'", s)
		o.u.ErrorStackData(fmt.Sprintf("[getParams][File %s][Entry %s] invalid params", fs, en), p, e)

		return nil, nil, e
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

// parseValueSpecType resolves the type of an *ast.ValueSpec node encountered
// during constant or variable declarations. It returns the audtps.CodeType classification,
// a human-readable string representation, and a boolean indicating whether resolution succeeded.
//
// The function follows a prioritized lookup strategy:
//  1. If vs.Type is non-nil, it uses the explicitly declared type on the ValueSpec.
//  2. If vs.Type is nil, it attempts to inherit the type from a preceding ValueSpec
//     within the same *ast.GenDecl block by calling getInheritedValueSpec. This covers
//     grouped const/var blocks where the type is only declared once at the top
//     (e.g., "const (a int = 1; b = 2)" where b inherits int from a).
//  3. If a resolved type expression (explicit or inherited) is found, it delegates
//     to parseExprCode to classify and stringify the type.
//  4. If the type cannot be resolved through any of the above methods, it returns
//     EntryNone, an empty string, and false.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during parsing.
//   - afs: The AST File structure providing access to the source tree for inherited
//     type resolution within the same GenDecl block.
//   - vs: The AST ValueSpec node representing the constant or variable declaration.
//
// Returns:
//   - audtps.CodeType: The code type classification derived from the resolved expression.
//   - string: A string representation of the type suitable for documentation.
//   - bool: True if a valid type could be resolved; false otherwise.
func (o *mdl) parseValueSpecType(ctx context.Context, afs *ast.File, vs *ast.ValueSpec) (audtps.CodeType, string, bool) {
	if vs == nil {
		return audtps.EntryNone, "", false
	}

	tpe := vs.Type
	if tpe == nil {
		tpe, _ = o.getInheritedValueSpec(afs, vs)
	}

	if tpe != nil {
		return o.parseExprCode(ctx, tpe)
	}

	return audtps.EntryNone, "", false
}

// getInheritedValueSpec walks the AST of a Go file to find the most recent type and/or
// value expressions declared before the given ValueSpec within the same *ast.GenDecl.
//
// In Go, grouped const and var blocks allow the type to be declared once and then
// inherited by subsequent declarations on the same line or following lines (e.g.,
// "const (a int = 1; b = 2)" where b inherits int from a). This function inspects
// the AST tree using ast.Inspect, iterating through all GenDecl nodes. For each
// GenDecl, it iterates through the ValueSpec entries in declaration order, recording
// the last non-nil Type and the most recent set of Values encountered before the
// target ValueSpec. When the target is found, it stops the walk and returns the
// recorded type and values.
//
// Parameters:
//   - afs: The AST File node providing the root for the traversal.
//   - targetVs: The target ValueSpec for which the inherited type and values are resolved.
//
// Returns:
//   - ast.Expr: The most recently declared type expression before the target ValueSpec.
//     Nil if no preceding type expression was found.
//   - []ast.Expr: The most recent set of value expressions from the preceding
//     ValueSpec. Nil if no preceding value expressions were found.
func (o *mdl) getInheritedValueSpec(afs *ast.File, targetVs *ast.ValueSpec) (ast.Expr, []ast.Expr) {
	if afs == nil || targetVs == nil {
		return nil, nil
	}

	var (
		fnd bool

		oldTpe ast.Expr
		oldVal []ast.Expr
	)

	ast.Inspect(afs, func(n ast.Node) bool {
		if fnd {
			return false
		}

		gd, ok := n.(*ast.GenDecl)

		if !ok {
			return true
		}

		for _, spec := range gd.Specs {
			vs, k := spec.(*ast.ValueSpec)
			if !k {
				continue
			}

			if vs.Type != nil {
				oldTpe = vs.Type
			}

			if len(vs.Values) > 0 {
				oldVal = vs.Values
			}

			if vs == targetVs {
				fnd = true
				return false
			}
		}
		return true
	})

	return oldTpe, oldVal
}

// parseExprTypeFromExpr deduces the underlying Go type from a value expression
// by examining the AST expression node structure. It returns the audtps.CodeType
// classification, a human-readable string representation, and a boolean indicating
// whether deduction succeeded.
//
// This function handles the following expression types:
//   - *ast.BasicLit: Basic literals are mapped by their token kind (INT to int,
//     FLOAT to float64, STRING to string, CHAR to rune, IMAG to complex128).
//   - *ast.Ident: Pre-declared identifiers are resolved (true/false to bool,
//     nil to any, iota to int). Other identifiers fall back to parseExprCode.
//   - *ast.BinaryExpr: Binary operators (e.g., iota + 1), recurse on the left operand.
//   - *ast.CompositeLit: Composite literals (e.g., []string{"a", "b"}) resolve
//     the underlying type from t.Type when present.
//   - *ast.UnaryExpr: Unary operators (e.g., -3.14, !true), recurse on the operand.
//   - *ast.FuncLit: Anonymous function literals resolve the function type from t.Type.
//   - *ast.ParenExpr: Parenthesized expressions, recurse on the inner expression.
//   - *ast.CallExpr: Native allocations like new(T) return *T; make(T, ...) returns T.
//     Other calls delegate to parseExprCode on the function expression.
//   - *ast.SelectorExpr, *ast.IndexExpr, *ast.IndexListExpr: Selector and generic
//     index expressions delegate to parseExprCode for type resolution.
//
// When an expression cannot be classified, the function returns EntryNone, an
// empty string, and false.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during parsing.
//   - expr: The AST Expression node whose underlying type should be deduced.
//
// Returns:
//   - audtps.CodeType: The code type classification of the deduced type.
//   - string: A string representation of the deduced type.
//   - bool: True if the type could be deduced; false otherwise.
func (o *mdl) parseExprTypeFromExpr(ctx context.Context, expr ast.Expr) (audtps.CodeType, string, bool) {
	if expr == nil {
		return audtps.EntryNone, "", false
	}

	switch t := expr.(type) {
	case *ast.BasicLit:
		// Basic literals (numbers, strings, characters, etc.)
		switch t.Kind {
		case token.INT:
			return audtps.EntryPrimitive, "int", true
		case token.FLOAT:
			return audtps.EntryPrimitive, "float64", true // Default for untyped floating-point numbers in Go
		case token.STRING:
			return audtps.EntryPrimitive, "string", true
		case token.CHAR:
			return audtps.EntryPrimitive, "rune", true
		case token.IMAG:
			return audtps.EntryPrimitive, "complex128", true
		}
	case *ast.Ident:
		// Pre-declared identifiers such as booleans (true/false) or nil
		if t.Name == "true" || t.Name == "false" {
			return audtps.EntryPrimitive, "bool", true
		}
		if t.Name == "nil" {
			return audtps.EntryPrimitive, "any", true
		}
		if t.Name == "iota" {
			return audtps.EntryPrimitive, "int", true
		}
		// If it is another identifier, we attempt to resolve it via parseExprCode
		return o.parseExprCode(ctx, t)
	case *ast.BinaryExpr:
		// Binary expressions (e.g., iota + 1, x + y)
		return o.parseExprTypeFromExpr(ctx, t.X)
	case *ast.CompositeLit:
		// Composite literals (e.g., []string{"a", "b"} or MyStruct{Field: 1})
		if t.Type != nil {
			return o.parseExprCode(ctx, t.Type)
		}
	case *ast.UnaryExpr:
		// Unary expressions (ex: -3.14, !true)
		return o.parseExprTypeFromExpr(ctx, t.X)
	case *ast.FuncLit:
		// Anonymous functions / closures (ex: var f = func() { ... })
		if t.Type != nil {
			return o.parseExprCode(ctx, t.Type)
		}
	case *ast.ParenExpr:
		// Expressions in parentheses (e.g., (3.14))
		return o.parseExprTypeFromExpr(ctx, t.X)
	case *ast.CallExpr:
		// Managing native allocations: new(T) and make(T, ...)
		if ident, ok := t.Fun.(*ast.Ident); ok {
			if ident.Name == "new" && len(t.Args) == 1 {
				// new(T) returns a pointer *T
				if tcd, tst, k := o.parseExprCode(ctx, t.Args[0]); k {
					return tcd, "*" + tst, true
				}
			} else if ident.Name == "make" && len(t.Args) > 0 {
				// make(T, ...) returns T
				return o.parseExprCode(ctx, t.Args[0])
			}
		}
		// For any other function call (e.g. regexp.MustCompile(...)), parse the function expression as code
		return o.parseExprCode(ctx, t.Fun)
	case *ast.SelectorExpr, *ast.IndexExpr, *ast.IndexListExpr:
		// Function calls, selectors, etc., from which the type can be extracted
		return o.parseExprCode(ctx, t)
	}

	return audtps.EntryNone, "", false
}

// isTypeParam determines whether the given identifier is a generic type parameter
// defined in the TypeParams clause of a parent TypeSpec or generic function.
//
// The function walks the AST using ast.Inspect and examines two categories of nodes:
//   - *ast.TypeSpec: iterates through all type parameter fields in t.TypeParams and
//     checks whether any name matches the identifier's name. TypeSpec covers generic
//     structs, interfaces, and type aliases.
//   - *ast.FuncDecl: iterates through all type parameter fields in t.Type.TypeParams
//     when non-nil, checking whether any name matches the identifier's name; this
//     covers generic function declarations.
//
// Parameters:
//   - afs: The AST File node providing the root for the traversal.
//   - idt: The identifier to test against the type parameter definitions.
//
// Returns:
//   - bool: True if the identifier corresponds to a generic type parameter; false
//     if it does not, or if either input is nil.
func (o *mdl) isTypeParam(afs *ast.File, idt *ast.Ident) bool {
	if afs == nil || idt == nil {
		return false
	}

	var found bool

	ast.Inspect(afs, func(n ast.Node) bool {
		if found {
			return false
		}

		switch t := n.(type) {
		case *ast.TypeSpec:
			if t.TypeParams != nil {
				for _, field := range t.TypeParams.List {
					for _, name := range field.Names {
						if name.Name == idt.Name {
							found = true
							return false
						}
					}
				}
			}
		case *ast.FuncDecl:
			if t.Type != nil && t.Type.TypeParams != nil {
				for _, field := range t.Type.TypeParams.List {
					for _, name := range field.Names {
						if name.Name == idt.Name {
							found = true
							return false
						}
					}
				}
			}
		}
		return true
	})

	return found
}
