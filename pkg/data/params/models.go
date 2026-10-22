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

// File models.go contains the concrete implementation of the Params
// interface.
//
// The mdl type stores the name, the code type, the type string, the summary
// and the input and output parameters of a parameter set.

package params

import (
	"fmt"

	audtps "github.com/nabbar/auditor/pkg/data/types"
)

// mdl represents a parameter configuration model.
//
// This struct implements the Params interface and serves as the concrete
// implementation for the parameter configuration management. It holds:
//   - nm: The name of the parameter configuration.
//   - tc: The code type (audtps.CodeType) that categorizes the parameter
//     configuration.
//   - ts: The specific type string that provides an additional context for
//     the code type.
//   - sm: The summary or the description that explains what the configuration
//     represents.
//   - inp: The list of the input parameter configurations.
//   - out: The list of the output parameter configurations.
//
// The mdl struct is designed to be used internally by the package and is not
// exposed directly to the external consumers of the API.
type mdl struct {
	// nm represents the name of the parameter configuration, when it is set.
	//
	// The name is optional: a parameter may be described by its type only.
	nm string

	// tc represents the code type associated with this parameter
	// configuration.
	//
	// This field stores the categorization of the parameter configuration
	// using the audtps.CodeType enum values. It determines what kind of
	// parameter configuration this instance represents.
	tc audtps.CodeType

	// ts represents the specific type string for the code type.
	//
	// This field provides an additional contextual information about the code
	// type. For some code types, a type string is required to fully specify
	// the configuration.
	ts string

	// sm represents the summary or the description of this parameter
	// configuration.
	//
	// This field contains a human readable description that explains what this
	// parameter configuration represents and what its purpose is within the
	// system.
	sm string

	// inp represents the list of the input parameter configurations.
	//
	// This field stores the parameter configurations that serve as inputs to
	// this parameter configuration. These are typically the upstream
	// dependencies or the sources of data that feed into this configuration.
	inp []Params

	// out represents the list of the output parameter configurations.
	//
	// This field stores the parameter configurations that serve as outputs from
	// this parameter configuration. These are typically the downstream
	// consumers or the destinations of the data produced by this
	// configuration.
	out []Params
}

// IsEmpty checks if the parameter configuration is empty.
//
// Returns true if the formatted type string representation is empty (length
// < 1), indicating that no meaningful configuration has been set. This method
// is useful for the validation purposes, to ensure that the configurations are
// properly initialized.
//
// Returns:
//   - bool: true when the configuration carries no usable type information.
func (o *mdl) IsEmpty() bool {
	return len(o.GetTypeString()) < 1
}

// IsEqual compares this parameter configuration with another one for
// equality.
//
// Returns true if both configurations have an identical type and an identical
// type string, which is what the comparison of their string representations
// checks. This method is useful to determine whether two parameter
// configurations are equivalent.
//
// Parameters:
//   - f: The Params instance to compare against.
//
// Returns:
//   - bool: true if the string representations are identical, false
//     otherwise.
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

// SetSummary updates the summary description of this parameter
// configuration.
//
// This method allows the modification of the descriptive information
// associated with the parameter configuration. The summary provides a context
// about what the configuration represents and about its purpose within the
// system.
//
// Parameters:
//   - p: The summary description string to assign.
func (o *mdl) SetSummary(p string) {
	o.sm = p
}

// GetSummary retrieves the summary description of this parameter
// configuration.
//
// Returns the current summary string that describes the parameter
// configuration. This method allows an access to the descriptive information
// associated with the configuration.
//
// Returns:
//   - string: The current summary description of the parameter configuration.
func (o *mdl) GetSummary() string {
	return o.sm
}

// GetType retrieves the code type and its associated type string.
//
// Returns the code type (audtps.CodeType) and the type string (string) values.
// The code type represents the category or the classification of the
// parameter configuration, while the type string provides an additional
// contextual information about the type.
//
// Returns:
//   - audtps.CodeType: The code type representing the category or the
//     classification.
//   - string: The type string providing additional contextual information.
func (o *mdl) GetType() (audtps.CodeType, string) {
	return o.tc, o.ts
}

// GetTypeString generates a formatted string representation of the code
// type and of its associated type string.
//
// If the code type requires a type string, the result is formatted as
// "CodeType: TypeString", otherwise it is just the CodeType.String()
// representation. This method provides a standardized way to represent the
// parameter configurations in a readable format, making it suitable for the
// logging, the debugging and the user interfaces.
//
// Returns:
//   - string: The formatted string representation of the code type and of the
//     type string.
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
// satisfies the fmt.Stringer interface and provides a convenient way to
// convert the parameter configurations to their string form for the display
// or the logging purposes.
//
// Returns:
//   - string: The string representation of the parameter configuration.
func (o *mdl) String() string {
	return o.GetTypeString()
}

// GetInput retrieves the list of the input parameter configurations.
//
// Returns a copy of the input parameter configurations slice, ensuring that
// the modifications of the returned slice do not affect the internal state.
// This method is useful to access the input parameters without risking
// unintended side effects.
//
// Returns:
//   - []Params: A copy of the slice containing the input parameter
//     configurations.
func (o *mdl) GetInput() []Params {
	n := make([]Params, len(o.inp))
	copy(n, o.inp)
	return n
}

// SetInput sets the list of the input parameter configurations.
//
// This method appends the provided parameter configurations to the existing
// input list. If the current input list is empty, it initializes a new slice
// with a capacity equal to the number of the provided parameters. If the
// remaining capacity is insufficient, it extends the slice by creating a new
// slice with an increased capacity. The nil parameters are ignored and are
// not added to the list.
//
// Parameters:
//   - p: A variadic list of Params instances to add as the input parameters.
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

// DelInput removes all the input parameter configurations.
//
// This method clears all the input parameter configurations by clearing the
// slice and by setting it to nil. After calling this method, the input list
// is empty and the configuration does not hold any reference to the removed
// parameters anymore.
func (o *mdl) DelInput() {
	clear(o.inp)
	o.inp = nil
}

// GetOutput retrieves the list of the output parameter configurations.
//
// Returns a copy of the output parameter configurations slice, ensuring that
// the modifications of the returned slice do not affect the internal state.
// This method is useful to access the output parameters without risking
// unintended side effects.
//
// Returns:
//   - []Params: A copy of the slice containing the output parameter
//     configurations.
func (o *mdl) GetOutput() []Params {
	n := make([]Params, len(o.out))
	copy(n, o.out)
	return n
}

// SetOutput sets the list of the output parameter configurations.
//
// This method appends the provided parameter configurations to the existing
// output list. If the current output list is empty, it initializes a new
// slice with a capacity equal to the number of the provided parameters. If
// the remaining capacity is insufficient, it extends the slice by creating a
// new slice with an increased capacity. The nil parameters are ignored and
// are not added to the list.
//
// Parameters:
//   - p: A variadic list of Params instances to add as the output parameters.
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

// DelOutput removes all the output parameter configurations.
//
// This method clears all the output parameter configurations by clearing the
// slice and by setting it to nil. After calling this method, the output list
// is empty and the configuration does not hold any reference to the removed
// parameters anymore.
func (o *mdl) DelOutput() {
	clear(o.out)
	o.out = nil
}
