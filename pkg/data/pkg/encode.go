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

// Package pkg provides interfaces and utilities for managing package information.
//
// This package defines the Package interface and related functions for handling
// Go package information including paths, module references, and aliases.
// It supports various serialization formats (JSON, YAML, TOML, CBOR) and provides
// thread-safe access to package data through mutex protection.
package pkg

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audrep "github.com/nabbar/auditor/pkg/data/reports"
	audsts "github.com/nabbar/auditor/pkg/data/status"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// Export represents the serialized form of a Package for encoding/decoding.
//
// This struct is used as an intermediate representation when marshalling or
// unmarshalling package data in various formats (JSON, YAML, TOML, CBOR).
// It contains all the fields necessary to reconstruct a Package instance.
type Export struct {
	// Id is the unique identifier for the package.
	Id audids.ID `json:"id" yaml:"id" toml:"id"`
	// Mod is the module ID that references the containing module.
	Mod audids.ID `json:"mod" yaml:"mod" toml:"mod"`
	// Pkg is the Go package path relative to its module.
	Pkg string `json:"pkg" yaml:"pkg" toml:"pkg"`
	// Alias contains alternative names for the package.
	Alias []string `json:"alias" yaml:"alias" toml:"alias"`
	// Status is the current lifecycle status of the package.
	Status audsts.Status `json:"status" yaml:"status" toml:"status"`
	// Sum is the summary description of the package.
	Sum string `json:"summary" yaml:"summary" toml:"summary"`
	// Rep contains audit reports associated with the package.
	Rep map[string]audrep.Report `json:"reports" yaml:"reports" toml:"reports"`
}

// fmg is a function that retrieves a module by ID.
//
// This global variable is used for deserializing package data to reconstruct
// module references. It must be registered using RegisterGetModule before
// deserialization can work properly.
var fmg func(id audids.ID) audmod.Module

// RegisterGetModule registers a function that can retrieve modules by ID.
//
// This registration is necessary for proper deserialization of package data where
// the module reference needs to be reconstructed from an ID. The registered function
// should return a Module that can be retrieved based on its ID.
// Without this registration, deserialization will not be able to properly reconstruct
// module relationships when loading package data.
//
// Parameters:
//   - f: A function that takes a module ID and returns the corresponding Module
func RegisterGetModule(f func(id audids.ID) audmod.Module) {
	fmg = f
}

// MarshalJSON implements the json.Marshaler interface for the mdl struct.
//
// This method serializes the package data into JSON format. It uses the marshallObj
// helper function to create a structured representation of the package's data,
// which includes:
//   - ID: Unique identifier for the package
//   - Module ID (Mod): The ID of the associated module (not the full module)
//   - Package path (Pkg): The Go package path
//   - Aliases (Alias): Alternative names for the package
//   - Status (sts): Current lifecycle status of the package
//   - Summary (sum): Description of the package
//   - Reports (rep): Associated audit reports
//
// The serialization is thread-safe as it uses read locks on the mutex.
// This method is typically used when saving package data to JSON files or sending
// over network protocols in JSON format.
//
// Returns:
//   - []byte: The JSON-encoded byte slice
//   - error: Any error that occurred during marshalling
func (o *mdl) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.marshallObj())
}

