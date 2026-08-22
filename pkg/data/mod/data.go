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
	"maps"
	"slices"

	audids "github.com/nabbar/auditor/pkg/data/id"
	audrep "github.com/nabbar/auditor/pkg/data/reports"
	audsts "github.com/nabbar/auditor/pkg/data/status"
)

// GetID returns the unique identifier for this module.
//
// The ID is derived from the module path via audids.GenID at construction time
// and remains constant for the lifetime of the module. It can be used as a
// stable key in maps, databases, or any data structure that requires a
// reliable, immutable identifier.
//
// Returns:
//   - audids.ID: the unique identifier for this module instance.
func (o *mdl) GetID() audids.ID {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.id
}

// IsEmpty reports whether the module is empty or uninitialized.
//
// It returns true if any of the following conditions hold:
//   - The module ID is less than 1 (invalid or unset).
//   - The module path is empty (zero length).
//   - The module is not vendored and has no associated file path.
//
// Use IsEmpty to validate that a Module instance contains sufficient data
// before performing operations that require a well-formed module.
//
// Returns:
//   - bool: true if the module is considered empty or uninitialized.
func (o *mdl) IsEmpty() bool {
	return o.GetID() < 1 || len(o.GetMod()) < 1 || (!o.IsVendor() && len(o.GetFile()) < 1)
}

// IsEqual reports whether this module is equivalent to another module f.
//
// Two modules are considered equal when all of the following hold:
//   - Neither module is empty (IsEmpty returns false for both).
//   - Their IDs match.
//   - Their module paths match.
//   - Their file paths match.
//   - Their vendor flags match.
//
// IsEqual is useful for deduplication and comparison operations in module
// management systems.
//
// Parameters:
//   - f: the Module instance to compare against.
//
// Returns:
//   - bool: true if both modules are equal, false otherwise.
func (o *mdl) IsEqual(f Module) bool {
	return !o.IsEmpty() && !f.IsEmpty() && o.GetID() == f.GetID() && o.GetMod() == f.GetMod() && o.GetFile() == f.GetFile() && o.IsVendor() && f.IsVendor()
}

// Merge updates this module with data from another module f.
//
// Merge performs the following operations:
//  1. Type-asserts f to *mdl; returns false if the assertion fails or f is nil.
//  2. Verifies that both modules share the same ID and module path; returns
//     false if they differ, since merging across different modules is not
//     supported.
//  3. If the source file path differs, updates fgm and recomputes the vendor
//     flag from the new file path.
//  4. If the source summary is non-empty and differs, updates the summary.
//  5. If the source status is not audsts.None and differs, updates the status.
//  6. Copies all non-empty reports from the source module into the receiver's
//     report map, overwriting any existing report with the same name.
//
// The merge operation is thread-safe: the receiver's mutex is locked for the
// entire duration of the update.
//
// Parameters:
//   - f: the Module instance whose data should be merged into the receiver.
//
// Returns:
//   - bool: true if the merge succeeded; false if the type assertion failed,
//     the source is nil, or the IDs/module paths do not match.
func (o *mdl) Merge(f any) bool {
	var (
		k bool
		m *mdl
	)

	if m, k = f.(*mdl); !k || m == nil {
		return false
	}

	o.mx.Lock()
	defer o.mx.Unlock()

	if m.id != o.id {
		return false
	}

	if m.mod != o.mod {
		return false
	}

	if m.fgm != o.fgm {
		o.fgm = m.fgm
	}

	o.vdr = len(o.fgm) < 1

	if len(m.sum) > 0 && m.sum != o.sum {
		o.sum = m.sum
	}

	if m.sts != audsts.None && m.sts != o.sts {
		o.sts = m.sts
	}

	if o.rep == nil {
		o.rep = make(map[string]audrep.Report)
	}

	for _, i := range m.rep {
		if len(i.Content) > 0 {
			o.rep[i.Name] = i
		}
	}

	return true
}

// SetStatus updates the lifecycle status of this module.
//
// The status indicates the module's current state within the audit or
// management process (e.g. Pending, Active, Inactive). The operation is
// thread-safe.
//
// Parameters:
//   - s: the new Status value to assign.
func (o *mdl) SetStatus(s audsts.Status) {
	o.mx.Lock()
	defer o.mx.Unlock()

	o.sts = s
}

// GetStatus returns the current lifecycle status of this module.
//
// The status value can be used to filter, prioritize, or classify modules
// based on their current condition in workflows.
//
// Returns:
//   - audsts.Status: the current status of this module instance.
func (o *mdl) GetStatus() audsts.Status {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.sts
}

// SetSummary updates the summary description of this module.
//
// The summary provides a brief description of what the module represents
// or contains. It is useful in user interfaces or reports where quick
// identification of modules is needed. The operation is thread-safe.
//
// Parameters:
//   - p: the new summary text to assign.
func (o *mdl) SetSummary(p string) {
	o.mx.Lock()
	defer o.mx.Unlock()

	o.sum = p
}

// GetSummary returns the summary description of this module.
//
// The summary is a brief textual overview of the module's purpose or content.
//
// Returns:
//   - string: the current summary description of this module instance.
func (o *mdl) GetSummary() string {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.sum
}

// SetReports adds or replaces a report in the module's report collection.
//
// The report is stored under its Name key. If the report map is nil, it is
// initialized before insertion. This method is thread-safe.
//
// Parameters:
//   - r: the Report instance to add or update.
func (o *mdl) SetReports(r audrep.Report) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if len(o.rep) < 1 {
		o.rep = make(map[string]audrep.Report)
	}

	o.rep[r.Name] = r
}

// DelReports removes a specific report by name from this module.
//
// If the supplied name is non-empty, only the report with that name is removed.
// If the name is empty, all reports are cleared and the map is re-initialized.
// The operation is thread-safe.
//
// Parameters:
//   - name: the name of the report to remove; an empty string removes all reports.
func (o *mdl) DelReports(name string) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if len(name) < 1 {
		o.rep = make(map[string]audrep.Report)
	} else {
		delete(o.rep, name)
	}
}

// GetReports returns the report with the specified name, or the zero value
// if no such report exists.
//
// Parameters:
//   - name: the name of the report to retrieve.
//
// Returns:
//   - audrep.Report: the requested report instance, or the zero value if not found.
func (o *mdl) GetReports(name string) audrep.Report {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.rep[name]
}

// LstReports returns a slice of all reports associated with this module.
//
// The order of reports in the returned slice is not guaranteed and may vary
// between calls, since the underlying map iteration order is unspecified.
//
// Returns:
//   - []audrep.Report: a slice of all report instances associated with this module.
func (o *mdl) LstReports() []audrep.Report {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return slices.Collect(maps.Values(o.rep))
}
