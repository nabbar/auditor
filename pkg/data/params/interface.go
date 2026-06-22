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

// Package params provides interfaces and utilities for managing parameter configurations.
//
// This package defines the core interface for parameter configuration management
// and implements a concrete type that satisfies this interface. It supports multiple
// serialization formats (JSON, YAML, TOML, CBOR) and provides methods for handling
// code types, type strings, and summary information.
package params

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	audtps "github.com/nabbar/auditor/pkg/data/types"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// FuncParams represents a function type that creates and returns a new Params instance.
//
// This function type is typically used as a factory pattern to create instances of
// the Params interface. It allows for flexible instantiation of parameter configurations
// without direct dependency on concrete implementations.
type FuncParams func() Params

// Params represents an interface for handling parameter configurations.
//
// The Params interface provides a comprehensive set of methods for managing
// parameter configuration data including serialization/deserialization capabilities,
// code type information, and summary descriptions. It implements multiple standard
// Go interfaces to support various data formats and string representations.
//
// This interface enables consistent handling of parameter configurations across
// different parts of the application while maintaining flexibility in implementation.
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
	//   - audtps.CodeType: The code type representing the category or classification
	//     of the parameter configuration.
	//   - string: The type string providing additional contextual information about the type.
	GetType() (audtps.CodeType, string)

	// GetTypeString generates a formatted string representation of the code type and its associated type string.
	//
	// If the code type requires a type string, it formats as "CodeType: TypeString",
	// otherwise it just returns CodeType.String().
	// This method provides a standardized way to represent parameter configurations in a readable format.
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

	// SetSummary updates the summary description of this parameter configuration.
	//
	// The summary provides a concise description of what the parameter configuration represents
	// and its purpose within the system. This is useful for documentation, user interfaces,
	// and debugging purposes.
	//
	// Parameters:
	//   - string: The summary description to assign.
	SetSummary(string)

	// GetSummary retrieves the summary description of this parameter configuration.
	//
	// Returns:
	//   - string: The current summary string that describes the parameter configuration.
	GetSummary() string

	// GetInput retrieves the list of input parameter configurations.
	//
	// Returns:
	//   - []Params: A slice of Params instances representing the input parameters.
	GetInput() []Params

	// SetInput sets the list of input parameter configurations.
	//
	// Parameters:
	//   - ...Params: A variadic list of Params instances to set as input parameters.
	SetInput(...Params)

	// DelInput removes all input parameter configurations.
	DelInput()

	// GetOutput retrieves the list of output parameter configurations.
	//
	// Returns:
	//   - []Params: A slice of Params instances representing the output parameters.
	GetOutput() []Params

	// SetOutput sets the list of output parameter configurations.
	//
	// Parameters:
	//   - ...Params: A variadic list of Params instances to set as output parameters.
	SetOutput(...Params)

	// DelOutput removes all output parameter configurations.
	DelOutput()
}

// New creates and returns a new Params instance with the specified source type and code type.
//
// This constructor function initializes a new parameter configuration with the given
// source type and code type. The summary is initialized as an empty string, allowing
// for later population of descriptive information.
//
// Parameters:
//   - srcType: A string representing the source type of the parameter configuration.
//   - srcCode: An audtps.CodeType value that categorizes the parameter configuration.
//
// Returns:
//   - Params: A new instance implementing the Params interface, with the name and summary
//     initialized to empty strings, and input/output slices initialized to nil.
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

// Empty creates and returns a new Params instance with default/empty values.
//
// This constructor function initializes a new parameter configuration with all fields
// set to their zero or empty values. The code type is set to audtps.EntryNone,
// indicating no specific entry type. The name, type string, and summary are empty
// strings, and input/output slices are nil.
//
// Returns:
//   - Params: A new instance implementing the Params interface, with all fields
//     initialized to empty or zero values.
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
