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

// Package generic provides the generic interface for AST (Abstract Syntax Tree) parsers
// that are used across different programming languages in the auditor tool.
//
// This package defines a standardized interface for Abstract Syntax Tree (AST) parsing
// operations, enabling consistent handling of source code analysis across multiple
// programming languages. The interface abstracts language-specific parsing details,
// allowing uniform processing of module and package structures regardless of the
// underlying implementation.
//
// Key aspects of this package include:
//   - Generic AST interface definition for standardized parsing operations
//   - Module-level parsing capabilities for extracting high-level structure information
//   - Package-level parsing for analyzing individual package constructs
//   - Resource management through io.Closer interface implementation
//   - Error handling for parsing failures and validation issues
//
// The generic interface serves as the foundation for language-specific implementations
// in the auditor tool, ensuring that all supported languages follow consistent patterns
// for AST processing while allowing for specialized behavior in each language implementation.
package generic

import (
	"context"
	"io"

	audmod "github.com/nabbar/auditor/pkg/data/mod"
)

// AST represents an Abstract Syntax Tree parser interface for various programming languages.
//
// The AST interface defines a contract for parsing source code into structured representations
// that can be analyzed by the auditor tool. This interface enables consistent handling of
// different programming languages while maintaining language-specific capabilities.
//
// Key features of this interface include:
//   - Standardized parsing operations across multiple programming languages
//   - Module-level parsing capabilities for extracting high-level structure information
//   - Package-level parsing for analyzing individual package constructs
//   - Resource management through io.Closer interface implementation
//   - Error handling for parsing failures and validation issues
//
// The AST interface serves as the foundation for language-specific implementations in the auditor tool,
// ensuring that all supported languages follow consistent patterns for AST processing while allowing
// for specialized behavior in each language implementation.
//
// Usage considerations:
//   - Implementations must properly handle context cancellation and timeouts
//   - All methods should return meaningful error messages for debugging purposes
//   - Resource cleanup through Close() method is essential for preventing memory leaks
type AST interface {
	// Closer interface implementation
	//
	// Provides the Close method to clean up resources when the AST is no longer needed.
	// This method ensures proper resource management and prevents memory leaks by clearing
	// cached data, references, and other state information that may have been accumulated
	// during parsing operations. It implements the standard io.Closer interface for consistent
	// resource handling across different language parsers.
	//
	// The Close method is designed to release any resources held by the AST implementation,
	// including but not limited to cached parsing results, file handles, and internal buffers.
	// This ensures proper cleanup when the AST instance is no longer required.
	//
	// Returns:
	//   - error: Any error encountered during cleanup operations. While this implementation
	//     currently returns nil, it maintains compatibility with the io.Closer interface
	//     for consistent resource management practices across different language parsers.
	io.Closer

	// ParseModule parses the entire module structure from the source file.
	//
	// This method processes the complete module definition, including its metadata,
	// dependencies, and overall structure. It extracts information about the module's
	// name, version, and other relevant attributes that define the module's identity
	// within the language ecosystem. The ParseModule function is essential for establishing
	// the foundational context for subsequent package-level analysis.
	//
	// This operation typically involves reading and interpreting the module declaration
	// at the top level of a source file or project structure, extracting key information
	// such as module name, version constraints, dependencies, and other metadata that
	// describes the module's characteristics.
	//
	// Context handling:
	//   - The context parameter allows for cancellation and timeout control during parsing
	//   - Implementations should respect context cancellation to prevent hanging operations
	//
	// Returns:
	//   - audmod.Module: The parsed Module instance containing all extracted module information,
	//     or nil if parsing fails due to invalid file content or missing module declarations
	//   - error: Any error encountered during module parsing, including file read errors,
	//     invalid module declarations, or parsing failures that prevent successful extraction
	//     of module information. Error messages should provide sufficient detail for debugging.
	ParseModule(context.Context) (audmod.Module, error)

	// ParsePackage parses package-level structures from the source file.
	//
	// This method handles parsing of packages, which may include imports, exports,
	// and other package-level constructs. It processes the package's content to extract
	// relevant information for analysis, including function declarations, type definitions,
	// and dependency relationships. The ParsePackage function is crucial for understanding
	// the structure and relationships within individual packages.
	//
	// Package parsing typically involves analyzing the source code at the package level,
	// extracting constructs such as:
	//   - Function and method definitions
	//   - Type declarations (structs, interfaces, etc.)
	//   - Variable and constant declarations
	//   - Import statements and their relationships
	//   - Exported identifiers and visibility information
	//
	// Context handling:
	//   - The context parameter allows for cancellation and timeout control during parsing
	//   - Implementations should respect context cancellation to prevent hanging operations
	//
	// Returns:
	//   - error: Any error encountered during package parsing, including file read errors,
	//     invalid package structures, or parsing failures that prevent successful extraction
	//     of package information. Returns nil if parsing completes successfully.
	//     Error messages should provide sufficient detail for debugging purposes.
	ParsePackage(context.Context) error

	// UpdatePackage updates the package structure with new information.
	//
	// This method allows for refreshing or updating package data after initial parsing.
	// It may be used to incorporate changes made to the source file since the last parse
	// operation, ensuring that package-level analysis reflects current code state.
	//
	// Context handling:
	//   - The context parameter allows for cancellation and timeout control during update
	//   - Implementations should respect context cancellation to prevent hanging operations
	//
	// Returns:
	//   - error: Any error encountered during package update operations, including file read errors,
	//     invalid package structures, or update failures that prevent successful synchronization.
	//     Returns nil if the update completes successfully. Error messages should provide sufficient
	//     detail for debugging purposes.
	UpdatePackage(context.Context) error

	// GetFile returns the name of the file being parsed.
	//
	// This method provides access to the filename of the source file that is currently
	// being processed by the AST parser. It allows callers to identify which file's content
	// is represented by the current AST instance.
	//
	// Returns:
	//   - string: The name of the file being parsed, including extension but excluding path
	GetFile() string

	// GetDirectory returns the directory path of the file being parsed.
	//
	// This method provides access to the directory containing the source file that is currently
	// being processed by the AST parser. It allows callers to determine the location of the file
	// within the filesystem structure.
	//
	// Returns:
	//   - string: The directory path of the file being parsed, excluding filename
	GetDirectory() string

	// GetFullPath returns the complete absolute path of the file being parsed.
	//
	// This method provides access to the full filesystem path of the source file that is currently
	// being processed by the AST parser. It allows callers to obtain the complete path including
	// directory and filename, which can be useful for debugging, logging, or file system operations.
	//
	// Returns:
	//   - string: The complete absolute path of the file being parsed, including directory and filename
	GetFullPath() string
}
