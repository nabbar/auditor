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

//
// Package pkg/ast/generic:
// ========================
//
// Package generic provides the generic interface for AST (Abstract Syntax Tree) parsers
//
// This package defines the foundational interface for Abstract Syntax Tree parsing operations
// across different programming languages in the auditor tool. It establishes a standardized
// contract that all language-specific AST implementations must follow, ensuring consistent
// functionality while allowing for specialized behavior in each language parser.
//
// Key features include:
//   - Standardized parsing interface for different programming languages
//   - Module-level parsing capabilities for extracting high-level structure information
//   - Package-level parsing for analyzing individual package constructs
//   - Resource management through io.Closer interface implementation
//   - Error handling for parsing failures and validation issues
//
// The generic AST interface serves as the foundation for language-specific implementations
// in the auditor tool, ensuring that all supported languages follow consistent patterns for
// AST processing while allowing for specialized behavior in each language implementation.
//
// The import path for this package is "github.com/nabbar/auditor/pkg/ast/generic" and the alias is "astgen".
//
// `type AST interface`:
//   - `io.Closer`:
//      Provides the Close method to clean up resources when the AST is no longer needed
//      Implements resource management by clearing cached data, references, and state information
//      Ensures proper cleanup of parsing resources to prevent memory leaks
//
//   - `ParseModule(context.Context) (audmod.Module, error)`:
//      Parses the entire module structure from the source file, returning a Module object and any error that occurred
//      Parameter:
//       - context.Context: Context allows for cancellation and timeout control during parsing operations
//      Returns:
//       - audmod.Module: The parsed Module instance containing all extracted module information,
//         or nil if parsing fails due to invalid file content or missing module declarations
//       - error: Any error encountered during module parsing, including file read errors,
//         invalid module declarations, or parsing failures that prevent successful extraction
//         of module information
//
//   - `ParsePackage(context.Context) error`:
//      Parses package-level structures from the source file, handling imports, exports, and other package-level constructs
//      Parameter:
//       - context.Context: Context allows for cancellation and timeout control during parsing operations
//      Return:
//       - error: Any error encountered during package parsing, including file read errors,
//         invalid package structures, or parsing failures that prevent successful extraction
//         of package information. Returns nil if parsing completes successfully
//
//   - `UpdatePackage(context.Context) error`:
//      Updates package structure with new information after initial parsing
//      Parameter:
//       - context.Context: Context allows for cancellation and timeout control during update operations
//      Return:
//       - error: Any error encountered during package update operations, including file read errors,
//         invalid package structures, or update failures that prevent successful synchronization
//
//   - `GetFile() string`:
//      Returns the name of the file being parsed
//      Parameter:
//       - None: This method operates on the internal state of the AST instance
//      Return:
//       - string: The name of the file being parsed, including extension but excluding path
//
//   - `GetDirectory() string`:
//      Returns the directory path of the file being parsed
//      Parameter:
//       - None: This method operates on the internal state of the AST instance
//      Return:
//       - string: The directory path of the file being parsed, excluding filename
//
//   - `GetFullPath() string`:
//      Returns the complete absolute path of the file being parsed
//      Parameter:
//       - None: This method operates on the internal state of the AST instance
//      Return:
//       - string: The complete absolute path of the file being parsed, including directory and filename
//
//

package generic
