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

// Package src provides a set of interfaces and utilities for managing source code content in various formats.
//
// This package defines a comprehensive framework for handling source code data, enabling flexible
// encoding and decoding across multiple formats, including JSON, YAML, TOML, and CBOR. It provides
// specialized implementations of the Source interface, optimized for efficient memory usage through
// hexadecimal encoding.
package src

import (
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// MarshalJSON implements the json.Marshaler interface.
//
// It serializes the internal hex-encoded string representation of the source
// content into a JSON-compatible format. The serialized output is a JSON string
// containing the hex-encoded data, preserving the original data integrity while
// providing compatibility with JSON-based systems and configuration files.
//
// Example:
//
//	source := New([]byte("Hello, World!"))
//	jsonData, err := source.MarshalJSON()
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Println(string(jsonData)) // Output: "\"48656c6c6f2c20576f726c6421\""
func (o *mdl) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.d)
}

// UnmarshalJSON implements the json.Unmarshaler interface.
//
// It deserializes JSON data back into the internal hex-encoded representation
// of source content. It expects the JSON input to be a string value containing
// the hex-encoded data. The method parses the JSON string and assigns it to the
// internal field.
//
// Example:
//
//	var source Source
//	json.Unmarshal([]byte(`"48656c6c6f2c20576f726c6421"`), &source)
//	// The source now contains the decoded content
func (o *mdl) UnmarshalJSON(p []byte) error {
	var r string

	if e := json.Unmarshal(p, &r); e != nil {
		return e
	}

	o.d = r

	return nil
}

// MarshalYAML implements the yaml.Marshaler interface.
//
// It serializes the internal hex-encoded string representation into a YAML
// format that is human-readable and compatible with YAML-based configuration
// files. The output maintains the original data integrity while providing
// clear, readable representations for YAML users. This is particularly useful
// for configuration files where readability is important.
//
// Example:
//
//	source := New([]byte("Hello, World!"))
//	yamlData, err := source.MarshalYAML()
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Println(string(yamlData)) // Output: "48656c6c6f2c20576f726c6421\n"
func (o *mdl) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(o.d)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface.
//
// It deserializes YAML data back into the internal hex-encoded representation
// of source content. It extracts the string value from the provided yaml.Node
// and parses it as the hex-encoded data. The method ensures that YAML
// configuration files can properly deserialize source data.
//
// Example:
//
//	var source Source
//	yaml.Unmarshal([]byte("48656c6c6f2c20576f726c6421"), &source)
//	// The source now contains the decoded content
func (o *mdl) UnmarshalYAML(value *yaml.Node) error {
	var r string

	if e := yaml.Unmarshal([]byte(value.Value), &r); e != nil {
		return e
	}

	o.d = r

	return nil
}

// MarshalTOML implements the TOML marshaler interface.
//
// It serializes the internal hex-encoded string representation into a TOML
// format that is human-readable and compatible with TOML-based configuration
// files. The output maintains the original data integrity while providing
// clear, readable representations for TOML users. TOML's simplicity makes it
// ideal for configuration files where clarity and ease of use are priorities.
//
// Example:
//
//	source := New([]byte("Hello, World!"))
//	tomlData, err := source.MarshalTOML()
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Println(string(tomlData)) // Output: "48656c6c6f2c20576f726c6421\n"
func (o *mdl) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.d)
}

// UnmarshalTOML implements the TOML unmarshaler interface.
//
// It deserializes TOML data back into the internal hex-encoded representation
// of source content. It accepts TOML values that are either string representations
// or byte arrays and handles parsing to maintain compatibility with TOML-based
// systems. The method ensures that TOML configuration files can properly
// deserialize source data.
//
// Example:
//
//	var source Source
//	toml.Unmarshal([]byte("48656c6c6f2c20576f726c6421"), &source)
//	// The source now contains the decoded content
func (o *mdl) UnmarshalTOML(i interface{}) error {
	var (
		e error
		k bool
		p []byte
		s string
		r string
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

	o.d = r

	return nil
}

// MarshalCBOR implements CBOR marshaling using the fxamacker/cbor library.
//
// It serializes the internal hex-encoded string representation into a compact
// binary format using the CBOR (Concise Binary Object Representation) encoding
// standard. CBOR provides efficient binary serialization that is smaller than
// JSON while maintaining compatibility with various systems. The method leverages
// the fxamacker/cbor library for robust and standardized CBOR implementation.
//
// See also: github.com/fxamacker/cbor
//
// Example:
//
//	source := New([]byte("Hello, World!"))
//	cborData, err := source.MarshalCBOR()
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Printf("%x\n", cborData) // Output: 6448656c6c6f2c20576f726c6421
func (o *mdl) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(o.d)
}

// UnmarshalCBOR implements CBOR unmarshaling using the fxamacker/cbor library.
//
// It deserializes CBOR data back into the internal hex-encoded representation
// of source content. It uses the CBOR decoding capabilities from the
// fxamacker/cbor library to properly parse binary data and reconstruct the
// original source information. The method ensures compatibility with CBOR-based
// systems and protocols.
//
// See also: github.com/fxamacker/cbor
//
// Example:
//
//	var source Source
//	cborData := []byte{0x64, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x21}
//	source.UnmarshalCBOR(cborData)
//	// The source now contains the decoded content
func (o *mdl) UnmarshalCBOR(p []byte) error {
	var r string

	if e := cbor.Unmarshal(p, &r); e != nil {
		return e
	}

	o.d = r

	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
//
// It provides binary serialization of source content using CBOR encoding
// internally. It offers a compact representation suitable for storage or
// network transmission while maintaining compatibility with Go's binary
// marshaling interface. The implementation leverages the efficient CBOR
// format to produce smaller serialized data compared to other formats.
//
// Example:
//
//	source := New([]byte("Hello, World!"))
//	binaryData, err := source.MarshalBinary()
//	if err != nil {
//	    // Handle error
//	}
//	// binaryData can be stored or transmitted in compact binary format
func (o *mdl) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
//
// It provides binary deserialization of source content using CBOR decoding
// internally. It reconstructs source data from compact binary representations,
// making it suitable for efficient storage and transmission scenarios. The
// implementation leverages the robust CBOR decoding capabilities to properly
// parse binary data back into usable source content.
//
// Example:
//
//	var source Source
//	binaryData := []byte{0x64, 0x48, 0x65, 0x6c, 0x6c, 0x6f, 0x2c, 0x20, 0x57, 0x6f, 0x72, 0x6c, 0x64, 0x21}
//	source.UnmarshalBinary(binaryData)
//	// The source now contains the decoded content
func (o *mdl) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}
