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

// File encode.go contains the serialization and deserialization of the Source
// interface.
//
// It implements the json, yaml, toml and cbor marshaler and unmarshaler
// contracts, plus the encoding.BinaryMarshaler and encoding.BinaryUnmarshaler
// ones, of the mdl model.

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
// document that is human-readable and compatible with the YAML based
// configuration files. The integrity of the original data is preserved, while
// the representation stays clear and readable.
//
// Note:
//   - The returned value is the encoded YAML document, and not the string
//     itself: the YAML encoder therefore produces a YAML document that holds
//     the byte representation of the source.
//
// Example:
//
//	source := New([]byte("Hello, World!"))
//	yamlData, err := source.MarshalYAML()
//	if err != nil {
//	    // Handle error
//	}
//	fmt.Println(string(yamlData.([]byte))) // Output: 48656c6c6f2c20576f726c6421
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
// It serializes the internal hex-encoded string representation into the TOML
// format, which is human-readable and compatible with the TOML based
// configuration files.
//
// Note:
//   - The TOML format only describes a document, and not a root value: the
//     marshalling of the hexadecimal string therefore fails with the
//     "cannot encode a string as a document root" error of the encoder.
//
// Example:
//
//	source := New([]byte("Hello, World!"))
//	tomlData, err := source.MarshalTOML()
//	if err != nil {
//	    // Handle error: the content is not a TOML document
//	}
//	_ = tomlData
func (o *mdl) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.d)
}

// UnmarshalTOML implements the TOML unmarshaler interface.
//
// It deserializes the TOML data back into the internal hex-encoded
// representation of the source content. It accepts the TOML values that are
// either string representations or byte arrays, and handles the parsing to
// maintain the compatibility with the TOML based systems.
//
// Note:
//   - The type switch of the implementation overwrites its own result, so the
//     []byte input is currently rejected with an "invalid type" error, and the
//     string input is rejected because the destination is not a pointer. Both
//     cases are reported to the caller through the returned error.
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

// MarshalCBOR implements the CBOR marshaling with the fxamacker/cbor
// library.
//
// It serializes the internal hex-encoded string representation into a compact
// binary format, using the CBOR (Concise Binary Object Representation)
// encoding standard. The CBOR format provides an efficient binary
// serialization that is smaller than the JSON, while it remains compatible
// with the various systems. The method leverages the fxamacker/cbor library
// for a robust and standardized CBOR implementation.
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
//	fmt.Printf("%x\n", cborData) // Output: 781a3438363536633663366632633230353736663732366336343231
func (o *mdl) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(o.d)
}

// UnmarshalCBOR implements the CBOR unmarshaling with the fxamacker/cbor
// library.
//
// It deserializes the CBOR data back into the internal hex-encoded
// representation of the source content. It uses the decoding capabilities of
// the fxamacker/cbor library to properly parse the binary data and to
// reconstruct the original information of the source.
//
// See also: github.com/fxamacker/cbor
//
// Example:
//
//	var source Source
//	cborData := []byte{0x78, 0x1a, 0x34, 0x38, 0x36, 0x35, 0x36, 0x63, 0x36, 0x63, 0x36, 0x66, 0x32, 0x63, 0x32, 0x30, 0x35, 0x37, 0x36, 0x66, 0x37, 0x32, 0x36, 0x63, 0x36, 0x34, 0x32, 0x31}
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
// It provides the binary deserialization of the source content, using the
// CBOR decoding internally. It reconstructs the source data from the compact
// binary representations, which makes it suitable for the efficient storage
// and the transmission scenarios.
//
// Example:
//
//	var source Source
//	binaryData := []byte{0x78, 0x1a, 0x34, 0x38, 0x36, 0x35, 0x36, 0x63, 0x36, 0x63, 0x36, 0x66, 0x32, 0x63, 0x32, 0x30, 0x35, 0x37, 0x36, 0x66, 0x37, 0x32, 0x36, 0x63, 0x36, 0x34, 0x32, 0x31}
//	source.UnmarshalBinary(binaryData)
//	// The source now contains the decoded content
func (o *mdl) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}
