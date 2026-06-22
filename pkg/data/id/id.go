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
// element handled by the auditor.
//
// An ID is a 64 bit unsigned integer computed from the coordinates of the
// element it designates (its module path, its package path, its name, ...) with
// the FNV-1a hash function. The result is stable: the same coordinates always
// produce the same identifier, in the same process and in any other process,
// which makes the identifiers usable as map keys, as file names or as database
// keys, and makes the analysis of a project reproducible.
//
// The hash is not a cryptographic one: it is a fast and non allocating hash,
// designed to be computed on every element discovered by an analysis. The
// identifiers are therefore not a substitute for a cryptographic digest when
// the content of an element has to be authenticated.
//
// # Serialization
//
// The package also provides the serialization of the ID type for the json,
// yaml, toml and cbor formats, plus the binary encoding interfaces, so that an
// identifier can be stored in any of the supported formats and read back
// without any loss.
package id

import "hash/fnv"

// ID represents a 64-bit unsigned integer identifier.
//
// The type is designed to be lightweight, efficient and deterministic for the
// identification of the code elements. ID is the primary identifier type of
// the auditor system: it guarantees consistent references across the runs of
// an analysis.
//
// The zero value is reserved for the invalid or unset identifiers. The
// collections and the managers of the auditor rely on that convention to
// detect the missing references.
type ID uint64

// Uint64 returns the underlying 64-bit unsigned integer value of the ID.
//
// This method allows the direct numeric manipulation of the identifier,
// enabling the arithmetic operations and the comparisons with the other
// uint64 values for the integration with the external systems.
//
// Returns:
//   - uint64: The raw 64-bit unsigned integer representation of this ID.
func (o ID) Uint64() uint64 {
	return uint64(o)
}

// GetID converts a uint64 value to an ID type.
//
// This function ensures the type safety by converting the numeric value to the
// ID type, allowing a seamless integration with the systems that work with
// uint64 values directly. The conversion preserves the numeric value while
// keeping the deterministic properties of the ID type.
//
// It is the counterpart of the Uint64 method, and is used when an identifier
// is read back from a serialized document.
//
// Parameters:
//   - i: The uint64 value to convert to an ID type.
//
// Returns:
//   - ID: An ID type built from the provided uint64 value.
func GetID(i uint64) ID {
	return ID(i)
}

// GenID generates a deterministic identifier from one or more string inputs.
//
// This function computes a consistent hash value from the input strings with
// the FNV-1a hash algorithm, which ensures reproducible results across the
// different runs while keeping the identifiers distinct for the different
// input combinations. The identifiers are therefore usable to reference the
// elements of an analysis from one run to another.
//
// The empty strings are ignored by the hashing, so they never influence the
// final identifier. A call without any argument, or with only empty strings,
// therefore returns the hash of an empty content.
//
// The input strings are joined with the "::" separator before being hashed.
// The separator makes the concatenation unambiguous for the common cases, but
// it does not make the function order independent: the hash function is
// applied sequentially, so the order of the arguments is significant.
//
// Parameters:
//   - s ...string: The input strings used to generate the unique identifier.
//
// Returns:
//   - ID: A deterministic identifier computed from the input strings.
//
// Notes:
//   - The same ordered list of strings always produces the same identifier,
//     whatever the process and whatever the run.
//   - Two different ordered lists of strings are highly likely to produce two
//     different identifiers, but the hash is not injective: a collision remains
//     possible, although it is extremely unlikely for the volumes handled by
//     the auditor.
//   - The order of the arguments matters: GenID("a", "b") and GenID("b", "a")
//     produce two different identifiers.
//   - GenID("a", "b") and GenID("a::b") produce the same identifier, because
//     the separator is inserted before the hashing.
func GenID(s ...string) ID {
	var (
		k bool
		h = fnv.New64a()
	)

	// FNV-1a Write never returns an error, the error is safely ignored
	// because the hash.Hash contract guarantees that it is always nil.
	for i := range s {
		if len(s[i]) < 1 {
			continue
		}

		if k {
			// Join the existing hash with "::" + new string to maintain a
			// consistent hashing order.
			_, _ = h.Write([]byte("::" + s[i]))
		} else {
			// Initial hash calculation for the first non empty string of the
			// list: it is written without any separator.
			_, _ = h.Write([]byte(s[i]))
			k = true
		}
	}

	return ID(h.Sum64())
}
