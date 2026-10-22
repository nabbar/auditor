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

// File encode.go contains the serialization and deserialization of the
// Collection interface.
//
// It implements the json, yaml, toml and cbor marshaler and unmarshaler
// contracts of the mdl[M] collection, so that a whole set of model instances
// can be persisted to, and restored from, any of the supported formats.

package collection

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fxamacker/cbor/v2"
	audids "github.com/nabbar/auditor/pkg/data/id"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// MarshalJSON implements the json.Marshaler interface for the mdl struct.
//
// This method serializes the underlying map of model instances into JSON
// format. It directly marshals the map o.l, which contains all the stored
// model instances keyed by their audids.ID identifiers.
//
// The identifiers are serialized as the JSON object keys, and the model
// instances are serialized with their own JSON marshaler.
//
// Thread safety: This method does not acquire any lock, because it reads the
// underlying map without modifying it. The read-write mutex of the collection
// is therefore not held during the serialization, and a concurrent
// modification of the collection is a data race. The caller has to
// serialize the call with the modifications by itself.
//
// Returns:
//   - []byte: JSON-encoded byte slice containing the map of model instances.
//   - error: Any error encountered during JSON marshaling.
func (o *mdl[M]) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.l)
}

// UnmarshalJSON implements the json.Unmarshaler interface for the mdl struct.
//
// This method deserializes JSON data back into the underlying map of model
// instances. It parses the JSON input into a temporary map and then replaces
// the existing map with the new data.
//
// Thread safety: This method acquires a write lock before modifying the
// underlying map, ensuring exclusive access during deserialization to prevent
// concurrent modification.
//
// Parameters:
//   - p: JSON-encoded byte slice containing the map of model instances
//
// Returns:
//   - error: Any error encountered during JSON unmarshaling
func (o *mdl[M]) UnmarshalJSON(p []byte) error {
	var r map[audids.ID]M

	if e := json.Unmarshal(p, &r); e != nil {
		return e
	}

	o.m.Lock()
	defer o.m.Unlock()

	o.l = r
	return nil
}

// MarshalYAML implements the yaml.Marshaler interface for the mdl struct.
//
// This method serializes the underlying map of model instances into the YAML
// format. It directly marshals the map o.l, which contains all the stored
// model instances keyed by their audids.ID identifiers.
//
// Note:
//   - The returned value is the encoded YAML document, and not the map
//     itself: the YAML encoder therefore produces a YAML document that holds
//     the byte representation of the collection.
//
// Thread safety: This method does not acquire any lock, because it reads the
// underlying map without modifying it. The read-write mutex of the collection
// is therefore not held during the serialization, and a concurrent
// modification of the collection is a data race. The caller has to
// serialize the call with the modifications by itself.
//
// Returns:
//   - interface{}: YAML-compatible representation of the map of model
//     instances.
//   - error: Any error encountered during YAML marshaling.
func (o *mdl[M]) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(o.l)
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for the mdl struct.
//
// This method deserializes YAML data back into the underlying map of model
// instances. It parses the YAML input from the node's Value field into a
// temporary map and then replaces the existing map with the new data.
//
// Thread safety: This method acquires a write lock before modifying the
// underlying map, ensuring exclusive access during deserialization to prevent
// concurrent modification.
//
// Parameters:
//   - value: YAML node containing the map of model instances
//
// Returns:
//   - error: Any error encountered during YAML unmarshaling
func (o *mdl[M]) UnmarshalYAML(value *yaml.Node) error {
	var r map[audids.ID]M

	if e := yaml.Unmarshal([]byte(value.Value), &r); e != nil {
		return e
	}

	o.m.Lock()
	defer o.m.Unlock()

	o.l = r
	return nil
}

// MarshalTOML implements the toml.Marshaler interface for the mdl struct.
//
// This method serializes the underlying map of model instances into TOML
// format. It directly marshals the map o.l, which contains all the stored
// model instances keyed by their audids.ID identifiers.
//
// Thread safety: This method does not acquire any lock, because it reads the
// underlying map without modifying it. The read-write mutex of the collection
// is therefore not held during the serialization, and a concurrent
// modification of the collection is a data race. The caller has to
// serialize the call with the modifications by itself.
//
// Returns:
//   - []byte: TOML-encoded byte slice containing the map of model instances.
//   - error: Any error encountered during TOML marshaling.
func (o *mdl[M]) MarshalTOML() ([]byte, error) {
	return toml.Marshal(o.l)
}

