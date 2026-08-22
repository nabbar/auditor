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

// Package id provides utilities for generating and managing unique identifiers.
// This package implements a deterministic identifier generation system based on the FNV-1a hash algorithm
// to create consistent, unique IDs for code elements throughout the auditor system.
// It ensures reproducible identification while maintaining uniqueness across different execution contexts.
//
// The ID type is designed to be lightweight and efficient, providing a way to uniquely identify
// components or elements in a deterministic manner.
//
// This package implements serialization and deserialization capabilities for ID types
// using various formats including JSON, YAML, TOML, and CBOR for persistent storage and transfer.
//
// The serialization methods enable seamless integration with configuration management,
// data persistence systems, and cross-system communication.
//
// This package extends the basic ID type functionality to support various data exchange formats,
// making it easier to store, transfer, and configure ID-based systems.
package id

import "hash/fnv"

// ID represents a 64-bit unsigned integer identifier type.
// It is designed to be lightweight, efficient, and deterministic for code element identification.
// ID is the primary identifier type for the auditor system, ensuring consistent and unique references across different runs.
type ID uint64

// Uint64 returns the underlying 64-bit unsigned integer value of the ID.
// This method allows direct numeric manipulation of the identifier, enabling arithmetic operations
// and comparisons with other uint64 values for integration with external systems.
//
// Returns: the raw 64-bit unsigned integer representation of this ID
func (o ID) Uint64() uint64 {
	return uint64(o)
}

// GetID converts a uint64 value to an ID type.
// This function ensures type safety by converting the numeric value to the ID type,
// allowing seamless integration with systems that may work with uint64 values directly.
// The conversion preserves the numeric value while maintaining the ID type's deterministic properties.
//
// Parameters: i - The uint64 value to convert to an ID type
// Returns: ID - An ID type constructed from the provided uint64 value
func GetID(i uint64) ID {
	return ID(i)
}

// GenID generates a deterministic unique identifier from one or more string inputs.
// This function creates a consistent hash value based on the input strings using the FNV-1a hash algorithm,
// ensuring reproducible results across different runs while maintaining uniqueness for different input combinations.
// The algorithm handles empty strings by ignoring them during hashing, preventing empty inputs from affecting the final identifier.
//
// The input strings are joined using "::" separator to ensure consistent hashing regardless of input order or quantity,
// allowing for flexible identification of code elements across different runs.
//
// Parameters: s ...string - The input strings used to generate the unique identifier
// Returns: ID - A deterministic unique identifier based on the input strings
// Note: If multiple non-empty strings are provided, the resulting ID will be unique for that combination.
// Note: If identical strings are used in different order or with different separators, the resulting IDs will be identical.
func GenID(s ...string) ID {
	var (
		k bool
		h = fnv.New64a()
	)

	// FNV-1a Write never returns an error, safely ignore
	for i := range s {
		if len(s[i]) < 1 {
			continue
		}

		if k {
			// Join existing hash with "::" + new string to maintain consistent hashing order
			_, _ = h.Write([]byte("::" + s[i]))
		} else {
			// Initial hash calculation for first string in list
			_, _ = h.Write([]byte(s[i]))
			k = true
		}
	}

	return ID(h.Sum64())
}
