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
	"errors"
	"flag"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	astgen "github.com/nabbar/auditor/pkg/ast/generic"
	auddbm "github.com/nabbar/auditor/pkg/data/manager"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	auduxi "github.com/nabbar/auditor/pkg/uxi"
)

// Lang defines the human-readable name for the Go language within this AST package.
//
// This constant provides a human-readable string representation of the Go language
// that is used throughout the AST parsing system to identify and display the language
// type. It serves as a consistent identifier for Go-specific operations and displays,
// ensuring that all Go-related functionality is clearly marked and distinguishable.
const Lang = "Go"

// flg represents a flag that controls whether Go language parsing is enabled or disabled.
//
// This global variable acts as a configuration switch for enabling or disabling
// Go language parsing functionality. It is initialized via command-line flags and
// can be used to conditionally enable or disable Go-specific AST processing. This allows
// users to control the inclusion of Go language support at runtime.
var flg bool

func init() {
	// Initialize the flag for controlling Go language parsing.
	//
	// This flag allows users to enable or disable Go language support at runtime,
	// providing flexibility in how the auditor tool handles different programming languages.
	flag.BoolVar(&flg, "lang-go", true, "Allow enable/disable parsing language Go")
}

// Identify determines whether a given file is a Go module file (go.mod).
//
// This function performs language-specific identification for Go module files by checking
// if the base name of the file equals "go.mod". Go module files are the standard configuration
// files used by Go modules to define dependencies and project structure. This identification
// method is crucial for proper language detection in source code analysis.
//
// Parameters:
//   - file: The full path to the file being checked for Go module identification.
//
// Returns:
//   - bool: True if the file is a Go module file (go.mod), false otherwise. This indicates
//     whether the specified file conforms to the standard Go module file naming convention.
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

// GetAST creates and returns an AST parser instance for Go module files.
//
// This function provides the core implementation for creating AST parsers specifically
// designed for Go module files. It validates that the provided file is a valid Go module
// file (go.mod) and initializes an appropriate AST structure for parsing Go module
// configuration. The returned AST parser is configured with language-specific settings
// to ensure accurate syntax analysis of Go module files.
//
// Parameters:
//   - uim: The UXI Manager instance used for user interaction and logging.
//   - mgr: The AST Manager instance for handling collections of AST elements and managing
//     the overall AST processing workflow.
//   - file: The full path to the go.mod file to parse for AST generation.
//
// Returns:
//   - astgen.AST: An AST parser instance configured for Go module parsing, or nil if
//     initialization fails. This provides a structured interface for analyzing Go module files.
//   - error: Any error encountered during AST creation, including invalid file names or
//     initialization errors. This includes cases where the file is not a valid go.mod file.
func GetAST(uim auduxi.Manager, mgr *auddbm.Linker, file string) (astgen.AST, error) {
	// Validate that the provided file is a valid Go module file before creating AST.
	if !Identify(file) {
		return nil, fmt.Errorf("invalid file name go.mod for file %s", file)
	}

	if mgr == nil || mgr.IsEmpty() {
		return nil, errors.New("invalid DB manager")
	}

	// Create and return a new AST parser instance configured for Go module files.
	return &mdl{
		x: sync.RWMutex{},
		u: uim,
		a: mgr,
		d: filepath.Dir(file),
		f: filepath.Base(file),
		m: func() audmod.Module {
			return nil
		},
		c: make(map[string][]byte),
	}, nil
}

// Pattern returns the file patterns associated with Go language files.
//
// This function provides a list of file pattern strings that represent the standard
// naming conventions and extensions for Go language files. These patterns are used
// for identifying files that belong to the Go programming language based on their
// naming conventions and extensions. The returned patterns include both module files
// and source code files.
//
// Returns:
//   - []string: A slice of file pattern strings representing Go language file extensions
//     and naming conventions, or nil if Go language parsing is disabled. This includes
//     standard Go module files (go.mod), Go source files (*.go), and test files (*test*.go).
func Pattern() []string {
	// Return patterns only if Go language parsing is enabled.
	if !flg {
		return nil
	}

	// Return the standard Go module file pattern.
	return []string{"go.mod"}
	// TODO: Uncomment and implement when support for test files is required.
	// return []string{"go.mod", "^*test*"}
}
