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

// File package.go contains the package level operations of the Manager
// interface.
//
// It creates, merges, retrieves, deletes, walks over, searches and lists the
// packages known by the manager.

package manager

import (
	"slices"
	"strings"

	audids "github.com/nabbar/auditor/pkg/data/id"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
)

// PkgNew creates a new Package associated with the specified module and file system path.
//
// This method directly creates a new package instance using the audpkg.New function,
// bypassing any registered callback functions since this implementation provides
// the default behavior for package creation. A package represents a logical grouping
// of code elements within a module that organizes related functionality.
//
// Parameters:
//   - mod: The module ID to associate with the new package
//   - path: The file system path for the package, usually representing the package directory
//
// Returns:
//   - A newly created Package instance that represents a logical grouping of code elements
func (o *mdl) PkgNew(mod audids.ID, path string) audpkg.Package {
	return audpkg.New(mod, path)
}

// PkgNewOrMerge creates a new package, or merges it with an existing one when
// it already exists.
//
// This method attempts to create a new package with the given module
// identifier and path. The new package is then stored in the collection: when
// a package with the same identity already exists, the data of the new package
// is merged into the existing one, so that no duplicate is created.
//
// Note:
//   - The method returns the newly created package, and not the instance that
//     is stored in the collection. When a merge occurred, the stored instance
//     is the one that has been updated.
//
// Parameters:
//   - mod: The identifier of the module to associate with the new package.
//   - path: The path of the file system of the package, usually the directory
//     of the package.
//
// Returns:
//   - audpkg.Package: The newly created Package instance, or nil when the
//     package could not be created.
func (o *mdl) PkgNewOrMerge(mod audids.ID, path string) audpkg.Package {
	p := audpkg.New(mod, path)
	if p == nil {
		return nil
	}

	o.pkg.Store(p)
	return p
}

// PkgGet retrieves a Package from the collection by its unique identifier.
//
// This method fetches an existing package from the internal package collection using
// its unique identifier. It provides access to previously created packages for further processing
// or analysis within the AST management system. If no package with the specified ID exists,
// this method returns nil.
//
// Parameters:
//   - id: The unique identifier of the package to retrieve
//
// Returns:
//   - The Package instance if found in the collection, otherwise nil
func (o *mdl) PkgGet(id audids.ID) audpkg.Package {
	return o.pkg.Load(id)
}

// PkgAdd stores a Package in the collection for later retrieval and processing.
//
// This method persists a package in the internal package collection for use during AST parsing operations.
// If a package with the same ID already exists and is not empty, it merges the new package with the existing one
// before storing to maintain data integrity. This approach ensures that packages can be incrementally built
// or updated rather than replaced entirely.
//
// Parameters:
//   - pkg: The Package instance to add to the collection
func (o *mdl) PkgAdd(pkg audpkg.Package) {
	o.pkg.Merge(pkg)
}

// PkgGetOrAdd retrieves an existing package or adds a new one if it doesn't exist.
//
// This method first attempts to find a package matching the provided package's identity.
// If found, it returns the existing package. If not found, it adds the provided package
// to the collection and returns it.
//
// Parameters:
//   - pkg: The Package instance to retrieve or add
//
// Returns:
//   - The existing or newly added Package instance
func (o *mdl) PkgGetOrAdd(pkg audpkg.Package) audpkg.Package {
	return o.pkg.GetOrStore(pkg)
}

// PkgDel removes a Package from the collection by its unique identifier.
//
// This method deletes a package from the internal package collection using its unique identifier.
// It also removes any associated entries for the package before deleting the package itself,
// ensuring that all related data is properly cleaned up. This cleanup process prevents
// orphaned entry references and maintains consistency in the AST structure.
//
// Parameters:
//   - id: The unique identifier of the package to delete
func (o *mdl) PkgDel(id audids.ID) {
	o.entDelForPkg(id)
	o.pkg.Delete(id)
}

// PkgLen returns the number of packages currently stored in the collection.
//
// This method provides a count of all packages managed by this manager instance,
// which can be useful for monitoring or debugging purposes. It accesses the underlying
// package collection's length method to return the total number of stored packages.
//
// Returns:
//   - The total number of packages in the internal collection
func (o *mdl) PkgLen() int {
	return o.pkg.Len()
}

// PkgWalk iterates over all packages in the collection using the provided callback function.
//
// This method executes a callback function for each package in the internal package collection,
// allowing for processing or analysis of all stored packages. The iteration continues until
// all packages have been processed or until the callback returns false, indicating early termination.
//
// Parameters:
//   - f: A callback function that receives a package and returns a boolean indicating
//     whether to continue iteration (true) or stop (false)
func (o *mdl) PkgWalk(f func(audpkg.Package) bool) {
	o.pkg.Walk(f)
}

