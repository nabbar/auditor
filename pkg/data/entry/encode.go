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

// File encode.go contains the serialization and deserialization of the Entry
// interface.
//
// It defines the Export intermediate structure shared by every format, and
// implements the json, yaml, toml and cbor marshaler and unmarshaler contracts
// of the mdl model.

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

// Export represents the serialized structure of a code entry.
//
// This struct is used as the intermediate representation during the
// serialization and the deserialization operations, across the different
// formats (JSON, YAML, TOML, CBOR). It holds all the relevant properties of
// a code entry in a format that is suitable for the serialization.
//
// Fields:
//   - Id: the unique identifier of the code entry.
//   - Pkg: the identifier of the package that contains the code entry.
//   - Name: the name of the code element.
//   - Code: the code type of the code element.
//   - Type: the type string of the code element.
//   - Src: the content of the source code.
//   - Lang: the identifier of the language of the source code.
//   - Inputs: the input parameters of the code element.
//   - Outputs: the output parameters of the code element.
//   - Depends: the identifiers of the dependencies of the code element.
//   - Status: the lifecycle status of the code element.
//   - Sum: the summary of the code element.
//   - Rep: the reports of the code element, keyed by their name.
type Export struct {
	// Id is the unique identifier of the code entry.
	Id audids.ID `json:"id" yaml:"id" toml:"id"`

	// Pkg is the identifier of the package that contains the code entry.
	Pkg audids.ID `json:"pkg" yaml:"pkg" toml:"pkg"`

	// Name is the name of the code element.
	Name string `json:"name" yaml:"name" toml:"name"`

	// Code is the code type of the code element.
	Code audtps.CodeType `json:"code" yaml:"code" toml:"code"`

	// Type is the type string of the code element.
	Type string `json:"type" yaml:"type" toml:"type"`

	// Src is the content of the source code.
	Src any `json:"src" yaml:"src" toml:"src"`

	// Lang is the identifier of the language of the source code.
	Lang string `json:"lang" yaml:"lang" toml:"lang"`

	// Inputs holds the input parameters of the code element.
	Inputs []any `json:"inputs" yaml:"inputs" toml:"inputs"`

	// Outputs holds the output parameters of the code element.
	Outputs []any `json:"outputs" yaml:"outputs" toml:"outputs"`

	// Depends holds the identifiers of the dependencies of the code element.
	Depends []audids.ID `json:"depends" yaml:"depends" toml:"depends"`

	// Status is the lifecycle status of the code element.
	Status audsts.Status `json:"status" yaml:"status" toml:"status"`

	// Sum is the summary of the code element.
	Sum string `json:"summary" yaml:"summary" toml:"summary"`

	// Rep holds the reports of the code element, keyed by their name.
	Rep map[string]audrep.Report `json:"reports" yaml:"reports" toml:"reports"`
}

// Global function pointers for package and entry resolution during unmarshaling
var (
	fpg func(id audids.ID) audpkg.Package
	feg func(id audids.ID) Entry
)

// RegisterGetPackage registers a function that retrieves the packages by
// their identifier.
//
// This registration is required to resolve the package that contains a code
// entry: the identifier of the package is stored in the entry, and the package
// itself is resolved on demand through the registered resolver. It is also
// required by the New constructor.
//
// The registered function is expected to return the package that matches the
// given identifier, or nil when the package is unknown.
//
// Note:
//   - The resolver is a package level variable that is not protected by any
//     mutex: it is expected to be registered once, during the initialization
//     of the application.
//   - When no resolver has been registered, the calls to New and to GetPkg
//     panic on the nil function pointer.
//
// Parameters:
//   - f: A function that takes an audids.ID and returns the matching
//     audpkg.Package.
func RegisterGetPackage(f func(id audids.ID) audpkg.Package) {
	fpg = f
}

// RegisterGetEntry registers a function that retrieves the code entries by
// their identifier.
//
// This registration is required to resolve the dependencies of a code entry:
// the dependencies are stored as identifiers, and they are resolved on demand
// through the registered resolver.
//
// The registered function is expected to return the code entry that matches
// the given identifier, or nil when the entry is unknown.
//
// Note:
//   - The resolver is a package level variable that is not protected by any
//     mutex: it is expected to be registered once, during the initialization
//     of the application.
//   - When no resolver has been registered, the call to GetDepend panics on
//     the nil function pointer.
//
// Parameters:
//   - f: A function that takes an audids.ID and returns the matching Entry.
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

// MarshalYAML implements the yaml.Marshaler interface for the serialization
// of a code entry.
//
// This method converts the code entry into a YAML representation that includes
// all the relevant properties. The YAML format is particularly useful for the
// configuration files and for the human readable data exchanges, because it
// provides a better readability than the JSON in many cases.
//
// Note:
//   - The returned value is the encoded YAML document, and not the Export
//     value itself: the YAML encoder therefore produces a YAML document that
//     holds the byte representation of the code entry.
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

// marshallObj prepares a structured representation of the code entry for the
// serialization.
//
// This helper function creates an Export structure that holds all the relevant
// properties of the code entry, in a format that is suitable for the
// serialization across the different formats.
//
// The dependencies are converted from the internal slice of identifiers to the
// slice of identifiers of the Export, and the parameters are copied as they
// are.
//
// Thread safety: The fields of the code entry are read without holding the
// mutex. A concurrent modification of the entry is therefore a data race, and
// the caller has to serialize the call with the modifications by itself.
//
// Returns:
//   - *Export: A pointer to an Export struct that holds the serialized
//     representation of the code entry.
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

// unmarshallApply applies the deserialized data to the code entry.
//
// This helper function processes the deserialized data, and converts it back
// into the internal representation of the code entry.
//
// The function performs the following steps:
//   - Validates that the input Export struct is not nil.
//   - Acquires the write lock of the mutex.
//   - Copies the basic fields: the identifier, the identifier of the package,
//     the name, the code type, the type string, the language, the status, the
//     summary and the reports.
//   - Copies the identifiers of the dependencies.
//   - Decodes the content of the source code with the CBOR encoding.
//   - Decodes the input and the output parameters with the CBOR encoding.
//
// The operation is thread-safe and protected by the write lock of the mutex.
//
// Note:
//   - The dependencies are kept as identifiers: they are not resolved during
//     the deserialization, but on demand, through the resolver that has been
//     registered with RegisterGetEntry.
//
// Parameters:
//   - v: A pointer to an Export struct that holds the deserialized data.
//
// Returns:
//   - error: An error if the input is nil, or if the CBOR marshaling or
//     unmarshaling fails.
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
