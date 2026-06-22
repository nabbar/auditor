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
// The data.go file contains the implementation of the Data interface methods for code entries,
// providing functionality for managing status, summaries, reports, and other metadata aspects
// of code elements within the audit system.
package entry

import (
	"maps"
	"slices"

	audids "github.com/nabbar/auditor/pkg/data/id"
	audprm "github.com/nabbar/auditor/pkg/data/params"
	audrep "github.com/nabbar/auditor/pkg/data/reports"
	audsts "github.com/nabbar/auditor/pkg/data/status"
	audtps "github.com/nabbar/auditor/pkg/data/types"
)

// GetID returns the unique identifier for this code entry.
//
// This function provides access to the globally unique identifier assigned to this code element,
// ensuring proper tracking and referencing within the code analysis system.
// The ID is thread-safe and protected by read-lock mutex during access.
//
// Returns:
//   - audids.ID: The unique identifier of this code entry.
func (o *mdl) GetID() audids.ID {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.id
}

// IsEmpty checks if the code entry is empty or uninitialized.
//
// Returns true if the ID is invalid (less than 1), the name is empty, or the name equals the full path.
// This function provides a mechanism to validate whether a code entry contains sufficient data
// for meaningful processing or analysis.
//
// Returns:
//   - bool: true if the entry is empty or uninitialized, false otherwise.
func (o *mdl) IsEmpty() bool {
	fn := o.GetFullPath()
	ln := o.GetName()
	return o.GetID() < 1 || len(ln) < 1 || ln == fn
}

// IsEqual compares this code entry with another entry for equality.
//
// Returns true if both entries have identical ID, package reference, name, type string, inputs, outputs, dependencies, and vendor status.
// This function performs a comprehensive comparison of code entry attributes to determine structural equivalence.
// It ensures thread-safe access to all properties being compared through appropriate mutex locking.
//
// Parameters:
//   - f: The Entry to compare against.
//
// Returns:
//   - bool: true if the entries are equal, false otherwise.
func (o *mdl) IsEqual(f Entry) bool {
	if o.IsEmpty() || f.IsEmpty() || o.GetID() != f.GetID() || o.GetPkgId() != f.GetPkgId() {
		return false
	}

	if o.GetName() != f.GetName() || o.GetTypeString() != f.GetTypeString() {
		return false
	}

	if slices.Compare(getSliceStringer(o.GetInputs()), getSliceStringer(f.GetInputs())) != 0 {
		return false
	}

	if slices.Compare(getSliceStringer(o.GetOutputs()), getSliceStringer(f.GetOutputs())) != 0 {
		return false
	}

	if slices.Compare(o.GetDependIds(), f.GetDependIds()) != 0 {
		return false
	}

	if v := o.IsVendor(); v != f.IsVendor() {
		return false
	} else if v {
		return true
	}

	if !o.GetSource().IsEqual(f.GetSource()) {
		return false
	}

	return true
}

// Merge merges data from another entry into this code entry.
//
// This function performs a comprehensive merge operation that updates all properties of this code entry
// with values from the provided entry, preserving the original identity while updating content.
// It ensures thread-safe access to data during the merge process through write-lock protection.
//
// The merge logic is as follows:
//   - If the IDs, package IDs, or names differ, the merge fails and returns false.
//   - If the type string is non-empty and the type is not EntryNone, and differs from the current entry, the type and type string are updated.
//   - The source code is merged using the Source.Merge method.
//   - If input parameters differ in length, they are replaced entirely. If lengths match, individual parameters are updated if they differ.
//   - If output parameters differ in length, they are replaced entirely. If lengths match, individual parameters are updated if they differ.
//   - If dependencies differ in length, they are replaced entirely.
//   - If the summary is non-empty and differs, it is updated.
//   - If the status is not None and differs, it is updated.
//   - Reports with non-empty content are added to the current entry's report map.
//
// Parameters:
//   - f: The entry to merge data from. Must be of type *mdl.
//
// Returns:
//   - bool: true if the merge was successful, false otherwise (e.g., if IDs differ or the type assertion fails).
func (o *mdl) Merge(f any) bool {
	var (
		k bool
		e *mdl
	)

	e, k = f.(*mdl)
	if !k {
		return false
	}

	o.mx.Lock()
	defer o.mx.Unlock()

	if e.id != o.id {
		return false
	}

	if e.pk != o.pk {
		return false
	}

	if e.nm != o.nm {
		return false
	}

	if len(e.ts) > 0 && e.tc != audtps.EntryNone && e.ts != o.ts && e.tc != o.tc {
		o.tc = e.tc
		o.ts = e.ts
	}

	o.src.Merge(e.src)

	if len(e.inp) > 0 && len(o.inp) != len(e.inp) {
		o.inp = make([]audprm.Params, len(e.inp))
		copy(o.inp, e.inp)
	} else if len(e.inp) > 0 {
		for j := range e.inp {
			if vc, _ := e.inp[j].GetType(); vc != audtps.EntryNone && e.inp[j].String() != o.inp[j].String() {
				o.inp[j] = e.inp[j]
			}
		}
	}

	if len(e.out) > 0 && len(o.out) != len(e.out) {
		o.out = make([]audprm.Params, len(e.out))
		copy(o.out, e.out)
	} else if len(e.out) > 0 {
		for j := range e.out {
			if vc, _ := e.out[j].GetType(); vc != audtps.EntryNone && e.out[j].String() != o.out[j].String() {
				o.out[j] = e.out[j]
			}
		}
	}

	if len(e.dep) > 0 && len(e.dep) != len(o.dep) {
		o.dep = make([]audids.ID, len(e.dep))
		copy(o.dep, e.dep)
	}

	if len(e.sum) > 0 && e.sum != o.sum {
		o.sum = e.sum
	}

	if e.sts != audsts.None && e.sts != o.sts {
		o.sts = e.sts
	}

	if o.rep == nil {
		o.rep = make(map[string]audrep.Report)
	}

	for _, i := range e.rep {
		if len(i.Content) > 0 {
			o.rep[i.Name] = i
		}
	}

	return true
}

