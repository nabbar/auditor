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
// an optional name, a summary, and the list of the parameters it consumes and
// produces.
//
// The parameters are therefore a tree: the signature of a function is
// described by a list of parameters, and each of them is able to describe its
// own signature. This is what allows the auditor to represent the nested
// structures of the complex signatures.
//
// # Serialization
//
// The package supports the json, yaml, toml and cbor formats.
package params

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	audtps "github.com/nabbar/auditor/pkg/data/types"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// FuncParams represents a function type that creates and returns a new
// Params instance.
//
// This function type is typically used as a factory, to create the instances
// of the Params interface. It allows a flexible instantiation of the
// parameter configurations without a direct dependency on the concrete
// implementation, and it gives the collections the ability to build a
// missing element on demand.
//
// A FuncParams value is invoked without any argument and yields a fully
// initialized Params.
type FuncParams func() Params

// Params represents an interface for handling the parameter configurations.
//
// The Params interface provides a comprehensive set of methods for managing
// the parameter configuration data, including the serialization and the
// deserialization capabilities, the code type information and the summary
// descriptions. It implements multiple standard Go interfaces to support the
// various data formats and the string representations.
//
// This interface enables a consistent handling of the parameter
// configurations across the different parts of the application, while keeping
// the flexibility of the implementation.
type Params interface {
	// Serialization and deserialization interfaces.
	json.Marshaler
	json.Unmarshaler
	yaml.Marshaler
	yaml.Unmarshaler
	toml.Marshaler
	toml.Unmarshaler
	cbor.Marshaler
	cbor.Unmarshaler

	// Stringer interface implementation for string representation.
	fmt.Stringer

	// GetType retrieves the code type and its associated type string.
	//
	// Returns:
	//   - audtps.CodeType: The code type representing the category or the
	//     classification of the parameter configuration.
	//   - string: The type string providing additional contextual information
	//     about the type.
	GetType() (audtps.CodeType, string)

	// GetTypeString generates a formatted string representation of the code
	// type and of its associated type string.
	//
	// If the code type requires a type string, the result is formatted as
	// "CodeType: TypeString", otherwise it is just the CodeType.String()
	// representation. This method provides a standardized way to represent the
	// parameter configurations in a readable format.
	//
	// Returns:
	//   - string: The formatted string representation of the code type and of
	//     the type string.
	GetTypeString() string

	// SetName sets the name of this parameter configuration.
	//
	// Parameters:
	//   - s: The name string to assign to this parameter configuration.
	SetName(s string)

	// GetName retrieves the name of this parameter configuration.
	//
	// Returns:
	//   - string: The current name of the parameter configuration.
	GetName() string

	// SetSummary updates the summary description of this parameter
	// configuration.
	//
	// The summary provides a concise description of what the parameter
	// configuration represents and of its purpose within the system. It is
	// useful for the documentation, for the user interfaces and for the
	// debugging purposes.
	//
	// Parameters:
	//   - string: The summary description to assign.
	SetSummary(string)

	// GetSummary retrieves the summary description of this parameter
	// configuration.
	//
	// Returns:
	//   - string: The current summary string that describes the parameter
	//     configuration.
	GetSummary() string

	// GetInput retrieves the list of the input parameter configurations.
	//
	// The returned slice is a copy of the internal one, so that the caller
	// cannot alter the state of the configuration by modifying it.
	//
	// Returns:
	//   - []Params: A slice of Params instances representing the input
	//     parameters.
	GetInput() []Params

	// SetInput sets the list of the input parameter configurations.
	//
	// The provided parameters are appended to the existing ones, and the nil
	// ones are ignored.
	//
	// Parameters:
	//   - ...Params: A variadic list of Params instances to set as the input
	//     parameters.
	SetInput(...Params)

	// DelInput removes all the input parameter configurations.
	//
	// The internal slice is cleared and released, so the configuration does
	// not hold any reference to the removed parameters anymore.
	DelInput()

	// GetOutput retrieves the list of the output parameter configurations.
	//
	// The returned slice is a copy of the internal one, so that the caller
	// cannot alter the state of the configuration by modifying it.
	//
	// Returns:
	//   - []Params: A slice of Params instances representing the output
	//     parameters.
	GetOutput() []Params

	// SetOutput sets the list of the output parameter configurations.
	//
	// The provided parameters are appended to the existing ones, and the nil
	// ones are ignored.
	//
	// Parameters:
	//   - ...Params: A variadic list of Params instances to set as the output
	//     parameters.
	SetOutput(...Params)

	// DelOutput removes all the output parameter configurations.
	//
	// The internal slice is cleared and released, so the configuration does
	// not hold any reference to the removed parameters anymore.
	DelOutput()
}

// New creates and returns a new Params instance with the specified source
// type and code type.
//
// This constructor function initializes a new parameter configuration with
// the given source type and code type. The summary is initialized as an empty
// string, allowing for a later population of the descriptive information.
//
// Parameters:
//   - srcType: A string representing the source type of the parameter
//     configuration.
//   - srcCode: An audtps.CodeType value that categorizes the parameter
//     configuration.
//
// Returns:
//   - Params: A new instance implementing the Params interface, with the name
//     and the summary initialized to empty strings, and the input and output
//     slices initialized to nil.
func New(srcType string, srcCode audtps.CodeType) Params {
	return &mdl{
		nm:  "",
		tc:  srcCode,
		ts:  srcType,
		sm:  "",
		inp: nil,
		out: nil,
	}
}

// Empty creates and returns a new Params instance with the default and empty
// values.
//
// This constructor function initializes a new parameter configuration with all
// the fields set to their zero or empty values. The code type is set to
// audtps.EntryNone, indicating that no specific entry type is known. The
// name, the type string and the summary are empty strings, and the input and
// the output slices are nil.
//
// It is the counterpart of the New constructor, without any code type, and it
// is used by the deserializers to build the objects they decode into.
//
// Returns:
//   - Params: A new instance implementing the Params interface, with all the
//     fields initialized to their empty or zero values.
func Empty() Params {
	return &mdl{
		nm:  "",
		tc:  audtps.EntryNone,
		ts:  "",
		sm:  "",
		inp: nil,
		out: nil,
	}
}
