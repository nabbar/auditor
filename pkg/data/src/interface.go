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

// Package src provides interfaces and utilities for managing source code content.
//
// This package defines the core interface for handling source code data in various formats,
// including JSON, YAML, TOML, and CBOR. It also provides utilities for encoding and decoding
// source content with support for hex-encoded storage to optimize memory usage.
package src

import (
	"encoding/hex"
	"encoding/json"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// Source represents an interface for handling source code content.
//
// The Source interface provides a unified way to work with different source formats
// while maintaining consistency across various data serialization methods. It supports
// multiple encoding formats including JSON, YAML, TOML, and CBOR through embedded interfaces.
//
// This abstraction allows for seamless conversion between formats while preserving
// the original content integrity and providing efficient handling of encoded data.
type Source interface {
	json.Marshaler
	json.Unmarshaler
	yaml.Marshaler
	yaml.Unmarshaler
	toml.Marshaler
	toml.Unmarshaler
	cbor.Marshaler
	cbor.Unmarshaler

	// IsEmpty checks if the source content is empty or invalid.
	//
	// This method determines whether the source contains valid data or if it's in an empty state.
	// It returns true when the content is either empty, corrupted, or cannot be decoded properly.
	// This check is essential for validating source integrity before processing.
	//
	// Returns:
	//   - true if the source content is empty or decoding fails
	//   - false otherwise
	IsEmpty() bool

	// IsEqual compares this source with another source for equality.
	//
	// This method performs a deep comparison between two Source instances to determine
	// if they contain identical content. It handles the comparison at the decoded byte level,
	// ensuring that even if different formats are used, the underlying data is compared correctly.
	//
	// The comparison considers both the encoded representation and the actual decoded content.
	// If either source is empty (as determined by IsEmpty), it returns false to prevent
	// comparing invalid or incomplete data.
	//
	// Parameters:
	//   - other: another Source instance to compare against
	//
	// Returns:
	//   - true if both sources contain identical content
	//   - false otherwise
	IsEqual(Source) bool

	// Merge merges the content from another source into this source.
	//
	// This method updates the internal state of this source by incorporating content from
	// another source. It returns true if the merge operation was successful, indicating
	// that the source content has been updated with new information.
	//
	// Parameters:
	//   - a: another Source instance to merge from
	//
	// Returns:
	//   - true if the merge operation succeeded
	//   - false if the merge failed (source is not of type *mdl or is empty)
	Merge(any) bool

	// Get retrieves the decoded source content as a byte slice.
	//
	// This method returns the actual decoded content of the source, converting from
	// its internal hex-encoded representation to raw bytes. It provides access to the
	// original data regardless of how it was encoded or stored internally.
	//
	// The returned byte slice contains the fully decoded content that can be used
	// for processing, analysis, or further serialization as needed.
	//
	// Returns:
	//   - []byte: the decoded source content
	//   - error: any error that occurred during decoding (nil if successful)
	Get() ([]byte, error)
}

// New creates and returns a new Source instance from raw byte data.
//
// This constructor initializes a new Source with the provided byte data. The input data
// is hex-encoded internally to optimize memory usage and provide efficient storage
// of binary content. The hex encoding ensures that the binary data can be safely stored
// and transmitted while maintaining its integrity.
//
// Parameters:
//   - data: raw byte data to be wrapped in a Source instance
//
// Returns:
//   - Source: a new Source instance containing the encoded data
//
// Example usage:
//
//	source := New([]byte("Hello, World!"))
func New(data []byte) Source {
	return &mdl{
		d: hex.EncodeToString(data),
	}
}

// Empty returns an empty Source instance.
//
// This function provides a convenient way to create an empty Source instance that can be
// used as a placeholder or for testing purposes. The returned instance contains an empty
// hex-encoded string representation, indicating that it does not contain any valid source content.
func Empty() Source {
	return &mdl{
		d: "",
	}
}
