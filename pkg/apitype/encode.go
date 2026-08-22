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

// Package apitype defines the API type enumeration and related utilities for
// identifying and converting between different AI/LLM API providers.
package apitype

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// MarshalJSON implements the json.Marshaler interface.
//
// This method serializes the ApiType value into a JSON representation
// using its string representation (lowercased). The serialization preserves
// the human-readable name of the API type for compatibility with JSON-based
// APIs and configuration systems.
//
// Returns:
//   - []byte: The JSON-encoded representation of the ApiType
//   - error: Any error that occurred during marshaling
func (o ApiType) MarshalJSON() ([]byte, error) {
	return json.Marshal(strings.ToLower(o.String()))
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// This method deserializes a JSON representation back into an ApiType value.
// It accepts both string representations (e.g., "ollama", "openia") and numeric
// values representing the underlying uint8 representation of the API type.
// The parsed value is converted to the appropriate ApiType using the Parse
// or ParseUint8 functions for validation and type safety.
//
// Parameters:
//   - p: The JSON-encoded data to be unmarshaled
//
// Returns:
//   - error: Any error that occurred during unmarshaling
func (o *ApiType) UnmarshalJSON(p []byte) error {
	var (
		s string
		i uint8
	)

	// Attempt to unmarshal as a string first
	if e := json.Unmarshal(p, &s); e == nil && s != "" {
		*o = Parse(s)
		return nil
	}

	// If string unmarshaling failed, attempt to unmarshal as a uint8
	if e := json.Unmarshal(p, &i); e != nil {
		return e
	}

	*o = ParseUint8(i)
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface.
//
// This method serializes the ApiType value into a YAML representation
// using its string representation (lowercased). The serialization preserves
// the human-readable name of the API type for compatibility with YAML-based
// configuration systems while maintaining readability.
//
// Returns:
//   - interface{}: The YAML-encoded representation of the ApiType
//   - error: Any error that occurred during marshaling
func (o ApiType) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(strings.ToLower(o.String()))
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
//
// This method deserializes a YAML representation back into an ApiType value.
// It accepts both string representations (e.g., "ollama", "openia") and numeric
// values representing the underlying uint8 representation of the API type.
// The parsed value is converted to the appropriate ApiType using the Parse
// or ParseUint8 functions for validation and type safety.
//
// Parameters:
//   - value: The YAML node containing the data to be unmarshaled
//
// Returns:
//   - error: Any error that occurred during unmarshaling
func (o *ApiType) UnmarshalYAML(value *yaml.Node) error {
	var (
		s string
		i uint8
	)

	// Attempt to unmarshal as a string first
	if e := yaml.Unmarshal([]byte(value.Value), &s); e == nil && s != "" {
		*o = Parse(s)
		return nil
	}

	// If string unmarshaling failed, attempt to unmarshal as a uint8
	if e := yaml.Unmarshal([]byte(value.Value), &i); e != nil {
		return e
	}

	*o = ParseUint8(i)

	return nil
}

// MarshalTOML implements the TOML marshaler interface.
//
// This method serializes the ApiType value into a TOML representation
// using its string representation (lowercased). The serialization preserves
// the human-readable name of the API type for compatibility with TOML-based
// configuration systems while maintaining readability.
//
// Returns:
//   - []byte: The TOML-encoded representation of the ApiType
//   - error: Any error that occurred during marshaling
func (o ApiType) MarshalTOML() ([]byte, error) {
	return toml.Marshal(strings.ToLower(o.String()))
}

// UnmarshalTOML implements the TOML unmarshaler interface.
//
// This method deserializes a TOML representation back into an ApiType value.
// It accepts both []byte and string representations of the API type.
// The parsed value is converted to the appropriate ApiType using the Parse
// or ParseUint8 functions for validation and type safety.
//
// Parameters:
//   - i: The TOML data to be unmarshaled (can be []byte or string)
//
// Returns:
//   - error: Any error that occurred during unmarshaling
func (o *ApiType) UnmarshalTOML(i interface{}) error {
	var (
		e error
		f func(m interface{}) error
		k bool
		p []byte
		s string
	)

	// Determine the type of input and set up the appropriate unmarshal function
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

	// Validate that the input type is supported
	if !k || f == nil {
		return fmt.Errorf("invalid type: %T", i)
	}

	var (
		j uint8
		b string
	)

	// Attempt to unmarshal as a string first
	if e = f(&b); e == nil && b != "" {
		*o = Parse(b)
		return nil
	}

	// If string unmarshaling failed, attempt to unmarshal as a uint8
	if e = f(&j); e != nil {
		return e
	}

	*o = ParseUint8(j)

	return nil
}

// MarshalCBOR implements CBOR marshaling using the fxamacker/cbor library.
//
// This method serializes the ApiType value into a CBOR (Concise Binary Object
// Representation) format. CBOR is a binary serialization format that provides
// compact encoding while maintaining compatibility with JSON-like structures.
// The serialization uses the string representation (lowercased) of the ApiType
// for efficient binary representation.
//
// Returns:
//   - []byte: The CBOR-encoded representation of the ApiType
//   - error: Any error that occurred during marshaling
func (o ApiType) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(strings.ToLower(o.String()))
}

// UnmarshalCBOR implements CBOR unmarshaling using the fxamacker/cbor library.
//
// This method deserializes a CBOR representation back into an ApiType value.
// It decodes CBOR data containing the string or numeric representation of the
// API type and converts it to the appropriate ApiType using the Parse or
// ParseUint8 functions for validation.
//
// Parameters:
//   - p: The CBOR-encoded data to be unmarshaled
//
// Returns:
//   - error: Any error that occurred during unmarshaling
func (o *ApiType) UnmarshalCBOR(p []byte) error {
	var (
		s string
		i uint8
	)

	// Attempt to unmarshal as a string first
	if e := cbor.Unmarshal(p, &s); e == nil && s != "" {
		*o = Parse(s)
		return nil
	}

	// If string unmarshaling failed, attempt to unmarshal as a uint8
	if e := cbor.Unmarshal(p, &i); e != nil {
		return e
	}

	*o = ParseUint8(i)
	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
//
// This method provides binary serialization of ApiType values using CBOR encoding.
// It serves as a convenient wrapper around the MarshalCBOR method, making ApiType
// compatible with Go's standard binary marshaling interfaces for use in network
// communication or storage systems.
//
// Returns:
//   - []byte: The binary representation of the ApiType
//   - error: Any error that occurred during marshaling
func (o ApiType) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
//
// This method provides binary deserialization of ApiType values using CBOR decoding.
// It serves as a convenient wrapper around the UnmarshalCBOR method, making ApiType
// compatible with Go's standard binary unmarshaling interfaces for use in network
// communication or storage systems.
//
// Parameters:
//   - p: The binary data to be unmarshaled
//
// Returns:
//   - error: Any error that occurred during unmarshaling
func (o *ApiType) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}
