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

// File encode.go contains the serialization and deserialization of the
// Status type.
//
// It implements the marshaler and unmarshaler contracts for the json, yaml,
// toml and cbor encoding libraries, plus the BinaryMarshaler and
// BinaryUnmarshaler interfaces from the standard library.

package status

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// MarshalJSON implements the json.Marshaler interface.
//
// The status is serialized as its raw uint8 value, so that the binary bit
// field is preserved losslessly. The numeric form keeps the round trip
// exact and stays compatible with the documents produced by the other
// serializers of the auditor.
//
// Returns:
//   - []byte: The JSON-encoded representation of the status.
//   - error: Any error that occurred during the marshaling. It is always nil
//     for a valid status.
func (o Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Uint8())
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// The payload must be a JSON number: it is decoded into a uint8 and then
// converted into a Status with the Get helper. A JSON string is rejected
// with a type error.
//
// Parameters:
//   - p: The JSON-encoded byte slice to decode into the status.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling. A type error is
//     returned when the payload is not a number or when the number does not
//     fit in a uint8.
func (o *Status) UnmarshalJSON(p []byte) error {
	var r uint8

	if e := json.Unmarshal(p, &r); e != nil {
		return e
	}

	*o = Get(r)

	return nil
}

// MarshalYAML implements the yaml.Marshaler interface.
//
// The status is serialized as its raw uint8 value, which keeps the binary bit
// field lossless.
//
// Returns:
//   - interface{}: The YAML-encoded representation of the status, as a byte
//     slice holding the encoded scalar.
//   - error: Any error that occurred during the marshaling.
//
// Note:
//   - The yaml encoder expects the raw value to be returned by this method.
//     The current implementation returns the YAML document itself, which the
//     yaml package encodes as a sequence of bytes, so the round trip through
//     the yaml format is not lossless.
func (o Status) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(o.Uint8())
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
//
// The scalar value carried by the node is decoded into a uint8 and then
// converted into a Status with the Get helper. A node that holds a non
// numeric scalar is rejected with a type error.
//
// Parameters:
//   - value: The YAML node holding the scalar to decode into the status.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling.
func (o *Status) UnmarshalYAML(value *yaml.Node) error {
	var r uint8

	if e := yaml.Unmarshal([]byte(value.Value), &r); e != nil {
		return e
	}

	*o = Get(r)

	return nil
}

// MarshalTOML serializes the status as its raw uint8 value using the
// go-toml/v2 library.
//
// Returns:
//   - []byte: The TOML-encoded representation of the status.
//   - error: Any error that occurred during the marshaling.
//
// Note:
//   - The go-toml/v2 library refuses to encode a scalar as the root of a
//     document, so this method reports an error when the status is encoded
//     on its own. The status has to be encoded as a field of a structure to
//     be accepted by the library.
func (o Status) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.Uint8())
}

// UnmarshalTOML deserializes a Status from an interface that must be either
// a byte slice or a string containing a TOML-encoded uint8.
//
// The two accepted input forms are handled by the same decoding logic, which
// keeps the compatibility with the various TOML data sources.
//
// Parameters:
//   - i: The TOML data to decode, either a []byte or a string. Any other type
//     is rejected with an "invalid type" error.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling.
//
// Note:
//   - The type detection is performed by two successive type assertions, and
//     the result of the last one is the one that is tested: a []byte input
//     is therefore reported as an invalid type.
//   - The decoded value is assigned to the receiver through a local variable
//     that is passed by value to the TOML decoder, so the receiver is left
//     unchanged and the status keeps its previous value.
func (o *Status) UnmarshalTOML(i interface{}) error {
	var (
		e error
		k bool
		p []byte
		s string
		r uint8
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

	*o = Get(r)

	return nil
}

// MarshalCBOR serializes the status as its raw uint8 value using the
// cbor/v2 library.
//
// The binary format is the canonical representation of the status: it is the
// one used by MarshalBinary and by UnmarshalBinary.
//
// Returns:
//   - []byte: The CBOR-encoded representation of the status.
//   - error: Any error that occurred during the marshaling.
func (o Status) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(o.Uint8())
}

// UnmarshalCBOR deserializes a Status from a CBOR-encoded byte slice
// containing a uint8.
//
// Parameters:
//   - p: The CBOR-encoded byte slice to decode into the status.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling. A type error is
//     returned when the payload does not hold an unsigned integer.
func (o *Status) UnmarshalCBOR(p []byte) error {
	var r uint8

	if e := cbor.Unmarshal(p, &r); e != nil {
		return e
	}

	*o = Get(r)

	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
//
// The CBOR format is used as the canonical binary representation of the
// status, so the method simply delegates to MarshalCBOR.
//
// Returns:
//   - []byte: The binary representation of the status.
//   - error: Any error that occurred during the marshaling.
func (o Status) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
//
// The input is expected to be a CBOR-encoded uint8 value, and the method
// simply delegates to UnmarshalCBOR.
//
// Parameters:
//   - p: The binary byte slice to decode into the status.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling.
func (o *Status) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}
