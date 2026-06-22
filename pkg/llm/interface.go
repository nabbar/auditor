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

// Package llm provides an isolated, production-grade local engine interface for
// orchestrating multi-pass source code audits, structural cataloging, and automated
// security reviews leveraging high-performance local Language Models (LLMs) via Ollama.
//
// The package acts as a high-level abstraction layer, decoupling the business logic
// of code auditing from the underlying implementation details of prompt engineering,
// HTTP transport protocols, and JSON response unmarshaling.
package llm

import (
	"context"
	"fmt"
	"sync"

	audids "github.com/nabbar/auditor/pkg/data/id"
	auddbm "github.com/nabbar/auditor/pkg/data/manager"
	llmcfg "github.com/nabbar/auditor/pkg/llmconfig"
	audloc "github.com/nabbar/auditor/pkg/local"
	auduim "github.com/nabbar/auditor/pkg/uxi"
)

// Manager defines the contract for LLM-driven operations on auditor
// targets. Each method receives a context for cancellation/timeout and
// the target's unique identifier.
type Manager interface {
	// Analyze performs a targeted LLM-based analysis on the code element
	// identified by the given ID. It retrieves the entry from the database,
	// loads the appropriate prompt template based on the code type and
	// language, renders the prompt with the entry data, and sends it to
	// the LLM for processing.
	//
	// Parameters:
	//   - context.Context: the context for cancellation and timeout control.
	//   - audids.ID: the unique identifier of the code element to analyze.
	//
	// Returns an error if the analysis fails. The error is also logged
	// via the UI manager.
	//
	// The method performs the following steps:
	//  1. Retrieves the entry from the database.
	//  2. Determines the code type and language.
	//  3. Loads the prompt template for the given code type and language.
	//  4. Renders the prompt with the entry data.
	//  5. Sends the rendered prompt to the LLM via the call method.
	Analyze(context.Context, audids.ID) error

	// Report generates an LLM-based report for the target identified by the
	// given ID. Currently not implemented.
	//
	// Parameters:
	//   - context.Context: the context for cancellation/timeout.
	//   - audids.ID: the auditor entry ID to generate the report for.
	//
	// Returns an error if the operation fails. Currently panics as the
	// implementation is pending.
	Report(context.Context, audids.ID) error

	// Package executes an LLM-based packaging operation on the target
	// identified by the given ID. Currently not implemented.
	//
	// Parameters:
	//   - context.Context: the context for cancellation/timeout.
	//   - audids.ID: the auditor package ID to generate the report for.
	//
	// Returns an error if the operation fails. Currently panics as the
	// implementation is pending.
	Package(context.Context, audids.ID) error

	// Module executes an LLM-based module operation on the target identified
	// by the given ID. Currently not implemented.
	//
	// Parameters:
	//   - context.Context: the context for cancellation/timeout.
	//   - audids.ID: the auditor module ID to generate the report for.
	//
	// Returns an error if the operation fails. Currently panics as the
	// implementation is pending.
	Module(context.Context, audids.ID) error
}

// New creates and returns a new Manager implementation backed by the
// provided LLM configuration, UI manager, local manager, and database
// linker.
//
// Parameters:
//   - cfg: the LLM configuration containing endpoint URL, model name,
//     and timeout settings.
//   - uim: the UI manager used to display progress or errors to the user;
//     must not be nil.
//   - loc: the local manager used to access local resources; must not be nil.
//   - dbm: the database linker used to persist or retrieve data; must not be nil.
//
// Returns a Manager instance on success, or an error if any of the
// required dependencies (uim, loc, dbm) is nil.
func New(cfg llmcfg.Config, uim auduim.Manager, loc audloc.Manager, dbm *auddbm.Linker) (Manager, error) {
	if uim == nil {
		return nil, fmt.Errorf("UI manager is empty")
	}

	if loc == nil {
		return nil, fmt.Errorf("local manager is empty")
	}

	if dbm == nil {
		return nil, fmt.Errorf("DB manager is empty")
	}

	if cfg.Default.Host == nil {
		return nil, fmt.Errorf("llm config is empty")
	}

	return &mdl{
		cfg: cfg,
		uim: uim,
		loc: loc,
		dbm: dbm,
		lim: &lmt{
			mux: sync.Mutex{},
			lst: make(map[string]*mrq),
		},
	}, nil
}
