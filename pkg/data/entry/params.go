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
// The params.go file contains methods for managing input parameters, output parameters,
// and dependencies of code entries, providing functionality for manipulating parameter lists
// and dependency relationships within the audit system.
package entry

import (
	"slices"

	audids "github.com/nabbar/auditor/pkg/data/id"
	audprm "github.com/nabbar/auditor/pkg/data/params"
)

// GetInputs returns the input parameters for this code entry.
//
// This function retrieves all input parameter definitions associated with this code element,
// providing access to function arguments or required parameters for execution.
// The returned slice is thread-safe and protected by read-lock mutex during access.
//
// Returns:
//   - []audprm.Params: A slice of input parameters for this code entry.
func (o *mdl) GetInputs() []audprm.Params {
	o.mx.RLock()
	defer o.mx.RUnlock()
	return o.inp
}

// GetOutputs returns the output parameters for this code entry.
//
// This function retrieves all output parameter definitions associated with this code element,
// providing access to return values or results produced by execution of this code element.
// The returned slice is thread-safe and protected by read-lock mutex during access.
//
// Returns:
//   - []audprm.Params: A slice of output parameters for this code entry.
func (o *mdl) GetOutputs() []audprm.Params {
	o.mx.RLock()
	defer o.mx.RUnlock()
	return o.out
}

// GetDepend returns the dependencies of this code entry.
//
// Returns a slice of Entry instances that this entry depends on, resolving function references
// to their actual Entry objects through the dependency functions.
// This function handles nil checks and properly resolves dependencies by executing the FuncEnt functions.
// The returned slice is thread-safe and protected by read-lock mutex during access.
//
// Returns:
//   - []Entry: A slice of Entry instances representing the dependencies of this code entry.
func (o *mdl) GetDepend() []Entry {

	var (
		l = o.GetDependIds()
		r = make([]Entry, len(l))
	)

	if len(l) == 0 {
		return r
	}

	for i := range l {
		if l[i] == 0 {
			continue
		} else if v := feg(l[i]); v == nil {
			continue
		} else {
			r = append(r, v)
		}
	}

	return r
}

// GetDependIds returns the IDs of the dependencies of this code entry.
//
// This function retrieves the IDs of the dependencies of this code entry.
// The returned slice is thread-safe and protected by read-lock mutex during access.
//
// Returns:
//   - []audids.ID: A slice of IDs representing the dependencies of this code entry.
func (o *mdl) GetDependIds() []audids.ID {
	o.mx.RLock()
	defer o.mx.RUnlock()
	p := make([]audids.ID, len(o.dep))
	copy(p, o.dep)
	return p
}

// AddInputs adds new input parameters to this code entry.
//
// This function appends new parameter definitions to the existing input parameters slice,
// allowing for dynamic addition of input specifications to this code element.
// It ensures thread-safe access to the input parameters through write-lock protection.
//
// Parameters:
//   - f: One or more input parameters to add.
func (o *mdl) AddInputs(f ...audprm.Params) {
	o.mx.Lock()
	defer o.mx.Unlock()

	for i := 0; i < len(f); i++ {
		if f[i] == nil || len(f[i].String()) < 1 {
			continue
		}
		o.inp = o.addParams(o.inp, f[i])
	}
}

// AddOutputs adds new output parameters to this code entry.
//
// This function appends new parameter definitions to the existing output parameters slice,
// allowing for dynamic addition of output specifications to this code element.
// It ensures thread-safe access to the output parameters through write-lock protection.
//
// Parameters:
//   - f: One or more output parameters to add.
func (o *mdl) AddOutputs(f ...audprm.Params) {
	o.mx.Lock()
	defer o.mx.Unlock()

	for i := 0; i < len(f); i++ {
		if f[i] == nil || len(f[i].String()) < 1 {
			continue
		}
		o.out = o.addParams(o.out, f[i])
	}
}

// AddDepend adds new dependencies to this code entry.
//
// This function appends new function dependency references to the existing dependencies slice,
// establishing relationships between code elements and their required dependencies.
// It ensures thread-safe access to the dependencies through write-lock protection.
//
// Parameters:
//   - f: One or more dependency IDs to add.
func (o *mdl) AddDepend(f ...audids.ID) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if o.dep == nil {
		o.dep = make([]audids.ID, 0)
	}

	for i := range f {
		if f[i] != 0 {
			o.dep = append(o.dep, f[i])
		}
	}
}

// DelInputs clears all input parameters from this code entry.
//
// This function removes all input parameter definitions, resetting the input parameters
// to an empty slice for this code element.
// It ensures thread-safe access to the input parameters through write-lock protection.
func (o *mdl) DelInputs() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.inp = make([]audprm.Params, 0)
}

// DelOutputs clears all output parameters from this code entry.
//
// This function removes all output parameter definitions, resetting the output parameters
// to an empty slice for this code element.
// It ensures thread-safe access to the output parameters through write-lock protection.
func (o *mdl) DelOutputs() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.out = make([]audprm.Params, 0)
}

// DelDepend removes specific dependencies by ID from this code entry.
//
// If no IDs are provided, it clears all dependencies associated with this code element.
// This function filters out dependencies that match the specified IDs, preserving
// remaining dependencies and maintaining the integrity of dependency relationships.
// It ensures thread-safe access to the dependencies through write-lock protection.
//
// Parameters:
//   - id: One or more dependency IDs to remove. If empty, all dependencies are removed.
func (o *mdl) DelDepend(id ...audids.ID) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if o.dep == nil {
		return
	}

	if len(id) < 1 {
		o.dep = make([]audids.ID, 0)
		return
	}

	var n = make([]audids.ID, 0, len(o.dep))

	for j := range o.dep {
		if o.dep[j] == 0 {
			continue
		}

		if !slices.Contains(id, o.dep[j]) {
			n = append(n, o.dep[j])
		}
	}

	o.dep = n
}

// addParams adds new parameters to an existing slice of parameters.
//
// This helper function combines existing parameters with new ones, ensuring proper
// handling of nil values and maintaining parameter integrity during addition operations.
// It returns a new slice containing both the original and newly added parameters.
//
// Parameters:
//   - l: The existing slice of parameters.
//   - f: One or more new parameters to add.
//
// Returns:
//   - []audprm.Params: A new slice containing the original and newly added parameters.
func (o *mdl) addParams(l []audprm.Params, f ...audprm.Params) []audprm.Params {
	var res = make([]audprm.Params, len(l), len(l)+len(f))

	copy(res, l)

	for i := range f {
		if f[i] != nil {
			res = append(res, f[i])
		}
	}

	return res
}
