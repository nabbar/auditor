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

package manager

import (
	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	audtps "github.com/nabbar/auditor/pkg/data/types"
)

// Linker is a struct that holds function pointers for all manager operations.
//
// It provides a way to link the manager's methods to external code, allowing for
// dependency injection and testing. Each field corresponds to a method on the
// Manager interface, enabling callers to access manager functionality through
// function references rather than direct method calls.
//
// This design pattern is useful for:
//   - Decoupling code from specific manager implementations
//   - Facilitating unit testing by allowing mock implementations
//   - Enabling runtime configuration of manager behavior
//   - Supporting plugin architectures where manager operations can be swapped
type Linker struct {
	// Merge combines the elements of another manager into this one.
	Merge func(Manager)

	// ModNew creates a new Module with the specified name and file path.
	ModNew func(name string, path string) audmod.Module

	// ModNewOrMerge creates a new module or merges with an existing one if it already exists.
	ModNewOrMerge func(name string, path string) audmod.Module

	// ModGet retrieves a Module by its unique identifier.
	ModGet func(id audids.ID) audmod.Module

	// ModAdd stores a Module in the collection for later retrieval and processing.
	ModAdd func(mod audmod.Module)

	// ModGetOrAdd retrieves an existing module or adds a new one if it doesn't exist.
	ModGetOrAdd func(mod audmod.Module) audmod.Module

	// ModDel removes a Module from the collection by its unique identifier.
	ModDel func(id audids.ID)

	// ModLen returns the number of modules currently stored in the collection.
	ModLen func() int

	// ModWalk iterates over all modules in the collection using the provided callback function.
	ModWalk func(fn func(mod audmod.Module) bool)

	// PkgNew creates a new Package associated with the specified module and path.
	PkgNew func(mod audids.ID, path string) audpkg.Package

	// PkgNewOrMerge creates a new package or merges with an existing one if it already exists.
	PkgNewOrMerge func(mod audids.ID, path string) audpkg.Package

	// PkgGet retrieves a Package by its unique identifier.
	PkgGet func(id audids.ID) audpkg.Package

	// PkgAdd stores a Package in the collection for later retrieval and processing.
	PkgAdd func(pkg audpkg.Package)

	// PkgGetOrAdd retrieves an existing package or adds a new one if it doesn't exist.
	PkgGetOrAdd func(pkg audpkg.Package) audpkg.Package

	// PkgDel removes a Package from the collection by its unique identifier.
	PkgDel func(id audids.ID)

	// PkgLen returns the number of packages currently stored in the collection.
	PkgLen func() int

	// PkgWalk iterates over all packages in the collection using the provided callback function.
	PkgWalk func(fn func(pkg audpkg.Package) bool)

	// PkgSearch performs an exact match search for a package by its import path.
	PkgSearch func(path string) audpkg.Package

	// PkgSearchLike performs a partial match search for packages by their import path.
	PkgSearchLike func(path string) audpkg.Package

	// EntNew creates a new Entry within a package using the specified parameters.
	EntNew func(pkg audids.ID, name string, typ string, codeType audtps.CodeType, lang string, src []byte) audent.Entry

	// EntNewOrMerge creates a new entry or merges with an existing one if it already exists.
	EntNewOrMerge func(pkg audids.ID, name string, typ string, codeType audtps.CodeType, lang string, src []byte) audent.Entry

	// EntGet retrieves an Entry by its unique identifier.
	EntGet func(id audids.ID) audent.Entry

	// EntAdd stores an Entry in the collection for later retrieval and processing.
	EntAdd func(ent audent.Entry)

	// EntGetOrAdd retrieves an existing entry or adds a new one if it doesn't exist.
	EntGetOrAdd func(ent audent.Entry) audent.Entry

	// EntDel removes an Entry from the collection by its unique identifier.
	EntDel func(id audids.ID)

	// EntLen returns the number of entries currently stored in the collection.
	EntLen func() int

	// EntWalk iterates over all entries in the collection using the provided callback function.
	EntWalk func(fn func(ent audent.Entry) bool)
}

// IsEmpty returns true if the Linker has no function pointers assigned.
//
// This method checks whether all function pointers in the Linker are nil.
// It is useful for determining if the Linker has been properly initialized
// with valid function references before attempting to use it.
//
// Returns:
//   - true if any function pointer is nil, indicating the Linker is not fully configured
//   - false if all function pointers are assigned, indicating the Linker is ready for use
func (o *Linker) IsEmpty() bool {
	if o.Merge == nil {
		return true
	}

	if o.ModNew == nil || o.ModNewOrMerge == nil || o.ModGet == nil || o.ModAdd == nil || o.ModGetOrAdd == nil || o.ModDel == nil || o.ModLen == nil || o.ModWalk == nil {
		return true
	}

	if o.PkgNew == nil || o.PkgNewOrMerge == nil || o.PkgGet == nil || o.PkgAdd == nil || o.PkgGetOrAdd == nil || o.PkgDel == nil || o.PkgLen == nil || o.PkgWalk == nil || o.PkgSearch == nil || o.PkgSearchLike == nil {
		return true
	}

	if o.EntNew == nil || o.EntNewOrMerge == nil || o.EntGet == nil || o.EntAdd == nil || o.EntGetOrAdd == nil || o.EntDel == nil || o.EntLen == nil || o.EntWalk == nil {
		return true
	}

	return false
}

// newLinker creates a new Linker instance bound to the given manager.
//
// This function populates a Linker struct with function pointers that reference
// the methods of the provided manager instance. It returns both the manager itself
// and the newly created Linker, allowing callers to access manager operations
// through either direct method calls or through the Linker's function pointers.
//
// Parameters:
//   - m: The manager instance to bind the Linker to
//
// Returns:
//   - The manager instance (m)
//   - A new Linker instance with all function pointers pointing to m's methods
func newLinker(m *mdl) (Manager, *Linker) {
	return m, &Linker{
		Merge:         m.Merge,
		ModNew:        m.ModNew,
		ModNewOrMerge: m.ModNewOrMerge,
		ModGet:        m.ModGet,
		ModAdd:        m.ModAdd,
		ModGetOrAdd:   m.ModGetOrAdd,
		ModDel:        m.ModDel,
		ModLen:        m.ModLen,
		ModWalk:       m.ModWalk,
		PkgNew:        m.PkgNew,
		PkgNewOrMerge: m.PkgNewOrMerge,
		PkgGet:        m.PkgGet,
		PkgAdd:        m.PkgAdd,
		PkgGetOrAdd:   m.PkgGetOrAdd,
		PkgDel:        m.PkgDel,
		PkgLen:        m.PkgLen,
		PkgWalk:       m.PkgWalk,
		PkgSearch:     m.PkgSearch,
		PkgSearchLike: m.PkgSearchLike,
		EntNew:        m.EntNew,
		EntNewOrMerge: m.EntNewOrMerge,
		EntGet:        m.EntGet,
		EntAdd:        m.EntAdd,
		EntGetOrAdd:   m.EntGetOrAdd,
		EntDel:        m.EntDel,
		EntLen:        m.EntLen,
		EntWalk:       m.EntWalk,
	}
}
