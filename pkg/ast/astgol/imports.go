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
	"go/build"
	"strings"

	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	audtps "github.com/nabbar/auditor/pkg/data/types"
)

// getDepend resolves dependencies for imported packages in AST expressions.
//
// This function analyzes selector expressions (like fmt.Println) to determine
// the package dependency and create appropriate entry references. It handles
// both standard library imports and third-party package imports, including
// aliased imports. The getDepend function is crucial for identifying and tracking
// external dependencies in Go source code during AST analysis.
//
// Key features:
//   - Extracts the package name from selector expressions using resolveRootIdent
//   - Resolves import paths and manages aliases through comprehensive import path analysis
//   - Creates or retrieves package/module entries for dependencies with proper error handling
//   - Builds entry references for function calls in imported packages to track usage
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - fs: The file path containing the AST node, used for logging and error reporting.
//   - afs: The AST File node containing the imports and declarations, providing context
//     for import resolution and package identification.
//   - fun: The AST Selector expression representing a function call (e.g., fmt.Println),
//     specifying the exact dependency to resolve.
//
// Returns:
//   - audent.Entry: The resolved dependency entry, or nil if resolution fails due to
//     missing imports or invalid package references. This provides a mechanism for
//     tracking external dependencies and their usage in source code.
func (o *mdl) getDepend(ctx context.Context, fs string, afs *ast.File, fun *ast.SelectorExpr) audent.Entry {
	// extract name of package
	var (
		err error

		pkc string // real name of package from depend (Ex: "fmt" from fmt.Println)
		ipp string // store full path of the import
		als string // store alias to update package registered

		pkg audpkg.Package
		ent audent.Entry
	)

	// extract real package name / alias from dependency selector
	if idt := o.resolveRootIdent(ctx, fun.X); idt != nil {
		pkc = idt.Name
	} else {
		return nil
	}

	// search complete import path and manage alias
	for _, imp := range afs.Imports {
		if ctx.Err() != nil {
			return nil
		}

		// if import has alias (ex: import log "github.com/sirupsen/logrus")
		if imp.Name != nil && imp.Name.Name == pkc {
			ipp = strings.Trim(imp.Path.Value, `"`)
			als = pkc
			break
		}

		// import is standard (ex: import "net/http")
		if imp.Name == nil {
			pth := strings.Trim(imp.Path.Value, `"`)
			if o.guessPkgName(pth) == pkc {
				ipp = pth
				break
			}
		}
	}

	if len(ipp) < 1 {
		o.u.Info("[getDepend][File %s] no import found for dependency %s", fs, pkc)
		// ipp not found so, no a package
		return nil
	}

	// Trying to retrieve package / module for founded package
	pkg = o.getPackage(ctx, fs, ipp)

	if pkg == nil || pkg.IsEmpty() {
		o.u.Info("[getDepend][File %s] Skip nil dependency", fs)
		// cannot add package or module
		return nil
	}

	if len(als) > 0 {
		pkg.AddAlias(als)
		o.u.Info("[getDepend][File %s] Add/Update package %s for dependency", fs, pkg.GetFullPath())
		o.a.PkgAdd(pkg)
	}

	if ctx.Err() != nil {
		return nil
	}

	// If the receiver is a function call (e.g., GetCobra().NewCommand)
	if cx, l := fun.X.(*ast.CallExpr); l {
		// We look at what is being called (e.g., audpkg.GetCobra)
		if in, k := cx.Fun.(*ast.SelectorExpr); k {
			// Resolution is delegated to the internal call (GetCobra)
			// instead of the final member (NewCommand)
			return o.getDepend(ctx, fs, afs, in)
		}
	}

	// add or retrieve entry
	// func name = fun.Sel.Name, Ex: "Println"
	o.u.Info("[getDepend][File %s] Find entry %s for dependencies", fs, fun.Sel.Name)

	if ent, err = o.findEntry(pkg.GetID(), fun.Sel.Name); err != nil {
		o.u.ErrorStackData(fmt.Sprintf("[getDepend][File %s] Cannot find entry %s for dependency", fs, fun.Sel.Name), fun, err)
		return nil
	}

	if ent == nil || ent.IsEmpty() {
		o.u.ErrorStackData(fmt.Sprintf("[getDepend][File %s] Cannot find entry %s for dependency", fs, fun.Sel.Name), fun, errors.New("empty entry"))
		return nil
	}

	return ent
}

