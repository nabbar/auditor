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

// Package params provides data structures and utilities for managing parameter configurations.
//
// This package contains the concrete implementation of the Params interface defined in the
// interface.go file. It provides a model structure that implements all required methods
// for handling parameter configurations including serialization, comparison, and
// descriptive information management.
package params

import (
	"fmt"

	audtps "github.com/nabbar/auditor/pkg/data/types"
)

// mdl represents a parameter configuration model that stores code type, type string, and summary information.
//
// This struct implements the Params interface and serves as the concrete implementation
// for parameter configuration management. It holds three key pieces of information:
//   - tc: The code type (audtps.CodeType) that categorizes the parameter configuration
//   - ts: The specific type string that provides additional context for the code type
//   - sm: The summary or description that explains what the configuration represents
//
// The mdl struct is designed to be used internally by the package and is not exposed
// directly to external consumers of the API.
type mdl struct {
	// nm represent the params name if set
	nm string

	// tc represents the code type associated with this parameter configuration.
	//
	// This field stores the categorization of the parameter configuration using
	// the audtps.CodeType enum values. It determines what kind of parameter configuration
	// this instance represents (e.g., error, warning, info, etc.).
	tc audtps.CodeType

	// ts represents the specific type string for the code type.
	//
	// This field provides additional contextual information about the code type.
	// For some code types, a type string may be required to fully specify the configuration.
	ts string

	// sm represents the summary or description of this parameter configuration.
	//
	// This field contains a human-readable description that explains what this
	// parameter configuration represents and its purpose within the system.
	sm string

	// inp represents the list of input parameter configurations.
	//
	// This field stores the parameter configurations that serve as inputs to this
	// parameter configuration. These are typically upstream dependencies or sources
	// of data that feed into this configuration.
	inp []Params

	// out represents the list of output parameter configurations.
	//
	// This field stores the parameter configurations that serve as outputs from this
	// parameter configuration. These are typically downstream consumers or destinations
	// of data produced by this configuration.
	out []Params
}

// IsEmpty checks if the parameter configuration is empty.
//
// Returns true if the formatted type string representation is empty (length < 1),
// indicating that no meaningful configuration has been set. This method is useful
// for validation purposes to ensure configurations are properly initialized.
func (o *mdl) IsEmpty() bool {
	return len(o.GetTypeString()) < 1
}

// IsEqual compares this parameter configuration with another for equality.
//
// Returns true if both configurations have identical type and type string values,
// effectively comparing their string representations. This method is useful for
// determining whether two parameter configurations are equivalent.
//
// Parameters:
//   - f: The Params instance to compare against.
//
// Returns:
//   - bool: true if the string representations are identical, false otherwise.
func (o *mdl) IsEqual(f Params) bool {
	return o.String() == f.String()
}

// SetName sets the name of this parameter configuration.
//
// Parameters:
//   - s: The name string to assign to this parameter configuration.
func (o *mdl) SetName(s string) {
	o.nm = s
}

// GetName retrieves the name of this parameter configuration.
//
// Returns:
//   - string: The current name of the parameter configuration.
func (o *mdl) GetName() string {
	return o.nm
}

// SetSummary updates the summary description of this parameter configuration.
//
// This method allows modification of the descriptive information associated with
// the parameter configuration. The summary provides context about what the
// configuration represents and its purpose within the system.
//
// Parameters:
//   - p: The summary description string to assign.
func (o *mdl) SetSummary(p string) {
	o.sm = p
}

// GetSummary retrieves the summary description of this parameter configuration.
//
// Returns the current summary string that describes the parameter configuration.
// This method allows access to the descriptive information associated with the configuration.
//
// Returns:
//   - string: The current summary description of the parameter configuration.
func (o *mdl) GetSummary() string {
	return o.sm
}

// GetType retrieves the code type and its associated type string.
//
// Returns the code type (audtps.CodeType) and type string (string) values.
// The code type represents the category or classification of the parameter configuration,
// while the type string provides additional contextual information about the type.
//
// Returns:
//   - audtps.CodeType: The code type representing the category or classification.
//   - string: The type string providing additional contextual information.
func (o *mdl) GetType() (audtps.CodeType, string) {
	return o.tc, o.ts
}

