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

// Package pkg/ast/astgol:
// ========================
//
// The astgol sub-package provides Go-specific AST parsing capabilities for the auditor tool.
//
// This package implements the AST interface for Go language files, specifically handling:
//   - Go source file identification and Go module file (go.mod) detection
//   - Module-level parsing from go.mod to extract project-level metadata and vendor dependencies
//   - Package-level parsing using golang.org/x/tools/go/packages to load full ASTs
//   - Complete symbol extraction (functions, structs, interfaces, primitives, custom types, constants, variables)
//   - Source code content retrieval with thread-safe caching mechanisms
//   - Multi-step parsing pipeline with global cross-file dependency resolution
//
// The import path for this package is "github.com/nabbar/auditor/pkg/ast/astgol" and the alias is "astgol".
//
// `const Lang = "Go"`
//   - Defines the human-readable string representation of the Go programming language used across the auditor tool.
//
// `func Identify(file string) bool`:
//   - Determines whether a given file is a valid Go module definition file (go.mod).
//   - Excludes vendor and test directories and honors the runtime flag -lang-go.
//
// `func GetAST(uim auduxi.Manager, mgr *auddbm.Linker, file string) (astgen.AST, error)`:
//   - Creates and initializes a new AST module parser instance (*mdl) for the specified go.mod file.
//   - Registers all set/update handlers for supported code entry types.
//
// `func Pattern() []string`:
//   - Returns the standard file pattern slice (e.g., []string{"go.mod"}) recognized for Go project entry points.
//   - Returns nil if Go language parsing is disabled via runtime flag.
//
// `func (o *mdl) Close() error`:
//   - Implements io.Closer to clear file/source caches and reset internal parser state to avoid memory leaks.
//
// `func (o *mdl) Step1(ctx context.Context) error`:
//   - Prepares vendor dependencies (`go mod tidy/download/vendor`), reads go.mod, registers the root module
//     and all vendor modules in the database manager.
//
// `func (o *mdl) Step2(ctx context.Context) error`:
//   - Loads and parses vendor packages contained in the module using x/tools/go/packages.
//
// `func (o *mdl) Step3(ctx context.Context) error`:
//   - Loads and parses local (non-vendor) packages, extracting declarations and caching ASTs for Step 4.
//
// `func (o *mdl) Step4(ctx context.Context) error`:
//   - Executes the second pass across all cached local ASTs to perform cross-file and cross-package dependency resolution.
//
// `func (o *mdl) GetFile() string`:
//   - Returns the base filename of the go.mod file being parsed.
//
// `func (o *mdl) GetDirectory() string`:
//   - Returns the working directory path where the Go module resides.
//
// `func (o *mdl) GetFullPath() string`:
//   - Returns the full absolute path of the go.mod file being parsed.
//
// AST Golang Dataflow :
// The processing of a Go module by your parser follows a multi-phase descending and iterative approach:
//   1 - Initialization & Module Discovery (Step1): Runs vendor preparation commands, reads go.mod, registers main and dependency modules.
//   2 - Vendor Package Parsing (Step2): Uses x/tools/go/packages to discover and register exported symbols from vendor packages.
//   3 - Local Package Parsing (Step3): Loads local Go packages, extracts exported/internal declarations (functions, types, vars, consts), and caches ASTs.
//   4 - Declaration Processing (processNode): Routes AST nodes to dedicated set<Type> functions to extract source code, types, and signature params.
//   5 - Update & Local Dependency Analysis (upd<Type>): Updates existing entries with parameter types, inputs, outputs, and direct dependencies.
//   6 - Global Resolution (Step4): Walks all cached local AST declarations to resolve global cross-file calls, interfaces, and symbol references.
//
// AST Golang process schematic:
// [Entry] GetAST(go.mod) -> Initialize *mdl
//   │
//   ├──> Step1(ctx)
//   │    ├──> initVendor() -> `go mod tidy`, `download`, `vendor`
//   │    ├──> getFileSrc() -> Read go.mod into memory
//   │    └──> parseGoMod() -> Register main module & vendor module references
//   │
//   ├──> Step2(ctx) [Vendor Packages]
//   │    └──> parsePackage(isVendor = true) -> Load astgop.Load()
//   │         └──> processNode() -> set<Type> (Exported symbols only)
//   │
//   ├──> Step3(ctx) [Local Packages]
//   │    └──> parsePackage(isVendor = false) -> Load astgop.Load()
//   │         ├──> addAST() -> Cache AST syntax in o.s
//   │         └──> processNode() -> set<Type> (All local symbols)
//   │
//   └──> Step4(ctx) [Global Resolution]
//        └──> Loop over cached ASTs in listAST()
//             └──> processNode(isUpdate = true)
//                  ├──> updFunc() -> ast.Inspect(Body) for *ast.CallExpr & *ast.SelectorExpr
//                  │    ├──> [Local Call]  -> getLocDepend()
//                  │    └──> [External Call] -> getDepend() -> resolveRootIdent() / guessPkgName()
//                  └──> updStruct() / updInterface() / updConstant() / updVariable() / updCustom() / updPrimitive()
//
// AST golang ignored items :
// The parser is designed with a specific audit perspective. Therefore, it excludes certain items in Go code:
//   - In vendor directories:
//      - Private/Unexported symbols: Any declaration not starting with an uppercase letter (isExportedNode == false) is ignored.
//      - Internal directories: Packages under vendor/.../internal/... are filtered out during package listing.
//
//   - At the file & directory level:
//      - Vendor & Test directories: Excluded during file identification (Identify) and package discovery (getPkgList).
//      - Files without declarations: If len(afs.Decls) < 1, the file is skipped.
//      - Blank identifiers (_): Functions, types, constants, or variables named '_' are ignored.
//      - Anonymous functions / Closures: Inline closures inside function bodies are skipped as entry declarations.
//
//   - In function bodies & business logic:
//      - Control flow & expressions: Statements like if, for, switch, assignments, and arithmetic operations are ignored.
//        Only symbol references (*ast.CallExpr, *ast.SelectorExpr, type references) are inspected for dependency tracing.
//      - GoRoot Standard Library: Dependencies pointing to Goroot (e.g., "fmt", "net/http") are skipped during package resolution.
//      - Unresolvable imports: Imports that cannot be matched to registered packages are skipped.
//
//   - In generic type parameters:
//      - Generic type constraints (e.g., [T any]) are identified via isTypeParam and skipped during local dependency resolution.
//

package astgol
