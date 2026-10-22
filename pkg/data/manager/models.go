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

// File models.go contains the concrete implementation of the Manager
// interface.
//
// The mdl type wires the three collections of the manager (modules, packages
// and entries) to the user interface manager, which is used to report the
// progress of the operations, and provides the status mutators that address
// an element, whatever its kind, by identifier.

package manager

import (
	"path/filepath"

	audcol "github.com/nabbar/auditor/pkg/data/collection"
	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	auduxi "github.com/nabbar/auditor/pkg/uxi"
	semtps "github.com/nabbar/golib/semaphore/types"
)

// mdl represents the concrete implementation of the Manager interface.
//
// This struct holds the state of the manager, that is the user interface
// manager and the three collections of the data layer.
//
// The mdl struct is designed to be flexible and extensible, so that different
// implementations of the processing operations can be used interchangeably
// while the interface stays consistent.
//
// Key characteristics:
//   - It implements the Manager interface, and therefore the contract of the
//     user interface manager.
//   - It centralizes the management of the elements of the data layer: the
//     modules, the packages and the entries.
//   - It encapsulates the user interface manager, that is used for the user
//     interactions and the progress reporting.
//   - It uses the generic collections of the collection package, for a type
//     safe storage of the elements.
type mdl struct {
	// sem is the embedded UI manager that provides context and semaphore functionality
	// for user interface interactions and concurrent operation management during AST processing.
	// This component handles user feedback, progress reporting, and synchronization mechanisms
	// that are essential for managing long-running AST operations.
	sem auduxi.Manager

	// mod represents a collection of module instances managed by this implementation.
	// Modules are top-level code units that contain multiple packages and represent complete
	// source code projects or libraries in the context of AST processing.
	mod audcol.Collection[audmod.Module]

	// pkg represents a collection of package instances managed by this implementation.
	// Packages are logical groupings of code elements within modules that organize related functionality.
	pkg audcol.Collection[audpkg.Package]

	// ent represents a collection of entry instances managed by this implementation.
	// Entries are individual code constructs such as functions, variables, or types within packages.
	ent audcol.Collection[audent.Entry]
}

// Close cleans up the manager state.
//
// This method implements the io.Closer interface, but does not close the embedded UI manager
// as it should be managed by the main UI manager to avoid resource conflicts. The cleanup
// performed by this method is minimal and focuses on any internal state management that
// might be required during the manager's lifecycle.
//
// This implementation ensures that resources are properly released while avoiding
// conflicts with external UI manager management. It provides a clean interface for
// terminating the manager's operations without interfering with the main application's
// user interface lifecycle.
//
// Returns:
//   - An error if cleanup operations fail, otherwise nil
func (o *mdl) Close() error {
	// closing ui manager must be in main ui manager not in usage
	return nil
}

// NewBar creates a new progress bar for tracking task completion.
//
// This method delegates to the embedded UI manager to provide visual feedback during AST processing.
// It allows for monitoring the progress of various operations such as parsing, analysis, or reporting,
// giving users a clear indication of how long tasks will take and their current status.
//
// The progress bar functionality is essential for providing user experience during intensive
// AST processing operations where users need visibility into operation completion percentages
// and estimated remaining time.
//
// Parameters:
//   - taskName: The name of the task for which the progress bar is created, used for labeling
//     and identification purposes in the user interface
//   - totalItems: The total number of items to be processed (used for calculating completion percentage)
//     This value determines the scale and accuracy of the progress tracking
//
// Returns:
//   - A semaphore progress bar instance that can be used to track task completion
//     and provide real-time feedback to users during processing
func (o *mdl) NewBar(taskName string, totalItems int) semtps.SemBar {
	return o.sem.NewBar(taskName, totalItems)
}

// Clean removes all elements from the manager's collections.
//
// This method clears the entry, package, and module collections in that order,
// ensuring a complete reset of the manager's state. It is useful when the manager
// needs to be reused for a new parsing session or when all previously collected
// AST elements should be discarded.
func (o *mdl) Clean() {
	o.ent.Clean()
	o.pkg.Clean()
	o.mod.Clean()
}

// Merge combines the elements of another manager into this one.
//
// This method iterates over all modules, packages, and entries in the provided
// manager m and adds each element to the corresponding collection in this manager.
// The iteration order is: modules first, then packages, then entries.
//
// Parameters:
//   - m: The manager whose elements should be merged into this manager.
//     All modules, packages, and entries from m will be added to this manager's collections.
func (o *mdl) Merge(m Manager) {
	m.ModWalk(func(mod audmod.Module) bool {
		o.ModAdd(mod)
		return true
	})
	m.PkgWalk(func(pkg audpkg.Package) bool {
		o.PkgAdd(pkg)
		return true
	})
	m.EntWalk(func(ent audent.Entry) bool {
		o.EntAdd(ent)
		return true
	})
}

// SearchID computes the identifier of an element from path parts.
//
// The parts of the path are joined with the operating system path
// separator, and the resulting string is hashed by the audids.GenID
// helper. The returned identifier uniquely maps to the path and can be
// used to look up the corresponding element in the manager's
// collections.
//
// Parameters:
//   - s: The parts of the path of the element.
//
// Returns:
//   - audids.ID: The identifier of the element designated by the path.
func (o *mdl) SearchID(s ...string) audids.ID {
	return audids.GenID(filepath.Join(s...))
}

