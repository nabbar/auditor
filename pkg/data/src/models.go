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

// Package src provides a set of interfaces and utilities for managing source code content in various formats.
//
// This package defines a comprehensive framework for handling source code data, enabling flexible
// encoding and decoding across multiple formats, including JSON, YAML, TOML, and CBOR. It provides
// specialized implementations of the Source interface, optimized for efficient memory usage through
// hexadecimal encoding.
package src

import (
	"encoding/hex"
	"slices"
)

// mdl is the internal implementation of the Source interface.
//
// It stores source content as a hexadecimal-encoded string, which allows binary data
// to be safely transmitted and stored as printable characters while preserving data
// integrity. This encoding strategy also enables efficient serialization and
// deserialization across different data formats.
type mdl struct {
	// d is the hexadecimal-encoded representation of the source content.
	//
	// The raw byte data is encoded into a hex string to provide safe storage and
	// transmission of binary content. Hex encoding ensures that all bytes can be
	// represented as printable ASCII characters, preserving the original data
	// integrity.
	d string
}

// IsEmpty checks whether the source content is empty or invalid.
//
// It attempts to decode the hex-encoded content and returns true if either:
//  1. The decoding fails (indicating corrupted or invalid hex data).
//  2. The decoded content is empty (zero-length byte slice).
//
// Returns:
//   - true if the content is empty, decoding fails, or contains no valid data.
//   - false otherwise.
func (o *mdl) IsEmpty() bool {
	s, e := o.Get()
	return e != nil || len(s) < 1
}

// IsEqual compares this source model with another source model for equality.
//
// It performs a deep comparison between two Source instances by decoding both
// and comparing the resulting byte slices at the byte level. This ensures that
// even if different encoding formats were used, the underlying data is compared
// correctly.
//
// The comparison logic follows these steps:
//  1. If either source is empty (as determined by IsEmpty), returns false.
//  2. Retrieves the decoded content from both sources using Get().
//  3. Compares the byte slices using slices.Compare.
//
// If either source cannot be decoded properly, or if the comparison fails, it
// returns false. This ensures that only valid and comparable data is considered
// for equality testing.
//
// Parameters:
//   - f: another Source instance to compare against.
//
// Returns:
//   - true if both models contain identical content.
//   - false otherwise (including when either source is empty or decoding fails).
func (o *mdl) IsEqual(f Source) bool {
	if o.IsEmpty() || f.IsEmpty() {
		return false
	}

	if d, e := f.Get(); e != nil {
		return false
	} else if t, e := o.Get(); e != nil {
		return false
	} else {
		return slices.Compare(d, t) == 0
	}
}

// Merge merges the content from another source into this source.
//
// It updates the internal state of this source by copying the hex-encoded content
// from another *mdl instance. The merge succeeds only if the provided argument is
// a non-nil *mdl and the content differs from the current source.
//
// Parameters:
//   - a: the source to merge from. Must be of type *mdl; otherwise, the merge fails.
//
// Returns:
//   - true if the merge operation succeeded (the argument is a valid *mdl).
//   - false if the merge failed (the argument is not of type *mdl or is nil).
func (o *mdl) Merge(a any) bool {
	var (
		k bool
		s *mdl
	)

	if s, k = a.(*mdl); !k || s == nil {
		return false
	}

	if s.d != o.d {
		o.d = s.d
	}

	return true
}

// Get retrieves the decoded source content as a byte slice.
//
// It converts the internal hex-encoded string representation back into its original
// binary form using hex.DecodeString. This provides direct access to the raw source
// data for processing or analysis.
//
// Returns:
//   - []byte: the decoded source content (may be empty if the original was empty).
//   - error: any error that occurred during decoding (nil if successful).
func (o *mdl) Get() ([]byte, error) {
	return hex.DecodeString(o.d)
}