// PkgSearch performs an exact match search for a package by its full import path.
//
// This method searches through all packages in the internal collection to find one
// whose full path exactly matches the specified search path. It is useful when looking
// for a specific package by its precise import path, such as during import resolution
// or dependency analysis.
//
// Parameters:
//   - path: The full import path to search for (e.g., "github.com/example/project/pkg")
//
// Returns:
//   - The Package instance if found, otherwise nil
func (o *mdl) PkgSearch(path string) audpkg.Package {
	var ids audids.ID

	o.PkgWalk(func(pkg audpkg.Package) bool {
		if pkg.GetFullPath() == path {
			ids = pkg.GetID()
			return false
		}

		return true
	})

	return o.PkgGet(ids)
}

// PkgSearchLike performs a partial match search for the packages by their
// import paths.
//
// This method searches through all the packages of the collection, and keeps
// the ones whose full path is contained in the given search string. It returns
// the package with the longest matching path, which is particularly useful
// when dealing with the partial or the ambiguous package references, such as
// during the resolution or the completion of a package name.
//
// Note:
//   - The containment is tested in that direction: the full path of the
//     package has to be a part of the search string.
//
// Parameters:
//   - path: The partial import path to search for, for example
//     "example/project/pkg".
//
// Returns:
//   - audpkg.Package: The Package instance if found, otherwise nil.
func (o *mdl) PkgSearchLike(path string) audpkg.Package {
	var (
		pth string
		ids audids.ID
	)

	o.PkgWalk(func(pkg audpkg.Package) bool {
		ful := pkg.GetFullPath()

		if !strings.Contains(path, ful) {
			return true
		}

		// Update the best match if this package has a longer matching path
		// or if no match has been found yet
		if len(pth) == 0 || len(ful) > len(pth) {
			pth = ful
			ids = pkg.GetID()
		}

		return true
	})

	return o.PkgGet(ids)
}

// PkgKeys returns the identifiers of the packages of the collection.
//
// Returns:
//   - []audids.ID: The identifiers of the stored packages.
func (o *mdl) PkgKeys() []audids.ID {
	return o.pkg.Keys()
}

// PkgModIds returns the identifiers of the modules that contain the packages
// of the collection.
//
// The packages are filtered on their vendor status, and the duplicate
// identifiers are removed.
//
// Parameters:
//   - vendor: true to take the vendored packages into account, false to
//     ignore them.
//
// Returns:
//   - []audids.ID: The identifiers of the modules that hold the matching
//     packages.
func (o *mdl) PkgModIds(vendor bool) []audids.ID {
	var ids = make([]audids.ID, 0, o.mod.Len())

	o.pkg.Walk(func(pkg audpkg.Package) bool {
		if pkg == nil || pkg.IsEmpty() {
			return true
		}

		if !vendor && pkg.IsVendor() {
			return true
		}

		if mid := pkg.GetModId(); !slices.Contains(ids, mid) {
			ids = append(ids, mid)
		}

		return true
	})

	return ids
}

// pkgDelForMod deletes all packages associated with a specific module.
//
// This internal helper method removes all packages from the collection that belong
// to the specified module ID. It is called during module deletion to ensure that
// all related package data is properly cleaned up before removing the module itself.
// This prevents orphaned package references and maintains data consistency.
//
// Parameters:
//   - id: The unique identifier of the module for which packages should be deleted
func (o *mdl) pkgDelForMod(id audids.ID) {
	var ids = make([]audids.ID, 0, o.pkg.Len())

	o.PkgWalk(func(pkg audpkg.Package) bool {
		if pkg.GetMod().GetID() == id {
			ids = append(ids, pkg.GetID())
		}
		return true
	})

	for _, k := range ids {
		o.PkgDel(k)
	}
}

// PkgList returns the full paths of the packages of the collection.
//
// The packages are filtered on the identifier of their module, and the
// returned list is sorted.
//
// Parameters:
//   - mod: The identifier of the module that has to be taken into account.
//     The zero identifier matches every module.
//
// Returns:
//   - []string: The sorted full paths of the matching packages.
func (o *mdl) PkgList(mod audids.ID) []string {
	var l = make([]string, 0, o.pkg.Len())

	o.pkg.Walk(func(pkg audpkg.Package) bool {
		if pkg == nil || pkg.IsEmpty() {
			return true
		}

		if mod > 0 && pkg.GetModId() != mod {
			return true
		}

		l = append(l, pkg.GetFullPath())
		return true
	})

	slices.Sort(l)
	return l
}
