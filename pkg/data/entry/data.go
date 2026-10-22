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

// File data.go contains the implementation of the auddat.Data contract for
// the code entries.
//
// It provides the identity, the emptiness check, the equality check, the
// merge, the status, the summary and the report management of a code entry,
// that is the part of the Entry interface shared by every model of the
// auditor data layer.

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

// SetStatusNone clears all status bits for this code entry.
//
// This function resets the status to the default "None" state where no
// processing phase is active. After this call, all other status bits
// are cleared and the instance returns to its initial state.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusNone() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetNone()
}

// SetStatusParsed marks the AST parse phase as completed.
//
// This function sets bit 0 to indicate that the AST parsing phase has
// successfully completed for this code entry.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusParsed() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetParsed()
}

// SetStatusVendor marks the AST vendor resolution phase as completed.
//
// This function sets bit 1 to indicate that the AST vendor resolution
// phase has successfully completed for this code entry.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusVendor() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetVendor()
}

// SetStatusAnalyzed marks the summary analysis phase as completed.
//
// This function sets bit 2 to indicate that the summary analysis phase
// has successfully completed for this code entry.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusAnalyzed() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetAnalyzed()
}

// SetStatusAudited1 marks the first audit report as generated.
//
// This function sets bit 3 to indicate that the first audit report
// has been generated for this code entry.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusAudited1() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetAudited1()
}

// SetStatusAudited2 marks the second audit report as generated.
//
// This function sets bit 4 to indicate that the second audit report
// has been generated for this code entry.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusAudited2() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetAudited2()
}

// SetStatusAudited3 marks the third audit report as generated.
//
// This function sets bit 5 to indicate that the third audit report
// has been generated for this code entry.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusAudited3() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetAudited3()
}

// SetStatusAudited4 marks the fourth audit report as generated.
//
// This function sets bit 6 to indicate that the fourth audit report
// has been generated for this code entry.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusAudited4() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetAudited4()
}

// SetStatusOutdated marks the item as out of date.
//
// This function sets bit 7 to indicate that this code entry has
// been flagged as out of date.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusOutdated() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetOutdated()
}

// UnSetStatusParsed clears the AST parse phase status bit.
//
// This function clears bit 0 from the code entry status, removing
// the indication that AST parsing has completed. Other status bits
// remain unchanged.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) UnSetStatusParsed() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.UnSetParsed()
}

// UnSetStatusVendor clears the AST vendor resolution phase status bit.
//
// This function clears bit 1 from the code entry status, removing
// the indication that vendor resolution has completed. Other status
// bits remain unchanged.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) UnSetStatusVendor() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.UnSetVendor()
}

// UnSetStatusAnalyzed clears the summary analysis phase status bit.
//
// This function clears bit 2 from the code entry status, removing
// the indication that summary analysis has completed. Other status
// bits remain unchanged.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) UnSetStatusAnalyzed() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.UnSetAnalyzed()
}

// UnSetStatusAudited1 clears the first audit report status bit.
//
// This function clears bit 3 from the code entry status, removing
// the indication that the first audit report was generated. Other
// status bits remain unchanged.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) UnSetStatusAudited1() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.UnSetAudited1()
}

// UnSetStatusAudited2 clears the second audit report status bit.
//
// This function clears bit 4 from the code entry status, removing
// the indication that the second audit report was generated. Other
// status bits remain unchanged.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) UnSetStatusAudited2() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.UnSetAudited2()
}

// UnSetStatusAudited3 clears the third audit report status bit.
//
// This function clears bit 5 from the code entry status, removing
// the indication that the third audit report was generated. Other
// status bits remain unchanged.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) UnSetStatusAudited3() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.UnSetAudited3()
}

// UnSetStatusAudited4 clears the fourth audit report status bit.
//
// This function clears bit 6 from the code entry status, removing
// the indication that the fourth audit report was generated. Other
// status bits remain unchanged.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) UnSetStatusAudited4() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.UnSetAudited4()
}

// UnSetStatusOutdated clears the out of date status bit.
//
// This function clears bit 7 from the code entry status, removing
// the out of date flag. Other status bits remain unchanged.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) UnSetStatusOutdated() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.UnSetOutdated()
}

// IsStatusNone returns true when all status bits are cleared.
//
// This function checks whether all status bits are zero, indicating
// that no processing phase is currently active on this code entry.
// It uses a write-lock because the implementation acquires o.mx.Lock()
// instead of o.mx.RLock(), even though the operation is a read-only
// query.
//
// Returns:
//   - bool: true if the status equals zero (no active phases), false otherwise
func (o *mdl) IsStatusNone() bool {
	o.mx.Lock()
	defer o.mx.Unlock()
	return o.sts.IsNone()
}