// getPackage looks up a package by its full import path from the managed collection.
//
// This function first validates that the import path is non-empty and then uses Go's
// build.Default.Import to check whether the path refers to the Go standard library
// (Goroot), which is skipped. It then searches the managed package collection via
// o.a.PkgSearch(pth) to find an existing package entry. If the package is not found
// or is empty, nil is returned. The getPackage function does not create packages;
// it only retrieves an already-registered package from the internal registry.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - fs: The file path containing the AST node, used for logging and error reporting.
//   - pth: The full import path for the package (e.g., "github.com/sirupsen/logrus"),
//     which uniquely identifies the package within the Go ecosystem.
//
// Returns:
//   - audpkg.Package: The package instance if found in the collection, or nil if the
//     package does not exist in the registry or if the import path refers to a
//     Go standard library package.
func (o *mdl) getPackage(ctx context.Context, fs string, pth string) audpkg.Package {
	if len(pth) < 1 {
		return nil
	}

	// Attempt to resolve the import path using Go's build package
	if p, e := build.Default.Import(pth, "", build.FindOnly); e != nil {
		o.u.ErrorStack(fmt.Sprintf("[getPackage][File %s] identify root package trigger error for %s", fs, pth), e)
	} else if p.Goroot {
		o.u.Info("[getPackage][File %s] Skip GoRoot dependencies: %s", fs, pth)
		return nil
	}

	var pk = o.a.PkgSearch(pth)

	if ctx.Err() != nil {
		return nil
	}

	if pk == nil || pk.IsEmpty() {
		o.u.Info("[getPackage][File %s] Skip find package for dependency : %s", fs, pth)
		return nil
	}

	o.u.Info("[getPackage][File %s] Find package '%s' (id %d) for dependency '%s'", fs, pk.GetFullPath(), pk.GetID(), pth)
	return pk
}

// resolveRootIdent recursively traverses AST expressions to extract the base identifier.
//
// This function handles complex selector chains like obj.subObj.Method() or obj.Get().Method()
// by recursively following the expression tree until it reaches the base identifier. The
// resolveRootIdent function is crucial for resolving the root package name in complex
// selector expressions, enabling accurate dependency tracking and package identification.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - expr: The AST expression node to traverse, which may be a simple identifier or
//     complex selector chain that needs to be resolved to its base component.
//
// Returns:
//   - *ast.Ident: The base identifier if found, otherwise nil. This provides the root
//     package name for dependency resolution in selector expressions.
func (o *mdl) resolveRootIdent(ctx context.Context, expr ast.Expr) *ast.Ident {
	switch t := expr.(type) {
	case *ast.Ident:
		return t
	case *ast.SelectorExpr:
		return o.resolveRootIdent(ctx, t.X)
	case *ast.CallExpr:
		return o.resolveRootIdent(ctx, t.Fun)
	case *ast.IndexExpr:
		return o.resolveRootIdent(ctx, t.X)
	case *ast.TypeAssertExpr:
		return o.resolveRootIdent(ctx, t.X)
	case *ast.StarExpr:
		return o.resolveRootIdent(ctx, t.X)
	case *ast.ParenExpr:
		return o.resolveRootIdent(ctx, t.X)
	case *ast.IndexListExpr:
		return o.resolveRootIdent(ctx, t.X)
	default:
		return nil
	}
}

