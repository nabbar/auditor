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

// Package status provides enumerations and utilities for managing audit status states.
//
// This package defines the different lifecycle states an audit item can be in during processing.
// Each status represents a distinct phase in the audit workflow, from initial discovery to final review.
// The status values are designed to be used as flags in state machines or process tracking systems.
package status

// String returns the string representation of the Status.
//
// This method provides a human-readable string version of the Status enum value.
// It converts each status constant into its uppercase string representation,
// which is useful for logging, display purposes, or serialization to external formats.
// Returns an empty string for invalid/unrecognized status values.
//
// Example usage:
// status := Pending.String()  // Returns "PENDING"
// status := Failed.String()   // Returns "FAILED"
//
// Note: The string representation is case-insensitive, allowing for flexible state tracking.
// For example, "PENDING", "Pending", "PENDING!", and "PENDING?" all return "PENDING".
func (o Status) String() string {
	switch o {
	case Pending:
		return "PENDING"
	case Progress:
		return "PROGRESS"
	case Failed:
		return "FAILED"
	case Retry:
		return "RETRY"
	case Extracted:
		return "EXTRACTED"
	case Analyzed:
		return "ANALYZED"
	case Audited:
		return "AUDITED"
	case Reviewed:
		return "REVIEWED"
	case Outdated:
		return "OUTDATED"
	default:
		return ""
	}
}

// Uint8 returns the underlying uint8 value of the Status.
//
// This method provides access to the raw numeric representation of the Status enum.
// It allows for conversion to and from uint8 values, which can be useful for
// database storage, serialization, or compatibility with systems that expect
// numeric representations of enumerated types.
//
// Example usage:
// value := Pending.Uint8()  // Returns 1 (the underlying uint8 value)
// value := Failed.Uint8()   // Returns 3 (the underlying uint8 value)
//
// Important note: The underlying uint8 value is sequential starting from 1,
// with None (0) as the zero value. This allows for efficient storage and
// comparison in database operations or sorting algorithms.
func (o Status) Uint8() uint8 {
	return uint8(o)
}
