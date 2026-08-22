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

// Package status provides enumerations and utilities for managing audit status states.
//
// This package defines the different lifecycle states an audit item can be in during processing.
// Each status represents a distinct phase in the audit workflow, from initial discovery to final review.
// The status values are designed to be used as flags in state machines or process tracking systems.
package status

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// MarshalJSON implements the json.Marshaler interface for Status type.
//
// This method provides JSON serialization of Status values by converting them to their
// underlying uint8 representation. This approach ensures consistent serialization
// across different systems and maintains compatibility with JSON-based APIs.
//
// Example:
//
//	status := Pending
//	jsonData, err := status.MarshalJSON()
//	// Returns: []byte("1") where 1 represents Pending status
//
// Note: The JSON output is a simple numeric string ("1", "2", "3", etc.) rather than
// a JSON object or array. This format is compact and works well for JSON-based APIs
// and configuration files.
func (o Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.Uint8())
}

// UnmarshalJSON implements the json.Unmarshaler interface for Status type.
//
// This method handles JSON deserialization of Status values. It accepts numeric values
// representing valid Status codes and converts them back to Status enum values.
// Invalid values will result in None status being returned.
//
// Example:
//
//	var status Status
//	json.Unmarshal([]byte("3"), &status) // Sets status to Failed
//
// Tip: Be cautious when using this function, as invalid input values (like "abc")
// will silently return None, potentially masking critical state tracking issues.
func (o *Status) UnmarshalJSON(p []byte) error {
	var r uint8

	if e := json.Unmarshal(p, &r); e != nil {
		return e
	}

	*o = Get(r)

	return nil
}

// MarshalYAML implements the yaml.Marshaler interface for Status type.
//
// This method provides YAML serialization of Status values by converting them to their
// underlying uint8 representation. The serialized format maintains consistency and
// readability in YAML configuration files.
//
// Example:
//
//	status := Pending
//	yamlData, err := status.MarshalYAML()
//	// Returns: []byte("1") where 1 represents Pending status
//
// YAML output is a simple numeric value ("1", "2", etc.) which can be easily read
// back into a Status enum using the UnmarshalYAML method.
func (o Status) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(o.Uint8())
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for Status type.
//
// This method handles YAML deserialization of Status values. It accepts numeric values
// representing valid Status codes and converts them back to Status enum values.
// Invalid values will result in None status being returned.
//
// Example:
//
//	var status Status
//	yaml.Unmarshal([]byte("3"), &status) // Sets status to Failed
func (o *Status) UnmarshalYAML(value *yaml.Node) error {
	var r uint8

	if e := yaml.Unmarshal([]byte(value.Value), &r); e != nil {
		return e
	}

	*o = Get(r)

	return nil
}

// MarshalTOML implements the TOML marshaler interface for Status type.
//
// This method provides TOML serialization of Status values by converting them to their
// underlying uint8 representation. This ensures compatibility with TOML configuration files
// while maintaining consistent data representation.
//
// Example:
//
//	status := Pending
//	tomlData, err := status.MarshalTOML()
//	// Returns: []byte("1") where 1 represents Pending status
func (o Status) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.Uint8())
}

// UnmarshalTOML implements the TOML unmarshaler interface for Status type.
//
// This method handles TOML deserialization of Status values. It accepts numeric values
// representing valid Status codes and converts them back to Status enum values.
// Invalid values will result in None status being returned.
//
// Example:
//
//	var status Status
//	toml.Unmarshal(data, &status) // Sets status based on TOML value
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

// MarshalCBOR implements CBOR marshaling for Status type using the fxamacker/cbor library.
//
// This method provides compact binary serialization of Status values using CBOR format.
// CBOR is a more efficient encoding than JSON or YAML, especially for machine-to-machine
// communication and storage purposes. It preserves the exact status value as a numeric code.
//
// Example:
//
//	status := Pending
//	cborData, err := status.MarshalCBOR()
//	// Returns compact binary representation of the status value
func (o Status) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(o.Uint8())
}

// UnmarshalCBOR implements CBOR unmarshaling for Status type using the fxamacker/cbor library.
//
// This method decodes CBOR data containing numeric status representations and converts
// them back to Status enum values. It provides efficient deserialization for binary formats.
//
// Example:
//
//	var status Status
//	status.UnmarshalCBOR(cborData) // Decodes CBOR data into Status value
func (o *Status) UnmarshalCBOR(p []byte) error {
	var r uint8

	if e := cbor.Unmarshal(p, &r); e != nil {
		return e
	}

	*o = Get(r)

	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface for Status type.
//
// This method uses CBOR encoding internally to provide compact binary serialization
// of Status values. It's useful for storing status information in databases or
// transmitting over network protocols where efficiency is important.
//
// Example:
//
//	status := Pending
//	binaryData, err := status.MarshalBinary()
//	// Returns compact binary representation suitable for storage/transmission
func (o Status) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface for Status type.
//
// This method uses CBOR decoding internally to deserialize binary data into Status values.
// It's useful for reading status information from databases or network protocols where
// efficient binary representation is required.
//
// Example:
//
//	var status Status
//	status.UnmarshalBinary(binaryData) // Decodes binary data into Status value
func (o *Status) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}
