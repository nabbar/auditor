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

import "time"

// Deadline returns the time when the context will time out, if any.
// This implementation delegates to the underlying semaphore's deadline method.
func (o *mdl) Deadline() (deadline time.Time, ok bool) {
	return o.sem.Deadline()
}

// Done returns a channel that's closed when the context is canceled or times out.
// This implementation delegates to the underlying semaphore's done channel.
func (o *mdl) Done() <-chan struct{} {
	return o.sem.Done()
}

// Err returns the first accumulated error from the context, if any.
// This implementation delegates to the underlying semaphore's error method.
func (o *mdl) Err() error {
	return o.sem.Err()
}

// Value returns the value associated with this key in the context, if any.
// This implementation delegates to the underlying semaphore's value method.
func (o *mdl) Value(key any) any {
	return o.sem.Value(key)
}
