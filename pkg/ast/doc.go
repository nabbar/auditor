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

// Package ast provides language identification and AST (Abstract Syntax Tree)
// parsing capabilities for the programming languages supported by the auditor tool.
//
// This package defines the Lang enumeration type and the public API through which
// each language can be identified, its AST parser created, and file patterns
// retrieved. It delegates the actual parsing work to language-specific sub-packages
// (currently only Go via astgol). The package uses a modular architecture: each
// newly supported language adds its own sub-package alongside a Lang constant in the
// enumeration.
//
// The import path for this package is github.com/nabbar/auditor/pkg/ast. The
// conventional import alias is "audast".
//
// Supported Languages (Lang constants)
// =====================================
//
// The exported type Lang is a uint8 enumeration that enumerates every programming
// language the auditor tool recognizes:
//
//   - LangUnknown (iota 0): unknown or unsupported language
//   - LangAsm (iota 1): Assembly / Assembler
//   - LangC (iota 2): C programming language
//   - LangGo (iota 3): Go programming language -- the only language with a
//     complete AST parser implementation at this time
//   - LangCpp (iota 4): C++ programming language
//   - LangRust (iota 5): Rust programming language
//   - LangPython (iota 6): Python programming language
//   - LangJava (iota 7): Java programming language
//   - LangPhp (iota 8): PHP programming language
//   - LangJS (iota 9): JavaScript programming language
//
// Note that only LangGo has a working implementation of Identify and AST parsing.
// All other language constants exist so that new parsers can be added incrementally.
// For any language other than Go, Identify(file) always returns false, Pattern() always
// returns nil, and AST returns an error indicating the parser is not implemented.
//
//
// Public API
// ==========
//
// `func (Lang) String() string`:
//   Returns the human-readable language name. Implements the fmt.Stringer interface.
//   Returns "Assembler" for LangAsm, "C" for LangC, "Go" for LangGo (delegating to
//   astgol.Lang), "C++" for LangCpp, "Rust" for LangRust, "Python" for LangPython,
//   "Java" for LangJava, "PHP" for LangPhp, and "JavaScript" for LangJS. Returns the
//   value of audpkg.Unknown for any value that does not map to a known language.
//
// `func Parse(s string) Lang`:
//   Converts a human-readable language name into the corresponding Lang value.
//   Comparison is case-insensitive and delegates to each Lang's String() method,
//   so the exact casing does not matter. Returns LangUnknown when the string does
//   not match any known language.
//
// `func (Lang) Identify(file string) bool`:
//   Determines whether the given file path matches this language.
//   For LangGo, delegates to astgol.Identify which checks that:
//     - the file base name is exactly "go.mod"
//     - the file is not inside a "vendor" or "test" directory
//     - Go parsing is enabled (controlled by the -lang-go flag)
//   For all other languages, returns false.
//
// `func (Lang) AST(uim auduxi.Manager, mgr *auddbm.Linker, file string) (astgen.AST, error)`:
//   Creates and returns an AST parser instance for this language.
//   Parameters:
//     - uim: an auduxi.Manager used for logging, progress, and context.
//     - mgr: an auddbm.Linker that manages the database of parsed AST elements.
//     - file: the path to the source file to parse.
//   For LangGo, delegates to astgol.GetAST which validates the file and initializes
//   a *mdl instance. For all other languages, returns an error.
//
// `func (Lang) Pattern() []string`:
//   Returns the file pattern(s) used to recognize entry points for this language.
//   For LangGo, returns []string{"go.mod"} when Go parsing is enabled; returns nil
//   otherwise. For all other languages, returns nil.
//
// `func Identify(file string) Lang`:
//   Package-level convenience function that probes each known Lang constant in
//   declaration order and returns the first match. If no language matches, returns
//   LangUnknown. Because LangGo is the only supported language at the time of
//   writing, this function effectively identifies Go module files.
//
// `func Pattern() map[Lang][]string`:
//   Returns a map from Lang keys to their associated file pattern slices.
//   Languages whose Pattern() returns nil (or an empty slice) are omitted from the
//   map. Currently only LangGo has entries in this map.
//
// `func List() []Lang`:
//   Returns a slice of all supported Lang constants in declaration order. Useful
//   for iterating over the complete set of languages.
//
//
// Package ast/generic
// ====================
//
// The generic sub-package defines the AST interface that every language-specific
// parser must implement. It provides a four-step sequential pipeline that ensures a
// uniform driving contract across all languages.
//
// The import path is github.com/nabbar/auditor/pkg/ast/generic. The conventional
// alias is "astgen".
//
// `type AST interface`:
//   The unified contract for all AST parsers. It embeds io.Closer and exposes these
//   methods, which **must be called in order**:
//
//   - io.Closer (Close method):
//     Releases any cached state, buffer memory, file handles, or other resources
//     accumulated during parsing. Callers should invoke Close when parsing is
//     complete or when an error has terminated the pipeline.
//
//   - Step1(ctx context.Context) error:
//     Parses the root module or library definition. For Go this means reading go.mod,
//     extracting the module name and vendor dependencies, and registering them in the
//     database manager. Returns a descriptive error on failure.
//
//   - Step2(ctx context.Context) error:
//     Parses external (third-party or vendored) package structures. For Go this loads
//     vendor packages via x/tools/go/packages, extracting exported symbols and
//     registration.
//
//   - Step3(ctx context.Context) error:
//     Parses the local source packages. For Go this loads source packages (excluding
//     vendor), extracts functions, types, constants, variables, and caches the AST
//     syntax for the Step4 resolution pass.
//
//   - Step4(ctx context.Context) error:
//     Performs a global resolution pass across all cached local ASTs. For Go this
//     walks every declaration and resolves cross-file and cross-package call graphs,
//     interface implementations, and type references.
//
//   - GetFile() string:
//     Returns the base name of the parsed file (e.g. "go.mod").
//
//   - GetDirectory() string:
//     Returns the parent directory path of the parsed file.
//
//   - GetFullPath() string:
//     Returns the absolute file system path of the parsed file.
//
// All step methods accept a context.Context parameter and must honor context
// cancellation -- they should return as soon as ctx.Done() is signaled.
//
//
// Package ast/astgol
// ===================
//
// The astgol sub-package provides Go-specific AST parsing capabilities. It implements
// the astgen.AST interface and follows the four-step pipeline (Step1 through Step4).
// It identifies Go module files via the Identify function and creates a parser instance
// through GetAST.
//
// The import path is github.com/nabbar/auditor/pkg/ast/astgol. The conventional
// alias is "astgol".
//
// `const Lang = "Go"`:
//   The human-readable name string used by LangGo.String() to return "Go".
//
// `func Identify(file string) bool`:
//   Returns true when the file is a valid go.mod file. Checks performed:
//     - Global flag -lang-go must be true (default).
//     - The parent directory must not contain "vendor".
//     - The file path must not contain "test".
//     - filepath.Base(file) must equal "go.mod" exactly.
//
// `func GetAST(uim auduxi.Manager, mgr *auddbm.Linker, file string) (astgen.AST, error)`:
//   Validates the file via Identify, verifies the Linker manager is non-empty, then
//   constructs and returns a *mdl instance. The handler map is initialized with
//   set/update callback pairs for every AST entry type (Function, Struct, Interface,
//   Primitive, Custom, Constant, Variable).
//
// `func Pattern() []string`:
//   Returns []string{"go.mod"} when Go parsing is enabled. Returns nil when
//   disabled via the -lang-go flag. Note that only the module entry point file is
//   returned here; Go source files (.go) are handled through the x/tools/go/packages
//   loader during Step2/Step3, not via this pattern list.
//
//
// AST Go Dataflow
// ================
//
// The parser follows a four-step descending and iterative approach, going from
// the project root down to individual declarations and their interdependencies.
//
//   1 - Step1 (Module Discovery):
//      Initializes the vendor environment (go mod tidy/download/vendor), reads
//      go.mod, and registers the root module and all vendor module references
//      in the database manager.
//
//   2 - Step2 (Vendor Package Parsing):
//      Loads vendor packages via x/tools/go/packages. Only exported symbols
//      (declarations starting with an uppercase letter) are registered.
//      Internal sub-directories are excluded.
//
//   3 - Step3 (Local Package Parsing):
//      Loads local (non-vendor) Go source packages. For each file the parser
//      caches the AST syntax and processes every declaration through processNode,
//      which routes nodes to the appropriate set<Type> handlers (setFunc,
//      setStruct, setInterface, setPrimitive, setCustom, setConstant, setVariable).
//
//   4 - Step4 (Global Resolution):
//      Walks all cached local ASTs in a second pass, invoking upd<Type> handlers
//      (updFunc, updStruct, updInterface, updPrimitive, updCustom, updConstant,
//      updVariable) to resolve cross-file and cross-package symbol references,
//      call graphs, and interface associations.
//
// AST Go process schematic:
// [Entry] GetAST(go.mod) -> Initialize *mdl
//   |
//   ├──> Step1(ctx)
//   │    ├──> initVendor() -> `go mod tidy`, `download`, `vendor`
//   │    ├──> getFileSrc() -> Read go.mod into memory
//   │    └──> parseGoMod() -> Register main module & vendor module references
//   |
//   ├──> Step2(ctx) [Vendor Packages]
//   │    └──> parsePackage(isVendor = true) -> Load astgop.Load()
//   │         └──> processNode() -> set<Type> (Exported symbols only)
//   |
//   ├──> Step3(ctx) [Local Packages]
//   │    └──> parsePackage(isVendor = false) -> Load astgop.Load()
//   │         ├──> addAST() -> Cache AST syntax in o.s
//   │         └──> processNode() -> set<Type> (All local symbols)
//   |
//   └──> Step4(ctx) [Global Resolution]
//        └──> Loop over cached ASTs in listAST()
//             └──> processNode(isUpdate = true)
//                  ├──> updFunc() -> ast.Inspect(Body) for *ast.CallExpr & *ast.SelectorExpr
//                  │    ├──> [Local Call]  -> getLocDepend()
//                  │    └──> [External Call] -> getDepend() -> resolveRootIdent() / guessPkgName()
//                  └──> updStruct() / updInterface() / updConstant() / updVariable() / updCustom() / updPrimitive()
//
//
// AST Go ignored items
// =====================
//
// The parser is designed for an audit perspective and intentionally ignores many
// valid Go constructs:
//
//   - In vendor directories:
//     Private/unexported symbols: declarations not starting with an uppercase
//     letter (isExportedNode == false) are ignored. Internal directories
//     (vendor/.../internal/...) are filtered during package listing.
//
//   - At the file and directory level:
//     Vendor and test directories are excluded during Identify and getPkgList.
//     Files without declarations (len(afs.Decls) < 1) are skipped.
//     Blank identifiers ("_") for functions, types, constants, or variables are ignored.
//     Anonymous functions / closures inside function bodies are skipped as
//     entry-level declarations.
//
//   - In function bodies and business logic:
//     Control flow and expressions (if, for, switch, assignments, arithmetic) are
//     ignored. Only *ast.CallExpr and *ast.SelectorExpr are inspected for
//     dependency tracing. GoRoot standard library dependencies are skipped during
//     package resolution. Unresolvable imports that cannot be matched to registered
//     packages are skipped.
//
//   - In expression resolution (parseExprCode):
//     If the AST expression type falls into the default case (not handled by the
//     switch), parseExprCode returns (EntryNone, "unknown", false). For unknown
//     types in function parameters or return types this triggers an error and
//     causes the function analysis to fail.
//
//   - In chained selector resolution (resolveRootIdent):
//     If an unhandled expression appears in a chained selector, resolveRootIdent
//     returns nil and the dependency is silently ignored.
//
//   - In generic type parameters:
//     Generic type constraints identified via isTypeParam are skipped during
//     local dependency resolution. The parser supports generic index expressions
//     (IndexExpr, IndexListExpr) in expression resolution.
//
//

package ast
