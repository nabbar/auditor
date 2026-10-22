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
// the Go packages.
//
// It provides the identity, the emptiness check, the equality check, the
// merge, the status, the summary and the report management of a package, that
// is the part of the Package interface shared by every model of the auditor
// data layer.

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
// Returns true when one of these conditions is met:
//   - The package ID is invalid (lower than 1).
//   - The identifier of the containing module is zero, which means that the
//     package does not reference any module.
//
// This method is useful to validate whether a package instance contains valid
// data. It provides a quick way to determine if the package has been properly
// initialized.
//
// Returns:
//   - bool: true if the package is considered empty or uninitialized.
func (o *mdl) IsEmpty() bool {
	return o.GetID() < 1 || o.GetModId() == 0
}

// IsEqual compares this package with another package for equality.
//
// Returns true when both packages are not empty and when they share:
//   - The same unique identifier.
//   - The same package path.
//   - The same identifier of the containing module.
//
// Note:
//   - The modules themselves are not compared: only their identifiers are.
//     The comparison of the modules is the responsibility of the module
//     layer.
//
// This method is used to determine if two package instances represent the same
// logical package.
//
// Parameters:
//   - f: The Package to compare against.
//
// Returns:
//   - bool: true if both packages are equal, false otherwise.
func (o *mdl) IsEqual(f Package) bool {
	return !o.IsEmpty() && !f.IsEmpty() && o.GetID() == f.GetID() && o.GetPackage() == f.GetPackage() && o.GetModId() == f.GetModId()
}

// Merge compares and updates this package with another package.
//
// The merge operation:
//   - Updates the identifier of the module and the path of the package when
//     the module of the source is not vendored and the module of the receiver
//     is vendored.
//   - Merges the aliases of the source package, appending the new ones and
//     avoiding the duplicates.
//   - Updates the summary when the source has a non-empty summary that
//     differs from the current one.
//   - Updates the status when the status of the source is not None and differs
//     from the current status.
//   - Copies all the reports of the source package that have a non-empty
//     content.
//
// The merge is thread-safe and uses the write lock of the mutex.
//
// Note:
//   - The modules are resolved through the resolver that has been registered
//     with RegisterGetModule. When no resolver has been registered, the call
//     panics on the nil function pointer.
//
// Parameters:
//   - f: The package to merge from. It must be a *mdl value, otherwise the
//     merge fails.
//
// Returns:
//   - bool: true if the merge succeeded, false if the types do not match, if
//     the identifiers differ, or if the modules cannot be resolved.
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
// (Pending, Active, Inactive, etc.). This method is thread-safe and uses write-lock
// protection on the mutex. The status update is atomic and prevents race conditions
// during concurrent access.
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
// The update operation is thread-safe and uses write-lock protection on the mutex.
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
// This method is thread-safe and uses write-lock protection on the mutex.
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
// all reports associated with this package. This method is thread-safe and uses write-lock
// protection on the mutex. It ensures consistent state management even under concurrent
// access scenarios.
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
// This method provides read-only access to all package reports. Thread-safety is
// ensured through read-lock protection on the mutex.
//
// Returns:
//   - []audrep.Report: A slice of all reports associated with this package
func (o *mdl) LstReports() []audrep.Report {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return slices.Collect(maps.Values(o.rep))
}

// SetStatusNone clears all status bits for this package.
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
// successfully completed for this package.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusParsed() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetParsed()
}

// SetStatusVendor marks the AST vendor resolution phase as completed.
//
// This function sets bit 1 to indicate that the AST vendor resolution
// phase has successfully completed for this package.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusVendor() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetVendor()
}

// SetStatusAnalyzed marks the summary analysis phase as completed.
//
// This function sets bit 2 to indicate that the summary analysis phase
// has successfully completed for this package.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusAnalyzed() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetAnalyzed()
}

// SetStatusAudited1 marks the first audit report as generated.
//
// This function sets bit 3 to indicate that the first audit report
// has been generated for this package.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusAudited1() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetAudited1()
}

// SetStatusAudited2 marks the second audit report as generated.
//
// This function sets bit 4 to indicate that the second audit report
// has been generated for this package.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusAudited2() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetAudited2()
}

// SetStatusAudited3 marks the third audit report as generated.
//
// This function sets bit 5 to indicate that the third audit report
// has been generated for this package.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusAudited3() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetAudited3()
}

// SetStatusAudited4 marks the fourth audit report as generated.
//
// This function sets bit 6 to indicate that the fourth audit report
// has been generated for this package.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusAudited4() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetAudited4()
}

// SetStatusOutdated marks the item as out of date.
//
// This function sets bit 7 to indicate that this package has
// been flagged as out of date.
// It ensures thread-safe access through write-lock protection.
func (o *mdl) SetStatusOutdated() {
	o.mx.Lock()
	defer o.mx.Unlock()
	o.sts = o.sts.SetOutdated()
}

// UnSetStatusParsed clears the AST parse phase status bit.
//
// This function clears bit 0 from the package status, removing
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
// This function clears bit 1 from the package status, removing
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
// This function clears bit 2 from the package status, removing
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
// This function clears bit 3 from the package status, removing
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
// This function clears bit 4 from the package status, removing
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
// This function clears bit 5 from the package status, removing
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
// This function clears bit 6 from the package status, removing
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
// This function clears bit 7 from the package status, removing
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
// that no processing phase is currently active on this package.
// It ensures thread-safe access through write-lock protection.
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
// the AST parsing phase has completed for this package.
// It ensures thread-safe access through write-lock protection.
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
// the AST vendor resolution phase has completed for this package.
// It ensures thread-safe access through write-lock protection.
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
// the summary analysis phase has completed for this package.
// It ensures thread-safe access through write-lock protection.
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
// the first audit report has been generated for this package.
// It ensures thread-safe access through write-lock protection.
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
// the second audit report has been generated for this package.
// It ensures thread-safe access through write-lock protection.
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
// the third audit report has been generated for this package.
// It ensures thread-safe access through write-lock protection.
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
// the fourth audit report has been generated for this package.
// It ensures thread-safe access through write-lock protection.
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
// this package has been flagged as out of date.
// It ensures thread-safe access through write-lock protection.
//
// Returns:
//   - bool: true if the Outdated bit is set, false otherwise
func (o *mdl) IsStatusOutdated() bool {
	o.mx.Lock()
	defer o.mx.Unlock()
	return o.sts.IsOutdated()
}

// GetParentId returns the identifier of the parent of this package.
//
// The parent of a package is the module that contains it. The identifier of
// that module is returned, so that the hierarchy of the data model can be
// walked upwards.
//
// Returns:
//   - audids.ID: The identifier of the containing module.
func (o *mdl) GetParentId() audids.ID {
	return o.GetModId()
}