// SetStatusNone clears all the status bits of the element designated by the
// given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and clears every status bit, resetting the element
// to its initial unprocessed state.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) SetStatusNone(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.SetStatusNone(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.SetStatusNone(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.SetStatusNone(id)
	}
}

// SetStatusParsed marks the AST parse phase as completed for the element
// designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and sets the status bit that indicates the AST
// parsing phase completed successfully.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) SetStatusParsed(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.SetStatusParsed(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.SetStatusParsed(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.SetStatusParsed(id)
	}
}

// SetStatusVendor marks the AST vendor resolution phase as completed for the
// element designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and sets the status bit that indicates the AST
// vendor resolution phase completed successfully.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) SetStatusVendor(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.SetStatusVendor(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.SetStatusVendor(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.SetStatusVendor(id)
	}
}

// SetStatusAnalyzed marks the summary analysis phase as completed for the
// element designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and sets the status bit that indicates the summary
// analysis phase completed successfully.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) SetStatusAnalyzed(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.SetStatusAnalyzed(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.SetStatusAnalyzed(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.SetStatusAnalyzed(id)
	}
}

// SetStatusAudited1 marks the first audit report as generated for the element
// designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and sets the status bit that indicates the first audit
// report has been generated.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) SetStatusAudited1(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.SetStatusAudited1(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.SetStatusAudited1(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.SetStatusAudited1(id)
	}
}

// SetStatusAudited2 marks the second audit report as generated for the element
// designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and sets the status bit that indicates the second audit
// report has been generated.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) SetStatusAudited2(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.SetStatusAudited2(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.SetStatusAudited2(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.SetStatusAudited2(id)
	}
}

// SetStatusAudited3 marks the third audit report as generated for the element
// designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and sets the status bit that indicates the third audit
// report has been generated.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) SetStatusAudited3(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.SetStatusAudited3(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.SetStatusAudited3(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.SetStatusAudited3(id)
	}
}

// SetStatusAudited4 marks the fourth audit report as generated for the element
// designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and sets the status bit that indicates the fourth audit
// report has been generated.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) SetStatusAudited4(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.SetStatusAudited4(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.SetStatusAudited4(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.SetStatusAudited4(id)
	}
}

// SetStatusOutdated marks the element designated by the given identifier as
// out of date.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and sets the status bit that flags the element as out
// of date.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) SetStatusOutdated(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.SetStatusOutdated(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.SetStatusOutdated(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.SetStatusOutdated(id)
	}
}

// UnSetStatusParsed clears the AST parse phase status bit of the element
// designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and clears bit 0, which marks the AST parse phase.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) UnSetStatusParsed(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.UnSetStatusParsed(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.UnSetStatusParsed(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.UnSetStatusParsed(id)
	}
}

// UnSetStatusVendor clears the AST vendor resolution phase status bit of the
// element designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and clears bit 1, which marks the AST vendor
// resolution phase.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) UnSetStatusVendor(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.UnSetStatusVendor(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.UnSetStatusVendor(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.UnSetStatusVendor(id)
	}
}

// UnSetStatusAnalyzed clears the summary analysis phase status bit of the
// element designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and clears bit 2, which marks the summary analysis
// phase.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) UnSetStatusAnalyzed(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.UnSetStatusAnalyzed(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.UnSetStatusAnalyzed(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.UnSetStatusAnalyzed(id)
	}
}

// UnSetStatusAudited1 clears the first audit report status bit of the element
// designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and clears bit 3, which marks that the first audit
// report status.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) UnSetStatusAudited1(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.UnSetStatusAudited1(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.UnSetStatusAudited1(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.UnSetStatusAudited1(id)
	}
}

// UnSetStatusAudited2 clears the second audit report status bit of the element
// designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and clears bit 4, which marks that the second audit
// report status.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) UnSetStatusAudited2(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.UnSetStatusAudited2(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.UnSetStatusAudited2(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.UnSetStatusAudited2(id)
	}
}

// UnSetStatusAudited3 clears the third audit report status bit of the element
// designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and clears bit 5, which marks that the third audit
// report status.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) UnSetStatusAudited3(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.UnSetStatusAudited3(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.UnSetStatusAudited3(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.UnSetStatusAudited3(id)
	}
}

// UnSetStatusAudited4 clears the fourth audit report status bit of the element
// designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and clears bit 6, which marks that the fourth audit
// report status.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) UnSetStatusAudited4(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.UnSetStatusAudited4(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.UnSetStatusAudited4(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.UnSetStatusAudited4(id)
	}
}

// UnSetStatusOutdated clears the out-of-date status bit of the element
// designated by the given identifier.
//
// The method looks up the module, the package, or the entry that the
// identifier designates, and clears bit 7, which marks the out-of-date flag.
//
// Parameters:
//   - id: The unique identifier of the element to update.
//
// Returns:
//   - None
func (o *mdl) UnSetStatusOutdated(id audids.ID) {
	if o.mod.Load(id) != nil {
		o.mod.UnSetStatusOutdated(id)
	}

	if o.pkg.Load(id) != nil {
		o.pkg.UnSetStatusOutdated(id)
	}

	if o.ent.Load(id) != nil {
		o.ent.UnSetStatusOutdated(id)
	}
}
