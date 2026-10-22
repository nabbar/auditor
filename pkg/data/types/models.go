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

// File models.go contains the helper methods of the CodeType type.
//
// They render a code type as text, expose its numeric value, and tell whether
// the type requires an additional type string to be fully described.

package types

// String returns the string representation of the CodeType.
//
// The enumeration value is converted into its corresponding upper case string
// representation. The returned strings match the names of the constants
// defined in this package, which makes them easy to serialize or to display
// in the user interfaces.
//
// The complete list of the recognized representations is:
// "FUNCTION", "STRUCT", "INTERFACE", "PRIMITIVE", "CUSTOM", "CONSTANT",
// "VARIABLE" and "DEPEND".
//
// Returns:
//   - string: The upper case string representation of the code type, or an
//     empty string for the values that match no code type (EntryNone and the
//     out of range values).
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
	case EntryConstant:
		return "CONSTANT"
	case EntryVariable:
		return "VARIABLE"
	case EntryDepend:
		return "DEPEND"
	default:
		return ""
	}
}

// Uint8 returns the underlying uint8 value of the CodeType.
//
// The numeric value is the ordinal position of the code type inside the
// enumeration. It is useful for the serialization, for the storage, or for the
// interfaces that expect a number rather than a named enumeration.
//
// Returns:
//   - uint8: The underlying numeric representation of the CodeType.
func (o CodeType) Uint8() uint8 {
	return uint8(o)
}

// NeedString indicates whether this code type requires an additional string
// parameter to be fully described.
//
// The structural types, that is the functions, the structures and the
// interfaces, are fully described by their code type alone. The other types
// carry a type string that qualifies them: the concrete type of a primitive
// or of a custom type, the type of a constant, the type of a variable or the
// name of the referenced dependency.
//
// The unknown values, which include EntryNone, also require a type string,
// so that a malformed entry still exposes the information it may hold.
//
// Returns:
//   - bool: true when the code type requires an additional type string, false
//     otherwise.
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
	case EntryConstant:
		return true
	case EntryVariable:
		return true
	case EntryDepend:
		return true
	default:
		return true
	}
}
