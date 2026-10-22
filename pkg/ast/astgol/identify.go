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
	"flag"
	"fmt"
	"go/ast"
	"path/filepath"
	"strings"
	"sync"

	astgen "github.com/nabbar/auditor/pkg/ast/generic"
	audids "github.com/nabbar/auditor/pkg/data/id"
	auddbm "github.com/nabbar/auditor/pkg/data/manager"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audtps "github.com/nabbar/auditor/pkg/data/types"
	auduxi "github.com/nabbar/auditor/pkg/uxi"
)

// Lang is a constant that defines the human-readable name
// for the Go programming language within the AST parsing system.
// It provides a consistent string identifier used throughout
// the auditor framework to label Go-specific AST operations,
// display language type in output, and distinguish Go-related
// functionality from other supported languages.
const Lang = "Go"

// flg is a global boolean variable that controls whether Go
// language parsing is enabled or disabled within the auditor
// framework. It serves as a runtime configuration switch,
// allowing users to selectively enable or disable Go language
// support via the command-line "lang-go" flag. When set to
// false, all Go-specific AST identification and parsing
// operations are skipped.
var flg bool

// init initializes the command-line flag that controls Go
// language parsing. It registers the "lang-go" boolean flag
// with flag.BoolVar, binding it to the "flg" global variable
// with a default value of true. This means Go language parsing
// is enabled by default, and users can disable it at runtime
// by passing the -lang-go=false command-line argument.
func init() {
	flag.BoolVar(&flg, "lang-go", true, "Allow enable/disable parsing language Go")
}

// Identify determines whether a given file is a Go module
// file (go.mod). First, it checks if the "flg" variable is
// false and returns false if Go language parsing is disabled.
// Then, it checks the file's parent directory: if the directory
// path or any ancestor contains "vendor", the function returns
// false to exclude vendored Go module files from processing.
// Similarly, if the path contains "test", the function returns
// false to exclude test-related go.mod files. Finally, it
// verifies that the file's base name is exactly "go.mod"
// using filepath.Base. Only when all these checks pass does
// it return true.
func Identify(file string) bool {
	if !flg {
		return false
	}

	// Exclude vendor directories from processing
	if strings.Contains(filepath.Dir(file), "vendor") {
		return false
	}

	// Exclude test directories from processing
	if strings.Contains(filepath.Dir(file), "test") {
		return false
	}

	// Check if Go language parsing is enabled and verify if the file matches the go.mod pattern.
	return filepath.Base(file) == "go.mod"
}

// GetAST creates and returns an AST parser instance for Go
// module files. It first invokes "Identify" to validate that
// the supplied file is a valid go.mod file; if the validation
// fails, the function returns an error. It then verifies that
// the provided Linker manager is non-nil and non-empty; a
// missing or empty manager results in an error. Upon success,
// it constructs a "mdl" struct with a sync.RWMutex for
// concurrency safety, the UXI manager for logging, the Linker
// for AST element storage, the file's parent directory, the
// file base name, a nil Module factory, and a byte-cache map.
// It also populates the handler map with set/update callback
// pairs for each AST entry type: EntryFunction, EntryStruct,
// EntryInterface, EntryPrimitive, EntryCustom, EntryConstant,
// and EntryVariable.
func GetAST(uim auduxi.Manager, mgr *auddbm.Linker, file string) (astgen.AST, error) {
	// Validate that the provided file is a valid Go module file before creating AST.
	if !Identify(file) {
		return nil, fmt.Errorf("invalid file name go.mod for file %s", file)
	}

	if mgr == nil || mgr.IsEmpty() {
		return nil, errors.New("invalid DB manager")
	}

	// Create and return a new AST parser instance configured for Go module files.
	m := &mdl{
		x: sync.RWMutex{},
		u: uim,
		a: mgr,
		d: filepath.Dir(file),
		f: filepath.Base(file),
		m: func() audmod.Module {
			return nil
		},
		c: make(map[string][]byte),
	}

	m.h = map[audtps.CodeType]map[bool]func(context.Context, audids.ID, string, *ast.File, ast.Node){
		audtps.EntryFunction: {
			false: m.setFunc,
			true:  m.updFunc,
		},
		audtps.EntryStruct: {
			false: m.setStruct,
			true:  m.updStruct,
		},
		audtps.EntryInterface: {
			false: m.setInterface,
			true:  m.updInterface,
		},
		audtps.EntryPrimitive: {
			false: m.setPrimitive,
			true:  m.updPrimitive,
		},
		audtps.EntryCustom: {
			false: m.setCustom,
			true:  m.updCustom,
		},
		audtps.EntryConstant: {
			false: m.setConstant,
			true:  m.updConstant,
		},
		audtps.EntryVariable: {
			false: m.setVariable,
			true:  m.updVariable,
		},
	}

	return m, nil
}

// Pattern returns a slice of file pattern strings used for
// language identification of Go module files. If the "flg"
// variable is false (meaning Go language parsing is disabled
// via the lang-go command-line flag), the function returns
// nil. Otherwise, it returns a string slice containing the
// single entry "go.mod", which is the standard file name for
// Go module definitions. Go source files (*.go) and test
// files (*_test.go) are not listed here because they are
// handled through the ParsePackage loading mechanism rather
// than through this pattern-list-based detection.
func Pattern() []string {
	// Return patterns only if Go language parsing is enabled via the lang-go flag.
	if !flg {
		return nil
	}

	// Return the standard Go module file pattern.
	// This is the entry point for Go language detection in the auditor tool.
	return []string{"go.mod"}
}
