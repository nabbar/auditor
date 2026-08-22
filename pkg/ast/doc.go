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
// Package pkg/ast:
// ================
//
// This package provides language identification and AST (Abstract Syntax Tree) parsing capabilities
// for different programming languages supported by the auditor tool.
//
// This package serves as the core interface for language detection and parsing within the auditor system.
// It defines the Lang enum for supported languages, provides methods to identify files by language,
// and retrieve AST representations for analysis.
//
// The package follows a modular architecture where specific language implementations are handled by sub-packages,
// allowing for extensibility and clean separation of concerns.
//
// The import path for this package is : "github.com/nabbar/auditor/pkg/ast" and the alias is "audast"
//
// `type Lang uint8`:
//   - LangUnknown: Represents an unknown or unsupported language
//   - LangAsm: Represents Assembly language
//   - LangC: Represents C programming language
//   - LangGo: Represents Go programming language
//   - LangCpp: Represents C++ programming language
//   - LangRust: Represents Rust programming language
//   - LangPython: Represents Python programming language
//   - LangJava: Represents Java programming language
//   - LangPhp: Represents PHP programming language
//   - LangJS: Represents JavaScript programming language
//
// `func (Lang) String() string`:
//   - function to allow fmt.Stringer interface implementation
//   - Returns the human-readable name of the language
//   - Return result: string containing the language name
//
// `func (Lang) Identify(string) bool`:
//   - Determines whether a given file matches the language type
//   - Uses language-specific identification logic to detect if a file belongs to this programming language category
//   - Entry params:
//       - string :  is a file path allowing to identify the language associated
//   - Return result: bool indicating if file matches language
//
// `func (Lang) AST(astgen.Manager, string) (astgen.AST, error)`:
//   - Retrieves the Abstract Syntax Tree for a given file of this language type
//   - Returns an AST implementation specific to the language and handles errors appropriately
//   - Entry params:
//       - astgen.Manager: an instance of UI Manager to allow context following, log messages, progress bar creation. (See pkg/uxi for more details)
//       - string:  is a file path allowing to identify the language associated
//   - Return result: astgen.AST and error
//
// `func (o Lang) Pattern() []string`:
//   - Returns the file pattern or extension associated with this language type
//   - Provides a list of file patterns or extensions that are commonly used for files belonging to this programming language
//   - Returns a slice of file pattern strings representing the file extensions or naming conventions associated with this language type
//   - Returns nil if no patterns are defined for this language
//
// `func Identify(string) Lang`:
//   - Determines the programming language of a given file by checking all supported languages
//   - Iterates through all available language types and uses their Identify method to find the appropriate language for the specified file
//   - Entry params:
//       - string :  is a file path allowing to identify the language associated
//   - Return result: Lang representing the identified language
//
// `func Pattern() map[Lang][]string`:
//   - Returns a map of file patterns for all supported languages
//   - Aggregates the file pattern information for all supported programming languages
//   - Constructs a mapping from language types to their corresponding file patterns
//   - Useful for applications that need to identify files based on their extensions or naming conventions
//   - Returns a map where keys are language types and values are slices of file pattern strings representing the extensions or naming conventions associated with each language type
//
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
//   - `ParseModule() (audmod.Module, error)`:
//      Parses the entire module structure from the source file, returning a Module object and any error that occurred
//      Parameter:
//       - None: This method operates on the internal state of the AST instance
//      Returns:
//       - audmod.Module: The parsed Module instance containing all extracted module information,
//         or nil if parsing fails due to invalid file content or missing module declarations
//       - error: Any error encountered during module parsing, including file read errors,
//         invalid module declarations, or parsing failures that prevent successful extraction
//         of module information
//
//   - `ParsePackage() error`:
//      Parses package-level structures from the source file, handling imports, exports, and other package-level constructs
//      Parameter:
//       - None: This method operates on the internal state of the AST instance
//      Return:
//       - error: Any error encountered during package parsing, including file read errors,
//         invalid package structures, or parsing failures that prevent successful extraction
//         of package information. Returns nil if parsing completes successfully
//
//
// Package pkg/ast/astgol:
// ========================
//
// The astgol sub-package provides Go-specific AST parsing capabilities for the auditor tool.
//
// This package implements the AST interface for Go language files, specifically handling:
//   - go.mod file identification and module parsing
//   - Package-level parsing using golang.org/x/tools/go/packages
//   - Function and type declaration extraction from Go source files
//   - Source code content retrieval with caching mechanisms
//
// The import path for this package is "github.com/nabbar/auditor/pkg/ast/astgol" and the alias is "astgol".
//
// `const Lang = "Go"`
//   - Defines the human-readable language name for Go within the AST parsing system
//   - Provides a consistent identifier for Go-specific operations and displays
//   - Used throughout the package to identify and display the Go language type
//
// `func Identify(file string) bool`:
//   - Determines whether a given file is a Go module file (go.mod)
//   - Checks if the base name of the file equals "go.mod"
//   - Returns true only if the filename is exactly "go.mod"
//
// `func GetAST(astgen.Manager, string) (astgen.AST, error)`:
//   - Creates and returns a Go-specific AST parser instance for go.mod files
//   - Validates that the provided file is a valid Go module file
//   - Returns an AST implementation specific to Go language parsing
//   - Handles validation errors with descriptive messages
//
// `func Pattern() []string`:
//   - Returns the file patterns associated with Go language files
//   - Provides a list of file pattern strings representing Go language file extensions and naming conventions
//   - Includes standard Go module files (go.mod), Go source files (*.go), and test files (*test*.go)
//   - Returns nil if Go language parsing is disabled via command-line flag
//
// `func (mdl) Close() error`:
//   - Cleans up the AST parser state by clearing cached data and references
//   - Implements io.Closer interface for proper resource cleanup
//   - Clears the source code cache to free up memory
//   - Resets the module retrieval function to nil to prevent dangling references
//   - Clears filename and directory fields to reset state
//
// `func (mdl) ParseModule() (audmod.Module, error)`:
//   - Parses Go module information from the go.mod file
//   - Reads the go.mod file and extracts the module name
//   - Returns a Module object and stores it in the collection
//   - Implements proper error handling for missing or malformed module declarations
//
// `func (mdl) ParsePackage() error`:
//   - Parses Go module information from the go.mod file
//   - Reads the go.mod file and extracts the module name
//   - Returns a Module object and stores it in the collection
//   - Implements proper error handling for missing or malformed module declarations
//
// AST Golang Dataflow :
// The processing of a Go module by your parser follows a descending (Top-Down) approach, going from the project root to the code expression level.
//   1 - Initialization: The process begins with validating the root file using Identify (checks for go.mod name).
//       If valid, GetAST instantiates the main structure mdl which will serve as state context (containing cache, AST manager, paths).
//   2 - Module Analysis (ParseModule): The parser reads the go.mod file line by line to extract the module name (via the module directive) and
//       stores it in the global manager.
//   3 - Package Loading (ParsePackage): The parser delegates tree reading to golang.org/x/tools/go/packages.
//       It retrieves syntax (complete AST) of all valid Go files in the directory, then iterates over each file with parseFile.
//   4 - File Routing (parseFile): For each file, the system traverses top-level declarations (Decls) and acts as a router:
//         - *ast.FuncDecl are sent to parseFunction
//         - *ast.GenDecl are sent to parseDeclaration
//   5 - Entity Extraction:
//         - For Functions (parseFunction):
//            The parser retrieves exact source code (including comments via getSrc and getPosition),
//            analyzes input/output types via parseExprCode, and inspects the function body for calls (*ast.CallExpr) to establish dependencies.
//         - For Types (parseDeclaration -> parseType):
//            The parser isolates type specifications (*ast.TypeSpec), identifies whether it's an interface,
//            structure, primitive type or custom type via parseExprCode, and stores it.
//   6 - Dependency Resolution (getDepend): When an external function call is detected (*ast.SelectorExpr),
//       the parser climbs up the expression tree with resolveRootIdent. It cleans the import path with guessPkgName
//       (to handle aliases, versions /v2 and prefixes/suffixes) to link the function to the correct dependency in the manager.
//
// AST Golang process schematic:
// [Entry] GetAST(go.mod) -> Initialize *mdl
//   │
//   ├──> ParseModule()
//   │    └──> getFileSrc() -> Read go.mod into memory
//   │
//   └──> ParsePackage() -> Load astgop.Load()
//        └──> Loop through packages and syntax files
//             └──> parseFile(AST File)
//                  │
//                  ├──> [If FuncDecl] parseFunction()
//                  │    ├──> getSrc() / getPosition() -> Extract source code and documentation
//                  │    ├──> parseExprCode() -> Resolve parameter types (Params)
//                  │    ├──> parseExprCode() -> Resolve return types (Results)
//                  │    └──> ast.Inspect(Body) -> Search for *ast.CallExpr
//                  │         ├──> [If internal call] -> Add local dependency
//                  │         └──> [If external call] getDepend()
//                  │              ├──> resolveRootIdent() -> Extract called package name
//                  │              ├──> guessPkgName() -> Clean import (e.g., remove .v2, /v3)
//                  │              └──> getPackage() / addPackage() -> Store import in vendor
//                  │
//                  └──> [If GenDecl] parseDeclaration()
//                       └──> [Filter] If *ast.TypeSpec -> parseType()
//                            ├──> parseExprCode() -> Identify (Struct, Interface, Alias...)
//                            └──> Store type entity
//
// AST golang ignored items :
// The parser is designed with a specific audit perspective. Therefore, it excludes many valid data in Go.
// Here is the precise list of what is excluded or ignored by this code:
//   - In the go.mod file
//     Secondary directives: The ParseModule function only looks for lines starting with module.
//     It completely ignores require, replace, exclude and go (language version) directives.
//     Dependencies defined in go.mod are therefore not preloaded.
//
//   - At the global file parsing level
//      - Files without declarations: If len(afs.Decls) < 1, the file is silently ignored.
//      - Files without syntax: During package loading, if the Syntax list is empty, the file is ignored.
//      - Blank identifiers (_): Functions or types named _ generate an error and are rejected (considered invalid AST function or Declaration).
//
//   - In Global Declarations (parseDeclaration)
//      - Variables and Constants: In a GenDecl block (e.g., var x = 1 or const y = 2), the switch only processes *ast.TypeSpec.
//        The default case continues and deliberately ignores all global variable and constant declarations.
//
//   - In function bodies (parseFunction)
//      - Business logic and control structures: ast.Inspect focuses only on *ast.CallExpr (function calls).
//        All other code in the function is ignored by the analyzer (for loops, if/switch conditions, variable assignments :=, mathematical operations).
//      - Unresolved dependencies: In getDepend, if the import path doesn't match any import validated by guessPkgName or aliases (if len(ipp) < 1),
//        the function returns nil and the dependency is silently ignored.
//
//   - In expression resolution (parseExprCode)
//      - Unknown types: If the evaluated AST type is not handled by the switch (covering Ident, FuncType, StructType, InterfaceType, SelectorExpr, StarExpr,
//        ArrayType, MapType, Ellipsis and ChanType), it falls into the default case.
//      - Default consequence: The parser returns ok = false with "unknown". If this occurs during analysis of function parameters or return types,
//        it triggers fmt.Errorf("invalid params input...") or fmt.Errorf("invalid result type..."), causing the entire function analysis to fail.
//
//   - In complex chained calls (resolveRootIdent)
//      - Unhandled nodes: If a chained call contains an unhandled expression in the switch (Ident, SelectorExpr, CallExpr, IndexExpr, TypeAssertExpr, StarExpr),
//        the function returns nil. This causes a cascade nil return in getDepend, and the dependency is ignored.
//
//

package ast
