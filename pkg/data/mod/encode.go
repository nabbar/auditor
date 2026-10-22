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

// File encode.go contains the serialization and deserialization of the Module
// interface.
//
// It defines the Export intermediate structure shared by every format, and
// implements the json, yaml, toml and cbor marshaler and unmarshaler contracts
// of the mdl model.

package mod

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audrep "github.com/nabbar/auditor/pkg/data/reports"
	audsts "github.com/nabbar/auditor/pkg/data/status"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// Export is the serialization struct for a Module.
//
// It holds all the fields required to round-trip a Module through the JSON,
// YAML, TOML or CBOR encoding. Each field carries the struct tags of the
// corresponding serialization format.
//
// Fields:
//   - Id: the unique identifier of the module.
//   - Mod: the path of the module.
//   - File: the path of the file associated with the module.
//   - Vendor: the vendor status of the module.
//   - Status: the lifecycle status of the module.
//   - Sum: the summary of the module.
//   - Rep: the reports of the module, keyed by their name.
type Export struct {
	// Id is the unique identifier of the module.
	Id audids.ID `json:"id" yaml:"id" toml:"id"`

	// Mod is the path of the module.
	Mod string `json:"module" yaml:"module" toml:"module"`

	// File is the path of the file associated with the module.
	File string `json:"file" yaml:"file" toml:"file"`

	// Vendor reports whether the module is a vendored dependency.
	Vendor bool `json:"vendor" yaml:"vendor" toml:"vendor"`

	// Status is the lifecycle status of the module.
	Status audsts.Status `json:"status" yaml:"status" toml:"status"`

	// Sum is the summary of the module.
	Sum string `json:"summary" yaml:"summary" toml:"summary"`

	// Rep holds the reports of the module, keyed by their name.
	Rep map[string]audrep.Report `json:"reports" yaml:"reports" toml:"reports"`
}

// MarshalJSON implements the json.Marshaler interface for the mdl struct.
//
// It serializes the module data into the JSON format by delegating to
// marshallObj to build an Export value, and then encoding that value with
// json.Marshal.
//
// Thread safety: The fields of the module are read without holding the
// mutex, except for the vendor status, which is computed by IsVendor under
// the read lock. A concurrent modification of the module is therefore a data
// race, and the caller has to serialize the call with the modifications by
// itself.
//
// Returns:
//   - []byte: the JSON-encoded representation of the module data.
//   - error: any error that occurred during the serialization.
func (o *mdl) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.marshallObj())
}

// UnmarshalJSON implements the json.Unmarshaler interface for the mdl struct.
//
// It deserializes JSON data back into the module instance. The process is:
//  1. Create an empty Export value via unmarshallObj.
//  2. Parse the JSON input into that Export value.
//  3. Apply the parsed values to the current module instance via unmarshallApply.
//
// The operation is thread-safe: it acquires a write lock on the internal mutex
// through unmarshallApply.
//
// Parameters:
//   - p: the JSON-encoded byte slice containing module data.
//
// Returns:
//   - error: any error that occurred during deserialization.
func (o *mdl) UnmarshalJSON(p []byte) error {
	var r = o.unmarshallObj()

	if e := json.Unmarshal(p, &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalYAML implements the yaml.Marshaler interface for the mdl struct.
//
// It serializes the module data into the YAML format by delegating to
// marshallObj to build an Export value, and then encoding that value with
// yaml.Marshal.
//
// The returned value is the encoded YAML document, and not the Export value
// itself: the YAML encoder therefore produces a YAML document that holds the
// byte representation of the module.
//
// Thread safety: The fields of the module are read without holding the
// mutex, except for the vendor status, which is computed by IsVendor under
// the read lock. A concurrent modification of the module is therefore a data
// race, and the caller has to serialize the call with the modifications by
// itself.
//
// Returns:
//   - interface{}: the YAML-compatible representation of the module data.
//   - error: any error that occurred during the serialization.
func (o *mdl) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(o.marshallObj())
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for the mdl struct.
//
// It deserializes YAML data back into the module instance. The process is:
//  1. Create an empty Export value via unmarshallObj.
//  2. Parse the YAML input into that Export value.
//  3. Apply the parsed values to the current module instance via unmarshallApply.
//
// The operation is thread-safe: it acquires a write lock on the internal mutex
// through unmarshallApply.
//
// Parameters:
//   - value: the YAML node containing module data.
//
// Returns:
//   - error: any error that occurred during deserialization.
func (o *mdl) UnmarshalYAML(value *yaml.Node) error {
	var r = o.unmarshallObj()

	if e := yaml.Unmarshal([]byte(value.Value), &r); e != nil {
		return e
	}

	return o.unmarshallApply(r)
}

// MarshalTOML implements the TOML marshaler interface for the mdl struct.
//
// It serializes the module data into the TOML format by delegating to
// marshallObj to build an Export value, and then encoding that value with
// toml.Marshal.
//
// Thread safety: The fields of the module are read without holding the
// mutex, except for the vendor status, which is computed by IsVendor under
// the read lock. A concurrent modification of the module is therefore a data
// race, and the caller has to serialize the call with the modifications by
// itself.
//
// Returns:
//   - []byte: the TOML-encoded representation of the module data.
//   - error: any error that occurred during the serialization.
func (o *mdl) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.marshallObj())
}

