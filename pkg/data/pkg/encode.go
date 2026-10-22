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

// File encode.go contains the serialization and deserialization of the Package
// interface.
//
// It defines the Export intermediate structure shared by every format, the
// module resolver used to rebuild the reference to the containing module, and
// implements the json, yaml, toml and cbor marshaler and unmarshaler
// contracts of the mdl model.

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

// RegisterGetModule registers a function that can retrieve the modules by
// their identifier.
//
// This registration is required to resolve the module that contains a
// package: the identifier of the module is stored in the package, and the
// module itself is resolved on demand through the registered resolver. It is
// also required by the Merge method, and it is used by the New constructor.
//
// The registered function is expected to return the module that matches the
// given identifier, or nil when the module is unknown.
//
// Note:
//   - The resolver is a package level variable that is not protected by any
//     mutex: it is expected to be registered once, during the initialization
//     of the application.
//   - When no resolver has been registered, the calls to New and to Merge
//     panic on the nil function pointer.
//
// Parameters:
//   - f: A function that takes a module identifier and returns the matching
//     module.
func RegisterGetModule(f func(id audids.ID) audmod.Module) {
	fmg = f
}

// MarshalJSON implements the json.Marshaler interface for the mdl struct.
//
// This method serializes the data of the package into the JSON format. It uses
// the marshallObj helper function to build a structured representation of the
// data of the package, which includes:
//   - The unique identifier of the package.
//   - The identifier of the containing module, and not the module itself.
//   - The path of the package.
//   - The aliases of the package.
//   - The lifecycle status of the package.
//   - The summary of the package.
//   - The reports of the package.
//
// Thread safety: The fields of the package are read without holding the
// mutex. A concurrent modification of the package is therefore a data race,
// and the caller has to serialize the call with the modifications by itself.
//
// This method is typically used when the data of the package is saved to a
// JSON file, or sent over the network in the JSON format.
//
// Returns:
//   - []byte: The JSON-encoded byte slice.
//   - error: Any error that occurred during the marshalling.
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
// This method serializes the data of the package into the YAML format. The
// output includes all the fields of the package in a human-readable format,
// which is suitable for the configuration files. It uses the marshallObj
// helper function to build a structured representation of the data of the
// package, similar to the JSON serialization but formatted for the YAML.
//
// Note:
//   - The returned value is the encoded YAML document, and not the Export
//     value itself: the YAML encoder therefore produces a YAML document that
//     holds the byte representation of the package.
//
// Thread safety: The fields of the package are read without holding the
// mutex. A concurrent modification of the package is therefore a data race,
// and the caller has to serialize the call with the modifications by itself.
//
// Returns:
//   - interface{}: The YAML-encoded interface.
//   - error: Any error that occurred during the marshalling.
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
// This method serializes the data of the package into the TOML format. The
// TOML format is a configuration file format that is easy to read and to
// write. The output includes all the fields of the package in a structured
// way, which is suitable for the TOML configuration files.
//
// Thread safety: The fields of the package are read without holding the
// mutex. A concurrent modification of the package is therefore a data race,
// and the caller has to serialize the call with the modifications by itself.
//
// Returns:
//   - []byte: The TOML-encoded byte slice.
//   - error: Any error that occurred during the marshalling.
func (o *mdl) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.marshallObj())
}

// UnmarshalTOML implements the TOML unmarshaler interface for the mdl struct.
//
// This method deserializes the TOML data back into a package instance. It
// accepts both the []byte and the string representations that come from the
// TOML files. The deserialization:
//   - Creates an unmarshalling object.
//   - Parses the TOML input into this object.
//   - Applies the parsed values to the current package instance.
//
// Thread safety: The deserialization is thread-safe, as it uses the write lock
// of the mutex.
//
// Parameters:
//   - i: The TOML data to unmarshal. It must be either a []byte or a string.
//
// Returns:
//   - error: Any error that occurred during the unmarshalling, or an error if
//     the type of the input is invalid.
//
// Note:
//   - The type switch of the implementation overwrites its own result, so the
//     []byte input is currently rejected with an "invalid type" error, and the
//     string input is rejected because the destination is not a pointer. Both
//     cases are reported to the caller through the returned error.
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

// MarshalCBOR implements the CBOR marshaling with the fxamacker/cbor library
// for the mdl struct.
//
// This method serializes the data of the package into the CBOR (Concise Binary
// Object Representation) format. The CBOR format is a binary serialization
// format that is more compact than the JSON, while it stays human-readable
// once decoded.
//
// The method uses the marshallObj helper function to build a structured
// representation of the data of the package.
//
// See also: github.com/fxamacker/cbor
//
// Thread safety: The fields of the package are read without holding the
// mutex. A concurrent modification of the package is therefore a data race,
// and the caller has to serialize the call with the modifications by itself.
//
// Returns:
//   - []byte: The CBOR-encoded byte slice.
//   - error: Any error that occurred during the marshalling.
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

// MarshalBinary implements the encoding.BinaryMarshaler interface for the
// mdl struct.
//
// This method provides a compact binary serialization of the data of the
// package by delegating to the CBOR marshaling implementation.
//
// Thread safety: The fields of the package are read without holding the
// mutex. A concurrent modification of the package is therefore a data race,
// and the caller has to serialize the call with the modifications by itself.
//
// Returns:
//   - []byte: The binary-encoded byte slice.
//   - error: Any error that occurred during the marshalling.
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

// marshallObj creates a structured representation of the data of the package
// for the serialization.
//
// This helper function prepares the data of the package in a format suitable
// for the marshalling into the various formats (JSON, YAML, TOML, CBOR). It
// creates an Export struct that holds all the relevant fields of the package,
// with the appropriate tags for the different serialization formats.
//
// The identifier of the containing module is stored as it is: the module
// itself is not resolved during the marshalling.
//
// Thread safety: The fields of the package are read without holding the
// mutex. A concurrent modification of the package is therefore a data race,
// and the caller has to serialize the call with the modifications by itself.
//
// Returns:
//   - *Export: A structured representation of the data of the package.
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

// unmarshallObj creates an empty structured representation for the
// deserialization.
//
// This helper function prepares an empty Export struct that is used to
// receive the data during the unmarshalling operations. It is designed to
// match the structure of marshallObj, but without any initial value.
//
// The receiver is not read, and the mutex is therefore not acquired.
//
// Returns:
//   - *Export: An empty Export struct, ready for the unmarshalling.
func (o *mdl) unmarshallObj() *Export {
	return &Export{}
}

// unmarshallApply applies the deserialized data to the current package
// instance.
//
// This helper function takes the deserialized data and applies it to the
// current package instance. It updates all the fields of the package
// atomically, under the write lock of the mutex.
//
// The identifier of the containing module is stored as it is: the module
// itself is resolved on demand, through the resolver that has been registered
// with RegisterGetModule.
//
// Parameters:
//   - v: The Export struct containing the deserialized data.
//
// Returns:
//   - error: An error if the input is nil, nil on success.
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
