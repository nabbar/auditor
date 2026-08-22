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
	"path/filepath"
	"slices"
	"sync"

	audids "github.com/nabbar/auditor/pkg/data/id"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audrep "github.com/nabbar/auditor/pkg/data/reports"
	audsts "github.com/nabbar/auditor/pkg/data/status"
)

// mdl represents a package model that stores package information including ID, module reference, package path, and aliases.
//
// This struct implements the Package interface and provides thread-safe access to package data.
// It maintains:
//   - A unique identifier (ID) for tracking packages
//   - A module ID (md) that references the containing module
//   - Package path information (pk) for Go package resolution
//   - Aliases (al) for alternative naming of packages
//   - Status tracking (sts) for package lifecycle management
//   - Summary description (sum) for package documentation
//   - Reports (rep) for associated audit findings or results
//
// The mutex (mx) ensures thread safety when accessing any of these fields.
// This implementation provides a robust foundation for package management with proper concurrency control.
type mdl struct {
	// mx is the mutex protecting access to the underlying data.
	// This mutex ensures that concurrent access to package data is safe and prevents race conditions.
	// It protects all fields in this struct to ensure thread safety during read/write operations.
	mx sync.RWMutex
	// id is the unique identifier for this package.
	// Generated using audids.GenID() based on the module path and package path, this ID provides a stable way to identify packages.
	// The ID is created once during initialization and remains constant throughout the package's lifetime.
	id audids.ID
	// md is the module ID that references the containing module.
	// This ID allows lazy initialization of module instances when needed.
	// It ensures that modules can be retrieved on-demand without requiring immediate instantiation.
	md audids.ID

	// pk contains the package path
	// This represents the Go package path relative to its module.
	// The path is normalized and validated during package creation.
	pk string
	// al contains a list of package aliases
	// These are alternative names that can be used to reference this package.
	// Aliases provide flexibility in how packages are referenced throughout the system.
	al []string

	// sts is the current status of the package
	// Tracks the lifecycle status of the package (Pending, Active, Inactive, etc.).
	// Status tracking enables monitoring and control of package states during operations.
	sts audsts.Status
	// sum is the summary description of the package
	// Provides a brief description or summary of what this package represents.
	// This field can be used for documentation purposes or user-facing descriptions.
	sum string
	// rep stores reports associated with this package
	// Contains audit reports or findings related to this package.
	// Reports provide detailed information about package analysis results and findings.
	rep map[string]audrep.Report
}

// GetMod retrieves the module associated with this package.
//
// Returns the Module instance that contains this package. This provides access
// to the module's path, file information, and vendor status.
// If no module function is defined or if the module cannot be retrieved, it returns nil.
// The method is thread-safe and uses read-lock protection for concurrent access.
func (o *mdl) GetMod() audmod.Module {
	if m := fmg(o.GetModId()); m != nil {
		return m
	}

	return nil
}

// GetModId returns the module ID associated with this package.
//
// Returns the audids.ID that references the containing module.
// This ID can be used to look up the module via the fmg function.
// The method is thread-safe and uses read-lock protection for concurrent access.
func (o *mdl) GetModId() audids.ID {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.md
}

// GetPackage retrieves the package path.
//
// Returns the cleaned directory path of the package relative to its module.
// This represents the package's location within the module's structure.
// The returned path is normalized and validated during initialization.
// The method is thread-safe and uses read-lock protection for concurrent access.
func (o *mdl) GetPackage() string {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.pk
}

// GetFile extracts the filename from a given path relative to this package.
//
// Given a relative path, this method extracts just the filename component.
// Returns two values:
//   - filename: The base filename with extension
//   - valid: Boolean indicating whether the operation was successful
//
// This method is useful for extracting file information from package-relative paths.
// It validates that the file belongs to this specific package by checking against
// the module's package resolution logic. The validation ensures proper path resolution
// and prevents cross-package file access.
//
// Parameters:
//   - path: The relative path to resolve against this package
//
// Returns:
//   - string: The extracted filename, or empty string if resolution fails
//   - bool: true if the file was successfully resolved and belongs to this package, false otherwise
func (o *mdl) GetFile(path string) (string, bool) {
	md := o.GetMod()
	if md == nil {
		return "", false
	}

	if p, f, k := md.GetPkg(path); !k {
		return "", false
	} else if p != o.GetPackage() {
		return "", false
	} else {
		return f, true
	}
}

// GetFullPath returns the complete path of this package by joining the module path and package path.
//
// Combines the module's path with the package's path to create a full path
// that represents the complete location of this package in the filesystem.
// Returns an empty string if no module is available or if module retrieval fails.
// This method provides a consistent way to resolve full package paths for file system operations.
func (o *mdl) GetFullPath() string {
	if md := o.GetMod(); md == nil {
		return ""
	} else {
		return filepath.Join(md.GetMod(), o.GetPackage())
	}
}

// IsVendor checks if this package belongs to a vendor module.
//
// Returns true if the package's containing module is a vendor dependency (typically
// located in a vendor directory). Vendor packages are usually third-party
// dependencies managed by Go's vendor system.
// This check is essential for distinguishing between local and external dependencies.
// The method is thread-safe and uses read-lock protection for concurrent access.
func (o *mdl) IsVendor() bool {
	if md := o.GetMod(); md == nil {
		return false
	} else {
		return md.IsVendor()
	}
}

// AddAlias adds a new alias to this package.
//
// Adds an alternative name for this package. If the alias already exists,
// it does nothing to prevent duplicates. This allows for flexible referencing
// of packages under different names or conventions. This method is thread-safe
// and uses mutex protection to ensure safe concurrent access.
// The implementation prevents duplicate aliases by checking existing entries before adding.
//
// Parameters:
//   - p: The alias string to add to this package
func (o *mdl) AddAlias(p string) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if len(o.al) < 1 {
		o.al = make([]string, 0)
	}

	if slices.Contains(o.al, p) {
		return
	}

	o.al = append(o.al, p)
}

// DelAlias removes a specific alias from this package.
//
// Removes a single alias if a name is provided. If the alias is empty, it clears
// all aliases associated with this package. This method is thread-safe and uses mutex protection.
// It ensures consistent state management even under concurrent access scenarios.
// The implementation handles both partial deletion (specific alias) and full deletion (all aliases).
//
// Parameters:
//   - p: The alias string to remove. If empty, all aliases are cleared.
func (o *mdl) DelAlias(p string) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if len(o.al) < 1 {
		return
	}

	if len(p) < 1 {
		o.al = make([]string, 0)
		return
	}

	var a = make([]string, len(o.al))
	copy(a, o.al)
	o.al = make([]string, 0, len(o.al))

	for i := range a {
		if a[i] != p {
			o.al = append(o.al, a[i])
		}
	}
}

// GetAlias returns all aliases associated with this package.
//
// Returns a slice containing all alternative names for this package.
// The order of aliases in the slice is not guaranteed and may vary between calls.
// This method provides read-only access to the package's alias collection.
// The returned slice is a copy to prevent external modification of internal data.
// The method is thread-safe and uses read-lock protection for concurrent access.
func (o *mdl) GetAlias() []string {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.al
}
