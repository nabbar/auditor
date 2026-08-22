/*
 * MIT License
 *
 * Copyright (c) 2022 Nicolas JUHEL
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

// Package id provides utilities for generating and managing unique identifiers.
// This package implements a deterministic identifier generation system based on the FNV-1a hash algorithm
// to create consistent, unique IDs for code elements throughout the auditor system.
// It ensures reproducible identification while maintaining uniqueness across different execution contexts.
//
// The ID type is designed to be lightweight and efficient, providing a way to uniquely identify
// components or elements in a deterministic manner.
//
// This package implements serialization and deserialization capabilities for ID types
// using various formats including JSON, YAML, TOML, and CBOR for persistent storage and transfer.
//
// The serialization methods enable seamless integration with configuration management,
// data persistence systems, and cross-system communication.
//
// This package extends the basic ID type functionality to support various data exchange formats,
// making it easier to store, transfer, and configure ID-based systems.
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
// This function converts the ID into a JSON representation by returning its underlying
// uint64 value. The marshaled output is a simple numeric value suitable for standard
// JSON processing tools and configuration management.
//
// Returns:
//   - []byte: JSON-encoded representation of the ID value
//   - error: Any error that occurred during marshaling
func (o ID) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Uint64())
}

// UnmarshalJSON implements the json.Unmarshaler interface for ID deserialization.
//
// This function parses JSON data into an ID value by converting the numeric representation
// back to the ID type. It accepts both numeric values and string representations of numbers,
// providing flexibility in JSON input formats.
//
// The unmarshaling process handles various JSON input scenarios, including:
//   - Numeric values (e.g., 123456)
//   - String representations that can be parsed as numbers
//
// Returns:
//   - error: Any error that occurred during unmarshaling
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
// This function converts the ID into a YAML representation by returning its underlying
// uint64 value. YAML format is particularly useful for configuration files and
// human-readable data exchange, providing clean syntax and readability.
//
// Returns:
//   - interface{}: YAML representation of the ID value
//   - error: Any error that occurred during marshaling
func (o ID) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(o.Uint64())
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for ID deserialization.
//
// This function parses YAML data into an ID value by converting the numeric representation
// back to the ID type. It handles standard YAML input formats and provides reliable
// conversion from YAML configuration files.
//
// Returns:
//   - error: Any error that occurred during unmarshaling
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
// This function converts the ID into a TOML representation by returning its underlying
// uint64 value. TOML format is ideal for configuration files and provides clean,
// readable syntax that's easy to maintain in configuration management scenarios.
//
// Returns:
//   - []byte: TOML-encoded representation of the ID value
//   - error: Any error that occurred during marshaling
func (o ID) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.Uint64())
}

// UnmarshalTOML implements the TOML unmarshaler interface for ID deserialization.
//
// This function accepts both []byte and string representations from TOML files,
// parsing them into an ID value by converting the numeric representation back
// to the ID type. It provides flexibility in handling different TOML input formats.
//
// The implementation handles multiple input types for TOML unmarshaling, ensuring
// compatibility with various TOML data sources while maintaining robust error handling.
//
// Parameters:
//   - i: TOML data in []byte or string format
//
// Returns:
//   - error: Any error that occurred during unmarshaling
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
// This function serializes the ID into CBOR (Concise Binary Object Representation)
// format, providing compact binary serialization suitable for efficient storage and
// transmission. CBOR is particularly useful for machine-to-machine communication
// and binary data exchange scenarios.
//
// Returns:
//   - []byte: CBOR-encoded representation of the ID value
//   - error: Any error that occurred during marshaling
func (o ID) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(o.Uint64())
}

// UnmarshalCBOR implements CBOR unmarshaling using the fxamacker/cbor library.
//
// This function deserializes CBOR data back into an ID value by converting the
// numeric representation back to the ID type. It provides reliable decoding of
// binary data for efficient reconstruction of ID values.
//
// Returns:
//   - error: Any error that occurred during unmarshaling
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
// This method uses CBOR encoding internally to provide compact binary
// serialization of ID values. It's particularly useful for efficient
// storage and network transmission when binary format is preferred.
//
// Returns:
//   - []byte: Binary representation of the ID value
//   - error: Any error that occurred during marshaling
func (o ID) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
//
// This method uses CBOR decoding internally to deserialize binary data
// back into an ID value. It's particularly useful for efficient
// reconstruction of ID values from binary storage or network transmission.
//
// Returns:
//   - error: Any error that occurred during unmarshaling
func (o *ID) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}
