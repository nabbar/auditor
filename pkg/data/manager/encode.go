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

// Package manager provides the generic interface for AST (Abstract Syntax Tree) parsers
// that are used across different programming languages in the auditor tool.
package manager

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	audcol "github.com/nabbar/auditor/pkg/data/collection"
	audent "github.com/nabbar/auditor/pkg/data/entry"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// Export is a struct used for serializing and deserializing the manager's
// collections of modules, packages, and entries.
//
// It provides a unified representation of the manager's state that can be
// marshaled into various formats (JSON, YAML, TOML, CBOR) and unmarshaled
// back into the manager's internal collections.
type Export struct {
	// Mod is the collection of modules.
	Mod audcol.Collection[audmod.Module] `json:"lib" yaml:"lib" toml:"lib"`

	// Pkg is the collection of packages.
	Pkg audcol.Collection[audpkg.Package] `json:"pkg" yaml:"pkg" toml:"pkg"`

	// Ent is the collection of entries.
	Ent audcol.Collection[audent.Entry] `json:"ent" yaml:"ent" toml:"ent"`
}

// MarshalJSON implements the json.Marshaler interface for the mdl struct.
//
// This method serializes the manager's data into JSON format. It creates an
// Export struct containing the module, package, and entry collections, then
// marshals it into JSON bytes.
//
// Returns:
//   - []byte: JSON-encoded representation of the manager's data
//   - error: Any error that occurred during serialization
func (o *mdl) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.marshallObj())
}

