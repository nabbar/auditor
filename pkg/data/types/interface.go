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

import "strings"

// CodeType represents an enumeration of supported code entry types.
//
// Each CodeType corresponds to a specific category of code elements that can be analyzed
// by the auditor system. These types are used throughout the application to categorize
// and process different kinds of code structures, such as functions, structs, interfaces,
// primitives, and custom types.
type CodeType uint8

// CodeType represents an enumeration of supported code entry types.
//
// Each CodeType corresponds to a specific category of code elements that can be analyzed
// by the auditor system. These types are used throughout the application to categorize
// and process different kinds of code structures, such as functions, structs, interfaces,
// primitives, and custom types.
const (
	// EntryNone represents an invalid or unspecified code type.
	//
	// This constant is typically used as a default value when no valid code type
	// has been determined. It serves as a sentinel value to indicate an error state
	// or an uninitialized code type.
	EntryNone CodeType = iota

	// EntryFunction represents a function or method definition.
	//
	// This code type indicates that the analyzed element is a function or method,
	// which may include both standalone functions and methods belonging to structs
	// or interfaces. Functions are typically defined with parameters, return values,
	// and implementation blocks.
	EntryFunction

	// EntryStruct represents a struct definition.
	//
	// This code type indicates that the analyzed element is a struct declaration,
	// which defines a composite data type consisting of fields with specific types.
	// Structs are used to group related data together and can contain methods.
	EntryStruct

	// EntryInterface represents an interface definition.
	//
	// This code type indicates that the analyzed element is an interface declaration,
	// which defines a contract specifying method signatures that implementing types
	// must fulfill. Interfaces are used for defining behavior contracts in Go programs.
	EntryInterface

	// EntryPrimitive represents primitive data types.
	//
	// This code type indicates that the analyzed element is a primitive data type,
	// such as int, float, string, bool, etc. These are basic building blocks of
	// Go's type system and are not composite types.
	EntryPrimitive

	// EntryCustom represents custom or user-defined code types.
	//
	// This code type indicates that the analyzed element is a custom or user-defined
	// type that doesn't fit into the other predefined categories. This could include
	// aliases, type definitions, or other user-specific code constructs.
	EntryCustom

	// EntryDepend type defined a type who referred to a dependency type.
	EntryDepend
)

// Parse normalizes an incoming untrusted string descriptor into a validated operational
// CodeType enumeration instance.
//
// This function provides a safe way to convert string representations of code types
// into their corresponding enumerated values. It performs case-insensitive evaluation
// and trims whitespace to ensure robust parsing of user input or configuration values.
// The function returns EntryNone for unrecognized input strings, making it safe to use
// in contexts where invalid inputs might occur.
//
// Parameters:
//   - s: A string representing the code type name (e.g., "function", "struct", etc.)
//
// Returns:
//   - CodeType: The corresponding enumerated value if valid, otherwise EntryNone
func Parse(s string) CodeType {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case EntryFunction.String():
		return EntryFunction
	case EntryStruct.String():
		return EntryStruct
	case EntryInterface.String():
		return EntryInterface
	case EntryPrimitive.String():
		return EntryPrimitive
	case EntryCustom.String():
		return EntryCustom
	case EntryDepend.String():
		return EntryDepend
	default:
		return EntryNone
	}
}

// Get converts a uint8 value to a CodeType enumeration.
//
// This utility function provides a way to convert numeric values into their corresponding
// CodeType enumerations. It ensures type safety by validating that the input value corresponds
// to a known code type. If an invalid uint8 value is provided, it returns EntryNone as the
// default safe fallback.
//
// Parameters:
//   - i: A uint8 value representing a code type index
//
// Returns:
//   - CodeType: The corresponding enumerated value if valid, otherwise EntryNone
func Get(i uint8) CodeType {
	switch CodeType(i) {
	case EntryFunction:
		return EntryFunction
	case EntryStruct:
		return EntryStruct
	case EntryInterface:
		return EntryInterface
	case EntryPrimitive:
		return EntryPrimitive
	case EntryCustom:
		return EntryCustom
	case EntryDepend:
		return EntryDepend
	default:
		return EntryNone
	}
}

// EntryTypeList returns a slice of all valid code type string representations.
//
// This function provides a complete list of all supported code type names as strings,
// which can be used for validation purposes, user interface display, or configuration
// of allowed code types. The returned slice maintains consistent ordering with the
// enumeration constants defined in this package.
//
// Returns:
//   - []string: A slice containing string representations of all valid code types
func EntryTypeList() []string {
	return []string{
		EntryFunction.String(),
		EntryStruct.String(),
		EntryInterface.String(),
		EntryPrimitive.String(),
		EntryCustom.String(),
	}
}

func List() []CodeType {
	return []CodeType{
		EntryFunction,
		EntryStruct,
		EntryInterface,
		EntryPrimitive,
		EntryCustom,
	}
}
