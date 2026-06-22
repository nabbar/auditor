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

// Package params models the parameters of a code element: the arguments it
// receives and the values it returns.
//
// A Params instance carries the code type of the parameter, an optional type
// string that qualifies it (the concrete type of a primitive, for example),
// an optional name, a summary, and the list of the parameters it consumes
// and produces.
//
// The parameters form a tree: the signature of a function is described by
// a list of parameters, and each of them can describe its own signature in
// turn. This nesting allows the auditor to represent the complex and nested
// parameter structures of parameterized types.
//
// # Types
//
// The public types are:
//
//  - Params - the interface that encapsulates the state and the behaviour of
//    a parameter set. It embeds json.Marshaler, json.Unmarshaler,
//    yaml.Marshaler, yaml.Unmarshaler, toml.Marshaler, toml.Unmarshaler,
//    cbor.Marshaler, cbor.Unmarshaler, and fmt.Stringer, so that every
//    Params value can be serialized and deserialized in those formats and
//    can produce a string representation via its String() method.
//
//  - FuncParams - a factory function type that creates and returns a new
//    Params instance. It is typically used by collections to lazily build
//    a missing element on demand, without requiring a direct dependency on
//    the concrete implementation. It is called without arguments and returns
//    a fully initialized Params.
//
//  - Export - an intermediate struct used during serialization and
//    deserialization. It exposes six fields: Name (string), Code
//    (audtps.CodeType), Type (string), Sum (string), Input ([]Export),
//    and Output ([]Export). Its field tags ensure consistent serialization
//    across JSON, YAML, TOML, and CBOR formats.
//
// # Constructors
//
// Two constructor functions are provided:
//
//  - New(srcType string, srcCode audtps.CodeType) - creates a new Params
//    instance initialized with the given type string and code type. The
//    name and the summary are set to empty strings, and the input and output
//    slices are initialized to nil.
//
//  - Empty() - creates a new Params instance with all fields at their zero
//    values. The code type is set to audtps.EntryNone, and the type string,
//    the name, and the summary are empty. This constructor is primarily used
//    internally by the deserializers to instantiate a blank object before
//    populating it with decoded data.
//
// # Methods
//
// The Params interface exposes the following methods:
//
//  - GetType() (audtps.CodeType, string) - returns the code type and the
//    associated type string.
//
//  - GetTypeString() string - returns a formatted string representation of
//    the code type. If the code type requires a type string (as determined by
//    audtps.CodeType.NeedString()), the result is "CodeType: TypeString";
//    otherwise it is just CodeType.String().
//
//  - SetName(s string) / GetName() string - sets and gets the optional name
//    of the parameter.
//
//  - SetSummary(string) / GetSummary() string - sets and gets the human-
//    readable description of the parameter.
//
//  - GetInput() []Params - returns a copy of the input parameter list.
//    The caller cannot alter the internal state through the returned slice.
//
//  - SetInput(...Params) - appends the provided Params instances to the
//    existing input list. Nil values are silently ignored. The internal slice
//    capacity is extended automatically when needed.
//
//  - DelInput() - removes and releases all input parameters.
//
//  - GetOutput() []Params - returns a copy of the output parameter list.
//
//  - SetOutput(...Params) - appends the provided Params instances to the
//    existing output list. Nil values are silently ignored.
//
//  - DelOutput() - removes and releases all output parameters.
//
//  - IsEmpty() bool - returns true when the formatted type string is empty,
//    i.e. GetTypeString() returns a string with zero length. This can be
//    used to check whether a parameter was properly initialized.
//
//  - IsEqual(Params) bool - compares this instance with another by their
//    string representations (i.e. their String() values).
//
//  - MarshalJSON() / UnmarshalJSON([]byte) error - serializes and
//    deserializes to/from JSON.
//
//  - MarshalYAML() / UnmarshalYAML(*yaml.Node) error - serializes and
//    deserializes to/from YAML.
//
//  - MarshalTOML() / UnmarshalTOML(interface{}) error - serializes and
//    deserializes to/from TOML. UnmarshalTOML accepts both []byte and
//    string inputs.
//
//  - MarshalCBOR() / UnmarshalCBOR([]byte) error - serializes and
//    deserializes to/from CBOR (Concise Binary Object Representation).
//
//  - MarshalBinary() / UnmarshalBinary([]byte) error - implements the
//    encoding.BinaryMarshaler and encoding.BinaryUnmarshaler interfaces;
//    these delegate to the CBOR marshalers.
//
// # Serialization internals
//
// All serialization formats share a common intermediate representation based
// on the Export struct. When marshaling, the private mdl fields are mapped
// onto an Export value, then that value is fed to the respective encoder.
// Marshalers traverse nested Input and Output slices recursively, so a
// full tree is serialized in a single call.
//
// During deserialization, the raw data is first decoded into an Export
// struct. The values are then applied back to the mdl fields, and nested
// Input/Output entries are reconstructed by creating new mdl instances for
// each of them.
//
// # Data flow
//
// The parameter data follows a three-way pipeline across the package:
//
//  +-----------------------------------+                 +------------+
//  |  Source data (API, AST, etc.)     | --(New/Empty)-> |  mdl       |
//  +-----------------------------------+                 +------+-----+
//                                                        |            |
//                                                        +------+-----+
//                                                               |
//                                                               v
//                                               +-----------------------------+
//                                               |  Params interface (public)  |
//                                               +-----------------------------+
//                                                               |
//                  +------------------------+-------------------+
//                  |                        |                   |
//                  v                        v                   v
//           +--------------+       +-----------------+    +-----------+
//           | marshallObj()|       | unmarshallObj() |    | String()  |
//           | -> *Export   |       | -> *Export      |    | -> string |
//           +-------+------+       +---------+-------+    +-----+-----+
//                   |                        |                  |
//                   v                        v                  v
//    +--------------------------+       +---------+       +-----------+
//    | encoder fmt              |       | Export  |       | display   |
//    | (JSON, YAML, TOML, CBOR) |       | -> mdl  |       | / logging |
//    +--------------------------+       +---------+       +-----------+
//                                                               |
//                                                               |
//                                                               |
//                                                               |
//                                                               |
//                              +------------------+-------------+
//                              |                  |
//                              v                  v
//                          +-------+         +--------+
//                          | Bytes |         | Nested |
//                          +-------+         | params |
//                                            +--------+
//                                                |
//                                                v
//                                           +----------+
//                                           | New mdl  |
//                                           | params   |
//                                           +----------+
//
// In short: the public API builds and manipulates mdl objects through
// the Params interface. For serialization, the internal state is mapped
// to an Export intermediary, then encoded. For deserialization, the
// encoded data is decoded into an Export, then applied back to a new
// mdl instance. Nested parameters are handled recursively in both
// directions.

package params
