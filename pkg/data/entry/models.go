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

// File models.go contains the concrete model of the Entry interface.
//
// The mdl type stores the identity, the package reference, the name, the code
// type, the type string, the source code, the language, the parameters, the
// dependencies, the status, the summary and the reports of a single code
// element. Every field is protected by a sync.RWMutex, which makes the model
// safe for concurrent use.

package entry

import (
	"fmt"
	"path/filepath"
	"sync"

	audids "github.com/nabbar/auditor/pkg/data/id"
	audprm "github.com/nabbar/auditor/pkg/data/params"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	audrep "github.com/nabbar/auditor/pkg/data/reports"
	audsrc "github.com/nabbar/auditor/pkg/data/src"
	audsts "github.com/nabbar/auditor/pkg/data/status"
	audtps "github.com/nabbar/auditor/pkg/data/types"
)

// mdl represents a concrete implementation of the Entry interface.
//
// This struct provides a thread-safe model for storing comprehensive information
// about code elements including package references, names, types, source code,
// dependencies, and status. It uses read-write mutex protection to ensure safe
// concurrent access to all fields.
type mdl struct {
	// mx is the mutex protecting access to the underlying data to ensure thread safety
	// during concurrent reads and writes to this code entry model.
	mx sync.RWMutex

	// id is the unique identifier for this code entry, ensuring distinct identification
	// within the context of code analysis and reporting systems.
	id audids.ID

	// pk is the ID of the package associated with this code entry,
	// providing a reference to the containing package for lazy initialization.
	pk audids.ID

	// nm is the name of the code entry, representing the identifier or function name
	// within its containing package scope.
	nm string

	// tc is the code type of the entry, categorizing the element as a function, method,
	// variable, constant, etc., according to the defined types system.
	tc audtps.CodeType

	// ts is the specific type string for the code type, providing additional context
	// about the precise nature of the code element when required by the type system.
	ts string

	// src is the source code content associated with this entry, storing the actual
	// code text that can be analyzed or displayed for debugging purposes.
	src audsrc.Source

	// lng is the language string representing the programming language identifier
	// for the source code, supporting multi-language analysis capabilities.
	lng string

	// inp contains input parameters for this entry, describing the function arguments
	// or parameters required for execution of this code element.
	inp []audprm.Params

	// out contains output parameters for this entry, describing the return values
	// or results produced by execution of this code element.
	out []audprm.Params

	// dep contains the IDs of entries that this entry depends on, establishing relationships
	// between code elements and supporting dependency analysis and tracking.
	dep []audids.ID

	// sts is the current status of the code entry, indicating the processing state
	// such as pending, analyzed, or error conditions during code evaluation.
	sts audsts.Status

	// sum is the summary description of the code entry, providing a concise overview
	// of the purpose and functionality of this code element.
	sum string

	// rep stores reports associated with this code entry, maintaining analysis results
	// and findings related to this specific code element for reporting purposes.
	rep map[string]audrep.Report
}

// GetPkg retrieves the package associated with this code entry.
//
// This function accesses the package ID through read-lock protection and
// resolves the package object via the registered resolver. It returns nil
// if no package ID is set or the package cannot be retrieved. It provides
// access to the package information including its path and vendor status.
//
// Returns:
//   - audpkg.Package: The package instance associated with this entry, or nil if
//     the package ID is 0 or the package cannot be found.
func (o *mdl) GetPkg() audpkg.Package {
	if id := o.GetPkgId(); id == 0 {
		return nil
	} else if pk := fpg(id); pk == nil {
		return nil
	} else {
		return pk
	}
}

// GetPkgId retrieves the ID of the package associated with this code entry.
//
// This function safely accesses the package ID through read-lock protection.
//
// Returns:
//   - audids.ID: The ID of the package containing this entry, or 0 if no package is associated.
func (o *mdl) GetPkgId() audids.ID {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.pk
}

// GetName retrieves the name of this code entry.
//
// The function returns the identifier or function name for this code element,
// ensuring thread-safe access through read-lock protection.
//
// Returns:
//   - string: The name of this code entry (e.g., function name, variable name).
func (o *mdl) GetName() string {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.nm
}

// GetFullPath returns the complete path of this code entry by joining the package path and entry name.
//
// This function constructs the full path using filepath.Join for proper path construction,
// handling cases where package information may be unavailable. Returns an empty string
// if the package cannot be retrieved.
//
// Returns:
//   - string: The full path combining the package path and entry name, or an empty string
//     if the package cannot be retrieved.
func (o *mdl) GetFullPath() string {
	if pk := o.GetPkg(); pk == nil {
		return ""
	} else {
		return filepath.Join(pk.GetFullPath(), o.GetName())
	}
}

// SetType sets the code type of this code entry and the type string that
// qualifies it.
//
// The code type categorizes the code element, and the type string provides the
// additional information about its precise type. The operation is thread-safe.
//
// Parameters:
//   - c: The code type of the code element.
//   - s: The type string of the code element.
func (o *mdl) SetType(c audtps.CodeType, s string) {
	o.mx.Lock()
	defer o.mx.Unlock()

	o.tc = c
	o.ts = s
}

// GetType retrieves the code type and its associated type string.
//
// Returns both the code type enum value and the specific type string for detailed classification.
// This function provides access to the categorization of this code element for analysis purposes.
//
// Returns:
//   - audtps.CodeType: The code type enum value categorizing this entry.
//   - string: The specific type string providing additional context about the code element.
func (o *mdl) GetType() (audtps.CodeType, string) {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.tc, o.ts
}

// GetTypeString generates a formatted string representation of the code type
// and of its associated type string.
//
// When the type string is not empty, the result is formatted as "CodeType:
// TypeString", otherwise it is only the CodeType.String() representation.
//
// Note:
//   - The format does not depend on the CodeType.NeedString helper: the type
//     string is used as soon as it is not empty.
//
// Returns:
//   - string: The formatted string representation of the code type and of the
//     type string.
func (o *mdl) GetTypeString() string {
	if c, s := o.GetType(); len(s) > 0 {
		return fmt.Sprintf("%s: %s", c.String(), s)
	} else {
		return c.String()
	}
}

// GetSource retrieves the source code content associated with this entry.
//
// Returns a new empty source if no source is available, ensuring safe access to source data.
// This function provides access to the raw code text for analysis and debugging.
//
// Returns:
//   - audsrc.Source: The source code content, or a new empty source if none is available.
func (o *mdl) GetSource() audsrc.Source {
	o.mx.RLock()
	defer o.mx.RUnlock()

	if o.src == nil {
		return audsrc.New(nil)
	}

	return o.src
}

// GetLang retrieves the language identifier for the source code.
//
// Returns the programming language string to support multi-language analysis capabilities.
//
// Returns:
//   - string: The language identifier for the source code.
func (o *mdl) GetLang() string {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.lng
}

// IsVendor checks if this code entry belongs to a vendor package.
//
// Returns true if the associated package is marked as vendor, indicating third-party dependencies.
// This function provides information about whether the code element is from external libraries.
// Returns false if the package cannot be retrieved or if it's not a vendor package.
//
// Returns:
//   - bool: true if the entry belongs to a vendor package, false otherwise.
func (o *mdl) IsVendor() bool {
	if pk := o.GetPkg(); pk == nil {
		return false
	} else {
		return pk.IsVendor()
	}
}
