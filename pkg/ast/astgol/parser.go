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
	audtps "github.com/nabbar/auditor/pkg/data/types"
)

// primitives is the lookup table for Go primitive type names, used during AST parsing to
// classify type specifications as built-in primitives rather than user-defined types.
//
// The map contains all standard language types: boolean, string, integer variants (int, int8,
// int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr), floating-point types
// (float32, float64), complex types (complex64, complex128), numeric aliases (rune, byte),
// and structural special types (any, error, interface, func). A type identifier that appears
// in this map is considered a primitive and will not be treated as an importable or
// resolvable symbol.
var primitives = map[string]bool{
	"any":        true,
	"bool":       true,
	"byte":       true,
	"complex64":  true,
	"complex128": true,
	"error":      true,
	"func":       true,
	"float32":    true,
	"float64":    true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"interface":  true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"uintptr":    true,
	"rune":       true,
	"string":     true,
}

// builtins is the lookup table for Go built-in function names, used during AST parsing
// to distinguish language-provided operations from resolvable symbol references.
//
// The map contains all standard library-provided functions: append, cap, clear, close,
// complex, copy, delete, imag, iota, len, make, max, min, new, panic, print, println,
// real, and recover. A call expression whose identifier matches an entry in this map is
// treated as a language intrinsic and will not be treated as an importable or resolvable
// symbol.
var builtins = map[string]bool{
	"append":  true,
	"cap":     true,
	"clear":   true,
	"close":   true,
	"complex": true,
	"copy":    true,
	"delete":  true,
	"imag":    true,
	"iota":    true,
	"len":     true,
	"make":    true,
	"max":     true,
	"min":     true,
	"new":     true,
	"panic":   true,
	"print":   true,
	"println": true,
	"real":    true,
	"recover": true,
}

func (o *mdl) processNode(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, node ast.Node, isUpdate bool) {
	// If the file is part of the vendors, everything that is not exported (starting with an uppercase letter) is ignored.
	if strings.Contains(fs, "/vendor/") {
		if !isExportedNode(node) {
			return
		}
	}

	var tcd audtps.CodeType

	switch t := node.(type) {
	case *ast.FuncDecl:
		tcd = audtps.EntryFunction

	case *ast.TypeSpec:
		if t.Type == nil {
			return
		}

		switch t.Type.(type) {
		case *ast.StructType:
			tcd = audtps.EntryStruct
		case *ast.InterfaceType:
			tcd = audtps.EntryInterface
		default:
			// Using parseExprCode instead of getTypeName to identify the type
			tpc, _, iok := o.parseExprCode(ctx, t.Type)
			if !iok {
				o.u.ErrorStack("[processNode]trigger error on identify type", fmt.Errorf("[File %s] cannot identify type %s", fs, t.Type))
				return
			}
			tcd = tpc
		}

	case *ast.GenDecl:
		switch t.Tok {
		case token.CONST:
			tcd = audtps.EntryConstant
		case token.VAR:
			tcd = audtps.EntryVariable
		case token.TYPE:
			for _, spec := range t.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					o.processNode(ctx, pkg, fs, afs, ts, isUpdate)
				}
			}
			return
		default:
			return
		}

		for _, spec := range t.Specs {
			if vs, ok := spec.(*ast.ValueSpec); ok {
				if fnMap, exists := o.h[tcd]; exists {
					if fn, k := fnMap[isUpdate]; k && fn != nil {
						o.u.Info("calling processNode function for type %s", tcd)
						fn(ctx, pkg, fs, afs, vs)
					}
				}
			}
		}

		return

	default:
		return
	}

	if fnMap, exists := o.h[tcd]; exists {
		if fn, ok := fnMap[isUpdate]; ok && fn != nil {
			o.u.Info("calling processNode function for type %s", tcd)
			fn(ctx, pkg, fs, afs, node)
		}
	}
}

func (o *mdl) findEntry(pkg audids.ID, name string) (audent.Entry, error) {
	var (
		eid audids.ID
		ent audent.Entry

		pdt = o.a.PkgGet(pkg)
	)

	// retrieve package
	if pdt == nil || pdt.IsEmpty() {
		return nil, fmt.Errorf("cannot find package for id: %d", pkg)
	}

	// retrieve entry
	eid = o.a.SearchID(pdt.GetFullPath(), name)
	if eid < 1 {
		return nil, fmt.Errorf("cannot find entry '%s' for package: %s (id: %d)", name, pdt.GetFullPath(), pkg)
	}

	ent = o.a.EntGet(eid)
	if ent == nil || ent.IsEmpty() {
		return nil, fmt.Errorf("cannot find entry '%s' (id: %d) for package: %s (id: %d)", name, eid, pdt.GetFullPath(), pkg)
	}

	return ent, nil
}

func (o *mdl) parseFile(ctx context.Context, pkg audids.ID, fs string, afs *ast.File) error {
	if len(afs.Decls) < 1 {
		return nil
	}

	for _, adc := range afs.Decls {
		switch i := adc.(type) {
		case *ast.FuncDecl:
			o.processNode(ctx, pkg, fs, afs, i, false)
		case *ast.GenDecl:
			o.processNode(ctx, pkg, fs, afs, i, false)
		}
	}

	return nil
}

// isExportedNode checks whether the given AST node represents an exported Go identifier,
// defined as an identifier whose first character is an uppercase Unicode letter. The
// function examines the node by type and delegates to ast.IsExported for the name
// extraction, handling *ast.FuncDecl, *ast.TypeSpec, *ast.ValueSpec, *ast.GenDecl, and
// returning true for any other node not covered by the switch.
func isExportedNode(node ast.Node) bool {
	if node == nil {
		return false
	}

	switch t := node.(type) {
	case *ast.FuncDecl:
		if t.Name != nil {
			return ast.IsExported(t.Name.Name)
		}
	case *ast.TypeSpec:
		if t.Name != nil {
			return ast.IsExported(t.Name.Name)
		}
	case *ast.ValueSpec:
		// For blocks of constants or variables, if one of the identifiers is exported, the block is relevant.
		for _, name := range t.Names {
			if name != nil && ast.IsExported(name.Name) {
				return true
			}
		}
		return false
	case *ast.GenDecl:
		// For a block of general declarations (const, var, type), its internal specifications are inspected.
		for _, spec := range t.Specs {
			if isExportedNode(spec) {
				return true
			}
		}
		return false
	}

	// By default, if the node does not fall into these categories, it is permitted.
	return true
}