// guessPkgName infers the Go package name from a full import path by applying a series
// of heuristics that account for Go versioning and naming conventions.
//
// The function splits the import path by "/" and starts from the last segment. It then
// applies the following corrections in order:
//
//   - Versioned module directories: when the last segment matches the pattern "/v<digit>"
//     (e.g., "github.com/go-redis/redis/v8"), the segment is replaced with the second-to-last
//     segment so that "v8" is skipped and "redis" is returned.
//   - gopkg.in versioning: when the segment contains a ".v<digits>" suffix
//     (e.g., "gopkg.in/yaml.v2", "gopkg.in/check.v1"), the suffix is stripped so that
//     "yaml" is returned.
//   - "go-" prefix: common in Go library module names
//     (e.g., "go-sqlite3", "go-redis"), stripped so that "sqlite3" or "redis" is returned.
//   - "-go" suffix: common in Go library module names
//     (e.g., "amqp-go", "mongo-go"), stripped so that "amqp" or "mongo" is returned.
//
// Parameters:
//   - pth: The full import path for the package (e.g., "github.com/go-redis/redis/v8",
//     "gopkg.in/yaml.v2").
//
// Returns:
//   - string: The inferred package name after all versioning and convention heuristics
//     have been applied. An empty string is returned if the path is empty or contains
//     no valid segments.
func (o *mdl) guessPkgName(pth string) string {
	prt := strings.Split(pth, "/")
	if len(prt) == 0 || (len(prt) == 1 && prt[0] == "") {
		return ""
	}

	chk := prt[len(prt)-1]

	// 1. Handle versioned directories (e.g., github.com/go-redis/redis/v8)
	if len(chk) >= 2 && chk[0] == 'v' && chk[1] >= '0' && chk[1] <= '9' && len(prt) > 1 {
		chk = prt[len(prt)-2]
	}

	// 2. Handle versioned package names (e.g., gopkg.in/yaml.v2)
	if idx := strings.LastIndex(chk, ".v"); idx > 0 {
		isVersion := true
		// Verify that everything after ".v" is numeric
		for i := idx + 2; i < len(chk); i++ {
			if chk[i] < '0' || chk[i] > '9' {
				isVersion = false
				break
			}
		}
		if isVersion {
			chk = chk[:idx] // keep only "yaml"
		}
	}

	// 3. Clean up common GitHub prefixes and suffixes ("go-" prefix, "-go" suffix)
	chk = strings.TrimPrefix(chk, "go-")
	chk = strings.TrimSuffix(chk, "-go")

	return chk
}

// getLocDepend resolves local dependencies for function calls within the same package.
//
// This function analyzes local function calls (e.g., MyFunction()) to determine their
// type classification and create appropriate entry references. It handles both built-in
// functions and user-defined functions, providing accurate dependency tracking for
// local code elements. The getLocDepend function is essential for identifying and
// tracking internal dependencies within Go source code.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - fs: The file path containing the AST node, used for logging and error reporting.
//   - pki: The package identifier for context association with local dependencies.
//   - exp: The AST Call expression representing a function call (e.g., MyFunction()),
//     specifying the exact dependency to resolve.
//
// Returns:
//   - audent.Entry: The resolved local dependency entry, or nil if resolution fails
//     due to built-in functions or invalid identifiers. This provides a mechanism for
//     tracking internal dependencies and their usage in source code.
func (o *mdl) getLocDepend(_ context.Context, fs string, pki audids.ID, exp *ast.CallExpr) audent.Entry {
	var (
		err error
		ent audent.Entry

		idt, ok = exp.Fun.(*ast.Ident)
	)

	if !ok || idt == nil {
		return nil
	}

	// Skip built-in functions and primitives as they don't require dependency tracking
	if primitives[idt.Name] || builtins[idt.Name] {
		return nil
	}

	// Ignore whether the identifier is a local variable or a parameter (e.g., closure 'f')
	if idt.Obj != nil && idt.Obj.Kind == ast.Var {
		return nil
	}

	o.u.Info("[getLocDepend][File %s] Find entry %s for local dependency", fs, idt.Name)
	if ent, err = o.findEntry(pki, idt.Name); err != nil {
		o.u.ErrorStackData(fmt.Sprintf("[getLocDepend][File %s] Cannot find entry %s for local dependency", fs, idt.Name), exp, err)
		return nil
	}

	if ent == nil || ent.IsEmpty() {
		o.u.ErrorStackData(fmt.Sprintf("[getLocDepend][File %s] Cannot find entry %s for local dependency", fs, idt.Name), exp, errors.New("empty entry"))
		return nil
	}

	return ent
}