// UnmarshalTOML implements the TOML unmarshaler interface for the mdl struct.
//
// It deserializes the TOML data back into the module instance. The input may
// be either a []byte or a string. The process is:
//  1. Create an empty Export value via unmarshallObj.
//  2. Parse the TOML input into that Export value.
//  3. Apply the parsed values to the current module instance via
//     unmarshallApply.
//
// If the input type is neither []byte nor string, an error is returned.
//
// Thread safety: The write lock of the internal mutex is acquired by
// unmarshallApply, for the whole duration of the update of the module.
//
// Parameters:
//   - i: the TOML input, either a []byte or a string.
//
// Returns:
//   - error: any error that occurred during the deserialization.
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
// It serializes the module data into the CBOR (Concise Binary Object
// Representation) format by delegating to marshallObj to build an Export
// value, and then encoding that value with cbor.Marshal.
//
// The CBOR format provides a compact binary representation, which is suitable
// for the efficient transmission and the storage of the data.
//
// Thread safety: The fields of the module are read without holding the
// mutex, except for the vendor status, which is computed by IsVendor under
// the read lock. A concurrent modification of the module is therefore a data
// race, and the caller has to serialize the call with the modifications by
// itself.
//
// Returns:
//   - []byte: the CBOR-encoded representation of the module data.
//   - error: any error that occurred during the serialization.
func (o *mdl) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(o.marshallObj())
}

// UnmarshalCBOR implements CBOR unmarshaling using the fxamacker/cbor library
// for the mdl struct.
//
// It deserializes CBOR data back into the module instance. The process is:
//  1. Create an empty Export value via unmarshallObj.
//  2. Decode the CBOR input into that Export value.
//  3. Apply the parsed values to the current module instance via unmarshallApply.
//
// The operation is thread-safe: it acquires a write lock on the internal mutex
// through unmarshallApply.
//
// Parameters:
//   - p: the CBOR-encoded byte slice containing module data.
//
// Returns:
//   - error: any error that occurred during deserialization.
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
// It provides a compact binary serialization of the module data by delegating
// to MarshalCBOR.
//
// Thread safety: The fields of the module are read without holding the
// mutex, except for the vendor status, which is computed by IsVendor under
// the read lock. A concurrent modification of the module is therefore a data
// race, and the caller has to serialize the call with the modifications by
// itself.
//
// Returns:
//   - []byte: the binary representation of the module data (CBOR-encoded).
//   - error: any error that occurred during the serialization.
func (o *mdl) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface for
// the mdl struct.
//
// It provides deserialization of binary data into module instances by
// delegating to UnmarshalCBOR. The operation is thread-safe: it acquires
// a write lock on the internal mutex through UnmarshalCBOR.
//
// Parameters:
//   - p: the binary-encoded byte slice containing module data (CBOR-encoded).
//
// Returns:
//   - error: any error that occurred during deserialization.
func (o *mdl) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}

// marshallObj creates a structured representation of the data of the module
// for the serialization.
//
// It returns an Export value populated with the current fields of the module.
// The returned value is used as the source of all the serialization methods
// (JSON, YAML, TOML, CBOR), which ensures the consistency between the
// formats.
//
// The fields are read without holding the mutex, except for the vendor
// status, which is computed by IsVendor under the read lock.
//
// Returns:
//   - *Export: a pointer to an Export struct containing the data of the
//     module.
func (o *mdl) marshallObj() *Export {
	return &Export{
		Id:     o.id,
		Mod:    o.mod,
		File:   o.fgm,
		Vendor: o.IsVendor(),
		Status: o.sts,
		Sum:    o.sum,
		Rep:    o.rep,
	}
}

// unmarshallObj creates an empty Export struct for deserialization.
//
// It returns a pointer to a zero-valued Export struct that will be used as
// the target for unmarshalling operations.
//
// Returns:
//   - *Export: a pointer to an empty Export struct suitable for deserialization.
func (o *mdl) unmarshallObj() *Export {
	return &Export{}
}

// unmarshallApply applies deserialized data to the current module instance.
//
// It takes an Export value containing deserialized data and writes all fields
// into the module instance under a write lock. If the supplied Export is nil,
// an error is returned.
//
// Parameters:
//   - v: the deserialized Export struct containing module information.
//
// Returns:
//   - error: nil on success, or an error if v is nil.
func (o *mdl) unmarshallApply(v *Export) error {
	if v == nil {
		return errors.New("invalid object")
	}

	o.mx.Lock()
	defer o.mx.Unlock()

	o.id = v.Id
	o.mod = v.Mod
	o.fgm = v.File
	o.sts = v.Status
	o.sum = v.Sum
	o.rep = v.Rep

	return nil
}