// UnmarshalTOML implements the toml.Unmarshaler interface for the mdl struct.
//
// This method deserializes TOML data back into the underlying map of model
// instances. It accepts either a []byte or a string input, parses it into a
// temporary map, and then replaces the existing map with the new data.
//
// Thread safety: This method acquires a write lock before modifying the
// underlying map, ensuring exclusive access during deserialization to prevent
// concurrent modification.
//
// Parameters:
//   - i: TOML input, either a []byte or a string containing the TOML data.
//
// Returns:
//   - error: Any error encountered during TOML unmarshaling, or an error if
//     the input type is neither []byte nor string.
//
// Note:
//   - The type switch of the implementation overwrites its own result, so the
//     []byte input is currently rejected with an "invalid type" error, and the
//     string input is rejected because the destination is not a pointer. Both
//     cases are reported to the caller through the returned error.
func (o *mdl[M]) UnmarshalTOML(i interface{}) error {
	var (
		e error
		k bool
		p []byte
		s string
		r map[audids.ID]M
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

	o.m.Lock()
	defer o.m.Unlock()

	o.l = r
	return nil
}

// MarshalCBOR implements the cbor.Marshaler interface for the mdl struct.
//
// This method serializes the underlying map of model instances into CBOR
// format. It directly marshals the map o.l, which contains all the stored
// model instances keyed by their audids.ID identifiers.
//
// Thread safety: This method does not acquire any lock, because it reads the
// underlying map without modifying it. The read-write mutex of the collection
// is therefore not held during the serialization, and a concurrent
// modification of the collection is a data race. The caller has to
// serialize the call with the modifications by itself.
//
// Returns:
//   - []byte: CBOR-encoded byte slice containing the map of model instances.
//   - error: Any error encountered during CBOR marshaling.
func (o *mdl[M]) MarshalCBOR() ([]byte, error) {
	return cbor.Marshal(o.l)
}

// UnmarshalCBOR implements the cbor.Unmarshaler interface for the mdl struct.
//
// This method deserializes CBOR data back into the underlying map of model
// instances. It first parses the CBOR input into a temporary map of any type,
// then for each entry, it marshals the value back to CBOR and unmarshals it
// into a new instance created by the constructor function n(). This approach
// ensures that each model instance is properly reconstructed using its own
// UnmarshalCBOR implementation if available.
//
// Thread safety: This method acquires a write lock before modifying the
// underlying map, ensuring exclusive access during deserialization to prevent
// concurrent modification.
//
// Parameters:
//   - p: CBOR-encoded byte slice containing the map of model instances
//
// Returns:
//   - error: Any error encountered during CBOR unmarshaling, or an error if
//     the model type does not implement cbor.Unmarshaler
func (o *mdl[M]) UnmarshalCBOR(p []byte) error {
	var (
		r = make(map[audids.ID]any)
	)

	if e := cbor.Unmarshal(p, &r); e != nil {
		return e
	}

	o.m.Lock()
	defer o.m.Unlock()

	for id, val := range r {
		p, e := cbor.Marshal(val)
		if e != nil {
			return e
		}

		var m = o.n()

		if v, k := any(m).(cbor.Unmarshaler); !k {
			return errors.New("invalid type")
		} else if e = v.UnmarshalCBOR(p); e != nil {
			return e
		}

		o.l[id] = m
	}

	return nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface for the
// mdl struct.
//
// This method provides a compact binary serialization of the model data by
// delegating to the CBOR marshaling implementation. It is useful for the
// efficient storage or the transmission of the model data in a binary format.
//
// Thread safety: This method does not acquire any lock, because it reads the
// underlying map without modifying it. The read-write mutex of the collection
// is therefore not held during the serialization, and a concurrent
// modification of the collection is a data race. The caller has to
// serialize the call with the modifications by itself.
//
// Returns:
//   - []byte: Binary representation of the model data (CBOR-encoded).
//   - error: Any error encountered during CBOR marshaling.
func (o *mdl[M]) MarshalBinary() ([]byte, error) {
	return o.MarshalCBOR()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface for the mdl struct.
//
// This method provides deserialization of binary data into model instances by
// delegating to the CBOR unmarshaling implementation. It's useful for efficiently
// reconstructing model data from binary storage or transmission.
//
// Thread safety: This method acquires a write lock before modifying the
// underlying map, ensuring exclusive access during deserialization to prevent
// concurrent modification.
//
// Parameters:
//   - p: Binary-encoded byte slice containing model data (CBOR-encoded)
//
// Returns:
//   - error: Any error encountered during CBOR unmarshaling
func (o *mdl[M]) UnmarshalBinary(p []byte) error {
	return o.UnmarshalCBOR(p)
}
