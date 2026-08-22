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

// Package types provides enumerations and utilities for code type management.
//
// This package defines a set of constants representing different code element types
// that can be analyzed within the auditor system. It includes functions for parsing
// string representations into these enumerated types, as well as utility methods
// for working with these types in a type-safe manner.
package types

// String returns the string representation of the CodeType.
//
// This method converts a CodeType enumeration value into its corresponding uppercase
// string representation. The returned strings match the constants defined in this package,
// making it easy to serialize code type information or display it in user interfaces.
// For invalid CodeType values (e.g., EntryNone), an empty string is returned.
//
// Returns:
//   - string: The uppercase string representation of the code type (FUNCTION, STRUCT, INTERFACE, PRIMITIVE, CUSTOM, DEPEND)
func (o CodeType) String() string {
	switch o {
	case EntryFunction:
		return "FUNCTION"
	case EntryStruct:
		return "STRUCT"
	case EntryInterface:
		return "INTERFACE"
	case EntryPrimitive:
		return "PRIMITIVE"
	case EntryCustom:
		return "CUSTOM"
	case EntryDepend:
		return "DEPEND"
	default:
		return ""
	}
}

// Uint8 returns the underlying uint8 value of the CodeType.
//
// This method provides a way to access the raw numeric representation of the CodeType
// enumeration. It's useful for serialization, storage, or when interfacing with systems
// that expect numeric values rather than named enumerations. The returned value corresponds
// to the ordinal position of the code type within its enumeration.
//
// Returns:
//   - uint8: The underlying numeric representation of the CodeType
func (o CodeType) Uint8() uint8 {
	return uint8(o)
}

// NeedString indicates whether this code type requires an additional string parameter.
//
// This method determines whether a specific code type requires an additional string
// parameter to fully specify its nature. For example, primitive types and custom types
// require a type string to identify their specific subtype, while structural types like
// functions, structs, and interfaces do not need additional information.
//
// Returns:
//   - bool: True if the code type requires an additional string parameter, false otherwise
func (o CodeType) NeedString() bool {
	switch o {
	case EntryFunction:
		return false
	case EntryStruct:
		return false
	case EntryInterface:
		return false
	case EntryPrimitive:
		return true
	case EntryCustom:
		return true
	case EntryDepend:
		return true
	default:
		return true
	}
}
