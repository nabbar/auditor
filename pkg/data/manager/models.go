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

// Package manager provides the generic interface for AST (Abstract Syntax Tree) parsers
// that are used across different programming languages in the auditor tool.
package manager

import (
	audcol "github.com/nabbar/auditor/pkg/data/collection"
	audent "github.com/nabbar/auditor/pkg/data/entry"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	auduxi "github.com/nabbar/auditor/pkg/uxi"
	semtps "github.com/nabbar/golib/semaphore/types"
)

// mdl represents the concrete implementation of the ManagerAdmin interface.
//
// This struct holds the state for the AST manager, including the embedded UI manager
// and function callbacks for managing collections of AST elements. It implements
// all the methods required by both the Manager and ManagerAdmin interfaces,
// providing a complete implementation for AST element management.
//
// The mdl struct is designed to be flexible and extensible, allowing different
// implementations of AST processing operations through callback functions.
// This design enables various AST parser backends to be used interchangeably
// while maintaining a consistent interface.
//
// Key characteristics:
//   - Implements both Manager and ManagerAdmin interfaces
//   - Provides centralized management of AST elements (modules, packages, entries)
//   - Encapsulates the UI manager for user interaction and progress tracking
//   - Uses generic data structures for type-safe collections of AST elements
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
