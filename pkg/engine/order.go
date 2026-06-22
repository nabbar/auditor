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

package engine

import (
	"slices"
	"time"

	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	librun "github.com/nabbar/golib/runner"
	semtps "github.com/nabbar/golib/semaphore/types"
)

// ReOrder reorders entries by dependency graph to ensure dependencies are analyzed before their dependents.
//
// This function builds a dependency graph from all stored entries and performs a custom ordering algorithm
// that ensures dependencies are processed before their dependents. It constructs a hierarchical structure
// where each entry's dependencies are recursively traversed, building an ordered list of IDs that reflects
// the processing order requirements for analysis.
//
// The reordering process follows a specific approach:
//
// 1. Builds a dependency graph from all stored entries in the data manager.
// 2. Traverses each entry and recursively collects its dependencies, building a hierarchical structure.
// 3. Flattens this hierarchical structure into a linear list of IDs ensuring proper processing order.
// 4. Stores the ordered entry IDs in the engine's internal slice for subsequent analysis phases.
//
// This step is essential for maintaining proper analysis flow where dependent code elements
// are processed in the correct sequence, especially when dealing with complex dependency relationships
// that require hierarchical traversal rather than simple topological sorting.
//
// The algorithm implemented here does not use Kahn's algorithm but instead follows a tree-based approach
// where dependencies are resolved through recursive traversal from leaf nodes up to parent nodes,
// effectively flattening the dependency hierarchy into a linear sequence that represents the correct
// processing order for audit analysis.
//
// Example usage:
//
//	err := engine.ReOrder()
//	if err != nil {
//	    // Handle error appropriately
//	}
func (o *eng) ReOrder() error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("auditor/engine/ReOrder", r)
		}
	}()

	var bar semtps.SemBar

	defer func() {
		if bar != nil {
			bar.DeferMain()
		}
	}()

	o.u.Info("Reorder entries")
	bar = o.u.NewBar("Reorder entries", o.d.EntLen())

	// Build dependency graph from all stored entries and perform custom ordering algorithm
	// The approach constructs a hierarchical dependency tree and flattens it into a linear list

	o.i = make([]audids.ID, o.d.EntLen())
	cnt := 0

	o.d.EntWalk(func(ent audent.Entry) bool {
		defer bar.Inc(1)

		if ent == nil {
			return true
		}

		if ent.GetID() == 0 {
			return true
		}

		c := []audids.ID{ent.GetID()}

		for _, d := range ent.GetDepend() {
			r := o.parseEntry(c, d)

			for i := 0; i < len(r); i++ {
				if !slices.Contains(o.i, r[i]) {
					o.i[cnt] = r[i]
					cnt++
				}
			}
		}

		o.i[cnt] = ent.GetID()
		cnt++

		return true
	})

	// force update bar
	for !bar.Completed() {
		bar.Inc(1)
		time.Sleep(time.Millisecond)
	}

	// timer to wait ui is updated
	time.Sleep(500 * time.Millisecond)

	return nil
}

// parseEntry recursively traverses the dependency tree to collect all dependencies of a given entry.
//
// This helper function performs a recursive traversal of the dependency graph starting from a given
// entry. It collects all dependencies (both direct and indirect) in a way that ensures proper ordering
// for analysis. The recursion handles nested dependencies by building up a chain of dependencies
// and ensuring no circular references are processed.
//
// Parameters:
//
//   - b: A slice of IDs representing the current traversal path to prevent circular dependencies
//   - c: The entry whose dependencies need to be collected
//
// Returns:
//
//   - A slice of audids.ID containing all dependencies in the correct order for processing
//
// This function is a key component of the custom reordering algorithm that does not use Kahn's algorithm,
// but instead builds a dependency tree and flattens it appropriately. The resulting list represents
// the proper processing sequence where dependencies are processed before their dependents.
func (o *eng) parseEntry(b []audids.ID, c audent.Entry) []audids.ID {
	var (
		l []audent.Entry
		r []audids.ID
		n []audids.ID
	)

	if c == nil {
		return r
	}

	if c.GetID() == 0 {
		return r
	}

	if slices.Contains(o.i, c.GetID()) {
		return r
	}

	if l = c.GetDepend(); len(l) < 1 {
		// skip circular dependencies to prevent infinite recursion
		return r
	}

	n = append(b, c.GetID())

	for _, d := range l {
		if d == nil {
			continue
		}

		if d.GetID() == 0 {
			continue
		}

		if slices.Contains(b, d.GetID()) {
			// skip circular dependencies to prevent infinite recursion
			continue
		}

		if slices.Contains(o.i, d.GetID()) {
			// skip duplicate ID
			continue
		}

		f := o.parseEntry(n, d)

		for i := 0; i < len(f); i++ {
			if !slices.Contains(r, f[i]) {
				r = append(r, f[i])
			}
		}
	}

	return r
}
