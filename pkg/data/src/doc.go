//
// MIT License
//
// Copyright (c) 2026 Nicolas JUHEL
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the \\"Software\\"), to deal
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

// Package src provides the model and interface for source code content handled
// by the auditor.
//
// A Source wraps a source file's raw byte content and stores it as a
// hexadecimal-encoded string internally. This encoding ensures that any binary
// data remains safe to store and to exchange through text-based formats such as
// JSON, YAML, TOML, or CBOR, without risk of corruption or encoding conflicts.
// Each byte is represented as two ASCII hex characters, so the internal string
// has a length double the original byte count. This trade increases memory
// usage in exchange for safe, format-agnostic storage.
//
// # Public API
//
// The package exposes the following public elements:
//
//   - Source: the interface that defines the contract for source content. It
//     embeds the standard marshaler/unmarshaler interfaces for JSON, YAML,
//     TOML, and CBOR, and declares four domain methods: IsEmpty, IsEqual,
//     Merge, and Get.
//
//   - New(data []byte): the constructor that creates a Source from raw bytes.
//     The input is hex-encoded and stored internally. The caller may pass any
//     sequence of bytes, including NUL bytes or non-UTF-8 data.
//
//   - Empty(): a convenience constructor that returns a Source pre-initialized
//     with empty content (an empty hex string).
//
// # Source Interface
//
// The Source interface declares four domain behaviors:
//
//   - Get() ([]byte, error): returns the raw decoded content. The internal
//     hex string is decoded back to bytes using hex.DecodeString. A non-nil
//     error indicates that the stored hex data is malformed and cannot be
//     decoded.
//
//   - IsEmpty() bool: returns true if the source holds no data or if decoding
//     the hex string fails. It delegates to Get internally and considers the
//     source empty when the decoded byte slice is zero-length or when an error
//     is returned.
//
//   - IsEqual(Source) bool: compares this source with another by decoding both
//     and performing a byte-level comparison via slices.Compare. If either side
//     is empty or cannot be decoded, false is returned.
//
//   - Merge(any) bool: copies the hex-encoded content from another Source into
//     the receiver, but only when the source is of the internal concrete type
//     and differs from the current content. A merge with identical content is a
//     no-op. Returns true on success, false if the argument is not a valid
//     source or is nil.
//
// # Serialization
//
// The internal model implements the marshaler and unmarshaler contracts for
// multiple formats. The serialized representation is always the hex-encoded
// string, so that the original content can be reconstructed faithfully.
//
// The supported formats are:
//
//   - JSON: the hex string is serialized as a JSON string value and can be
//     unmarshaled back from the same form.
//
//   - YAML: the hex string is serialized as a YAML string scalar and can be
//     unmarshaled back from the same form.
//
//   - CBOR: the hex string is serialized as a CBOR text string and can be
//     unmarshaled back. Binary marshaling via encoding.BinaryMarshaler also
//     delegates to CBOR internally, providing a compact binary representation
//     for efficient storage or network transmission.
//
//   - TOML: MarshalTOML returns an error because a TOML document must be a
//     table, not a bare string. UnmarshalTOML is also broken: the type-switch
//     implementation writes into a non-pointer target and fails for both
//     byte-slice and string inputs. These limitations are inherent to the
//     current implementation and TOML's structural requirements.
//
// # Data Flow
//
// The source content flows through the package as follows:
//
//   1. Ingestion: New(data []byte) receives raw bytes and encodes them to a
//      hex string via hex.EncodeToString. The resulting string is the sole
//      persistent representation inside the model.
//
//   2. Storage and exchange: the hex string can be serialized to any supported
//      format. Because it is a plain ASCII string, it can be embedded safely
//      in any text-based interchange format or database.
//
//   3. Retrieval: Get() decodes the hex string back to raw bytes via
//      hex.DecodeString, returning the original content. The caller is
//      responsible for handling any decode error.
//
//   4. Comparison: IsEqual(Source) decodes both sides and compares the byte
//      slices directly. IsEmpty() checks whether decoding fails or yields zero
//      bytes.
//
//   5. Merge: Merge(any) overwrites the receiver's hex string with the source's
//      hex string when the source is of the expected concrete type and the
//      content differs.
//
// # Example
//
//	import "github.com/nabbar/auditor/pkg/data/src"
//
//	// Create a source from raw bytes
//	source := src.New([]byte("Hello, World!"))
//
//	// Retrieve the decoded content
//	data, err := source.Get()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	_ = data // []byte{72, 101, 108, 108, 111, ...}
//
//	// Serialize to JSON
//	encoded, err := source.MarshalJSON()
//	_ = encoded // []byte(`...`)
//
//	// Compare with another source
//	source2 := src.New([]byte("Hello, World!"))
//	_ = source.IsEqual(source2) // true
//
//	// Check for empty source
//	_ = src.Empty().IsEmpty() // true
//
//	// Merge sources
//	source.Merge(source2) // true

package src
