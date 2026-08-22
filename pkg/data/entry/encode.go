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

// Package entry provides interfaces and utilities for managing code entry information.
//
// This package defines the Entry interface and related functions for handling
// Go code entry information including packages, names, types, source code, and dependencies.
// It supports various serialization formats (JSON, YAML, TOML, CBOR) and provides
// thread-safe access to entry data through mutex protection.
//
// The encode.go file contains methods for serializing and deserializing code entries
// in multiple formats including JSON, YAML, TOML, and CBOR. It implements the encoding
// interfaces and provides functionality for converting code entry data between
// internal representation and serialized formats.
package entry

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audprm "github.com/nabbar/auditor/pkg/data/params"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	audrep "github.com/nabbar/auditor/pkg/data/reports"
	audsrc "github.com/nabbar/auditor/pkg/data/src"
	audsts "github.com/nabbar/auditor/pkg/data/status"
	audtps "github.com/nabbar/auditor/pkg/data/types"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// Export represents the serialized structure of a code entry for encoding/decoding purposes.
//
// This struct is used as an intermediate representation during serialization and deserialization
// operations across different formats (JSON, YAML, TOML, CBOR). It contains all relevant
// properties of a code entry in a format suitable for serialization.
type Export struct {
	Id      audids.ID                `json:"id" yaml:"id" toml:"id"`
	Pkg     audids.ID                `json:"pkg" yaml:"pkg" toml:"pkg"`
	Name    string                   `json:"name" yaml:"name" toml:"name"`
	Code    audtps.CodeType          `json:"code" yaml:"code" toml:"code"`
	Type    string                   `json:"type" yaml:"type" toml:"type"`
	Src     any                      `json:"src" yaml:"src" toml:"src"`
	Lang    string                   `json:"lang" yaml:"lang" toml:"lang"`
	Inputs  []any                    `json:"inputs" yaml:"inputs" toml:"inputs"`
	Outputs []any                    `json:"outputs" yaml:"outputs" toml:"outputs"`
	Depends []audids.ID              `json:"depends" yaml:"depends" toml:"depends"`
	Status  audsts.Status            `json:"status" yaml:"status" toml:"status"`
	Sum     string                   `json:"summary" yaml:"summary" toml:"summary"`
	Rep     map[string]audrep.Report `json:"reports" yaml:"reports" toml:"reports"`
}

// Global function pointers for package and entry resolution during unmarshaling
var (
	fpg func(id audids.ID) audpkg.Package
	feg func(id audids.ID) Entry
)

// RegisterGetPackage registers a function to retrieve package information by ID.
//
// This function is used during deserialization to resolve package references from IDs.
// It should be called before any unmarshaling operations that require package resolution.
// The registered function will be used to convert package IDs back into package objects
// when reconstructing code entries from serialized data.
//
// Parameters:
//   - f: A function that takes an audids.ID and returns an audpkg.Package.
func RegisterGetPackage(f func(id audids.ID) audpkg.Package) {
	fpg = f
}

// RegisterGetEntry registers a function to retrieve entry information by ID.
//
// This function is used during deserialization to resolve dependency entries from IDs.
// It should be called before any unmarshaling operations that require entry resolution.
// The registered function will be used to convert dependency IDs back into entry objects
// when reconstructing code entries from serialized data.
//
// Parameters:
//   - f: A function that takes an audids.ID and returns an Entry.
func RegisterGetEntry(f func(id audids.ID) Entry) {
	feg = f
}

// MarshalJSON implements the json.Marshaler interface for code entry serialization.
//
// This function converts the code entry into a JSON representation that includes all
// relevant properties such as ID, package reference, name, code type, source, parameters,
// dependencies, status, summary, and reports. The marshaled output is designed to be
// human-readable and compatible with standard JSON processing tools.
//
// Example:
//
//	entry := New(...)
//	jsonData, err := json.Marshal(entry)
//	if err != nil {
//	    // handle error
//	}
func (o *mdl) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.marshallObj())
}