// IsStatusParsed returns true if the AST parse phase bit is set.
//
// This function checks whether bit 0 (Parsed) is set, indicating that
// the AST parsing phase has completed for this code entry.
// It uses a write-lock because the implementation acquires o.mx.Lock()
// instead of o.mx.RLock(), even though the operation is a read-only
// query.
//
// Returns:
//   - bool: true if the Parsed bit is set, false otherwise
func (o *mdl) IsStatusParsed() bool {
	o.mx.Lock()
	defer o.mx.Unlock()
	return o.sts.IsParsed()
}

// IsStatusVendor returns true if the AST vendor resolution bit is set.
//
// This function checks whether bit 1 (Vendor) is set, indicating that
// the AST vendor resolution phase has completed for this code entry.
// It uses a write-lock because the implementation acquires o.mx.Lock()
// instead of o.mx.RLock(), even though the operation is a read-only
// query.
//
// Returns:
//   - bool: true if the Vendor bit is set, false otherwise
func (o *mdl) IsStatusVendor() bool {
	o.mx.Lock()
	defer o.mx.Unlock()
	return o.sts.IsVendor()
}

// IsStatusAnalyzed returns true if the summary analysis bit is set.
//
// This function checks whether bit 2 (Analyzed) is set, indicating that
// the summary analysis phase has completed for this code entry.
// It uses a write-lock because the implementation acquires o.mx.Lock()
// instead of o.mx.RLock(), even though the operation is a read-only
// query.
//
// Returns:
//   - bool: true if the Analyzed bit is set, false otherwise
func (o *mdl) IsStatusAnalyzed() bool {
	o.mx.Lock()
	defer o.mx.Unlock()
	return o.sts.IsAnalyzed()
}

// IsStatusAudited1 returns true if the first audit report bit is set.
//
// This function checks whether bit 3 (Audited1) is set, indicating that
// the first audit report has been generated for this code entry.
// It uses a write-lock because the implementation acquires o.mx.Lock()
// instead of o.mx.RLock(), even though the operation is a read-only
// query.
//
// Returns:
//   - bool: true if the Audited1 bit is set, false otherwise
func (o *mdl) IsStatusAudited1() bool {
	o.mx.Lock()
	defer o.mx.Unlock()
	return o.sts.IsAudited1()
}

// IsStatusAudited2 returns true if the second audit report bit is set.
//
// This function checks whether bit 4 (Audited2) is set, indicating that
// the second audit report has been generated for this code entry.
// It uses a write-lock because the implementation acquires o.mx.Lock()
// instead of o.mx.RLock(), even though the operation is a read-only
// query.
//
// Returns:
//   - bool: true if the Audited2 bit is set, false otherwise
func (o *mdl) IsStatusAudited2() bool {
	o.mx.Lock()
	defer o.mx.Unlock()
	return o.sts.IsAudited2()
}

// IsStatusAudited3 returns true if the third audit report bit is set.
//
// This function checks whether bit 5 (Audited3) is set, indicating that
// the third audit report has been generated for this code entry.
// It uses a write-lock because the implementation acquires o.mx.Lock()
// instead of o.mx.RLock(), even though the operation is a read-only
// query.
//
// Returns:
//   - bool: true if the Audited3 bit is set, false otherwise
func (o *mdl) IsStatusAudited3() bool {
	o.mx.Lock()
	defer o.mx.Unlock()
	return o.sts.IsAudited3()
}

// IsStatusAudited4 returns true if the fourth audit report bit is set.
//
// This function checks whether bit 6 (Audited4) is set, indicating that
// the fourth audit report has been generated for this code entry.
// It uses a write-lock because the implementation acquires o.mx.Lock()
// instead of o.mx.RLock(), even though the operation is a read-only
// query.
//
// Returns:
//   - bool: true if the Audited4 bit is set, false otherwise
func (o *mdl) IsStatusAudited4() bool {
	o.mx.Lock()
	defer o.mx.Unlock()
	return o.sts.IsAudited4()
}

// IsStatusOutdated returns true if the out of date bit is set.
//
// This function checks whether bit 7 (Outdated) is set, indicating that
// this code entry has been flagged as out of date.
// It uses a write-lock because the implementation acquires o.mx.Lock()
// instead of o.mx.RLock(), even though the operation is a read-only
// query.
//
// Returns:
//   - bool: true if the Outdated bit is set, false otherwise
func (o *mdl) IsStatusOutdated() bool {
	o.mx.Lock()
	defer o.mx.Unlock()
	return o.sts.IsOutdated()
}

// GetParentId returns the identifier of the parent of this code entry.
//
// The parent of a code entry is the package that contains it. The identifier
// of that package is returned, so that the hierarchy of the data model can be
// walked upwards.
//
// Returns:
//   - audids.ID: The identifier of the containing package.
func (o *mdl) GetParentId() audids.ID {
	return o.GetPkgId()
}