// GetTypeString generates a formatted string representation of the code type and its associated type string.
//
// If the code type requires a type string, it formats as "CodeType: TypeString",
// otherwise it just returns CodeType.String().
// This method provides a standardized way to represent parameter configurations in a readable format,
// making it suitable for logging, debugging, and user interfaces.
//
// Returns:
//   - string: The formatted string representation of the code type and type string.
func (o *mdl) GetTypeString() string {
	if c, s := o.GetType(); c.NeedString() {
		return fmt.Sprintf("%s: %s", c.String(), s)
	} else {
		return c.String()
	}
}

// String returns a string representation of this parameter configuration.
//
// It uses the formatted type string as the primary representation. This method
// satisfies the fmt.Stringer interface and provides a convenient way to convert
// parameter configurations to their string form for display or logging purposes.
//
// Returns:
//   - string: The string representation of the parameter configuration.
func (o *mdl) String() string {
	return o.GetTypeString()
}

// GetInput retrieves the list of input parameter configurations.
//
// Returns a copy of the input parameter configurations slice, ensuring that
// modifications to the returned slice do not affect the internal state.
// This method is useful for accessing the input parameters without risking
// unintended side effects.
//
// Returns:
//   - []Params: A copy of the slice containing input parameter configurations.
func (o *mdl) GetInput() []Params {
	n := make([]Params, len(o.inp))
	copy(n, o.inp)
	return n
}

// SetInput sets the list of input parameter configurations.
//
// This method appends the provided parameter configurations to the existing input list.
// If the current input list is empty, it initializes a new slice with capacity equal to
// the number of provided parameters. If the remaining capacity is insufficient, it
// extends the slice by creating a new slice with increased capacity.
// Nil parameters are ignored and not added to the list.
//
// Parameters:
//   - p: A variadic list of Params instances to add as input parameters.
func (o *mdl) SetInput(p ...Params) {
	if len(o.inp) < 1 {
		o.inp = make([]Params, 0, len(p))
	}

	if (cap(o.inp) - len(o.inp)) < len(p) {
		// extend capacity
		n := make([]Params, len(o.inp), len(o.inp)+len(p))
		copy(n, o.inp)
		o.inp = n
	}

	for _, i := range p {
		if i != nil {
			o.inp = append(o.inp, i)
		}
	}
}

// DelInput removes all input parameter configurations.
//
// This method clears all input parameter configurations by clearing the slice
// and setting it to nil. After calling this method, the input list will be empty.
func (o *mdl) DelInput() {
	clear(o.inp)
	o.inp = nil
}

// GetOutput retrieves the list of output parameter configurations.
//
// Returns a copy of the output parameter configurations slice, ensuring that
// modifications to the returned slice do not affect the internal state.
// This method is useful for accessing the output parameters without risking
// unintended side effects.
//
// Returns:
//   - []Params: A copy of the slice containing output parameter configurations.
func (o *mdl) GetOutput() []Params {
	n := make([]Params, len(o.out))
	copy(n, o.out)
	return n
}

// SetOutput sets the list of output parameter configurations.
//
// This method appends the provided parameter configurations to the existing output list.
// If the current output list is empty, it initializes a new slice with capacity equal to
// the number of provided parameters. If the remaining capacity is insufficient, it
// extends the slice by creating a new slice with increased capacity.
// Nil parameters are ignored and not added to the list.
//
// Parameters:
//   - p: A variadic list of Params instances to add as output parameters.
func (o *mdl) SetOutput(p ...Params) {
	if len(o.out) < 1 {
		o.out = make([]Params, 0, len(p))
	}

	if (cap(o.out) - len(o.out)) < len(p) {
		// extend capacity
		n := make([]Params, len(o.out), len(o.out)+len(p))
		copy(n, o.out)
		o.out = n
	}

	for _, i := range p {
		if i != nil {
			o.out = append(o.out, i)
		}
	}
}

// DelOutput removes all output parameter configurations.
//
// This method clears all output parameter configurations by clearing the slice
// and setting it to nil. After calling this method, the output list will be empty.
func (o *mdl) DelOutput() {
	clear(o.out)
	o.out = nil
}
