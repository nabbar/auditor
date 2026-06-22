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

// Package generic defines the AST interface contract used by the auditor tool.
//
// The AST interface is the central abstraction that every language-specific parser
// must implement. It defines a four-step sequential parsing pipeline that is
// identical regardless of the target language, so the auditing pipeline can drive
// any parser uniformly.
//
//
// Four-step Sequential Pipeline
// =============================
//
// The AST interface exposes four step methods — Step1, Step2, Step3, Step4 —
// that must be called in strict order. Each step builds on the data produced by
// the previous step.
//
//   Step1: Parse the root module or library definition.
//
//   Step1 reads the top-level module descriptor (for example a go.mod, a
//   package.json, or an equivalent manifest) and extracts the module name, its
//   declared dependencies, and the overall structure. This step establishes the
//   foundational context that Step2 through Step4 rely on.
//
//   Step1 returns an error for file read failures, malformed module descriptors,
//   or any registration failure. A nil error indicates success.
//
//   Step2: Parse external (third-party) dependencies.
//
//   Step2 discovers and parses every package that the module imports from external
//   sources — whether vendored, linked, or fetched at build time. The purpose of
//   this step is to populate symbol definitions and export information for those
//   external packages so they are available during Step3 and Step4.
//
//   Step2 may only be called after a successful Step1. If the module has no
//   external dependencies, the implementation is expected to return nil.
//
//   Step3: Parse the local source packages.
//
//   Step3 analyzes the source files that belong directly to the module itself,
//   excluding external or third-party packages. It extracts function declarations,
//   type definitions, variable and constant bindings, and internal package
//   structure, storing the results in the auditor's data layer.
//
//   Step3 may only be called after both Step1 and Step2 have completed
//   successfully.
//
//   Step4: Perform global dependency resolution.
//
//   Step4 executes a resolution pass over every declaration and expression parsed
//   in the previous steps. It resolves inter-package symbol references, builds
//   call graphs, links type associations across packages, and establishes
//   cross-module linkages. After Step4 completes successfully, the AST data is
//   considered fully resolved and ready for analysis.
//
//   Step4 may only be called after Step1, Step2, and Step3 have all completed
//   successfully.
//
//
// Resource Cleanup
// ================
//
// The AST interface embeds io.Closer, so every implementation must provide a
// Close method. Implementers are required to release any cached state, buffer
// memory, file handles, or other resources held during parsing. Callers should
// invoke Close once parsing is complete or when an error has terminated the
// pipeline. This prevents memory leaks in long-running auditing sessions.
//
//
// Context Handling
// ================
//
// All step methods accept a context.Context parameter. Implementations must honor
// context cancellation: if ctx.Done() is signaled during any step, the step must
// return as soon as possible to avoid holding resources or hanging the pipeline.
//
//
// File Information Accessors
// ==========================
//
// The interface provides three accessor methods for the file being parsed:
//
//   - GetFile returns the base name of the file (file name with extension,
//     excluding any directory path components).
//   - GetDirectory returns the parent directory path of the file, with the file
//     name omitted.
//   - GetFullPath returns the absolute file system path, combining directory
//     and file name into a canonical absolute path.
//
//
// Usage
// =====
//
// The import path for this package is github.com/nabbar/auditor/pkg/ast/generic.
//
// Every language-specific AST parser in the auditor tool must implement the AST
// interface. The implementation details — lexing, parsing, and structuring the
// AST according to the rules of the target language — are delegated to the
// specific parser. The only contract is that the four steps are called in order
// and that Close is called when the parser is no longer needed.
//

package generic
