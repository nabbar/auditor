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

// Package entry provides interfaces and utilities for managing code entry information.
//
// This package defines the Entry interface and related functions for handling
// Go code entry information including packages, names, types, source code, and dependencies.
// It supports various serialization formats (JSON, YAML, TOML, CBOR) and provides
// thread-safe access to entry data through mutex protection.
//
// The Entry interface is designed to represent a single code element within a Go package,
// such as functions, variables, constants, or types. It encapsulates all relevant metadata
// about the code element and its relationships with other elements in the codebase.
package entry

import (
	"fmt"
)

// getSliceStringer converts a slice of fmt.Stringer instances to a slice of their string representations.
//
// This generic helper function takes a slice of any type T that implements the fmt.Stringer interface
// and returns a corresponding slice of string representations. For each element in the input slice,
// the function calls its String() method to generate the textual representation.
//
// The returned slice has the same length as the input slice, with each element at index i
// corresponding to the String() output of the element at index i in the input slice.
//
// This function is particularly useful for generating comparable string arrays from complex
// data structures when performing equality checks, sorting operations, or serialization.
//
// Parameters:
//   - l: A slice of type T where T implements the fmt.Stringer interface.
//
// Returns:
//   - A slice of strings containing the String() representation of each element in the input slice.
//
// Example usage:
//
//	params := []Params{param1, param2, param3}
//	strings := getSliceStringer(params)
//	// Returns []string{"param1", "param2", "param3"}
func getSliceStringer[T fmt.Stringer](l []T) []string {
	var r = make([]string, len(l))

	for i := range l {
		r[i] = l[i].String()
	}

	return r
}
