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

// Package types provides enumerations and utilities for code type management.
//
// This package defines a set of constants representing different code element types
// that can be analyzed within the auditor system. It includes functions for parsing
// string representations into these enumerated types, as well as utility methods
// for working with these types in a type-safe manner.
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
// This method serializes the CodeType value into a JSON representation
// using its underlying uint8 value. The serialization preserves the
// numeric representation of the code type for compatibility with JSON-based
// APIs and configuration systems.
//
// Returns:
//   - []byte: The JSON-encoded representation of the CodeType
//   - error: Any error that occurred during marshaling
func (o CodeType) MarshalJSON() ([]byte, error) {
	if o == EntryNone {
		v := audpkg.Unknown
		return json.Marshal(v)
	}

	return json.Marshal(strings.ToLower(o.String()))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// This method deserializes a JSON representation back into a CodeType value.
// It accepts numeric values representing the underlying uint8 representation
// of the code type. The parsed value is converted to the appropriate CodeType
// using the Get function for validation and type safety.
//
// Parameters:
//   - p: The JSON-encoded data to be unmarshaled
//
// Returns:
//   - error: Any error that occurred during unmarshaling
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
// This method serializes the CodeType value into a YAML representation
// using its underlying uint8 value. The serialization preserves the
// numeric representation of the code type for compatibility with YAML-based
// configuration systems while maintaining readability.
//
// Returns:
//   - interface{}: The YAML-encoded representation of the CodeType
//   - error: Any error that occurred during marshaling
func (o CodeType) MarshalYAML() (interface{}, error) {
	if o == EntryNone {
		v := audpkg.Unknown
		return yaml.Marshal(v)
	}

	return yaml.Marshal(strings.ToLower(o.String()))
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
//
// This method deserializes a YAML representation back into a CodeType value.
// It accepts numeric values representing the underlying uint8 representation
// of the code type. The parsed value is converted to the appropriate CodeType
// using the Get function for validation and type safety.
//
// Parameters:
//   - value: The YAML node containing the data to be unmarshaled
//
// Returns:
//   - error: Any error that occurred during unmarshaling
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
// This method serializes the CodeType value into a TOML representation
// using its underlying uint8 value. The serialization preserves the
// numeric representation of the code type for compatibility with TOML-based
// configuration systems while maintaining readability.
//
// Returns:
//   - []byte: The TOML-encoded representation of the CodeType
//   - error: Any error that occurred during marshaling
func (o CodeType) MarshalTOML() ([]byte, error) {
	if o == EntryNone {
		v := audpkg.Unknown
		return toml.Marshal(v)
	}

	return toml.Marshal(strings.ToLower(o.String()))
}

// UnmarshalTOML implements the TOML unmarshaler interface.
//
// This method deserializes a TOML representation back into a CodeType value.
// It accepts various data types including []byte and string representations
// of the code type. The parsed value is converted to the appropriate CodeType
// using the Get function for validation and type safety.
//
// Parameters:
//   - i: The TOML data to be unmarshaled (can be []byte or string)
//
// Returns:
//   - error: Any error that occurred during unmarshaling
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
// This method serializes the CodeType value into a CBOR (Concise Binary Object Representation)
// format. CBOR is a binary serialization format that provides compact encoding while maintaining
// compatibility with JSON-like structures. The serialization uses the underlying uint8 value
// of the CodeType for efficient binary representation.
//
// Returns:
//   - []byte: The CBOR-encoded representation of the CodeType
//   - error: Any error that occurred during marshaling
func (o CodeType) MarshalCBOR() ([]byte, error) {
	if o == EntryNone {
		v := audpkg.Unknown
		return cbor.Marshal(v)
	}

	return cbor.Marshal(strings.ToLower(o.String()))
}

// UnmarshalCBOR implements CBOR unmarshaling using the fxamacker/cbor library.
//
// This method deserializes a CBOR representation back into a CodeType value.
// It decodes CBOR data containing the numeric representation of the code type
// and converts it to the appropriate CodeType using the Get function for validation.
//
// Parameters:
//   - p: The CBOR-encoded data to be unmarshaled
//
// Returns:
//   - error: Any error that occurred during unmarshaling
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
// This method provides binary serialization of CodeType values using CBOR encoding.
// It serves as a convenient wrapper around the MarshalCBOR method, making CodeType
// compatible with Go's standard binary marshaling interfaces for use in network
// communication or storage systems.
//
// Returns:
//   - []byte: The binary representation of the CodeType
//   - error: Any error that occurred during marshaling
func (o CodeType) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
//
// This method provides binary deserialization of CodeType values using CBOR decoding.
// It serves as a convenient wrapper around the UnmarshalCBOR method, making CodeType
// compatible with Go's standard binary unmarshaling interfaces for use in network
// communication or storage systems.
//
// Parameters:
//   - p: The binary data to be unmarshaled
//
// Returns:
//   - error: Any error that occurred during unmarshaling
func (o *CodeType) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}
