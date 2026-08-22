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
package entry

import (
	"encoding/json"
	"path/filepath"
	"sync"

	"github.com/fxamacker/cbor/v2"
	auddat "github.com/nabbar/auditor/pkg/data"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audprm "github.com/nabbar/auditor/pkg/data/params"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	audsrc "github.com/nabbar/auditor/pkg/data/src"
	audsts "github.com/nabbar/auditor/pkg/data/status"
	audtps "github.com/nabbar/auditor/pkg/data/types"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// FuncEnt represents a function type that creates and returns a new Entry instance.
//
// This function type is typically used as a constructor for creating Entry instances
// in a factory pattern. It allows for flexible instantiation of Entry implementations
// by providing a way to create package instances when needed.
type FuncEnt func() Entry

// Entry represents an interface for handling code entry information.
//
// The Entry interface extends the base Data interface and provides methods for managing:
// - Package references (linking entries to their containing packages)
// - Entry names (identifiers for code elements)
// - Code types and associated type strings
// - Source code content
// - Dependencies on other code entries
// - Input/output parameters for functions/methods
//
// This interface supports multiple serialization formats including JSON, YAML, TOML, and CBOR
// making it suitable for configuration management and data persistence scenarios.
type Entry interface {
	json.Marshaler
	json.Unmarshaler
	yaml.Marshaler
	yaml.Unmarshaler
	toml.Marshaler
	toml.Unmarshaler
	cbor.Marshaler
	cbor.Unmarshaler

	// Data interface implementation for entry-specific operations.
	//
	// This embedded interface provides common data management capabilities such as:
	// - ID generation and management
	// - Status tracking
	// - Data representation and serialization
	auddat.Data[Entry]

	// GetPkg retrieves the package associated with this code entry.
	//
	// Returns the Package instance that contains this code entry. This provides access
	// to the package's path, module information, and vendor status.
	GetPkg() audpkg.Package
	GetPkgId() audids.ID

	// GetName retrieves the name of this code entry.
	//
	// Returns the identifier for this code element (function name, variable name, etc.)
	// within its package context.
	GetName() string

	// GetFullPath returns the complete path of this code entry by joining the package path and entry name.
	//
	// Combines the package's full path with the entry's name to create a complete filesystem
	// path that represents the location of this code element.
	GetFullPath() string

	// GetType retrieves the code type and its associated type string.
	//
	// Returns two values:
	//   - CodeType: The type of code element (function, variable, etc.)
	//   - TypeString: Additional type information for specific code types
	//
	// This allows for detailed categorization of code elements based on their nature.
	GetType() (audtps.CodeType, string)

	// GetTypeString generates a formatted string representation of the code type and its associated type string.
	//
	// If the code type requires a type string, it formats as "CodeType: TypeString", otherwise just returns CodeType.String().
	// This provides a human-readable representation of the code element's type for display or logging purposes.
	GetTypeString() string

	// GetSource retrieves the source code content associated with this entry.
	//
	// Returns the Source instance that contains the actual code content and related information
	// such as language, line numbers, etc.
	GetSource() audsrc.Source

	// GetLang retrieves the language identifier for the source code.
	GetLang() string

	// IsVendor checks if this code entry belongs to a vendor package.
	//
	// Returns true if the entry's containing package is a vendor dependency (typically
	// located in a vendor directory). Vendor entries are usually third-party
	// dependencies managed by Go's vendor system.
	IsVendor() bool

	// GetInputs returns the input parameters for this code entry.
	//
	// For functions or methods, these represent the parameters accepted by the code element.
	// Returns a slice of Params instances describing each input parameter.
	GetInputs() []audprm.Params

	// GetOutputs returns the output parameters for this code entry.
	//
	// For functions or methods, these represent the return values of the code element.
	// Returns a slice of Params instances describing each output parameter.
	GetOutputs() []audprm.Params

	// GetDepend returns the dependencies of this code entry.
	//
	// Returns a slice of Entry instances that this entry depends on. These are other
	// code elements that this entry references or calls during execution.
	GetDepend() []Entry

	// GetDependIds returns the dependencies of this code entry.
	//
	// Returns a slice of audids.ID pointer that this entry depends on. These represent
	// the dependency entries ID (not resolved entries) that this entry references.
	GetDependIds() []audids.ID

	// AddInputs adds new input parameters to this code entry.
	//
	// Adds one or more new input parameters to this code entry. This is useful for
	// updating function signatures or parameter lists dynamically.
	AddInputs(...audprm.Params)

	// AddOutputs adds new output parameters to this code entry.
	//
	// Adds one or more new output parameters to this code entry. This is useful for
	// updating function return types or parameter lists dynamically.
	AddOutputs(...audprm.Params)

	// AddDepend adds new dependencies to this code entry.
	//
	// Adds one or more new dependencies to this code entry. These represent other
	// code elements that this entry depends on.
	AddDepend(...audids.ID)

	// DelInputs clears all input parameters from this code entry.
	//
	// Removes all input parameters, resetting the parameter list to empty.
	DelInputs()

	// DelOutputs clears all output parameters from this code entry.
	//
	// Removes all output parameters, resetting the return value list to empty.
	DelOutputs()

	// DelDepend removes specific dependencies by ID from this code entry.
	//
	// Removes dependencies that match the provided IDs. If no IDs are provided,
	// it clears all dependencies.
	DelDepend(...audids.ID)
}

// New creates and returns a new Entry instance with the specified package function, source name, type, and source code.
//
// This constructor function initializes a new Entry with:
//   - An ID for the entry
//   - A source name (function name, variable name, etc.)
//   - A source type (additional type information)
//   - A code type (function, variable, etc.)
//   - Source code content and language information
//
// The entry's ID is generated based on the full path of the package and source name.
// Returns nil if any of these conditions are met:
//   - The provided ID is 0 (invalid)
//   - Cannot retrieve a valid package instance from the ID
//
// The returned Entry instance is thread-safe and uses a mutex for concurrent access.
//
// Parameters:
//   - id: The ID of the package containing this entry
//   - srcName: The name of the code element (function name, variable name, etc.)
//   - srcType: Additional type information for specific code types
//   - srcCode: The type of code element (function, variable, etc.)
//   - lng: Language identifier for the source code
//   - src: The actual source code content as bytes
//
// Returns:
//   - Entry: A new Entry instance or nil if initialization fails
func New(id audids.ID, srcName, srcType string, srcCode audtps.CodeType, lng string, src []byte) Entry {
	var pp string

	if id == 0 {
		return nil
	}

	if p := fpg(id); p == nil {
		return nil
	} else {
		pp = p.GetFullPath()
	}

	return &mdl{
		mx:  sync.RWMutex{},
		id:  audids.GenID(filepath.Join(pp, srcName)),
		pk:  id,
		nm:  srcName,
		tc:  srcCode,
		ts:  srcType,
		src: audsrc.New(src),
		lng: lng,
		inp: nil,
		out: nil,
		dep: nil,
		sts: audsts.Pending,
		sum: "",
		rep: nil,
	}
}

// Empty creates and returns a new Entry instance with all fields set to their zero values.
//
// This function is useful for creating placeholder entries or for testing purposes
// where a valid Entry instance is needed but no actual data is available.
//
// The returned Entry instance is thread-safe and uses a mutex for concurrent access.
//
// Returns:
//   - Entry: A new Entry instance with all fields initialized to their zero values
func Empty() Entry {
	return &mdl{
		mx:  sync.RWMutex{},
		id:  0,
		pk:  0,
		nm:  "",
		tc:  0,
		ts:  "",
		src: nil,
		lng: "",
		inp: nil,
		out: nil,
		dep: nil,
		sts: 0,
		sum: "",
		rep: nil,
	}
}
