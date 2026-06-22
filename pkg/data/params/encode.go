/***********************************************************************************************************************
 *
 *   MIT License
 *
 *   Copyright (c) 2022 Nicolas JUHEL
 *
 *   Permission is hereby granted, free of charge, to any person obtaining a copy
 *   of this software and associated documentation files (the "Software"), to deal
 *   in the Software without restriction, including without limitation the rights
 *   to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *   copies of the Software, and to permit persons to whom the Software is
 *   furnished to do so, subject to the following conditions:
 *
 *   The above copyright notice and this permission notice shall be included in all
 *   copies or substantial portions of the Software.
 *
 *   THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *   IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *   FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *   AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *   LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *   OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *   SOFTWARE.
 *
 *
 **********************************************************************************************************************/

// Package params provides encoding and decoding utilities for parameter configurations.
//
// This package implements various serialization interfaces (JSON, YAML, TOML, CBOR)
// for the parameter configuration model. It handles marshaling and unmarshaling of
// parameter configurations to different formats while maintaining consistency
// with the underlying data structure.
package params

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	audtps "github.com/nabbar/auditor/pkg/data/types"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// Export represents a structured export object for serialization of parameter configurations.
//
// This struct defines the fields that will be serialized across different formats
// (JSON, YAML, TOML, CBOR). It provides a consistent interface for serialization
// and deserialization operations.
//
// Fields:
//   - Name: The name of the parameter configuration.
//   - Code: The code type (audtps.CodeType) representing the category or classification.
//   - Type: The type string providing additional contextual information.
//   - Sum: The summary or description of the parameter configuration.
//   - Input: A slice of nested Export objects representing input parameter configurations.
//   - Output: A slice of nested Export objects representing output parameter configurations.
type Export struct {
	// Name represents the name of the parameter configuration.
	Name string `json:"name" yaml:"name" toml:"name"`

	// Code represents the code type associated with this parameter configuration.
	Code audtps.CodeType `json:"code" yaml:"code" toml:"code"`

	// Type represents the specific type string for the code type.
	Type string `json:"type" yaml:"type" toml:"type"`

	// Sum represents the summary or description of this parameter configuration.
	Sum string `json:"summary" yaml:"summary" toml:"summary"`

	// Input represents the list of input parameter configurations.
	Input []Export `json:"input" yaml:"input" toml:"input"`

	// Output represents the list of output parameter configurations.
	Output []Export `json:"output" yaml:"output" toml:"output"`
}

