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

// Package types classifies the code elements handled by the auditor.
//
// A CodeType enumerates the kinds of elements an analysis can produce:
// functions, structures, interfaces, primitive types, custom types, constants,
// variables and dependencies. Each value has a stable textual representation,
// used by the serialized formats, and the package provides the helpers needed
// to convert a CodeType to and from that representation.
//
// The textual representation is the upper case name of the code type
// ("FUNCTION", "STRUCT", ...). The parsing of a code type is case
// insensitive, so a value coming from a configuration file may be written in
// any case.
//
// # Serialization
//
// The package provides the serialization of the CodeType type for the json,
// yaml, toml and cbor formats. A code type is serialized as its lower case
// textual representation, and can be read back either from that
// representation or from its numeric value.
package types

import "strings"

// CodeType represents an enumeration of the supported code entry types.
//
// Each CodeType corresponds to a specific category of the code elements that
// can be analyzed by the auditor. These types are used throughout the
// application to categorize and to process the different kinds of code
// structures, such as the functions, the structures, the interfaces, the
// primitives and the custom types.
//
// The type is an uint8, so it stays small and can be stored as a single byte
// in the serialized documents.
type CodeType uint8

// The enumeration of the supported code entry types.
//
// The constants are declared in a single block because they belong to the
// same enumeration and share the same type. Their numeric value follows the
// declaration order, which is the order used by the Get helper to rebuild a
// code type from its ordinal position.
//
// The first constant, EntryNone, is the default value of a CodeType and the
// zero value of the type.
const (
	// EntryNone represents an invalid or unspecified code type.
	//
	// This constant is typically used as a default value when no valid code
	// type has been determined. It serves as a sentinel value to indicate an
	// error state or an uninitialized code type.
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

	// EntryConstant represents a constant declaration.
	//
	// This code type indicates that the analyzed element is a constant, that
	// is a named value bound at compile time. A constant requires an
	// additional type string to be fully described.
	EntryConstant

	// EntryVariable represents a variable declaration.
	//
	// This code type indicates that the analyzed element is a variable, that
	// is a named storage location whose value may change at run time. A
	// variable requires an additional type string to be fully described.
	EntryVariable

	// EntryDepend represents a type that refers to a dependency.
	//
	// This code type indicates that the analyzed element is not a real code
	// element but a placeholder for an element that lives in another package
	// or in another module. It is used to keep the references resolvable when
	// the referenced element has not been parsed yet.
	EntryDepend
)

// Parse normalizes an incoming untrusted string descriptor into a validated
// operational CodeType enumeration instance.
//
// The comparison is case insensitive and the surrounding spaces are trimmed,
// so that a value coming from a configuration file may be written in any
// case and with any surrounding spacing.
//
// Parameters:
//   - s: A string representing the code type name (for example "function",
//     "struct", ...).
//
// Returns:
//   - CodeType: The corresponding enumerated value when the name is
//     recognized, EntryNone otherwise. EntryNone is the safe fallback used
//     for the invalid inputs.
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
	case EntryConstant.String():
		return EntryConstant
	case EntryVariable.String():
		return EntryVariable
	case EntryDepend.String():
		return EntryDepend
	default:
		return EntryNone
	}
}

// Get converts a uint8 value to a CodeType enumeration.
//
// The conversion is validated: an unknown ordinal position is mapped to
// EntryNone instead of producing an invalid enumeration value.
//
// Parameters:
//   - i: A uint8 value representing a code type index.
//
// Returns:
//   - CodeType: The corresponding enumerated value when the index is known,
//     EntryNone otherwise.
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
	case EntryConstant:
		return EntryConstant
	case EntryVariable:
		return EntryVariable
	case EntryDepend:
		return EntryDepend
	default:
		return EntryNone
	}
}

// EntryTypeList returns a slice of the string representations of the code
// types.
//
// The returned slice can be used for the validation purposes, for the display
// in the user interfaces, or for the configuration of the allowed code types.
// It keeps the same order as the enumeration constants.
//
// The list holds the seven code types that describe a real code element. The
// EntryDepend type is not part of it, because a dependency is a placeholder
// produced by the resolution of a reference and not a code element the
// auditor may be asked to analyze.
//
// Returns:
//   - []string: The string representations of the analyzable code types.
func EntryTypeList() []string {
	return []string{
		EntryFunction.String(),
		EntryStruct.String(),
		EntryInterface.String(),
		EntryPrimitive.String(),
		EntryCustom.String(),
		EntryConstant.String(),
		EntryVariable.String(),
	}
}

// List returns a slice of the analyzable code types.
//
// The returned slice is the numeric counterpart of EntryTypeList and keeps
// the same order. It is useful to iterate over the code types, for example to
// build the filter of an analysis.
//
// The EntryDepend type is not part of the list, for the same reason as in
// EntryTypeList.
//
// Returns:
//   - []CodeType: The analyzable code types.
func List() []CodeType {
	return []CodeType{
		EntryFunction,
		EntryStruct,
		EntryInterface,
		EntryPrimitive,
		EntryCustom,
		EntryConstant,
		EntryVariable,
	}
}
