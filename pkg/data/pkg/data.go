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
	"maps"
	"slices"

	audids "github.com/nabbar/auditor/pkg/data/id"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audrep "github.com/nabbar/auditor/pkg/data/reports"
	audsts "github.com/nabbar/auditor/pkg/data/status"
)

// GetID returns the unique identifier for this package.
//
// This method provides access to the package's unique identifier which is generated
// based on the module path and package path. The ID remains constant throughout
// the package's lifetime and can be used for reliable identification in data structures
// or comparisons. The returned ID is thread-safe and protected by read-lock mechanism.
//
// Returns:
//   - audids.ID: The unique identifier of this package
func (o *mdl) GetID() audids.ID {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.id
}

// IsEmpty checks if the package is empty or uninitialized.
//
// Returns true if any of these conditions are met:
//   - The package ID is invalid (< 1)
//   - The module reference is nil (no module available)
//   - The package path is too short (less than 3 characters)
//
// This method is useful for validating whether a package instance contains valid data.
// It provides a quick way to determine if the package has been properly initialized.
func (o *mdl) IsEmpty() bool {
	return o.GetID() < 1 || o.GetModId() == 0
}

// IsEqual compares this package with another package for equality.
//
// Returns true if both packages have identical:
//   - ID (unique identifier)
//   - Package path
//   - Module reference (using IsEqual on the module)
//
// This method is used to determine if two package instances represent the same logical package.
// It performs a comprehensive comparison across all relevant fields, including the module relationship.
// The comparison is thread-safe and uses read-lock protection for concurrent access.
//
// Parameters:
//   - f: The Package to compare against
//
// Returns:
//   - bool: true if both packages are equal, false otherwise
func (o *mdl) IsEqual(f Package) bool {
	return !o.IsEmpty() && !f.IsEmpty() && o.GetID() == f.GetID() && o.GetPackage() == f.GetPackage() && o.GetModId() == f.GetModId()
}

// Merge compares and updates this package with another package.
//
// Updates the current package's data with values from the provided package (f).
// This method performs a merge operation that:
//   - Updates package path and module ID if the source module is not a vendor module
//     and the current module is a vendor module
//   - Merges aliases from the source package (appending new ones, avoiding duplicates)
//   - Updates the summary if the source has a non-empty summary different from the current one
//   - Updates the status if the source status is not None and differs from the current status
//   - Copies all reports from the source package that have non-empty content
//
// Returns true if both packages are equal (as determined by IsEqual), indicating
// that the merge operation was successful. The merge is thread-safe and uses mutex protection.
//
// Parameters:
//   - f: The interface to merge from. Must be a *mdl type; otherwise returns false.
//
// Returns:
//   - bool: true if the merge was successful, false if the types don't match or IDs differ
func (o *mdl) Merge(f any) bool {
	var (
		k  bool
		p  *mdl
		mn audmod.Module
		mo audmod.Module
	)

	if p, k = f.(*mdl); !k || p == nil {
		return false
	}

	o.mx.Lock()
	defer o.mx.Unlock()

	if p.id != o.id {
		return false
	}

	mn = fmg(p.md)
	mo = fmg(o.md)

	if mo == nil || mn == nil {
		return false
	}

	if mo.IsVendor() && !mn.IsVendor() {
		o.md = p.md
		o.pk = p.pk
	}

	if len(o.al) < 1 {
		o.al = make([]string, 0)
	}

	for _, a := range p.al {
		if !slices.Contains(o.al, a) {
			o.al = append(o.al, a)
		}
	}

	if len(p.sum) > 0 && p.sum != o.sum {
		o.sum = p.sum
	}

	if p.sts != audsts.None && p.sts != o.sts {
		o.sts = p.sts
	}

	if o.rep == nil {
		o.rep = make(map[string]audrep.Report)
	}

	for _, i := range p.rep {
		if len(i.Content) > 0 {
			o.rep[i.Name] = i
		}
	}

	return true
}

// SetStatus updates the status of this package.
//
// Sets a new status for the package, which can indicate its current lifecycle state
// (Pending, Active, Inactive, etc.). This method is thread-safe and uses mutex protection.
// The status update is atomic and prevents race conditions during concurrent access.
//
// Parameters:
//   - s: The new status to set for this package
func (o *mdl) SetStatus(s audsts.Status) {
	o.mx.Lock()
	defer o.mx.Unlock()

	o.sts = s
}

// GetStatus retrieves the current status of this package.
//
// Returns the current lifecycle status of the package. This provides insight into
// the package's state within the audit or management process.
// The method is thread-safe and uses read-lock protection for concurrent access.
//
// Returns:
//   - audsts.Status: The current status of this package
func (o *mdl) GetStatus() audsts.Status {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.sts
}

// SetSummary updates the summary description of this package.
//
// Sets a new summary that describes what this package represents or contains.
// This is useful for providing context about the package's purpose or content.
// The update operation is thread-safe and uses mutex protection.
//
// Parameters:
//   - p: The new summary string to set for this package
func (o *mdl) SetSummary(p string) {
	o.mx.Lock()
	defer o.mx.Unlock()

	o.sum = p
}

// GetSummary retrieves the summary description of this package.
//
// Returns the current summary description that provides a brief overview of
// what this package represents or contains. The method is thread-safe and uses read-lock protection.
//
// Returns:
//   - string: The current summary description of this package
func (o *mdl) GetSummary() string {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.sum
}

// SetReports adds or updates a report associated with this package.
//
// Adds or replaces a report in the package's report collection. If no reports exist,
// it initializes the report map before adding the new report.
// This method is thread-safe and uses mutex protection for concurrent access.
//
// Parameters:
//   - r: The Report to add or update. The report is stored using its Name field as the key.
func (o *mdl) SetReports(r audrep.Report) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if len(o.rep) < 1 {
		o.rep = make(map[string]audrep.Report)
	}

	o.rep[r.Name] = r
}

// DelReports removes a specific report by name from this package.
//
// Removes a single report if a name is provided. If the name is empty, it clears
// all reports associated with this package. This method is thread-safe and uses mutex protection.
// It ensures consistent state management even under concurrent access scenarios.
//
// Parameters:
//   - name: The name of the report to remove. If empty, all reports are cleared.
func (o *mdl) DelReports(name string) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if len(name) < 1 {
		o.rep = make(map[string]audrep.Report)
	} else {
		delete(o.rep, name)
	}
}

// GetReports retrieves a specific report by name from this package.
//
// Returns the report with the specified name if it exists. If no such report exists,
// it returns an empty Report struct (zero value). The method is thread-safe and uses read-lock protection.
//
// Parameters:
//   - name: The name of the report to retrieve
//
// Returns:
//   - audrep.Report: The report with the given name, or a zero-value Report if not found
func (o *mdl) GetReports(name string) audrep.Report {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.rep[name]
}

// LstReports returns a slice of all reports associated with this package.
//
// Returns a slice containing all reports currently associated with this package.
// The order of reports in the slice is not guaranteed and may vary between calls.
// This method provides read-only access to all package reports and is thread-safe.
//
// Returns:
//   - []audrep.Report: A slice of all reports associated with this package
func (o *mdl) LstReports() []audrep.Report {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return slices.Collect(maps.Values(o.rep))
}
