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

// File linker.go contains the Linker struct, which holds function pointers
// that reference the methods of the Manager interface.
//
// The Linker provides a way to inject the manager's operations into other
// components of the application, and to replace them for testing, without
// requiring callers to hold a direct reference to the manager instance.

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
	// SearchID computes the identifier of an element from its path.
	//
	// The parts of the path are joined with the path separator of the
	// operating system, and the result is hashed by the audids.GenID helper.
	SearchID func(...string) audids.ID

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

	// ModKeys returns the identifiers of the modules of the collection.
	ModKeys func() []audids.ID

	// ModList returns the paths of the modules of the collection.
	//
	// The modules are filtered on their vendor status, and the returned list
	// is sorted.
	ModList func(vendor bool) []string

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

	// PkgKeys returns the identifiers of the packages of the collection.
	PkgKeys func() []audids.ID

	// PkgModIds returns the identifiers of the modules that contain the
	// packages of the collection.
	//
	// The packages are filtered on their vendor status, and the duplicate
	// identifiers are removed.
	PkgModIds func(vendor bool) []audids.ID

	// PkgList returns the full paths of the packages of the collection.
	//
	// The packages are filtered on the identifier of their module, and the
	// returned list is sorted.
	PkgList func(mod audids.ID) []string

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

	// EntKeys returns the identifiers of the entries of the collection.
	EntKeys func() []audids.ID

	// EntPkgIds returns the identifiers of the packages that contain the
	// entries of the collection.
	//
	// The entries are filtered on their vendor status, and the duplicate
	// identifiers are removed.
	EntPkgIds func(vendor bool) []audids.ID

	// EntDepPkgIds returns the identifiers of the packages that hold the
	// dependencies of the entries of the collection.
	//
	// The entries are filtered on their vendor status, the dependencies are
	// resolved on demand, and the duplicate identifiers are removed.
	EntDepPkgIds func(vendor bool) []audids.ID

	// EntList returns the full paths of the entries of the collection.
	//
	// The entries are filtered on the identifier of their package, and the
	// returned list is sorted.
	EntList func(pkg audids.ID) []string

	// SetStatusNone clears all status bits for the model instance.
	//
	// This method resets the status to the default "None" state where no
	// processing phase is active. After this call, all other status bits
	// are cleared and the instance returns to its initial state.
	SetStatusNone func(audids.ID)

	// SetStatusParsed marks the AST parse phase as completed.
	//
	// This method sets bit 0 to indicate that the AST parsing phase has
	// successfully completed for the model instance.
	SetStatusParsed func(audids.ID)

	// SetStatusVendor marks the AST vendor resolution phase as completed.
	//
	// This method sets bit 1 to indicate that the AST vendor resolution
	// phase has successfully completed for the model instance.
	SetStatusVendor func(audids.ID)

	// SetStatusAnalyzed marks the summary analysis phase as completed.
	//
	// This method sets bit 2 to indicate that the summary analysis phase
	// has successfully completed for the model instance.
	SetStatusAnalyzed func(audids.ID)

	// SetStatusAudited1 marks the first audit report as generated.
	//
	// This method sets bit 3 to indicate that the first audit report
	// has been generated for the model instance.
	SetStatusAudited1 func(audids.ID)

	// SetStatusAudited2 marks the second audit report as generated.
	//
	// This method sets bit 4 to indicate that the second audit report
	// has been generated for the model instance.
	SetStatusAudited2 func(audids.ID)

	// SetStatusAudited3 marks the third audit report as generated.
	//
	// This method sets bit 5 to indicate that the third audit report
	// has been generated for the model instance.
	SetStatusAudited3 func(audids.ID)

	// SetStatusAudited4 marks the fourth audit report as generated.
	//
	// This method sets bit 6 to indicate that the fourth audit report
	// has been generated for the model instance.
	SetStatusAudited4 func(audids.ID)

	// SetStatusOutdated marks the element designated by the given identifier
	// as out of date.
	//
	// This method sets bit 7 to indicate that the model instance has
	// been flagged as out of date.
	SetStatusOutdated func(audids.ID)

	// UnSetStatusParsed clears the AST parse phase status bit.
	//
	// This method clears bit 0 from the model instance status, removing
	// the indication that AST parsing has completed. Other status bits
	// remain unchanged.
	UnSetStatusParsed func(audids.ID)

	// UnSetStatusVendor clears the AST vendor resolution phase status bit.
	//
	// This method clears bit 1 from the model instance status, removing
	// the indication that vendor resolution has completed. Other status
	// bits remain unchanged.
	UnSetStatusVendor func(audids.ID)

	// UnSetStatusAnalyzed clears the summary analysis phase status bit.
	//
	// This method clears bit 2 from the model instance status, removing
	// the indication that summary analysis has completed. Other status
	// bits remain unchanged.
	UnSetStatusAnalyzed func(audids.ID)

	// UnSetStatusAudited1 clears the first audit report status bit.
	//
	// This method clears bit 3 from the model instance status, removing
	// the indication that the first audit report was generated. Other
	// status bits remain unchanged.
	UnSetStatusAudited1 func(audids.ID)

	// UnSetStatusAudited2 clears the second audit report status bit.
	//
	// This method clears bit 4 from the model instance status, removing
	// the indication that the second audit report was generated. Other
	// status bits remain unchanged.
	UnSetStatusAudited2 func(audids.ID)

	// UnSetStatusAudited3 clears the third audit report status bit.
	//
	// This method clears bit 5 from the model instance status, removing
	// the indication that the third audit report was generated. Other
	// status bits remain unchanged.
	UnSetStatusAudited3 func(audids.ID)

	// UnSetStatusAudited4 clears the fourth audit report status bit.
	//
	// This method clears bit 6 from the model instance status, removing
	// the indication that the fourth audit report was generated. Other
	// status bits remain unchanged.
	UnSetStatusAudited4 func(audids.ID)

	// UnSetStatusOutdated clears the out of date status bit.
	//
	// This method clears bit 7 from the model instance status, removing
	// the out of date flag. Other status bits remain unchanged.
	UnSetStatusOutdated func(audids.ID)
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
		Merge:               m.Merge,
		SearchID:            m.SearchID,
		ModNew:              m.ModNew,
		ModNewOrMerge:       m.ModNewOrMerge,
		ModGet:              m.ModGet,
		ModAdd:              m.ModAdd,
		ModGetOrAdd:         m.ModGetOrAdd,
		ModDel:              m.ModDel,
		ModLen:              m.ModLen,
		ModWalk:             m.ModWalk,
		ModKeys:             m.ModKeys,
		ModList:             m.ModList,
		PkgNew:              m.PkgNew,
		PkgNewOrMerge:       m.PkgNewOrMerge,
		PkgGet:              m.PkgGet,
		PkgAdd:              m.PkgAdd,
		PkgGetOrAdd:         m.PkgGetOrAdd,
		PkgDel:              m.PkgDel,
		PkgLen:              m.PkgLen,
		PkgWalk:             m.PkgWalk,
		PkgSearch:           m.PkgSearch,
		PkgSearchLike:       m.PkgSearchLike,
		PkgKeys:             m.PkgKeys,
		PkgModIds:           m.PkgModIds,
		PkgList:             m.PkgList,
		EntNew:              m.EntNew,
		EntNewOrMerge:       m.EntNewOrMerge,
		EntGet:              m.EntGet,
		EntAdd:              m.EntAdd,
		EntGetOrAdd:         m.EntGetOrAdd,
		EntDel:              m.EntDel,
		EntLen:              m.EntLen,
		EntWalk:             m.EntWalk,
		EntKeys:             m.EntKeys,
		EntPkgIds:           m.EntPkgIds,
		EntDepPkgIds:        m.EntDepPkgIds,
		EntList:             m.EntList,
		SetStatusNone:       m.SetStatusNone,
		SetStatusParsed:     m.SetStatusParsed,
		SetStatusVendor:     m.SetStatusVendor,
		SetStatusAnalyzed:   m.SetStatusAnalyzed,
		SetStatusAudited1:   m.SetStatusAudited1,
		SetStatusAudited2:   m.SetStatusAudited2,
		SetStatusAudited3:   m.SetStatusAudited3,
		SetStatusAudited4:   m.SetStatusAudited4,
		SetStatusOutdated:   m.SetStatusOutdated,
		UnSetStatusParsed:   m.UnSetStatusParsed,
		UnSetStatusVendor:   m.UnSetStatusVendor,
		UnSetStatusAnalyzed: m.UnSetStatusAnalyzed,
		UnSetStatusAudited1: m.UnSetStatusAudited1,
		UnSetStatusAudited2: m.UnSetStatusAudited2,
		UnSetStatusAudited3: m.UnSetStatusAudited3,
		UnSetStatusAudited4: m.UnSetStatusAudited4,
		UnSetStatusOutdated: m.UnSetStatusOutdated,
	}
}
