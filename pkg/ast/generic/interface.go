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

// Package generic provides the standard interface for AST (Abstract Syntax Tree)
// parsers used across multiple programming languages in the auditor tool.
//
// Every language-specific AST parser implementation must satisfy the AST interface
// contract so that the central auditing pipeline can drive parsing uniformly,
// regardless of the target language. The interface is organized around a four-step
// sequential pipeline:
//
//	Step1: Parse the root module definition, extracting its name, dependencies,
//	        and high-level structure. This establishes the context for all
//	        subsequent steps.
//
//	Step2: Parse external (third-party or vendored) package structures that the
//	        module depends on. Symbol definitions and export information from
//	        those packages become available to the local source parsing step.
//
//	Step3: Parse the local source packages of the module itself, extracting
//	        types, functions, variables, constants, and internal structure.
//
//	Step4: Perform a global resolution pass across all parsed AST data, resolving
//	        inter-package symbol references, call graphs, type associations, and
//	        cross-module linkages.
//
// The interface also embeds io.Closer so that every parser implementation must
// provide resource cleanup when parsing is no longer needed.
package generic

import (
	"context"
	"io"
)

// AST defines the contract for an Abstract Syntax Tree parser that can be used
// by the auditor tool regardless of the source language. Each implementation
// is responsible for lexing, parsing, and structuring the AST according to the
// rules of its target language, while providing a uniform four-step pipeline
// and resource cleanup.
//
// The four steps must be called sequentially in order: Step1, Step2, Step3,
// Step4. Step1 establishes the module context. Step2 parses external
// dependencies. Step3 parses the local source. Step4 resolves all cross-references.
//
// All methods that accept a context.Context parameter must honor context
// cancellation and return as soon as ctx.Done() is signaled, to avoid holding
// resources or hanging the pipeline.
//
// Resource cleanup: the embedded io.Closer contract requires implementations
// to release any cached state, buffer memory, file handles, or other resources
// held during parsing when Close is called. This is essential to prevent
// memory leaks in long-running auditing sessions.
type AST interface {

	// Closer implements io.Closer for resource cleanup.
	//
	// Implementations must release any state accumulated during parsing,
	// including cached data, buffer memory, and file handles. Callers should
	// invoke Close once parsing is complete or when an error has terminated
	// the pipeline.
	io.Closer

	// Step1 parses the root module/library definition.
	//
	// Step1 reads the top-level module descriptor and extracts the module name,
	// declared dependencies, and overall structure. This step establishes the
	// foundational context that Step2 through Step4 rely on; it must be called
	// first and is expected to run to completion before proceeding.
	//
	// A non-nil error may be returned for file read failures, malformed
	// module descriptors, or registration failures. Implementers must return
	// a descriptive error that identifies the root cause.
	Step1(ctx context.Context) error

	// Step2 parses external (third-party) dependencies.
	//
	// Step2 discovers and parses every package that the module imports from
	// external sources -- whether vendored, linked, or fetched at build time.
	// This step populates symbol definitions and export information for those
	// external packages, making them available for resolution during Step3
	// and Step4.
	//
	// Step2 may be called only after a successful Step1. If the module has
	// no external dependencies, the implementation is expected to return nil.
	Step2(ctx context.Context) error

	// Step3 parses the local source packages.
	//
	// Step3 analyzes the source code files that belong directly to the
	// module itself (excluding external or third-party packages). It extracts
	// function declarations, type definitions, variable and constant
	// bindings, and internal package structure, storing the results in the
	// auditor's data layer.
	//
	// Step3 may be called only after both Step1 and Step2 have completed
	// successfully.
	Step3(ctx context.Context) error

	// Step4 performs global dependency resolution across all parsed AST data.
	//
	// Step4 executes a resolution pass over every declaration and expression
	// parsed in the previous steps. It resolves inter-package symbol
	// references, builds call graphs, links type associations across
	// packages, and establishes cross-module linkages. After Step4 completes
	// successfully, the AST data is considered fully resolved and ready for
	// analysis.
	//
	// Step4 may be called only after Step1, Step2, and Step3 have all
	// completed successfully.
	Step4(ctx context.Context) error

	// GetFile returns the base name of the file being parsed.
	//
	// The returned string contains only the file name with its extension,
	// excluding any directory path components.
	GetFile() string

	// GetDirectory returns the directory path containing the file being parsed.
	//
	// The returned string is the path to the parent directory of the source
	// file, with the file name omitted.
	GetDirectory() string

	// GetFullPath returns the absolute file system path of the file being parsed.
	//
	// The returned string combines the directory path and file name into a
	// canonical absolute path, suitable for file system lookups or debug
	// output.
	GetFullPath() string
}