// resolveBaseTypeExpr removes type modifiers (pointers, slices, channels, parens)
// to extract the core type expression (SelectorExpr, Ident, MapType, etc.).
//
// This function recursively strips away type modifiers such as pointers (*), slices ([]),
// channels (<-), and parentheses () to isolate the fundamental type expression. It is
// essential for accurately analyzing complex type structures in Go code by identifying
// the base type that underlies composite types.
//
// Parameters:
//   - expr: The AST expression representing a type, which may include modifiers like pointers or slices.
//
// Returns:
//   - ast.Expr: The core type expression after removing all modifiers. This provides
//     access to the fundamental type for further analysis and dependency resolution.
func (o *mdl) resolveBaseTypeExpr(expr ast.Expr) ast.Expr {
	if expr == nil {
		return nil
	}

	for {
		switch t := expr.(type) {
		case *ast.StarExpr:
			expr = t.X
		case *ast.ArrayType:
			expr = t.Elt
		case *ast.Ellipsis:
			expr = t.Elt
		case *ast.ChanType:
			expr = t.Value
		case *ast.ParenExpr:
			expr = t.X
		default:
			return expr
		}
	}
}

// getLocTypeDepend resolves a local type identifier, inspecting its declaration
// to assign a precise CodeType (struct, interface, etc.) instead of EntryNone.
//
// This function analyzes local type declarations to determine their exact classification
// and create appropriate entry references. It provides accurate type information by
// examining the AST TypeSpec declaration directly, ensuring that types are properly
// categorized rather than defaulting to generic classifications. The getLocTypeDepend
// function is crucial for precise type analysis in Go source code.
//
// Key features:
//   - Examines AST TypeSpec declarations directly for accurate type classification
//   - Handles both built-in types and user-defined types appropriately
//   - Provides precise CodeType assignment instead of generic EntryNone classification
//   - Creates or updates entry objects with appropriate metadata for storage
//   - Implements proper error handling for invalid identifiers or missing declarations
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - fs: The file path containing the AST node, used for logging and error reporting.
//   - pki: The package identifier for context association with local type dependencies.
//   - idt: The AST Identifier representing a local type (e.g., MyStruct, MyInterface),
//     specifying the exact type to resolve.
//
// Returns:
//   - audent.Entry: The resolved local type dependency entry, or nil if resolution fails
//     due to built-in types or invalid identifiers. This provides accurate type information
//     for dependency tracking and analysis.
func (o *mdl) getLocTypeDepend(ctx context.Context, fs string, pki audids.ID, idt *ast.Ident) audent.Entry {
	// Skip built-in types and primitives as they don't require dependency tracking
	if idt == nil || primitives[idt.Name] || builtins[idt.Name] {
		return nil
	}

	// Skip generic type parameters (e.g., [M any] on mdl[M] or associated methods)
	if o.isTypeParam(o.s[fs].a, idt) {
		o.u.InfoData("[getLocTypeDepend][File %s] Skip generic type entry %s for local type dependency", idt, fs, idt.Name)
		return nil
	}

	var (
		err error
		ent audent.Entry
		tcd audtps.CodeType // default is EntryNone
		tps string
	)

	// 1. Try to identify exactly local type from AST TypeSpec declaration
	if idt.Obj != nil && idt.Obj.Kind == ast.Typ {
		if ts, ok := idt.Obj.Decl.(*ast.TypeSpec); ok && ts.Type != nil {
			tcd, tps, _ = o.parseExprCode(ctx, ts.Type)
		}
	}

	// 2. If type still not defined, try a direct expression analysis
	if tcd == audtps.EntryNone {
		if c, s, ok := o.parseExprCode(ctx, idt); ok && c != audtps.EntryNone {
			tcd = c
			tps = s
		}
	}

	o.u.Info("[getLocTypeDepend][File %s] Find entry %s for local type dependency (code type: %s - %s)", fs, idt.Name, tcd.String(), tps)

	if ent, err = o.findEntry(pki, idt.Name); err != nil {
		o.u.ErrorStackData(fmt.Sprintf("[getLocTypeDepend][File %s] Cannot find entry %s for local type dependency", fs, idt.Name), idt, err)
		return nil
	}

	if ent == nil || ent.IsEmpty() {
		o.u.ErrorStackData(fmt.Sprintf("[getLocTypeDepend][File %s] Cannot find entry %s for local type dependency", fs, idt.Name), idt, errors.New("empty entry"))
		return nil
	}

	return ent
}

