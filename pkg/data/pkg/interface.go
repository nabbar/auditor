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

// Package pkg provides interfaces and utilities for managing package information.
//
// This package defines the Package interface and related functions for handling
// Go package information including paths, module references, and aliases.
// It supports various serialization formats (JSON, YAML, TOML, CBOR) and provides
// thread-safe access to package data through mutex protection.
package pkg

import (
	"encoding/json"
	"path/filepath"
	"sync"

	"github.com/fxamacker/cbor/v2"
	auddat "github.com/nabbar/auditor/pkg/data"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audsts "github.com/nabbar/auditor/pkg/data/status"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// FuncPkg represents a function type that creates and returns a new Package instance.
//
// This function type is typically used as a constructor for creating Package instances
// in a factory pattern. It allows for flexible instantiation of Package implementations
// by providing a way to create module instances when needed.
// The function signature ensures consistent creation of Package objects with proper initialization.
type FuncPkg func() Package

// Package represents an interface for handling package information.
//
// The Package interface extends the base Data interface and provides methods for managing:
// - Package paths (directory structure within modules)
// - Module references (linking packages to their containing modules)
// - Aliases (alternative names for packages)
// - Vendor status (whether the package belongs to a vendor module)
//
// This interface supports multiple serialization formats including JSON, YAML, TOML, and CBOR
// making it suitable for configuration management and data persistence scenarios.
// The interface is designed to be thread-safe through embedded mutex protection mechanisms.
type Package interface {
	json.Marshaler
	json.Unmarshaler
	yaml.Marshaler
	yaml.Unmarshaler
	toml.Marshaler
	toml.Unmarshaler
	cbor.Marshaler
	cbor.Unmarshaler

	// Data interface implementation for package-specific operations.
	//
	// This embedded interface provides common data management capabilities such as:
	// - ID generation and management
	// - Status tracking
	// - Data representation and serialization
	// It ensures consistent behavior across different package implementations.
	auddat.Data[Package]

	// GetMod retrieves the module associated with this package.
	//
	// Returns the Module instance that contains this package. This provides access
	// to the module's path, file information, and vendor status.
	// The returned module is immutable and represents the package's containing module structure.
	GetMod() audmod.Module
	GetModId() audids.ID

	// GetPackage retrieves the package path.
	//
	// Returns the cleaned directory path of the package relative to its module.
	// This represents the package's location within the module's structure.
	// The path is normalized using filepath.Clean to ensure consistent formatting.
	GetPackage() string

	// GetFile extracts the filename from a given path relative to this package.
	//
	// Given a relative path, this method extracts just the filename component.
	// Returns two values:
	//   - filename: The base filename with extension
	//   - valid: Boolean indicating whether the operation was successful
	//
	// This method is useful for extracting file information from package-relative paths.
	// It handles edge cases such as empty paths or invalid file specifications.
	GetFile(string) (string, bool)

	// GetFullPath returns the complete path of this package by joining the module path and package path.
	//
	// Combines the module's path with the package's path to create a full path
	// that represents the complete location of this package in the filesystem.
	// The resulting path is suitable for file system operations and path resolution.
	GetFullPath() string

	// IsVendor checks if this package belongs to a vendor module.
	//
	// Returns true if the package's containing module is a vendor dependency (typically
	// located in a vendor directory). Vendor packages are usually third-party
	// dependencies managed by Go's vendor system.
	// This check is essential for distinguishing between local and external dependencies.
	IsVendor() bool

	// AddAlias adds a new alias to this package.
	//
	// Adds an alternative name for this package. If the alias already exists,
	// it does nothing to prevent duplicates. This allows for flexible referencing
	// of packages under different names or conventions.
	// Aliases are stored in a slice and maintained in a thread-safe manner.
	AddAlias(string)

	// DelAlias removes a specific alias from this package.
	//
	// Removes a single alias if a name is provided. If the alias is empty, it clears
	// all aliases associated with this package. This method is thread-safe and uses mutex protection.
	// It ensures consistent state management even under concurrent access scenarios.
	DelAlias(string)

	// GetAlias returns all aliases associated with this package.
	//
	// Returns a slice containing all alternative names for this package.
	// The order of aliases in the slice is not guaranteed and may vary between calls.
	// This method provides read-only access to the package's alias collection.
	GetAlias() []string
}

// New creates and returns a new Package instance with the specified module ID and path.
//
// This constructor function initializes a new Package with:
//   - A module ID that references the containing module
//   - A package path that must be valid within the module's structure
//
// The package path is validated against the module's package structure using GetPkg method.
// Returns nil if any of these conditions are met:
//   - The provided ID is zero (invalid)
//   - Cannot retrieve a valid module instance from the ID
//   - The module is empty
//   - The provided path is invalid or cannot be resolved within the module
//
// The returned Package instance is thread-safe and uses a mutex for concurrent access.
// This ensures safe operations when multiple goroutines access the package simultaneously.
//
// Parameters:
//   - id: The module ID that references the containing module
//   - path: The package path relative to the module
//
// Returns:
//   - Package: A new Package instance or nil if initialization fails
//     The returned instance is fully initialized with proper ID generation,
//     status tracking, and thread-safety mechanisms.
func New(id audids.ID, path string) Package {
	var (
		ok bool
		mp string
	)

	if id == 0 {
		return nil
	}

	if m := fmg(id); m == nil || m.IsEmpty() {
		return nil
	} else if path, _, ok = m.GetPkg(path); !ok {
		return nil
	} else {
		mp = m.GetMod()
	}

	return &mdl{
		mx:  sync.RWMutex{},
		id:  audids.GenID(filepath.Join(mp, path)),
		md:  id,
		pk:  path,
		al:  make([]string, 0),
		sts: audsts.Pending,
		sum: "",
		rep: nil,
	}
}

// Empty returns an empty Package instance with zeroed fields.
//
// This function creates a Package instance where all fields are set to their
// zero values: ID is 0, module ID is 0, package path is empty, aliases are nil,
// status is 0, sum is empty, and representation is nil.
//
// This is useful for creating placeholder or default Package instances that can
// be checked for emptiness using the IsEmpty method.
func Empty() Package {
	return &mdl{
		mx:  sync.RWMutex{},
		id:  0,
		md:  0,
		pk:  "",
		al:  nil,
		sts: 0,
		sum: "",
		rep: nil,
	}
}