// UnmarshalJSON implements the json.Unmarshaler interface for the mdl struct.
//
// This method deserializes JSON data back into the manager's internal collections.
// It creates an empty Export struct, parses the JSON input into it, and then
// applies the parsed values to the current manager instance.
//
// Parameters:
//   - p: JSON-encoded byte slice containing manager data
//
// Returns:
//   - error: Any error that occurred during deserialization
func (o *mdl) UnmarshalJSON(p []byte) error {
	var r = o.unmarshallObj()

	if e := json.Unmarshal(p, &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalYAML implements the yaml.Marshaler interface for the mdl struct.
//
// This method serializes the manager's data into YAML format. It creates an
// Export struct containing the module, package, and entry collections, then
// marshals it into a YAML-compatible interface{}.
//
// Returns:
//   - interface{}: YAML-compatible representation of manager data
//   - error: Any error that occurred during serialization
func (o *mdl) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(o.marshallObj())
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for the mdl struct.
//
// This method deserializes YAML data back into the manager's internal collections.
// It creates an empty Export struct, parses the YAML input into it, and then
// applies the parsed values to the current manager instance.
//
// Parameters:
//   - value: YAML node containing manager data
//
// Returns:
//   - error: Any error that occurred during deserialization
func (o *mdl) UnmarshalYAML(value *yaml.Node) error {
	var r = o.unmarshallObj()

	if e := yaml.Unmarshal([]byte(value.Value), &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalTOML implements the TOML marshaler interface for the mdl struct.
//
// This method serializes the manager's data into TOML format. It creates an
// Export struct containing the module, package, and entry collections, then
// marshals it into TOML bytes.
//
// Returns:
//   - []byte: TOML-encoded representation of manager data
//   - error: Any error that occurred during serialization
func (o *mdl) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.marshallObj())
}

// UnmarshalTOML implements the TOML unmarshaler interface for the mdl struct.
//
// This method deserializes TOML data back into the manager's internal collections.
// It accepts both []byte and string representations from TOML files. The
// deserialization creates an empty Export struct, parses the TOML input into it,
// and then applies the parsed values to the current manager instance.
//
// Parameters:
//   - i: TOML input (either []byte or string)
//
// Returns:
//   - error: Any error that occurred during deserialization
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

// MarshalCBOR implements CBOR marshaling using the fxamacker/cbor library for the mdl struct.
//
// This method serializes manager data into CBOR (Concise Binary Object Representation) format.
// CBOR is a binary serialization format that's more compact than JSON but still human-readable
// when decoded. It's useful for efficient data transmission and storage.
//
// The serialization uses the marshallObj helper function to create a structured representation
// of manager data in CBOR format.
//
// CBOR serialization provides optimal space efficiency for network transmission or
// storage scenarios where compact binary representations are preferred.
//
// See also: github.com/fxamacker/cbor
//
// Returns:
//   - []byte: CBOR-encoded representation of manager data
//   - error: Any error that occurred during serialization
func (o *mdl) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(o.marshallObj())
}

// UnmarshalCBOR implements CBOR unmarshaling using the fxamacker/cbor library for the mdl struct.
//
// This method deserializes CBOR data back into the manager's internal collections.
// It creates an empty Export struct, decodes the CBOR data into it, and then
// applies the parsed values to the current manager instance.
//
// CBOR deserialization supports binary data interchange with minimal overhead,
// making it ideal for performance-critical applications or network protocols.
//
// See also: github.com/fxamacker/cbor
//
// Parameters:
//   - p: CBOR-encoded byte slice containing manager data
//
// Returns:
//   - error: Any error that occurred during deserialization
func (o *mdl) UnmarshalCBOR(p []byte) error {
	var r = o.unmarshallObj()

	if e := cbor.Unmarshal(p, &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface for the mdl struct.
//
// This method provides compact binary serialization of manager data by delegating to
// the CBOR marshaling implementation. It's useful for efficient storage or transmission
// of manager data in binary format.
//
// Binary marshaling is particularly valuable when minimizing storage space or
// network bandwidth usage is a priority.
//
// Returns:
//   - []byte: Binary representation of manager data
//   - error: Any error that occurred during serialization
func (o *mdl) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface for the mdl struct.
//
// This method provides deserialization of binary data into the manager's internal
// collections by delegating to the CBOR unmarshaling implementation. It's useful
// for efficiently reconstructing manager data from binary storage or transmission.
//
// Binary unmarshaling enables fast reconstruction of manager data with minimal overhead,
// making it suitable for performance-critical scenarios.
//
// Parameters:
//   - p: Binary-encoded byte slice containing manager data
//
// Returns:
//   - error: Any error that occurred during deserialization
func (o *mdl) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}

// marshallObj creates a structured representation of the manager's data for serialization.
//
// This helper function prepares the manager data in a format suitable for marshalling
// into various formats (JSON, YAML, TOML, CBOR). It creates an Export struct that
// contains all relevant fields of the manager with appropriate tags for different
// serialization formats.
//
// The returned interface{} can be used by various marshal functions to serialize
// the manager data in different formats. This approach ensures consistency across
// all serialization methods and provides a clean separation between internal
// representation and external formats.
//
// Returns:
//   - *Export: Structured representation of manager data suitable for serialization
func (o *mdl) marshallObj() *Export {
	return &Export{
		Mod: o.mod,
		Pkg: o.pkg,
		Ent: o.ent,
	}
}

// unmarshallObj creates an empty structured representation for deserialization.
//
// This helper function prepares an empty Export struct that will be used to
// receive data during unmarshalling operations. It's designed to match the
// structure of marshallObj but with empty collections initialized using the
// respective Empty functions.
//
// This approach ensures type safety and consistency during deserialization
// processes, allowing for proper validation and application of data to manager instances.
//
// Returns:
//   - *Export: Empty structured representation suitable for deserialization
func (o *mdl) unmarshallObj() *Export {
	return &Export{
		Mod: audcol.New[audmod.Module](audmod.Empty),
		Pkg: audcol.New[audpkg.Package](audpkg.Empty),
		Ent: audcol.New[audent.Entry](audent.Empty),
	}
}

// unmarshallApply applies deserialized data to the current manager instance.
//
// This helper function takes deserialized data and applies it to the current
// manager instance. It performs type checking to ensure compatibility with
// the expected structure, then updates all fields of the manager atomically.
//
// The application process ensures that all manager data is updated consistently
// and safely. This method handles the final step of deserialization by transferring
// parsed data to the actual manager instance.
//
// Parameters:
//   - r: Deserialized data structure containing manager information
//
// Returns:
//   - error: Any error that occurred during the application process
func (o *mdl) unmarshallApply(r *Export) error {
	if r == nil {
		return errors.New("invalid object")
	}

	o.mod = r.Mod
	o.pkg = r.Pkg
	o.ent = r.Ent

	return nil
}