// UnmarshalJSON implements the json.Unmarshaler interface for code entry deserialization.
//
// This function parses JSON data into a code entry structure, reconstructing all properties
// including package references, dependencies, parameters, and reports. It handles the
// conversion of serialized data back into proper code entry objects with correct references.
//
// Example:
//
//	var entry Entry
//	err := json.Unmarshal(jsonData, &entry)
//	if err != nil {
//	    // handle error
//	}
func (o *mdl) UnmarshalJSON(p []byte) error {
	var r = o.unmarshallObj()

	if e := json.Unmarshal(p, &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalYAML implements the yaml.Marshaler interface for code entry serialization.
//
// This function converts the code entry into a YAML representation that includes all
// relevant properties. YAML format is particularly useful for configuration files and
// human-readable data exchange, providing better readability than JSON in many cases.
//
// Example:
//
//	entry := New(...)
//	yamlData, err := yaml.Marshal(entry)
//	if err != nil {
//	    // handle error
//	}
func (o *mdl) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(o.marshallObj())
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for code entry deserialization.
//
// This function parses YAML data into a code entry structure, reconstructing all properties
// including package references, dependencies, parameters, and reports. It handles the
// conversion of serialized data back into proper code entry objects with correct references.
//
// Example:
//
//	var entry Entry
//	err := yaml.Unmarshal(yamlData, &entry)
//	if err != nil {
//	    // handle error
//	}
func (o *mdl) UnmarshalYAML(value *yaml.Node) error {
	var r = o.unmarshallObj()

	if e := yaml.Unmarshal([]byte(value.Value), &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalTOML implements the TOML marshaler interface for code entry serialization.
//
// This function converts the code entry into a TOML representation that includes all
// relevant properties. TOML format is ideal for configuration files and provides
// clean, readable syntax that's easy to maintain.
//
// Example:
//
//	entry := New(...)
//	tomlData, err := toml.Marshal(entry)
//	if err != nil {
//	    // handle error
//	}
func (o *mdl) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.marshallObj())
}

// UnmarshalTOML implements the TOML unmarshaler interface for code entry deserialization.
//
// This function accepts both []byte and string representations from TOML files,
// parsing them into a code entry structure. It reconstructs all properties including
// package references, dependencies, parameters, and reports with proper reference resolution.
//
// Example:
//
//	var entry Entry
//	err := toml.Unmarshal(tomlData, &entry)
//	if err != nil {
//	    // handle error
//	}
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

// MarshalCBOR implements CBOR marshaling using the fxamacker/cbor library.
//
// This function serializes the code entry into CBOR (Concise Binary Object Representation)
// format, providing compact binary serialization suitable for efficient storage and
// transmission. CBOR is particularly useful for machine-to-machine communication.
//
// Example:
//
//	entry := New(...)
//	cborData, err := entry.MarshalCBOR()
//	if err != nil {
//	    // handle error
//	}
func (o *mdl) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(o.marshallObj())
}

// UnmarshalCBOR implements CBOR unmarshaling using the fxamacker/cbor library.
//
// This function deserializes CBOR data back into a code entry structure,
// reconstructing all properties including package references, dependencies,
// parameters, and reports with proper reference resolution.
//
// Example:
//
//	var entry Entry
//	err := entry.UnmarshalCBOR(cborData)
//	if err != nil {
//	    // handle error
//	}
func (o *mdl) UnmarshalCBOR(p []byte) error {
	var r = o.unmarshallObj()

	if e := cbor.Unmarshal(p, &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
//
// This method uses CBOR encoding internally to provide compact binary
// serialization of code entry data. It's particularly useful for efficient
// storage and network transmission when binary format is preferred.
//
// Example:
//
//	entry := New(...)
//	binaryData, err := entry.MarshalBinary()
//	if err != nil {
//	    // handle error
//	}
func (o *mdl) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
//
// This method uses CBOR decoding internally to deserialize binary data
// back into a code entry structure. It's particularly useful for efficient
// reconstruction of code entries from binary storage or network transmission.
//
// Example:
//
//	var entry Entry
//	err := entry.UnmarshalBinary(binaryData)
//	if err != nil {
//	    // handle error
//	}
func (o *mdl) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}

// marshallObj prepares a structured representation of the code entry for serialization.
//
// This helper function creates an Export structure that contains all relevant properties
// of the code entry in a format suitable for serialization across different formats.
// It handles conversion of dependencies from FuncEnt references to audids.ID values
// and ensures proper package reference resolution.
//
// Returns:
//   - *Export: A pointer to an Export struct containing the serialized representation of the code entry.
func (o *mdl) marshallObj() *Export {
	res := &Export{
		Id:      o.id,
		Pkg:     o.pk,
		Name:    o.nm,
		Code:    o.tc,
		Type:    o.ts,
		Src:     o.src,
		Lang:    o.lng,
		Inputs:  make([]any, len(o.inp)),
		Outputs: make([]any, len(o.out)),
		Depends: make([]audids.ID, len(o.dep)),
		Status:  o.sts,
		Sum:     o.sum,
		Rep:     o.rep,
	}

	copy(res.Depends, o.dep)

	for i := 0; i < len(o.inp); i++ {
		res.Inputs[i] = o.inp[i]
	}

	for i := 0; i < len(o.out); i++ {
		res.Outputs[i] = o.out[i]
	}

	return res
}

// unmarshallObj creates an empty Export structure for deserialization.
//
// This helper function prepares a structure that matches the serialization format
// to receive incoming data during deserialization. It provides the appropriate
// field types and tags needed for proper unmarshaling across different formats.
//
// Returns:
//   - *Export: A pointer to an empty Export struct ready to receive deserialized data.
func (o *mdl) unmarshallObj() *Export {
	return &Export{}
}

// unmarshallApply applies deserialized data to the code entry structure.
//
// This helper function processes the deserialized data by converting it back
// into the internal representation of the code entry. It handles the conversion
// of dependency IDs back into FuncEnt references using the registered resolution functions.
// The operation is thread-safe and protected by write-lock mutex during execution.
//
// The function performs the following steps:
//   - Validates that the input Export struct is not nil.
//   - Acquires a write lock on the mutex.
//   - Copies basic fields (ID, package ID, name, code type, type string, language, status, summary, reports).
//   - Initializes the source code field using CBOR marshaling/unmarshaling.
//   - Initializes input parameters using CBOR marshaling/unmarshaling for each parameter.
//   - Initializes output parameters using CBOR marshaling/unmarshaling for each parameter.
//
// Parameters:
//   - v: A pointer to an Export struct containing the deserialized data.
//
// Returns:
//   - error: An error if the input is nil, or if CBOR marshaling/unmarshaling fails.
func (o *mdl) unmarshallApply(v *Export) error {
	if v == nil {
		return errors.New("invalid object")
	}

	o.mx.Lock()
	defer o.mx.Unlock()

	o.id = v.Id
	o.pk = v.Pkg
	o.nm = v.Name
	o.tc = v.Code
	o.ts = v.Type
	o.src = audsrc.Empty()
	o.lng = v.Lang
	o.dep = make([]audids.ID, len(v.Depends))
	o.sts = v.Status
	o.sum = v.Sum
	o.rep = v.Rep

	copy(o.dep, v.Depends)

	var src = audsrc.Empty()
	if p, e := cbor.Marshal(v.Src); e != nil {
		return e
	} else if i, k := src.(cbor.Unmarshaler); !k {
		return fmt.Errorf("invalid type: %T", i)
	} else if e = i.UnmarshalCBOR(p); e != nil {
		return e
	}

	var lpr = make([]audprm.Params, len(v.Inputs))
	for i := 0; i < len(v.Inputs); i++ {
		var prm = audprm.Empty()

		if p, e := cbor.Marshal(v.Inputs[i]); e != nil {
			return e
		} else if a, k := prm.(cbor.Unmarshaler); !k {
			return fmt.Errorf("invalid type: %T", i)
		} else if e = a.UnmarshalCBOR(p); e != nil {
			return e
		}

		lpr[i] = prm
	}

	copy(o.inp, lpr)

	lpr = make([]audprm.Params, len(v.Outputs))
	for i := 0; i < len(v.Outputs); i++ {
		var prm = audprm.Empty()

		if p, e := cbor.Marshal(v.Outputs[i]); e != nil {
			return e
		} else if a, k := prm.(cbor.Unmarshaler); !k {
			return fmt.Errorf("invalid type: %T", i)
		} else if e = a.UnmarshalCBOR(p); e != nil {
			return e
		}

		lpr[i] = prm
	}

	copy(o.out, lpr)

	return nil
}
