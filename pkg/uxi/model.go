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

// Package uxi provides a unified interface and structural monitoring orchestrator
// that combines thread-safe structured diagnostics logging with concurrent worker progress tracking.
// It serves as a facade layer to simplify interactions between logging subsystems and
// semaphore-based concurrency control mechanisms.
//
// The package offers a Manager interface that encapsulates logging capabilities, progress
// visualization, and concurrent execution management. It supports both standard logging
// operations and advanced progress tracking for multithreaded tasks.
package uxi

import (
	"fmt"
	"strings"

	liblog "github.com/nabbar/golib/logger"
	libsem "github.com/nabbar/golib/semaphore"
	semtps "github.com/nabbar/golib/semaphore/types"
)

const taskLen = 120

// mdl serves as the private structural capsule backing the global Manager facade.
// It encapsulates the logger and semaphore components, ensuring that all log routing
// and concurrency state synchronization are abstracted away from the public API.
type mdl struct {
	log liblog.Logger
	sem libsem.Semaphore
}

// Close performs a graceful cleanup of the structural resources. It synchronizes
// terminal worker states via the semaphore controller and ensures all log entries
// are flushed/closed before returning.
func (o *mdl) Close() error {
	_ = o.log.Close()
	o.sem.DeferMain()
	return nil
}

// NewBar maps a localized tracking state engine line. It interfaces with the
// semaphore controller to return a visual progress bar (semtps.SemBar) initialized
// with the specified task name and total item count.
func (o *mdl) NewBar(taskName string, totalItems int) semtps.SemBar {
	if l := len(taskName); l < taskLen {
		taskName = fmt.Sprintf("%s%s", taskName, strings.Repeat(" ", taskLen-l))
	} else if l > 120 {
		taskName = taskName[l-(taskLen):]
		taskName = "..." + taskName[3:]
	}

	return o.sem.BarNumber(taskName, "", int64(totalItems), false, nil)
}
