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

// Package engine orchestrates the multi-pass structural analysis, asynchronous
// dependency resolution, linguistic LLM reasoning, and final report generation
// for automated source code auditing.
package engine

import (
	"io"

	auddbm "github.com/nabbar/auditor/pkg/data/manager"
	audllm "github.com/nabbar/auditor/pkg/llm"
	auduim "github.com/nabbar/auditor/pkg/uxi"
)

// Options aggregates localized filesystem layout configurations and operational
// override behaviors requested by the executive execution layer. It serves as
// the configuration schema for the engine's lifecycle management.
type Options struct {
	// RepoPath defines the absolute or relative root directory of the workspace
	// to be audited.
	RepoPath string
}

// Engine defines the explicit public pipeline interface facade for controlling
// the execution lifecycle sequence of the automated multi-pass source auditor.
// Implementations are expected to be thread-safe and maintain consistency
// across asynchronous analysis passes.
type Engine interface {
	io.Closer

	// ScanFiles parses local Go source files, resolves complex
	// multi-module layouts, extracts code signatures (AST), and catalogs initial
	// metadata targets into the relational storage engine.
	ScanFiles() error

	// ReOrder will ordered all entries by dependencies to analyze dependencies before entries
	ReOrder() error

	// Analyze processes the cataloged elements using parallel
	// asynchronous worker threads. It leverages specialized LLM inference engines
	// to perform deep semantic analysis and property evaluation on the code.
	Analyze() error

	// Review maps security assertions, cross-verifies structural
	// logic against defined safety policies, and validates cumulative cryptographic
	// or operational vulnerabilities discovered during the analysis phase.
	Review() error

	// Report extracts audited elements from the data matrices, consolidates
	// individual evaluation blocks into a coherent narrative, and persists
	// the unified markdown document to the configured output pathway.
	Report(string) error
}

// New instantiates a concrete thread-safe implementation container for the Engine
// interface. It leverages polymorphic abstractions for database management,
// LLM inference processing, and user interface feedback.
func New(d *auddbm.Linker, l audllm.Manager, u auduim.Manager, o Options) Engine {
	return &eng{
		d: d,
		l: l,
		u: u,
		o: o,
		c: &astcol{},
	}
}
