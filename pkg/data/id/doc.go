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

// Package id provides the deterministic identifiers used to reference every
// element tracked by the auditor system.
//
// # Overview
//
// Every item discovered during an analysis -- a module, a package, a source
// file, a function, a symbol -- is assigned a unique ID. The ID is a 64-bit
// unsigned integer computed from the element's coordinates (module path,
// package path, name, etc.) using the FNV-1a hash function. FNV-1a is a
// fast, non-allocating hash: it can be called on every element without
// measurable overhead and avoids the garbage collection pressure that would
// come from allocating a hash context per call.
//
// Because the hash is deterministic, the same coordinates always produce the
// same identifier across different processes and different executions of the
// auditor. This property makes IDs usable as map keys, file names, or
// database keys, and it makes the audit reproducible: the same input tree
// always produces the same set of IDs, in the same order.
//
// The FNV-1a hash is not cryptographic. It should never be used to
// authenticate content. Its purpose is purely identification: to give a
// stable, short numeric handle to a domain element that can be stored,
// transmitted, or compared without ambiguity within the auditor ecosystem.
//
// # Type Definition
//
// The ID type is defined as:
//
//     type ID uint64
//
// It wraps a standard uint64 with a named type to carry the semantic
// meaning of an identifier and to prevent accidental arithmetic with raw
// integers. The zero value of ID (i.e. ID(0)) is reserved to signal an
// invalid or unset identifier. Throughout the auditor, a zero-valued ID
// always means "no identifier". Collections and managers rely on this
// convention to detect missing references.
//
// Two accessor methods bridge the ID type with a raw uint64:
//
//   - ID.Uint64() returns the underlying uint64 value for numeric
//     comparison, storage, or integration with external systems.
//
//   - GetID(uint64) converts a raw uint64 back into an ID, completing the
//     round-trip. It is used when an identifier is loaded from serialized
//     data.
//
// # Identifier Generation
//
// The GenID function is the primary way to create an identifier. It accepts
// one or more variadic string arguments and computes a deterministic ID from
// them:
//
//     func GenID(s ...string) ID
//
// Input strings are processed sequentially. The first non-empty string is
// written directly into the FNV-1a hash. Each subsequent non-empty string
// is prefixed with the "::" separator before being written. Empty strings
// are silently skipped. This sequential write model means the order of
// arguments is significant: GenID("a", "b") and GenID("b", "a") produce
// different results.
//
// Because the separator is embedded before each subsequent string, calling
// GenID("a", "b") yields the same hash as GenID("a::b") -- both produce
// the byte sequence "a::b" fed to the hash.
//
// Important properties of GenID:
//
//   - Determinism: the same ordered list of strings always produces the same
//     ID, regardless of the process or run.
//   - Order sensitivity: GenID("a", "b") != GenID("b", "a").
//   - Empty-string filtering: only non-empty strings contribute to the hash.
//   - Collision resistance: FNV-1a has collisions, though they are
//     astronomically unlikely for the volumes handled by the auditor.
//   - Not injective: two different lists may, in principle, collide.
//
// # Data Flow
//
// The processing of identifiers inside the auditor follows this flow:
//
//     raw coordinates (string values)
//         --> GenID(s, ...)
//             --> FNV-1a hash with sequential "::"-separated writes
//         --> ID (uint64)
//             --> stored / indexed / transmitted
//
// In practice, the auditor calls GenID with the coordinates it extracts
// from the Go type information: for a module the input is the module path,
// for a package it is the module path and the package path, for a source
// file it is the module path, the package path, and the file name. The
// ordering of these inputs is part of the ID specification and must remain
// consistent across all callers.
//
// Once an ID is created, it can be used directly as a key in maps, compared
// numerically via Uint64(), or converted via GetID(uint64) after being read
// from any of the supported storage formats.
//
// # Serialization
//
// The ID type implements marshaler and unmarshaler interfaces for five
// serialization formats plus the standard Go binary encoding interfaces.
//
// ## JSON
//
// MarshalJSON serializes the ID as a JSON number (its raw uint64 value).
// UnmarshalJSON expects a JSON number; passing a JSON string will cause a
// type error. The numeric form is chosen for brevity and because it is
// lossless: any uint64 value round-trips correctly.
//
// Example output: 1234567890
//
// ## YAML
//
// MarshalYAML converts the ID to a YAML scalar. The result is a byte
// slice containing the encoded YAML document for the scalar value.
// UnmarshalYAML expects a node whose scalar value is a numeric string that
// can be parsed as a uint64.
//
// ## TOML
//
// MarshalTOML serializes the ID as a TOML integer. UnmarshalTOML accepts
// either a []byte or a string holding a TOML document and parses the
// integer value. Any other input type is rejected with an "invalid type"
// error.
//
// ## CBOR
//
// MarshalCBOR serializes the ID using the fxamacker/cbor library. The
// result is a compact binary representation suitable for network or
// storage use. UnmarshalCBOR decodes the CBOR data back into the ID.
//
// ## Binary
//
// MarshalBinary and UnmarshalBinary implement the standard
// encoding.BinaryMarshaler and encoding.BinaryUnmarshaler interfaces.
// Both delegate to the CBOR implementation, so binary and CBOR produce
// identical byte representations.
//
// # Public API Summary
//
//     type ID uint64                           -- identifier type
//     (o ID) Uint64() uint64                   -- extract raw value
//     func GetID(i uint64) ID                  -- construct from raw value
//     func GenID(s ...string) ID               -- generate from coordinates
//
//     // Serialization interfaces (all on *ID or ID receiver):
//     (o ID) MarshalJSON() ([]byte, error)    // json.Marshaler
//     (*o *ID) UnmarshalJSON(p []byte) error  // json.Unmarshaler
//     (o ID) MarshalYAML() (any, error)       // yaml.Marshaler
//     (*o *ID) UnmarshalYAML(value *yaml.Node) error
//     (o ID) MarshalTOML() ([]byte, error)    // toml.Marshaler
//     (*o *ID) UnmarshalTOML(i any) error     // toml.Unmarshaler
//     (o ID) MarshalCBOR() ([]byte, error)    // cbor.Marshaler
//     (*o *ID) UnmarshalCBOR(p []byte) error  // cbor.Unmarshaler
//     (o ID) MarshalBinary() ([]byte, error)  // encoding.BinaryMarshaler
//     (*o *ID) UnmarshalBinary(p []byte) error // encoding.BinaryUnmarshaler
//
package id
