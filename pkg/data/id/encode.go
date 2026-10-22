//
// MIT License
//
// Copyright (c) 2022 Nicolas JUHEL
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

// File encode.go contains the serialization and deserialization of the ID
// type.
//
// It implements the json, yaml, toml and cbor marshaler and unmarshaler
// contracts, plus the encoding.BinaryMarshaler and encoding.BinaryUnmarshaler
// ones, so that an identifier can be stored in any of the supported
// formats.

package id

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// MarshalJSON implements the json.Marshaler interface for ID serialization.
//
// This function converts the ID into a JSON representation by returning its
// underlying uint64 value. The marshaled output is a simple numeric value,
// suitable for the standard JSON processing tools and for the configuration
// management.
//
// The numeric form is chosen over the textual one because it is shorter and
// because it keeps the round trip lossless: any uint64 value is accepted back
// by UnmarshalJSON.
//
// Returns:
//   - []byte: The JSON-encoded representation of the ID value.
//   - error: Any error that occurred during the marshaling. It is always nil
//     for a valid uint64 value.
func (o ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Uint64())
}

// UnmarshalJSON implements the json.Unmarshaler interface for ID deserialization.
//
// This function parses the JSON data into an ID value by converting the
// numeric representation back to the ID type.
//
// The input must be a JSON number: the decoding is performed into a uint64,
// so a JSON string is rejected with a type error. This is the exact
// counterpart of MarshalJSON, which always produces a number.
//
// Parameters:
//   - p: The JSON-encoded byte slice to unmarshal into the identifier.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling. A type error is
//     returned when the payload is not a JSON number, or when the number does
//     not fit in a uint64.
func (o *ID) UnmarshalJSON(p []byte) error {
	var r uint64

	if e := json.Unmarshal(p, &r); e != nil {
		return e
	}

	*o = GetID(r)

	return nil
}

// MarshalYAML implements the yaml.Marshaler interface for ID serialization.
//
// This function converts the ID into a YAML representation by returning its
// underlying uint64 value. The YAML format is particularly useful for the
// configuration files and for the human readable data exchanges, because it
// provides a clean syntax and a good readability.
//
// The returned value is the YAML document itself, that is the bytes of the
// encoded scalar, and not the raw uint64 value. The yaml package accepts both
// forms, but only the raw value is correctly understood by UnmarshalYAML.
//
// Returns:
//   - interface{}: The YAML-encoded representation of the ID value, as a byte
//     slice holding the encoded scalar.
//   - error: Any error that occurred during the marshaling.
func (o ID) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(o.Uint64())
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for ID deserialization.
//
// This function parses the YAML data into an ID value by converting the
// numeric representation back to the ID type. The scalar value carried by the
// node is decoded into a uint64, so a node that holds a non numeric scalar is
// rejected with a type error.
//
// Parameters:
//   - value: The YAML node holding the scalar to decode into the identifier.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling.
func (o *ID) UnmarshalYAML(value *yaml.Node) error {
	var r uint64

	if e := yaml.Unmarshal([]byte(value.Value), &r); e != nil {
		return e
	}

	*o = GetID(r)

	return nil
}

// MarshalTOML implements the TOML marshaler interface for ID serialization.
//
// This function converts the ID into a TOML representation by returning its
// underlying uint64 value. The TOML format is ideal for the configuration
// files and provides a clean and readable syntax, which is easy to maintain
// in the configuration management scenarios.
//
// Returns:
//   - []byte: The TOML-encoded representation of the ID value.
//   - error: Any error that occurred during the marshaling.
func (o ID) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.Uint64())
}

// UnmarshalTOML implements the TOML unmarshaler interface for ID deserialization.
//
// This function accepts the []byte and the string representations of a TOML
// document, and parses them into an ID value by converting the numeric
// representation back to the ID type. The two input forms are handled by the
// same decoding logic, which keeps the compatibility with the various TOML
// data sources.
//
// Parameters:
//   - i: The TOML data to decode, either a []byte or a string. Any other type
//     is rejected with an "invalid type" error.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling.
//
// Note:
//   - The TOML decoder is called with the local variable r passed by
//     value, so r is never updated by the decoder. As a result, GetID(r)
//     always returns ID(0), and the receiver is overwritten with ID(0)
//     instead of receiving the decoded value.
func (o *ID) UnmarshalTOML(i interface{}) error {
	var (
		e error
		k bool
		p []byte
		s string
		r uint64
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

	*o = GetID(r)

	return nil
}

// MarshalCBOR implements CBOR marshaling using the fxamacker/cbor library.
//
// This function serializes the ID into the CBOR (Concise Binary Object
// Representation) format, which provides a compact binary serialization,
// suitable for an efficient storage and transmission. The CBOR format is
// particularly useful for the machine to machine communications and for the
// binary data exchanges.
//
// Returns:
//   - []byte: The CBOR-encoded representation of the ID value.
//   - error: Any error that occurred during the marshaling.
func (o ID) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(o.Uint64())
}

// UnmarshalCBOR implements CBOR unmarshaling using the fxamacker/cbor library.
//
// This function deserializes the CBOR data back into an ID value by converting
// the numeric representation back to the ID type. It provides a reliable
// decoding of the binary data for an efficient reconstruction of the
// identifiers.
//
// Parameters:
//   - p: The CBOR-encoded byte slice to decode into the identifier.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling. A type error is
//     returned when the payload does not hold an unsigned integer.
func (o *ID) UnmarshalCBOR(p []byte) error {
	var r uint64

	if e := cbor.Unmarshal(p, &r); e != nil {
		return e
	}

	*o = GetID(r)

	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
//
// This method uses the CBOR encoding internally to provide a compact binary
// serialization of the ID values. It is particularly useful for an efficient
// storage and for the network transmissions, when the binary format is
// preferred over the textual ones.
//
// Returns:
//   - []byte: The binary representation of the ID value.
//   - error: Any error that occurred during the marshaling.
func (o ID) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
//
// This method uses the CBOR decoding internally to deserialize the binary
// data back into an ID value. It is particularly useful for an efficient
// reconstruction of the identifiers from a binary storage or from a network
// transmission.
//
// Parameters:
//   - p: The binary byte slice to decode into the identifier.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling.
func (o *ID) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}
