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
// CodeType type.
//
// It implements the json, yaml, toml and cbor marshaler and unmarshaler
// contracts of the enumeration. A code type is serialized as its lower case
// textual representation, and can be read back either from that
// representation or from its numeric value.

package types

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fxamacker/cbor/v2"
	audpkg "github.com/nabbar/auditor/pkg/generic"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// MarshalJSON implements the json.Marshaler interface.
//
// A code type is serialized as its lower case textual representation, for
// example "function" or "constant", which keeps the serialized documents
// readable and stable across the versions of the auditor.
//
// The unknown code types are serialized as the audpkg.Unknown marker.
//
// Returns:
//   - []byte: The JSON-encoded representation of the code type.
//   - error: Any error that occurred during the marshaling.
func (o CodeType) MarshalJSON() ([]byte, error) {
	if o == EntryNone {
		v := audpkg.Unknown
		return json.Marshal(v)
	}

	return json.Marshal(strings.ToLower(o.String()))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// The payload is first decoded as a JSON string and parsed with the Parse
// helper, which accepts the names in any case. When the string is unknown,
// the payload is decoded again as a JSON number and converted with the Get
// helper, so that both the textual and the numeric forms are accepted.
//
// Parameters:
//   - p: The JSON-encoded data to be unmarshaled.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling.
func (o *CodeType) UnmarshalJSON(p []byte) error {
	var (
		s string
		i uint8
	)

	if e := json.Unmarshal(p, &s); e == nil {
		*o = Parse(s)
	}

	if *o != EntryNone {
		return nil
	}

	if e := json.Unmarshal(p, &i); e != nil {
		return e
	}

	*o = Get(i)
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface.
//
// A code type is serialized as its lower case textual representation, which
// keeps the serialized documents readable and stable across the versions of
// the auditor.
//
// The unknown code types are serialized as the audpkg.Unknown marker.
//
// Returns:
//   - interface{}: The YAML-encoded representation of the code type.
//   - error: Any error that occurred during the marshaling.
//
// Note:
//   - The yaml encoder expects the raw value to be returned by this method.
//     The current implementation returns the YAML document itself, which the
//     yaml package encodes as a sequence of bytes, so the round trip through
//     the yaml format is not lossless.
func (o CodeType) MarshalYAML() (interface{}, error) {
	if o == EntryNone {
		v := audpkg.Unknown
		return yaml.Marshal(v)
	}

	return yaml.Marshal(strings.ToLower(o.String()))
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
//
// The scalar value carried by the node is first decoded as a string and
// parsed with the Parse helper. When the string is unknown, the scalar is
// decoded again as a number and converted with the Get helper, so that both
// the textual and the numeric forms are accepted.
//
// Parameters:
//   - value: The YAML node containing the data to be unmarshaled.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling.
func (o *CodeType) UnmarshalYAML(value *yaml.Node) error {
	var (
		s string
		i uint8
	)

	if e := yaml.Unmarshal([]byte(value.Value), &s); e == nil {
		*o = Parse(s)
	}

	if *o != EntryNone {
		return nil
	}

	if e := yaml.Unmarshal([]byte(value.Value), &i); e != nil {
		return e
	}

	*o = Get(i)

	return nil
}

// MarshalTOML implements the TOML marshaler interface.
//
// A code type is serialized as its lower case textual representation, which
// keeps the serialized documents readable and stable across the versions of
// the auditor.
//
// The unknown code types are serialized as the audpkg.Unknown marker.
//
// Returns:
//   - []byte: The TOML-encoded representation of the code type.
//   - error: Any error that occurred during the marshaling.
//
// Note:
//   - The go-toml/v2 library refuses to encode a scalar as the root of a
//     document, so this method reports an error when the code type is
//     encoded on its own. The code type has to be encoded as a field of a
//     structure to be accepted by the library.
func (o CodeType) MarshalTOML() ([]byte, error) {
	if o == EntryNone {
		v := audpkg.Unknown
		return toml.Marshal(v)
	}

	return toml.Marshal(strings.ToLower(o.String()))
}

// UnmarshalTOML implements the TOML unmarshaler interface.
//
// The TOML data is first decoded as a string and parsed with the Parse
// helper. When the string is unknown, the data is decoded again as a number
// and converted with the Get helper, so that both the textual and the
// numeric forms are accepted.
//
// Parameters:
//   - i: The TOML data to be unmarshaled. It can be either a []byte or a
//     string. Any other type is rejected with an "invalid type" error.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling.
func (o *CodeType) UnmarshalTOML(i interface{}) error {
	var (
		e error
		f func(m interface{}) error
		k bool
		p []byte
		s string
	)

	if p, k = i.([]byte); k {
		f = func(m interface{}) error {
			return toml.Unmarshal(p, m)
		}
	}

	if s, k = i.(string); k {
		f = func(m interface{}) error {
			return toml.Unmarshal([]byte(s), m)
		}
	}

	if !k || f == nil {
		return fmt.Errorf("invalid type: %T", i)
	}

	var (
		j uint8
		b string
	)

	if e = f(&b); e == nil {
		*o = Parse(b)
	}

	if *o != EntryNone {
		return nil
	}

	if e = f(&j); e != nil {
		return e
	}

	*o = Get(j)

	return nil
}

// MarshalCBOR implements CBOR marshaling using the fxamacker/cbor library.
//
// A code type is serialized as its lower case textual representation, which
// keeps the serialized documents readable and stable across the versions of
// the auditor.
//
// The unknown code types are serialized as the audpkg.Unknown marker.
//
// Returns:
//   - []byte: The CBOR-encoded representation of the code type.
//   - error: Any error that occurred during the marshaling.
func (o CodeType) MarshalCBOR() ([]byte, error) {
	if o == EntryNone {
		v := audpkg.Unknown
		return cbor.Marshal(v)
	}

	return cbor.Marshal(strings.ToLower(o.String()))
}

// UnmarshalCBOR implements CBOR unmarshaling using the fxamacker/cbor library.
//
// The payload is first decoded as a string and parsed with the Parse helper.
// When the string is unknown, the payload is decoded again as a number and
// converted with the Get helper, so that both the textual and the numeric
// forms are accepted.
//
// Parameters:
//   - p: The CBOR-encoded data to be unmarshaled.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling.
func (o *CodeType) UnmarshalCBOR(p []byte) error {
	var (
		s string
		i uint8
	)

	if e := cbor.Unmarshal(p, &s); e == nil {
		*o = Parse(s)
	}

	if *o != EntryNone {
		return nil
	}

	if e := cbor.Unmarshal(p, &i); e != nil {
		return e
	}

	*o = Get(i)
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
//
// The CBOR encoding is used as the canonical binary representation of the
// code type, so the method simply delegates to MarshalCBOR.
//
// Returns:
//   - []byte: The binary representation of the code type.
//   - error: Any error that occurred during the marshaling.
func (o CodeType) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
//
// The method simply delegates to UnmarshalCBOR, which is the canonical binary
// decoding of the code type.
//
// Parameters:
//   - p: The binary data to be unmarshaled.
//
// Returns:
//   - error: Any error that occurred during the unmarshaling.
func (o *CodeType) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}