// SetStatus updates the status of this code entry.
//
// This function modifies the processing state of this code element, allowing for tracking
// of analysis progress or completion status within the code audit system.
// It provides thread-safe access to the status field through write-lock protection.
//
// Parameters:
//   - s: The new status to set.
func (o *mdl) SetStatus(s audsts.Status) {
	o.mx.Lock()
	defer o.mx.Unlock()

	o.sts = s
}

// GetStatus retrieves the current status of this code entry.
//
// This function returns the processing state of this code element, indicating whether it
// is pending, analyzed, or in an error condition within the audit workflow.
// It ensures thread-safe access to the status field through read-lock protection.
//
// Returns:
//   - audsts.Status: The current status of this code entry.
func (o *mdl) GetStatus() audsts.Status {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.sts
}

// SetSummary updates the summary description of this code entry.
//
// This function modifies the descriptive text that provides a concise overview of the
// code element's purpose and functionality for reporting and documentation purposes.
// It ensures thread-safe access to the summary field through write-lock protection.
//
// Parameters:
//   - p: The new summary string.
func (o *mdl) SetSummary(p string) {
	o.mx.Lock()
	defer o.mx.Unlock()

	o.sum = p
}

// GetSummary retrieves the summary description of this code entry.
//
// This function returns the descriptive text that provides a concise overview of the
// code element's purpose and functionality for reporting and documentation purposes.
// It ensures thread-safe access to the summary field through read-lock protection.
//
// Returns:
//   - string: The summary description of this code entry.
func (o *mdl) GetSummary() string {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.sum
}

// SetReports adds or updates a report associated with this code entry.
//
// This function stores or modifies a report entry for this code element, allowing
// for comprehensive tracking of analysis results and findings within the audit system.
// It ensures thread-safe access to the reports map through write-lock protection.
//
// Parameters:
//   - r: The report to add or update.
func (o *mdl) SetReports(r audrep.Report) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if len(o.rep) < 1 {
		o.rep = make(map[string]audrep.Report)
	}

	o.rep[r.Name] = r
}

// DelReports removes a specific report by name from this code entry.
//
// If the name is empty, it clears all reports associated with this code element.
// This function provides mechanism for managing report entries and removing outdated or
// unnecessary analysis results from the code entry's reporting system.
// It ensures thread-safe access to the reports map through write-lock protection.
//
// Parameters:
//   - name: The name of the report to remove. If empty, all reports are removed.
func (o *mdl) DelReports(name string) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if len(name) < 1 {
		o.rep = make(map[string]audrep.Report)
	} else {
		delete(o.rep, name)
	}
}

// GetReports retrieves a specific report by name from this code entry.
//
// This function returns a particular analysis report associated with this code element,
// allowing for targeted access to specific findings or results within the audit system.
// It ensures thread-safe access to the reports map through read-lock protection.
//
// Parameters:
//   - name: The name of the report to retrieve.
//
// Returns:
//   - audrep.Report: The report with the specified name, or the zero value if not found.
func (o *mdl) GetReports(name string) audrep.Report {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.rep[name]
}

// LstReports returns a slice of all reports associated with this code entry.
//
// This function provides a comprehensive collection of all analysis reports related
// to this code element, enabling full reporting capabilities and iteration over results.
// It ensures thread-safe access to the reports map through read-lock protection.
//
// Returns:
//   - []audrep.Report: A slice containing all reports associated with this code entry.
func (o *mdl) LstReports() []audrep.Report {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return slices.Collect(maps.Values(o.rep))
}