// getTypeDepend resolves and returns a slice of dependency entries associated with
// a parameter or return type expression.
//
// This function analyzes AST expressions to determine their dependency relationships
// and creates appropriate entry references for tracking. It handles various type expressions
// including external dependencies (SelectorExpr), local dependencies (Ident), and complex
// composite types like maps, slices, and generic types. The getTypeDepend function is essential
// for comprehensive dependency analysis in Go source code.
//
// Key features:
//   - Handles external dependencies through SelectorExpr resolution
//   - Processes local dependencies through Ident resolution with getLocTypeDepend
//   - Manages complex composite types including maps, slices, and generic types
//   - Recursively analyzes nested type expressions for comprehensive dependency tracking
//   - Returns nil for primitive, builtin, or unresolvable types to avoid unnecessary processing
//   - For composite types (MapType, IndexExpr, IndexListExpr) it collects entries from
//     multiple sub-expressions and returns them aggregately
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - fs: The file path containing the AST node, used for logging and error reporting.
//   - afs: The AST File node representing the complete file structure, providing context for position calculations.
//   - pki: The package identifier for context association with type dependencies.
//   - expr: The AST expression representing a type specification (e.g., http.Request, []string).
//
// Returns:
//   - []audent.Entry: A slice of resolved dependency entries, or nil if the type is primitive,
//     builtin, or unresolvable. This provides comprehensive dependency tracking for complex type expressions.
func (o *mdl) getTypeDepend(ctx context.Context, fs string, afs *ast.File, pki audids.ID, expr ast.Expr) []audent.Entry {
	if expr == nil {
		return nil
	}

	baseExpr := o.resolveBaseTypeExpr(expr)
	if baseExpr == nil {
		return nil
	}

	switch t := baseExpr.(type) {
	case *ast.SelectorExpr:
		// External dependency (ex: http.Request, context.Context)
		return []audent.Entry{o.getDepend(ctx, fs, afs, t)}

	case *ast.Ident:
		// Local dependency (ex: MyStruct, MyInterface)
		return []audent.Entry{o.getLocTypeDepend(ctx, fs, pki, t)}

	case *ast.MapType:
		var res = make([]audent.Entry, 0)
		// For map, try to identify and find value type and after key type if value is not a dependency
		if ent := o.getTypeDepend(ctx, fs, afs, pki, t.Key); ent != nil {
			res = append(res, ent...)
		} else {
			o.u.InfoData(fmt.Sprintf("[getTypeDepend][File %s] Skip nil entry map key for type dependency", fs), expr, fmt.Errorf("empty entry"))
		}
		if ent := o.getTypeDepend(ctx, fs, afs, pki, t.Value); ent != nil {
			res = append(res, ent...)
		} else {
			o.u.InfoData(fmt.Sprintf("[getTypeDepend][File %s] Skip nil entry map value for type dependency", fs), expr, fmt.Errorf("empty entry"))
		}
		return res

	case *ast.IndexExpr:
		// Extract dependencies from base type AND generic argument
		var res = make([]audent.Entry, 0)
		if ent := o.getTypeDepend(ctx, fs, afs, pki, t.X); ent != nil {
			res = append(res, ent...)
		} else {
			o.u.InfoData(fmt.Sprintf("[getTypeDepend][File %s] Skip nil entry expression for type dependency", fs), expr, fmt.Errorf("empty entry"))
		}
		if ent := o.getTypeDepend(ctx, fs, afs, pki, t.Index); ent != nil {
			res = append(res, ent...)
		} else {
			o.u.InfoData(fmt.Sprintf("[getTypeDepend][File %s] Skip nil entry index for type dependency", fs), expr, fmt.Errorf("empty entry"))
		}
		return res

	case *ast.IndexListExpr:
		// Extract dependencies from base type AND all generic arguments
		var res = make([]audent.Entry, 0)
		if ent := o.getTypeDepend(ctx, fs, afs, pki, t.X); ent != nil {
			res = append(res, ent...)
		} else {
			o.u.InfoData(fmt.Sprintf("[getTypeDepend][File %s] Skip nil entry list expression for type dependency", fs), expr, fmt.Errorf("empty entry"))
		}
		for _, idx := range t.Indices {
			if ent := o.getTypeDepend(ctx, fs, afs, pki, idx); ent != nil {
				res = append(res, ent...)
			} else {
				o.u.InfoData(fmt.Sprintf("[getTypeDepend][File %s] Skip nil entry list indices for type dependency", fs), expr, fmt.Errorf("empty entry"))
			}
		}
		return res

	default:
		return nil
	}
}