// MarshalJSON implements the json.Marshaler interface.
//
// This method serializes the parameter configuration into JSON format. The underlying
// data structure is marshaled using a helper function that creates a structured export
// with code type, type string, and summary fields. This ensures consistent JSON output
// regardless of how the parameter configuration was initialized.
//
// The implementation follows Go's standard marshaler pattern by returning serialized
// bytes and any error that occurred during serialization.
//
// Returns:
//   - []byte: The serialized JSON representation of the parameter configuration.
//   - error: Any error encountered during serialization, or nil on success.
func (o *mdl) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.marshallObj())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// This method deserializes JSON data into a parameter configuration. It first unmarshals
// the JSON into an intermediate structure, then applies the values to the receiver's fields.
// The method handles both string representations and numeric values that may be present
// in the JSON input, ensuring compatibility with various JSON formats.
//
// Error handling is performed by propagating any issues encountered during the unmarshaling process.
//
// Parameters:
//   - p: The JSON byte slice to deserialize.
//
// Returns:
//   - error: Any error encountered during deserialization, or nil on success.
func (o *mdl) UnmarshalJSON(p []byte) error {
	var r = o.unmarshallObj()

	if e := json.Unmarshal(p, &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalYAML implements the yaml.Marshaler interface.
//
// This method serializes the parameter configuration into YAML format. It uses the
// standard YAML marshaler with the helper structure to produce well-formatted YAML output.
// The serialized data maintains the same structure as JSON but is formatted for better
// readability in YAML files, making it suitable for configuration management.
//
// Returns:
//   - interface{}: The serialized YAML representation of the parameter configuration.
//   - error: Any error encountered during serialization, or nil on success.
func (o *mdl) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(o.marshallObj())
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
//
// This method deserializes YAML data into a parameter configuration. It accepts
// string representations from YAML files and parses them appropriately. The implementation
// uses the intermediate structure approach to ensure compatibility with various YAML formats.
//
// Error handling ensures that any parsing issues are properly reported to the caller.
//
// Parameters:
//   - value: The YAML node to deserialize.
//
// Returns:
//   - error: Any error encountered during deserialization, or nil on success.
func (o *mdl) UnmarshalYAML(value *yaml.Node) error {
	var r = o.unmarshallObj()

	if e := yaml.Unmarshal([]byte(value.Value), &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalTOML implements the TOML marshaler interface.
//
// This method serializes the parameter configuration into TOML format. It leverages
// the go-toml library to produce structured TOML output that maintains consistency
// with other supported formats. The serialized data is suitable for use in TOML-based
// configuration files and provides clear, readable output.
//
// Returns:
//   - []byte: The serialized TOML representation of the parameter configuration.
//   - error: Any error encountered during serialization, or nil on success.
func (o *mdl) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.marshallObj())
}

// UnmarshalTOML implements the TOML unmarshaler interface.
//
// This method deserializes TOML data into a parameter configuration. It accepts
// both []byte and string representations from TOML files, parsing them appropriately.
// The implementation handles different input types and ensures proper error handling
// for invalid or malformed TOML data.
//
// Returns an error if the input type is not supported or if parsing fails.
//
// Parameters:
//   - i: The TOML data to deserialize, which can be either a []byte or string.
//
// Returns:
//   - error: Any error encountered during deserialization, or nil on success.
func (o *mdl) UnmarshalTOML(i interface{}) error {
	var (
		e error
		k bool
		p []byte
		s string
		r = o.unmarshallObj()
	)

	if p, k = i.([]byte); k {
		e = toml.Unmarshal(p, r)
	}

	if s, k = i.(string); k {
		e = toml.Unmarshal([]byte(s), r)
	}

	if !k {
		return fmt.Errorf("invalid type: %T", i)
	}

	if e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalCBOR implements CBOR marshaling using the fxamacker/cbor library.
//
// This method serializes the parameter configuration into CBOR (Concise Binary Object Representation)
// format. CBOR is a binary serialization format that is more efficient than JSON for machine-to-machine
// communication. The method uses the fxamacker/cbor library to perform the encoding, producing compact
// binary data while maintaining human readability when decoded.
//
// See also: github.com/fxamacker/cbor
//
// Returns:
//   - []byte: The serialized CBOR representation of the parameter configuration.
//   - error: Any error encountered during serialization, or nil on success.
func (o *mdl) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(o.marshallObj())
}

// UnmarshalCBOR implements CBOR unmarshaling using the fxamacker/cbor library.
//
// This method deserializes CBOR data into a parameter configuration. It decodes binary
// CBOR data that contains a string representation of a parameter configuration and parses
// it into the appropriate fields. The implementation leverages the fxamacker/cbor library
// for robust decoding capabilities.
//
// See also: github.com/fxamacker/cbor
//
// Parameters:
//   - p: The CBOR byte slice to deserialize.
//
// Returns:
//   - error: Any error encountered during deserialization, or nil on success.
func (o *mdl) UnmarshalCBOR(p []byte) error {
	var r = o.unmarshallObj()

	if e := cbor.Unmarshal(p, &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
//
// This method provides binary serialization of parameter configurations using CBOR encoding.
// It serves as a bridge between the standard Go binary marshaler interface and the CBOR
// implementation. The method delegates to MarshalCBOR for actual serialization, ensuring
// consistent behavior across different serialization formats.
//
// Example usage:
//
//	size := size.ParseUint64(1048576) // 1 MB
//	data, _ := size.MarshalBinary()
//	// data can be stored or transmitted in binary format
//
// Returns:
//   - []byte: The serialized binary representation of the parameter configuration.
//   - error: Any error encountered during serialization, or nil on success.
func (o *mdl) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
//
// This method provides binary deserialization of parameter configurations using CBOR decoding.
// It serves as a bridge between the standard Go binary unmarshaler interface and the CBOR
// implementation. The method delegates to UnmarshalCBOR for actual deserialization, ensuring
// consistent behavior across different serialization formats.
//
// Example usage:
//
//	var size size.Size
//	size.UnmarshalBinary(data)
//	fmt.Println(size.MegaBytes()) // Output: 1
//
// Parameters:
//   - p: The binary byte slice to deserialize.
//
// Returns:
//   - error: Any error encountered during deserialization, or nil on success.
func (o *mdl) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}

// marshallObj creates a structured export object for serialization.
//
// This helper function returns a structured representation of the parameter configuration
// that can be used by various marshalers. The Export struct defines the fields that will
// be serialized, including code type, type string, and summary information. It provides
// a consistent interface across different serialization formats (JSON, YAML, TOML, CBOR).
//
// The function recursively processes nested input and output parameter configurations,
// ensuring that all levels of the hierarchy are properly serialized.
//
// Returns:
//   - *Export: A pointer to an Export struct representing the serialized parameter configuration.
func (o *mdl) marshallObj() *Export {
	res := &Export{
		Name:   o.nm,
		Code:   o.tc,
		Type:   o.ts,
		Sum:    o.sm,
		Input:  make([]Export, 0, len(o.inp)),
		Output: make([]Export, 0, len(o.out)),
	}

	if len(o.inp) > 0 {
		for i := 0; i < len(o.inp); i++ {
			if o.inp[i] == nil {
				continue
			}

			m, k := o.inp[i].(*mdl)

			if !k {
				continue
			}

			if r := m.marshallObj(); r != nil {
				res.Input = append(res.Input, *r)
			}
		}
	}

	if len(o.out) > 0 {
		for i := 0; i < len(o.out); i++ {
			if o.out[i] == nil {
				continue
			}

			m, k := o.out[i].(*mdl)

			if !k {
				continue
			}

			if r := m.marshallObj(); r != nil {
				res.Output = append(res.Output, *r)
			}
		}
	}

	return res
}

// unmarshallObj creates an intermediate structure for deserialization.
//
// This helper function returns an empty structured representation that will be used
// to receive data during the unmarshaling process. It provides a consistent interface
// for various unmarshalers to populate with data from different formats.
//
// Returns:
//   - *Export: A pointer to an empty Export struct for receiving deserialized data.
func (o *mdl) unmarshallObj() *Export {
	return &Export{}
}

// unmarshallApply applies deserialized data to the receiver's fields.
//
// This helper function takes the intermediate structure populated during unmarshaling
// and applies its values to the actual fields of the parameter configuration. It handles
// type assertion and validation to ensure that the deserialized data is correctly mapped
// to the receiver's fields.
//
// The function recursively processes nested input and output parameter configurations,
// creating new mdl instances for each nested element and applying their deserialized data.
//
// Parameters:
//   - v: The Export struct containing deserialized data to apply.
//
// Returns:
//   - error: Any error encountered during the application of deserialized data, or nil on success.
//     Returns an error if the input is nil.
func (o *mdl) unmarshallApply(v *Export) error {
	if v == nil {
		return errors.New("invalid object")
	}

	o.nm = v.Name
	o.tc = v.Code
	o.ts = v.Type
	o.sm = v.Sum
	o.inp = make([]Params, 0, len(v.Input))
	o.out = make([]Params, 0, len(v.Output))

	if len(v.Input) > 0 {
		for i := 0; i < len(v.Input); i++ {
			var p = &mdl{
				nm:  "",
				tc:  0,
				ts:  "",
				sm:  "",
				inp: nil,
				out: nil,
			}

			if e := p.unmarshallApply(&v.Input[i]); e != nil {
				return e
			}

			o.inp = append(o.inp, p)
		}
	}

	if len(v.Output) > 0 {
		for i := 0; i < len(v.Output); i++ {
			var p = &mdl{
				nm:  "",
				tc:  0,
				ts:  "",
				sm:  "",
				inp: nil,
				out: nil,
			}

			if e := p.unmarshallApply(&v.Output[i]); e != nil {
				return e
			}

			o.out = append(o.out, p)
		}
	}

	return nil
}
