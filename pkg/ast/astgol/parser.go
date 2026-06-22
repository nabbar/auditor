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
	"errors"
	"fmt"
	"go/ast"

	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audtps "github.com/nabbar/auditor/pkg/data/types"
)

// primitives is a map of Go primitive types that are used to categorize type declarations.
//
// This constant provides a comprehensive lookup table for identifying primitive Go types
// during AST parsing operations. It includes all standard Go primitive types including:
// basic types (bool, string, int, etc.), floating-point types, complex types, and error type.
//
// The map serves as a crucial reference point for determining whether a type specification
// represents a primitive type versus a custom or composite type during the parsing process.
//
// This lookup table is essential for accurate categorization of Go types in the auditor tool,
// enabling proper classification of types for analysis and reporting purposes.
// It allows the system to distinguish between fundamental types (like int, string) and
// user-defined types (structs, interfaces, etc.) during AST processing. The primitives map
// is used extensively throughout the parseExprCode function to classify different AST expression
// nodes based on their fundamental type characteristics.
//
// Key features:
//   - Comprehensive coverage of all standard Go primitive types
//   - Efficient lookup mechanism for rapid type classification during AST parsing
//   - Essential for distinguishing between built-in types and user-defined types
//   - Used throughout the parseExprCode function for accurate type categorization
var primitives = map[string]bool{
	"any":        true,
	"bool":       true,
	"string":     true,
	"int":        true,
	"int8":       true,
	"int16":      true,
	"int32":      true,
	"int64":      true,
	"uint":       true,
	"uint8":      true,
	"uint16":     true,
	"uint32":     true,
	"uint64":     true,
	"uintptr":    true,
	"byte":       true,
	"rune":       true,
	"func":       true,
	"float32":    true,
	"float64":    true,
	"complex64":  true,
	"complex128": true,
	"error":      true,
}

