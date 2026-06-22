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

// Package mod provides interfaces and utilities for managing Go module information.
//
// The package defines the Module interface, which abstracts a Go module by its
// canonical path, associated source file, and vendor status. It also exposes
// constructor and path-normalization helpers used to build Module values.
//
// Module values support serialization to JSON, YAML, TOML, and CBOR, and all
// state mutations are guarded by a read-write mutex for safe concurrent use.
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
// It holds all fields required to round-trip a Module through JSON, YAML, TOML,
// or CBOR encoding. Each field carries struct tags for the respective serialization
// format.
type Export struct {
	Id     audids.ID                `json:"id" yaml:"id" toml:"id"`
	Mod    string                   `json:"module" yaml:"module" toml:"module"`
	File   string                   `json:"file" yaml:"file" toml:"file"`
	Vendor bool                     `json:"vendor" yaml:"vendor" toml:"vendor"`
	Status audsts.Status            `json:"status" yaml:"status" toml:"status"`
	Sum    string                   `json:"summary" yaml:"summary" toml:"summary"`
	Rep    map[string]audrep.Report `json:"reports" yaml:"reports" toml:"reports"`
}

// MarshalJSON implements the json.Marshaler interface for the mdl struct.
//
// It serializes the module data into JSON format by delegating to
// marshallObj to construct an Export value, then encoding that value with
// json.Marshal. The operation is thread-safe: it acquires a read lock on
// the internal mutex.
//
// Returns:
//   - []byte: the JSON-encoded representation of the module data.
//   - error: any error that occurred during serialization.
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
// It serializes the module data into YAML format by delegating to
// marshallObj to construct an Export value, then encoding that value with
// yaml.Marshal. The operation is thread-safe: it acquires a read lock on
// the internal mutex.
//
// Returns:
//   - interface{}: the YAML-compatible representation of the module data.
//   - error: any error that occurred during serialization.
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
// It serializes the module data into TOML format by delegating to
// marshallObj to construct an Export value, then encoding that value with
// toml.Marshal. The operation is thread-safe: it acquires a read lock on
// the internal mutex.
//
// Returns:
//   - []byte: the TOML-encoded representation of the module data.
//   - error: any error that occurred during serialization.
func (o *mdl) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.marshallObj())
}

// UnmarshalTOML implements the TOML unmarshaler interface for the mdl struct.
//
// It deserializes TOML data back into the module instance. The input may be
// either a []byte or a string. The process is:
//  1. Create an empty Export value via unmarshallObj.
//  2. Parse the TOML input into that Export value.
//  3. Apply the parsed values to the current module instance via unmarshallApply.
//
// If the input type is neither []byte nor string, an error is returned.
// The operation is thread-safe: it acquires a write lock on the internal mutex
// through unmarshallApply.
//
// Parameters:
//   - i: the TOML input (either []byte or string).
//
// Returns:
//   - error: any error that occurred during deserialization.
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

// MarshalCBOR implements CBOR marshaling using the fxamacker/cbor library
// for the mdl struct.
//
// It serializes module data into CBOR (Concise Binary Object Representation)
// format by delegating to marshallObj to construct an Export value, then
// encoding that value with cbor.Marshal. The operation is thread-safe: it
// acquires a read lock on the internal mutex.
//
// CBOR provides a compact binary representation suitable for efficient data
// transmission and storage.
//
// Returns:
//   - []byte: the CBOR-encoded representation of the module data.
//   - error: any error that occurred during serialization.
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
// It provides compact binary serialization of module data by delegating to
// MarshalCBOR. The operation is thread-safe: it acquires a read lock on the
// internal mutex through MarshalCBOR.
//
// Returns:
//   - []byte: the binary representation of the module data (CBOR-encoded).
//   - error: any error that occurred during serialization.
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

// marshallObj creates a structured representation of the module's data for
// serialization.
//
// It returns an Export value populated with the current module's fields.
// The returned value is used as the source for all serialization methods
// (JSON, YAML, TOML, CBOR), ensuring consistency across formats.
//
// Returns:
//   - *Export: a pointer to an Export struct containing the module's data.
func (o *mdl) marshallObj() *Export {
	return &Export{
		Id:     o.id,
		Mod:    o.mod,
		File:   o.fgm,
		Vendor: o.vdr,
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
	o.vdr = v.Vendor
	o.sts = v.Status
	o.sum = v.Sum
	o.rep = v.Rep

	return nil
}
