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
// A CodeType (uint8) enumerates the kinds of elements an analysis can produce:
// functions, structures, interfaces, primitive types, custom types, constants,
// variables, and dependencies. Each value has a stable textual representation
// used by the serialized formats, and the package provides the helpers needed
// to convert a CodeType to and from that representation.
//
// CodeType.String() returns the upper-case textual representation
// ("FUNCTION", "STRUCT", ...). All serialization formats produce the lower-case
// version ("function", "struct", ...). The parsing of a code type is case
// insensitive and trimming whitespace around the input is handled automatically.
//
// # Enumeration Values
//
// The following constants define the CodeType enumeration. Their numeric value
// follows the declaration order (iota) and is used by the Get helper:
//
// - EntryNone (0): invalid or unspecified code type. This is the zero value
//   of the type and serves as a sentinel for errors or uninitialized state.
//
// - EntryFunction (1): function or method definition, including both standalone
//   functions and methods bound to structs or interfaces.
//
// - EntryStruct (2): struct definition, a composite data type with typed
//   fields that may also carry methods.
//
// - EntryInterface (3): interface definition, a contract specifying method
//   signatures that implementing types must fulfill.
//
// - EntryPrimitive (4): primitive data type such as int, float, string, or
//   bool. These are the basic building blocks of Go's type system.
//
// - EntryCustom (5): custom or user-defined type, including aliases and type
//   definitions that do not fit the other predefined categories.
//
// - EntryConstant (6): constant declaration, a named value bound at compile
//   time. Requires an additional type string to be fully described.
//
// - EntryVariable (7): variable declaration, a named storage location whose
//   value may change at runtime. Requires an additional type string to be
//   fully described.
//
// - EntryDepend (8): placeholder for a dependency element that lives in
//   another package or module. Used when the referenced element has not yet
//   been parsed, keeping references resolvable.
//
// Some code types require an additional type string to be fully described.
// EntryPrimitive, EntryCustom, EntryConstant, EntryVariable, and EntryDepend
// all need a type string. EntryNone also requires a type string so that a
// malformed entry still exposes the information it may hold. EntryFunction,
// EntryStruct, and EntryInterface are self-describing and do not need a type
// string. Use the NeedString helper method to determine this at runtime.
//
// # Public API
//
// Conversion helpers:
//
// - CodeType.String() returns the upper-case textual representation of the
//   code type, or an empty string for unknown and out-of-range values.
//
// - CodeType.Uint8() returns the underlying numeric value of the CodeType.
//   Useful for storage, serialization, or interfaces expecting a number.
//
// - CodeType.NeedString() returns true when the code type requires an
//   additional type string to be fully described, false otherwise.
//
// Parsing helpers:
//
// - Parse(s string) converts a string descriptor into the corresponding
//   CodeType. The comparison is case insensitive and surrounding whitespace
//   is trimmed. Returns EntryNone for unrecognized input.
//
// - Get(i uint8) converts a numeric index to the corresponding CodeType. If
//   the index does not match a known value, EntryNone is returned instead of
//   an invalid enumeration value.
//
// List helpers:
//
// - EntryTypeList() returns the string representations of the analyzable code
//   types. Includes the seven real code element types but excludes EntryDepend
//   (a placeholder, not an analyzable element) and EntryNone (invalid). Order
//   follows the enumeration declarations.
//
// - List() returns the CodeType values for the analyzable code types. Same
//   composition and order as EntryTypeList, but in numeric form.
//
// # Serialization
//
// A code type is serialized as its lower-case textual representation and can be
// read back either from that representation or from its numeric value.
// The following serialization formats are supported:
//
// - JSON (json.Marshaler / json.Unmarshaler): the code type is serialized as
//   its lower-case textual representation. EntryNone is serialized as the
//   "unknown" marker from pkg/generic. On deserialization, the payload is
//   first read as a string and parsed with Parse. If unrecognized, it is read
//   as a uint8 and converted with Get, so both textual and numeric forms are
//   accepted.
//
// - YAML (yaml.Marshaler / yaml.Unmarshaler): same lower-case textual
//   representation as JSON. EntryNone uses the "unknown" marker. The same
//   two-step deserialization strategy applies (string first, then uint8).
//
// - TOML (MarshalTOML / UnmarshalTOML): same lower-case textual
//   representation. UnmarshalTOML accepts either []byte or string as input.
//   Note: the go-toml/v2 library cannot encode a scalar as the root of a
//   document; CodeType must be encoded as a field of a struct for MarshalTOML
//   to succeed.
//
// - CBOR (MarshalCBOR / UnmarshalCBOR): same lower-case textual
//   representation. The same two-step deserialization strategy applies.
//
// - Binary (MarshalBinary / UnmarshalBinary): implements
//   encoding.BinaryMarshaler / encoding.BinaryUnmarshaler by delegating to
//   MarshalCBOR / UnmarshalCBOR. CBOR is the canonical binary format for
//   CodeType.
//
// # Data Flow
//
// A CodeType flows through the auditor in three stages:
//
// 1. Creation and classification: the analysis phase discovers code elements
//    and assigns each one a CodeType value. The classification is driven by
//    the AST node encountered: function declarations map to EntryFunction,
//    struct literals to EntryStruct, and so on.
//
// 2. Processing and filtering: the code types are used to filter and group
//    elements during the analysis pipeline. The List and EntryTypeList
//    helpers provide the full set of analyzable types. NeedString tells
//    downstream consumers whether they must also store a type string for a
//    given CodeType.
//
// 3. Serialization and persistence: the code types are serialized into
//    JSON, YAML, TOML, CBOR, or binary format for reporting, storage, or
//    inter-process communication. Deserialization on the consumer side
//    accepts both the lower-case textual form and the numeric form, providing
//    backward compatibility across auditor versions.
//
// The flow can be visualized as:
//
//   AST Analysis -> CodeType assignment -> Filter/Group -> Serialization
//
//   Parse(string) or Get(uint8) -> CodeType -> String() or Uint8()
//
//   CodeType -> (Marshal) -> []byte -> (Unmarshal) -> CodeType

package types