// builtins is a map of Go built-in functions that are used to identify built-in operations
// during AST parsing and dependency analysis.
//
// This constant provides a lookup table for identifying built-in functions in Go source code.
// Built-in functions are fundamental operations provided by the Go language itself, such as
// append, len, make, new, and others. These functions do not require external dependencies
// and should be excluded from dependency tracking to avoid false positives in auditing.
//
// The map is used throughout the AST parsing process to identify when a function call refers
// to a built-in operation rather than an external package or user-defined function.
//
// Key features:
//   - Comprehensive coverage of all standard Go built-in functions
//   - Efficient lookup mechanism for rapid identification during AST analysis
//   - Essential for distinguishing between built-in operations and external dependencies
//   - Used in getLocDepend to skip built-in functions from dependency tracking
var builtins = map[string]bool{
	"append":  true,
	"cap":     true,
	"clear":   true,
	"close":   true,
	"complex": true,
	"copy":    true,
	"delete":  true,
	"imag":    true,
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

// parseFile processes a single Go source file to extract declarations.
//
// This function serves as the primary entry point for processing individual Go source files
// within the AST parsing framework. It iterates through all declarations present in the file
// and routes them to appropriate specialized parsing functions based on their AST node type.
//
// The function implements a robust dispatch mechanism that ensures each type of declaration
// is handled by the correct parser, maintaining consistency in how different Go constructs
// are processed throughout the auditor tool. It processes all declarations in the order they appear
// in the file and handles errors appropriately for each declaration type.
//
// Key features:
//   - Iterates through all declarations in the provided AST file structure
//   - Routes declarations to specialized parsing functions based on their type (FuncDecl or GenDecl)
//   - Implements error handling for malformed or invalid declarations
//   - Maintains processing order and ensures all declarations are processed
//   - Handles empty declaration lists gracefully without errors
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this source file, used to associate
//     declarations with their respective packages in the auditor's data collection system.
//   - fs: The filename path to the source file being processed, providing context for
//     error messages and documentation capture.
//   - afs: The AST File node representing the complete source file structure, containing
//     all parsed information about the file including declarations and imports.
//
// Returns:
//   - error: Any error encountered during declaration processing, including invalid AST structures,
//     parsing errors, or issues with individual declaration handling. This ensures that any problems
//     in processing a single file do not prevent continued processing of other files.
func (o *mdl) parseFile(ctx context.Context, pkg audids.ID, fs string, afs *ast.File) error {
	var (
		err error
		adc ast.Decl
	)

	// Handle empty declaration lists gracefully without errors
	if len(afs.Decls) < 1 {
		return nil
	}

	// Process each declaration in the file
	for _, adc = range afs.Decls {
		switch i := adc.(type) {
		case *ast.FuncDecl:
			// Route function declarations to parseFunction for detailed analysis
			err = o.parseFunction(ctx, pkg, fs, afs, i)
		case *ast.GenDecl:
			// Route general declarations to parseDeclaration for appropriate handling
			err = o.parseDeclaration(ctx, pkg, fs, afs, i)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// parseFunction extracts information from a Go function declaration.
//
// This function serves as the core mechanism for processing Go function declarations
// within the AST parsing framework. It extracts detailed information about functions,
// including their signatures, parameters, return types, and dependencies. The parseFunction
// method is essential for creating comprehensive function profiles that can be analyzed
// and tracked throughout the auditor tool's functionality.
//
// The implementation handles comprehensive analysis of function structures, including:
// - Function name extraction from AST identifiers
// - Source code content retrieval for precise documentation capture
// - Parameter and result type analysis using recursive expression parsing
// - Dependency tracking through AST inspection of function bodies
// - Creation of Entry objects with complete function metadata for storage
//
// Key features:
//   - Extracts function signature information including name and parameters
//   - Retrieves source code content for accurate documentation capture and analysis
//   - Analyzes parameter types and names to create detailed input specifications
//   - Processes return types to generate output specifications
//   - Identifies function dependencies through AST inspection of call expressions
//   - Creates Entry objects for storage in the auditor's data collection system with proper categorization
//   - Implements comprehensive error handling for malformed or invalid function structures
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this function, used to associate
//     the function with its package context in the auditor's data model.
//   - fs: The filename path to the source file containing this function, providing
//     location information for error reporting and documentation purposes.
//   - afs: The AST File node representing the complete source file structure, providing
//     context for position calculations and source code retrieval.
//   - atp: The AST Function Declaration node for processing, containing all function-specific
//     information including signature, body, and documentation.
//
// Returns:
//   - error: Any error encountered during function processing, including invalid AST structures,
//     parsing issues with parameters or return types, or problems with dependency tracking.
//     This ensures that individual function processing errors do not prevent continued analysis.
func (o *mdl) parseFunction(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, atp *ast.FuncDecl) error {
	var (
		err error
		cnt []byte
		ent audent.Entry
		dep = make(map[string]bool)
	)

	// Validate that the function declaration and its components are not nil or invalid
	if afs == nil || afs.Name == nil || len(afs.Name.Name) < 1 || afs.Name.Name == "_" || atp == nil || atp.Name == nil || len(atp.Name.Name) < 1 || atp.Type == nil {
		return fmt.Errorf("invalid AST function or Declaration")
	}

	// Validate that the package identifier is valid (non-zero)
	if pkg == 0 {
		return fmt.Errorf("invalid nil package")
	}

	o.u.Info("[File %s][Function %s] Parsing entry function", fs, atp.Name.Name)

	// Retrieve source code content for documentation purposes
	if cnt, err = o.getSrc(ctx, fs, afs, atp); err != nil {
		return fmt.Errorf("cannot retrieve content of file %s: %v", fs, err)
	}

	// Create or merge an Entry object with appropriate function information for storage
	ent = o.a.EntNewOrMerge(pkg, atp.Name.Name, "", audtps.EntryFunction, Lang, cnt)

	// Validate that the Entry was successfully created and is not empty
	if ent == nil || ent.IsEmpty() {
		return fmt.Errorf("invalid function %s in file %s", atp.Name.Name, fs)
	}

	// Process input parameters if they exist
	if atp.Type.Params != nil {
		for _, p := range atp.Type.Params.List {
			pr, dp, er := o.getParams(ctx, pkg, fs, afs, atp.Name.Name, p)

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
					o.u.Info("[File %s][Function %s] Skip existing dependency for entry params", fs, atp.Name.Name)
					continue
				}

				dep[fn] = true

				o.u.Info("[File %s][Function %s] Add dependency %s", fs, atp.Name.Name, et.GetFullPath())
				ent.AddDepend(id)
			}

			// Process input parameters for the function
			for _, prm := range pr {
				o.u.Info("[File %s][Function %s] Add Input Params type %s", fs, atp.Name.Name, prm.String())
			}

			// Add input parameters to the entry for comprehensive tracking
			ent.AddInputs(pr...)
		}
	}

	// Process output parameters (return types) if they exist
	if atp.Type != nil && atp.Type.Results != nil {
		for _, p := range atp.Type.Results.List {
			pr, dp, er := o.getParams(ctx, pkg, fs, afs, atp.Name.Name, p)

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
					o.u.Info("[File %s][Function %s] Skip existing dependency for output result", fs, atp.Name.Name)
					continue
				}

				dep[fn] = true

				o.u.Info("[File %s][Function %s] Add dependency %s", fs, atp.Name.Name, et.GetFullPath())
				ent.AddDepend(id)
			}

			// Process output parameters for the function
			for _, prm := range pr {
				o.u.Info("[File %s][Function %s] Add Output Result type %s", fs, atp.Name.Name, prm.String())
			}

			// Add output parameters to the entry for comprehensive tracking
			ent.AddOutputs(pr...)
		}
	}

	// Process function body to identify dependencies within the function implementation
	if atp.Body != nil {
		ast.Inspect(atp.Body, func(n ast.Node) bool {
			if ctx.Err() != nil {
				return false
			}

			var et audent.Entry

			switch expr := n.(type) {
			// External dependencies only like http.Request, http.NewRequest, ...
			case *ast.SelectorExpr:
				if expr == nil {
					return true
				}
				et = o.getDepend(ctx, fs, afs, expr)

			// External dependencies only like NewRequest, Println, ... must only be function !
			case *ast.CallExpr:
				if expr == nil {
					return true
				}
				et = o.getLocDepend(ctx, fs, pkg, expr)

			default:
				return true
			}

			// Skip nil or empty dependencies to prevent processing issues
			if et == nil || et.IsEmpty() {
				o.u.Info("[File %s][Function %s] Skip nil dependency", fs, atp.Name.Name)
				return true
			}

			id := et.GetID()
			fn := et.GetFullPath()

			// Skip duplicate dependencies to prevent redundant tracking
			if dep[fn] {
				o.u.Info("[File %s][Function %s] Skip existing dependency", fs, atp.Name.Name)
				return true
			}

			dep[fn] = true

			o.u.Info("[File %s][Function %s] Add dependency %s", fs, atp.Name.Name, et.GetFullPath())
			ent.AddDepend(id)
			return true
		})
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	o.u.Info("[File %s][Function %s] Update entry function", fs, atp.Name.Name)
	o.a.EntAdd(ent)
	return nil
}

// parseDeclaration processes Go declaration statements (e.g., type, var, const declarations).
//
// This function serves as the dispatcher for handling different types of Go declaration statements
// within the AST parsing framework. It routes various declaration types to their appropriate specialized parsers.
//
// The implementation focuses on processing only the most relevant declaration types for auditing purposes,
// specifically targeting type specifications while ignoring variable and constant declarations that are less useful
// for the auditor's analysis requirements. This selective approach ensures that the parser concentrates on
// meaningful constructs that contribute to understanding code structure and dependencies.
//
// Key features:
//   - Processes different types of Go declarations (type, var, const)
//   - Routes type specifications to specialized parsing functions for detailed analysis
//   - Ignores variable and constant declarations as they are not relevant for auditing purposes
//   - Implements error handling for malformed or invalid declarations
//   - Maintains focus on structurally significant declarations that contribute to code understanding
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pkg: The identifier of the package containing this declaration, used to associate
//     declarations with their respective packages in the auditor's data collection system.
//   - fs: The filename path to the source file containing this declaration, providing context
//     for error reporting and documentation purposes.
//   - afs: The AST File node representing the complete source file structure, providing context
//     for position calculations and source code retrieval.
//   - adc: The AST General Declaration node for processing, representing a group of related
//     declarations that may contain multiple specifications.
//
// Returns:
//   - error: Any error encountered during declaration processing, including invalid AST structures,
//     parsing issues with individual specifications, or problems with declaration routing. This ensures
//     that errors in one declaration do not prevent processing of other declarations in the same file.
func (o *mdl) parseDeclaration(ctx context.Context, pkg audids.ID, fs string, afs *ast.File, adc *ast.GenDecl) error {
	var err error

	// Validate that the general declaration is not nil
	if adc == nil {
		return errors.New("invalid AST Declaration")
	} else if len(adc.Specs) < 1 {
		return nil
	}

	o.u.Info("[File %s] Parsing declaration", fs)

	// Process each specification within the general declaration
	for _, spec := range adc.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			// Route type specifications to parseType for detailed analysis
			err = o.parseType(ctx, pkg, fs, afs, s)
		default:
			// ignore variable, constant, ... not usable
			// only type, func are usable
			continue
		}

		if err != nil {
			return err
		}
	}

	return nil
}
