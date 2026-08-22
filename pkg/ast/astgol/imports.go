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
	"go/ast"
	"go/build"
	"strings"

	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
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
		pkc string // real name of package from depend (Ex: "fmt" from fmt.Println)
		ipp string // store full path of the import
		als string // store alias to update package registered
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
		o.u.Info("[File %s] no import found for dependency %s", fs, pkc)
		// ipp not found so, no a package
		return nil
	}

	// Trying to retrieve package / module for founded package
	pk := o.getPackage(ctx, fs, ipp)

	if pk == nil || pk.IsEmpty() {
		o.u.Info("[File %s] Skip nil dependency", fs)
		// cannot add package or module
		return nil
	}

	if len(als) > 0 {
		pk.AddAlias(als)
		o.u.Info("[File %s] Add/Update package %s for dependency", fs, pk.GetFullPath())
		o.a.PkgAdd(pk)
	}

	if ctx.Err() != nil {
		return nil
	}

	// add or retrieve entry
	// func name = fun.Sel.Name, Ex: "Println"
	o.u.Info("[File %s] Preparing entry %s for dependencies", fs, fun.Sel.Name)
	return o.a.EntNewOrMerge(pk.GetID(), fun.Sel.Name, "", audtps.EntryNone, Lang, nil)
}

// getPackage retrieves or creates a package entry for the given import path.
//
// This function attempts to find an existing package in the collection based on
// the full import path. If no package is found, it creates a new one using addPackage.
// The getPackage function serves as the core mechanism for managing package references
// during AST analysis, ensuring that all imported packages are properly tracked and
// stored in the auditor's data structures.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - fs: The file path containing the AST node, used for logging and error reporting.
//   - pth: The full import path for the package (e.g., "github.com/sirupsen/logrus"),
//     which uniquely identifies the package within the Go ecosystem.
//
// Returns:
//   - audpkg.Package: The package instance, or nil if creation fails. This provides
//     a consistent interface for accessing package information regardless of
//     whether it was newly created or retrieved from existing data.
func (o *mdl) getPackage(ctx context.Context, fs string, pth string) audpkg.Package {
	if len(pth) < 1 {
		return nil
	}

	// Attempt to resolve the import path using Go's build package
	if p, e := build.Default.Import(pth, "", build.FindOnly); e != nil {
		o.u.ErrorStack("identify root package trigger error for "+pth, e)
	} else if p.Goroot {
		o.u.Info("[File %s] Skip GoRoot dependencies: %s", fs, pth)
		return nil
	}

	var pk = o.a.PkgSearch(pth)

	if ctx.Err() != nil {
		return nil
	}

	if pk == nil || pk.IsEmpty() {
		o.u.Info("[File %s] Add new package for dependency : %s", fs, pth)
		return o.addPackage(ctx, pth)
	}

	return pk
}

// addPackage creates a new package entry for the given import path.
//
// This function handles the creation of packages and their associated modules,
// including cases where the module needs to be created as a vendor module. It manages
// both direct package lookups and partial path matching for dependencies. The addPackage
// function is essential for maintaining complete dependency tracking in Go source code
// analysis, ensuring that all imported packages are properly registered and accessible.
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pth: The full import path for the package (e.g., "github.com/sirupsen/logrus"),
//     which specifies the location of the package within the Go module system.
//
// Returns:
//   - audpkg.Package: The newly created package instance, or nil if creation fails due to
//     invalid paths or resource constraints. This provides a mechanism for creating and
//     registering new packages in the auditor's collection system.
func (o *mdl) addPackage(ctx context.Context, pth string) audpkg.Package {
	var (
		pk = o.a.PkgSearchLike(pth)
		md audmod.Module
	)

	if pk != nil && !pk.IsEmpty() {
		return pk
	}

	if ctx.Err() != nil {
		return nil
	}

	// Create a new module entry for the import path, treating it as a vendor module
	md = o.a.ModNewOrMerge(pth, "") // add import as vendor module
	if md == nil || md.IsEmpty() {
		return nil // cannot add module, so cannot add package
	}

	// Create or merge a package entry associated with the module
	if p := o.a.PkgNewOrMerge(md.GetID(), pth); p == nil || p.IsEmpty() {
		return nil // cannot add package
	} else {
		return p
	}
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

// guessPkgName attempts to infer the package name from its import path.
//
// This function handles various naming conventions including versioned directories
// (e.g., /v2), gopkg.in extensions (e.g., yaml.v2), and common GitHub prefixes/suffixes
// (e.g., go-sqlite3, amqp-go). The guessPkgName function is essential for accurately
// determining package names when dealing with complex import paths that may contain
// version information or special naming conventions.
//
// Parameters:
//   - pth: The full import path for the package, which may include versioning information
//     or special naming conventions that need to be parsed to extract the actual package name.
//
// Returns:
//   - string: The inferred package name after processing version information and special
//     naming conventions. This provides a clean, standardized package name for consistent
//     tracking and identification in dependency analysis.
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
	if strings.HasPrefix(chk, "go-") {
		chk = strings.TrimPrefix(chk, "go-")
	}
	if strings.HasSuffix(chk, "-go") {
		chk = strings.TrimSuffix(chk, "-go")
	}

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
func (o *mdl) getLocDepend(ctx context.Context, fs string, pki audids.ID, exp *ast.CallExpr) audent.Entry {
	var (
		idt, ok = exp.Fun.(*ast.Ident)
		tcd     audtps.CodeType
		tps     string
	)

	if !ok || idt == nil {
		return nil
	}

	// Skip built-in functions and primitives as they don't require dependency tracking
	if primitives[idt.Name] || builtins[idt.Name] {
		return nil
	}

	if idt.Obj != nil {
		switch idt.Obj.Kind {
		case ast.Fun:
			tcd = audtps.EntryFunction
		case ast.Var, ast.Typ, ast.Con:
			tcd, tps, ok = o.parseExprCode(ctx, exp)
		default:
		}

		if tcd == audtps.EntryNone {
			return nil
		}
	}

	o.u.Info("[File %s] Add or update entry %s for local dependency", fs, idt.Name)
	return o.a.EntNewOrMerge(pki, idt.Name, tps, tcd, Lang, nil)
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

	var (
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

	o.u.Info("[File %s] Add or update entry %s for local type dependency (code type: %s - %s)", fs, idt.Name, tcd.String(), tps)
	return o.a.EntNewOrMerge(pki, idt.Name, tps, tcd, Lang, nil)
}

// getTypeDepend resolves and returns a single dependency entry associated with
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
		}
		if ent := o.getTypeDepend(ctx, fs, afs, pki, t.Value); ent != nil {
			res = append(res, ent...)
		}
		return res

	case *ast.IndexExpr:
		// Extract dependencies from base type AND generic argument
		var res = make([]audent.Entry, 0)
		if ent := o.getTypeDepend(ctx, fs, afs, pki, t.X); ent != nil {
			res = append(res, ent...)
		}
		if ent := o.getTypeDepend(ctx, fs, afs, pki, t.Index); ent != nil {
			res = append(res, ent...)
		}
		return res

	case *ast.IndexListExpr:
		// Extract dependencies from base type AND all generic arguments
		var res = make([]audent.Entry, 0)
		if ent := o.getTypeDepend(ctx, fs, afs, pki, t.X); ent != nil {
			res = append(res, ent...)
		}
		for _, idx := range t.Indices {
			if ent := o.getTypeDepend(ctx, fs, afs, pki, idx); ent != nil {
				res = append(res, ent...)
			}
		}
		return res

	default:
		return nil
	}
}
