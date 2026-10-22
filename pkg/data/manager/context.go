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

// File context.go contains the delegation of the context.Context contract to
// the embedded user interface manager.
//
// It lets a manager be used as a context, so that the cancellation and the
// deadlines of an analysis are propagated to every layer that holds a
// reference to the manager.

package manager

import (
	"time"
)

// Deadline returns the time when the context will time out.
//
// This method implements the context.Context interface, so that the manager
// can participate in cancellation and timeout mechanisms. It returns the time
// at which the context is expected to expire, which is useful for implementing
// timeouts for long-running operations.
//
// If no deadline is set, the zero time value is returned with ok set to
// false.
//
// Returns:
//   - deadline: The time when the context will time out, or the zero time
// when no deadline is set.
//   - ok: true when a deadline exists, false otherwise.
//
//	deadline, ok := ctx.Deadline()
//	if ok {
//	    // The context has a deadline
//	    fmt.Printf("The context expires at %v\n", deadline)
//	}
func (o *mdl) Deadline() (deadline time.Time, ok bool) {
	return o.sem.Deadline()
}

// Done returns a channel that is closed when the context is
// canceled or times out.
//
// This method implements the context.Context interface and provides support
// for cancellation signals and timeouts. The returned channel can be used to
// monitor cancellation events and to react accordingly during long-running
// operations. When the channel is closed, the context has been canceled or
// it has timed out.
//
// Returns:
//   - A read-only channel that is closed when the context is
// canceled or times out.
//
// Example usage:
//
//	select {
//	case <-ctx.Done():
//	    // The context was canceled or it timed out
//	    return
//	default:
//	    // The processing continues
//	}
func (o *mdl) Done() <-chan struct{} {
	return o.sem.Done()
}

// Err returns the error that caused the context to be canceled or to
// time out.
//
// This method implements the context.Context interface and allows the manager
// to propagate cancellation errors up the call stack. When the context is
// canceled or times out, this method returns the reason for the cancellation,
// which helps callers handle the error appropriately. If the context has not
// been canceled, nil is returned.
//
// Returns:
//   - The error that caused the context to be canceled or to time out, or
// nil when the context has not been canceled.
//
// Example usage:
//
//	err := ctx.Err()
//	if err != nil {
//	    // The context was canceled or it timed out
//	    fmt.Printf("The context error is: %v\n", err)
//	}
func (o *mdl) Err() error {
	return o.sem.Err()
}

// Value returns the value associated with the given key from the context.
//
// This method implements the context.Context interface and enables the manager
// to store and retrieve contextual information that may be needed during
// processing. It allows arbitrary values to be passed through the context,
// which is useful for configuration, logging, or any other contextual data.
//
// Parameters:
//   - key: The key to look up in the context. Keys are typically string
// or struct values.
//
// Returns:
//   - The value associated with the key, or nil when no value is found
// for the given key.
//
// Example usage:
//
//	value := ctx.Value("user-id")
//	if userID, ok := value.(string); ok {
//	    // The user identifier is used for the logging or for the processing
//	    fmt.Printf("The processing is done for the user: %s\n", userID)
//	}
func (o *mdl) Value(key any) any {
	return o.sem.Value(key)
}
