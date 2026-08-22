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

// Package manager provides the generic interface for AST (Abstract Syntax Tree) parsers
// that are used across different programming languages in the auditor tool.
package manager

import (
	"time"
)

// Deadline returns the time when the context will time out.
//
// This method implements the context.Context interface, allowing the manager to participate
// in context-based cancellation and timeout mechanisms. It provides information about when
// the context is expected to expire, which is useful for implementing timeouts in AST parsing operations.
// When a deadline is set, it allows for graceful termination of long-running parsing tasks
// that might otherwise block indefinitely.
//
// Returns:
//   - deadline: The time when the context will time out (zero value if no deadline)
//   - ok: A boolean indicating whether a deadline exists (true if deadline is set)
//
// Example usage:
//   - deadline, ok := ctx.Deadline()
//   - if ok {
//     // Context has a deadline
//     fmt.Printf("Context expires at %v\n", deadline)
//     }
func (o *mdl) Deadline() (deadline time.Time, ok bool) {
	return o.sem.Deadline()
}

// Done returns a channel that's closed when the context is canceled or times out.
//
// This method implements the context.Context interface, providing support for
// cancellation signals and timeouts throughout the AST parsing process. The returned
// channel can be used to monitor for cancellation events and respond appropriately
// during long-running parsing operations. When the channel is closed, it indicates
// that the context has been canceled or has timed out.
//
// Returns:
//   - A read-only channel that is closed when the context is canceled or times out
//
// Example usage:
//   - select {
//     case <-ctx.Done():
//     // Context was canceled or timed out
//     return
//     default:
//     // Continue with processing
//     }
func (o *mdl) Done() <-chan struct{} {
	return o.sem.Done()
}

// Err returns the error that caused the context to be canceled or timed out.
//
// This method implements the context.Context interface, allowing the manager to
// propagate cancellation errors up the call stack. When a context is canceled or times out,
// this method returns the reason for cancellation, which helps in proper error handling
// during AST processing operations. The returned error can be used to determine why
// a parsing operation was terminated.
//
// Returns:
//   - The error that caused the context to be canceled or timed out, or nil if not canceled
//
// Example usage:
//   - err := ctx.Err()
//   - if err != nil {
//     // Context was canceled or timed out
//     fmt.Printf("Context error: %v\n", err)
//     }
func (o *mdl) Err() error {
	return o.sem.Err()
}

// Value returns the value associated with this key in the context.
//
// This method implements the context.Context interface, enabling the manager
// to store and retrieve contextual information that may be needed during AST processing.
// It allows passing arbitrary values through the context, which can be useful for
// configuration, logging, or other contextual data required by parsing operations.
// The key-value pairs in the context provide a way to pass information down the call stack
// without requiring explicit function parameters.
//
// Parameters:
//   - key: The key to look up in the context (typically a string or struct)
//
// Returns:
//   - The value associated with the key, or nil if no value is found for the given key
//
// Example usage:
//   - value := ctx.Value("user-id")
//   - if userID, ok := value.(string); ok {
//     // Use the user ID for logging or processing
//     fmt.Printf("Processing for user: %s\n", userID)
//     }
func (o *mdl) Value(key any) any {
	return o.sem.Value(key)
}
