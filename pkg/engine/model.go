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

	audids "github.com/nabbar/auditor/pkg/data/id"
	auddbm "github.com/nabbar/auditor/pkg/data/manager"
	audllm "github.com/nabbar/auditor/pkg/llm"
	auduim "github.com/nabbar/auditor/pkg/uxi"
)

// eng acts as the central orchestrator structure, holding references to the
// database manager, LLM inference engine, user interface feedback layer,
// and execution options. It implements the Engine interface.
type eng struct {
	// d provides access to the database manager for storing and retrieving
	// audit data, entries, and structural information about the codebase.
	d *auddbm.Linker

	// l handles communication with Large Language Models (LLMs) for
	// linguistic reasoning and analysis tasks during the audit process.
	l audllm.Manager

	// u manages user interface feedback and logging operations,
	// providing real-time updates and status information to users during execution.
	u auduim.Manager

	// o contains configuration parameters that define how the engine should
	// operate, including repository paths, analysis settings, and operational modes.
	o Options

	// i stores the ordered list of entry IDs generated during the ReOrder
	// operation. This slice represents the processing sequence for audit analysis,
	// ensuring dependencies are analyzed before their dependents.
	i []audids.ID

	c ColAst
}

func (o *eng) Close() error {
	for _, l := range o.c.GetLang() {
		for _, a := range o.c.GetAst(l) {
			if c, k := a.(io.Closer); k && c != nil {
				_ = c.Close()
			}
		}
	}

	return nil
}