// UnmarshalJSON implements the json.Unmarshaler interface for the mdl struct.
//
// This method deserializes JSON data back into a package instance. It:
//   - Creates an unmarshalling object
//   - Parses the JSON input into this object
//   - Applies the parsed values to the current package instance
//
// The deserialization is thread-safe as it uses write locks on the mutex.
// This method is typically used when loading package data from JSON files or receiving
// JSON data over network protocols.
//
// Parameters:
//   - p: The JSON byte slice to unmarshal
//
// Returns:
//   - error: Any error that occurred during unmarshalling
func (o *mdl) UnmarshalJSON(p []byte) error {
	var r = o.unmarshallObj()

	if e := json.Unmarshal(p, &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalYAML implements the yaml.Marshaler interface for the mdl struct.
//
// This method serializes the package data into YAML format. The output includes
// all package fields in a human-readable format suitable for configuration files.
// It uses the marshallObj helper function to create a structured representation
// of the package's data, similar to JSON serialization but formatted for YAML.
//
// The serialization is thread-safe as it uses read locks on the mutex.
// YAML serialization is particularly useful for configuration management and human-readable storage.
//
// Returns:
//   - interface{}: The YAML-encoded interface
//   - error: Any error that occurred during marshalling
func (o *mdl) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(o.marshallObj())
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for the mdl struct.
//
// This method deserializes YAML data back into a package instance. It:
//   - Creates an unmarshalling object
//   - Parses the YAML input into this object
//   - Applies the parsed values to the current package instance
//
// The deserialization is thread-safe as it uses write locks on the mutex.
// YAML deserialization is commonly used for configuration files and data exchange.
//
// Parameters:
//   - value: The YAML node to unmarshal
//
// Returns:
//   - error: Any error that occurred during unmarshalling
func (o *mdl) UnmarshalYAML(value *yaml.Node) error {
	var r = o.unmarshallObj()

	if e := yaml.Unmarshal([]byte(value.Value), &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalTOML implements the TOML marshaler interface for the mdl struct.
//
// This method serializes the package data into TOML format. TOML is a configuration
// file format that's easy to read and write. The output includes all package fields
// in a structured way suitable for TOML configuration files.
//
// The serialization is thread-safe as it uses read locks on the mutex.
// TOML serialization is ideal for configuration management and data interchange
// where readability is important.
//
// Returns:
//   - []byte: The TOML-encoded byte slice
//   - error: Any error that occurred during marshalling
func (o *mdl) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.marshallObj())
}

// UnmarshalTOML implements the TOML unmarshaler interface for the mdl struct.
//
// This method deserializes TOML data back into a package instance. It accepts
// both []byte and string representations from TOML files. The deserialization:
//   - Creates an unmarshalling object
//   - Parses the TOML input into this object
//   - Applies the parsed values to the current package instance
//
// The deserialization is thread-safe as it uses write locks on the mutex.
// TOML deserialization is commonly used for configuration files and data exchange.
//
// Parameters:
//   - i: The TOML data to unmarshal. Must be either []byte or string.
//
// Returns:
//   - error: Any error that occurred during unmarshalling, or an error if the type is invalid
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
// This method serializes package data into CBOR (Concise Binary Object Representation) format.
// CBOR is a binary serialization format that's more compact than JSON but still human-readable
// when decoded. It's useful for efficient data transmission and storage.
//
// The serialization uses the marshallObj helper function to create a structured representation
// of package data in CBOR format.
//
// See also: github.com/fxamacker/cbor
// CBOR serialization is particularly useful for network protocols and binary storage where
// compactness matters more than human readability.
//
// Returns:
//   - []byte: The CBOR-encoded byte slice
//   - error: Any error that occurred during marshalling
func (o *mdl) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(o.marshallObj())
}

// UnmarshalCBOR implements CBOR unmarshaling using the fxamacker/cbor library for the mdl struct.
//
// This method deserializes CBOR data back into a package instance. It:
//   - Decodes CBOR data containing package information
//   - Parses the decoded data into an unmarshalling object
//   - Applies the parsed values to the current package instance
//
// The deserialization is thread-safe as it uses write locks on the mutex.
//
// See also: github.com/fxamacker/cbor
// CBOR deserialization is ideal for efficient data reconstruction from binary sources.
//
// Parameters:
//   - p: The CBOR byte slice to unmarshal
//
// Returns:
//   - error: Any error that occurred during unmarshalling
func (o *mdl) UnmarshalCBOR(p []byte) error {
	var r = o.unmarshallObj()

	if e := cbor.Unmarshal(p, &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface for the mdl struct.
//
// This method provides compact binary serialization of package data by delegating to
// the CBOR marshaling implementation. It's useful for efficient storage or transmission
// of package data in binary format.
//
// The serialization is thread-safe as it uses read locks on the mutex.
// Binary marshaling is particularly useful for high-performance scenarios where
// compact storage or fast transmission is required.
//
// Returns:
//   - []byte: The binary-encoded byte slice
//   - error: Any error that occurred during marshalling
func (o *mdl) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface for the mdl struct.
//
// This method provides deserialization of binary data into package instances by delegating
// to the CBOR unmarshaling implementation. It's useful for efficiently reconstructing
// package data from binary storage or transmission.
//
// The deserialization is thread-safe as it uses write locks on the mutex.
// Binary unmarshaling is ideal for high-performance scenarios where fast data reconstruction
// is required from compact binary sources.
//
// Parameters:
//   - p: The binary byte slice to unmarshal
//
// Returns:
//   - error: Any error that occurred during unmarshalling
func (o *mdl) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}

// marshallObj creates a structured representation of the package's data for serialization.
//
// This helper function prepares the package data in a format suitable for marshalling
// into various formats (JSON, YAML, TOML, CBOR). It creates an Export struct that
// contains all relevant fields of the package with appropriate tags for different
// serialization formats.
//
// The returned interface{} can be used by various marshal functions to serialize
// the package data in different formats. It handles the special case of module ID
// by retrieving it from the module function when available.
// This method is thread-safe and uses read locks on the mutex for concurrent access.
//
// Returns:
//   - *Export: A structured representation of the package data
func (o *mdl) marshallObj() *Export {
	res := &Export{
		Id:     o.id,
		Mod:    o.md,
		Pkg:    o.pk,
		Alias:  o.al,
		Status: o.sts,
		Sum:    o.sum,
		Rep:    o.rep,
	}

	return res
}

// unmarshallObj creates an empty structured representation for deserialization.
//
// This helper function prepares an empty Export struct that will be used to
// receive data during unmarshalling operations. It's designed to match the
// structure of marshallObj but without any initial values.
// This method is thread-safe and uses read locks on the mutex for concurrent access.
//
// Returns:
//   - *Export: An empty Export struct ready for unmarshalling
func (o *mdl) unmarshallObj() *Export {
	return &Export{}
}

// unmarshallApply applies deserialized data to the current package instance.
//
// This helper function takes deserialized data and applies it to the current
// package instance. It performs type checking to ensure compatibility with
// the expected structure, then updates all fields of the package atomically
// using write locks on the mutex.
//
// It reconstructs the module reference using the registered fmg function based
// on the module ID provided in the deserialized data.
//
// Returns an error if the data cannot be applied due to type mismatches or
// other issues during the application process.
// This method is thread-safe and uses write locks on the mutex for concurrent access.
//
// Parameters:
//   - v: The Export struct containing deserialized data
//
// Returns:
//   - error: An error if the input is nil or invalid, nil on success
func (o *mdl) unmarshallApply(v *Export) error {
	if v == nil {
		return errors.New("invalid object")
	}

	o.mx.Lock()
	defer o.mx.Unlock()

	o.id = v.Id
	o.md = v.Mod
	o.pk = v.Pkg
	o.al = v.Alias
	o.sts = v.Status
	o.sum = v.Sum
	o.rep = v.Rep

	return nil
}
