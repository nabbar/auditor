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

// Package manager provides the manager of the data collected by the auditor.
//
// The Manager holds the three collections of the data layer: the modules, the
// packages and the code entries. It exposes the creation, the retrieval, the
// addition, the deletion, the listing and the search of the elements of these
// collections, and it delegates the status mutators to the element that the
// identifier designates.
//
// The Manager also embeds the contract of the user interface manager, which
// provides the logging, the progress bars and the cancellation of the
// operations.
//
// # Linker
//
// The New constructor returns a Manager and a Linker. The Linker holds the
// function pointers of the operations of the manager, so that the operations
// can be injected into the other components of the application, and replaced
// for the tests.
//
// # Serialization
//
// The manager supports the json, yaml, toml and cbor formats.
package manager

import (
	"encoding/json"

	"github.com/fxamacker/cbor/v2"
	audcol "github.com/nabbar/auditor/pkg/data/collection"
	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	audtps "github.com/nabbar/auditor/pkg/data/types"
	auduxi "github.com/nabbar/auditor/pkg/uxi"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// Manager defines the interface for managing collections of AST elements.
//
// This interface provides methods to create, retrieve, add, and delete modules, packages,
// and entries within the AST processing system. It serves as a central point for managing
// the various components that make up the parsed code structure.
//
// The Manager interface is designed to abstract the operations on AST elements, enabling
// flexible implementations of different AST parsing strategies while maintaining a consistent
// API for interacting with the parsed data.
//
// This abstraction layer allows for:
//   - Uniform access patterns across different AST parsing backends
//   - Decoupling of parsing logic from storage mechanisms
//   - Extensibility through custom implementations
//   - Consistent handling of code elements regardless of language or parsing approach
type Manager interface {
	json.Marshaler
	json.Unmarshaler
	yaml.Marshaler
	yaml.Unmarshaler
	toml.Marshaler
	toml.Unmarshaler
	cbor.Marshaler
	cbor.Unmarshaler

	auduxi.Manager

	// Clean removes all the elements of the three collections of the manager.
	//
	// The entries, the packages and the modules are removed, in that order.
	//
	// Returns:
	//   - None
	Clean()
	// Merge combines the elements of another manager into this one.
	//
	// The modules, the packages and the entries of the other manager are added
	// to the corresponding collections of this manager.
	//
	// Parameters:
	//   - Manager: The manager whose elements have to be merged into this one.
	//
	// Returns:
	//   - None
	Merge(Manager)
	// SearchID computes the identifier of an element from its path.
	//
	// The parts of the path are joined with the path separator of the
	// operating system, and the result is hashed by the audids.GenID helper.
	//
	// Parameters:
	//   - ...string: The parts of the path of the element.
	//
	// Returns:
	//   - audids.ID: The identifier of the element designated by the path.
	SearchID(...string) audids.ID

	// ModNew creates a new Module with the specified name and file path.
	//
	// This method initializes a new module instance with the provided name and file path.
	// A module represents a complete source code unit that can contain multiple packages.
	// It is typically used when parsing Go source files to create top-level modules.
	//
	// Parameters:
	//   - name: The name of the module (e.g., "github.com/example/project")
	//   - path: The file path associated with the module, usually the main package file
	//
	// Returns:
	//   - A newly created Module instance that represents a complete source code unit
	ModNew(name string, path string) audmod.Module

	// ModNewOrMerge creates a new Module or merges with an existing one if it already exists.
	//
	// This method attempts to create a new module with the given name and path.
	// If a module with the same identity already exists in the collection, it merges
	// the new data into the existing module instead of creating a duplicate.
	//
	// Parameters:
	//   - name: The name of the module
	//   - path: The file path associated with the module
	//
	// Returns:
	//   - The newly created or merged Module instance
	ModNewOrMerge(name string, path string) audmod.Module

	// ModGet retrieves a Module by its unique identifier.
	//
	// This method fetches an existing module from the collection using its unique identifier.
	// It allows for retrieving previously created modules when needed for further processing.
	// If no module with the specified ID exists, this method returns nil.
	//
	// Parameters:
	//   - id: The unique identifier of the module to retrieve
	//
	// Returns:
	//   - The Module instance if found, otherwise nil
	ModGet(id audids.ID) audmod.Module

	// ModAdd stores a Module in the collection for later retrieval and processing.
	//
	// This method persists a module in the collection for use during AST parsing operations.
	// It is used to maintain state of modules that have been created during AST parsing.
	//
	// Parameters:
	//   - mod: The Module instance to add to the collection
	ModAdd(mod audmod.Module)

	// ModGetOrAdd retrieves an existing Module or adds a new one if it doesn't exist.
	//
	// This method first attempts to find a module matching the provided module's identity.
	// If found, it returns the existing module. If not found, it adds the provided module
	// to the collection and returns it.
	//
	// Parameters:
	//   - mod: The Module instance to retrieve or add
	//
	// Returns:
	//   - The existing or newly added Module instance
	ModGetOrAdd(mod audmod.Module) audmod.Module

	// ModDel removes a Module from the collection by its unique identifier.
	//
	// This method deletes a module from the collection using its unique identifier.
	// It is typically called when cleaning up or removing modules that are no longer needed,
	// such as during resource cleanup or when processing changes in source code structure.
	//
	// Parameters:
	//   - id: The unique identifier of the module to delete
	ModDel(id audids.ID)

	// ModLen returns the number of modules currently stored in the collection.
	//
	// This method provides a count of all modules managed by this manager instance,
	// which can be useful for monitoring or debugging purposes.
	//
	// Returns:
	//   - The total number of modules in the collection
	ModLen() int

	// ModWalk iterates over all modules in the collection using the provided callback function.
	//
	// This method executes a callback function for each module in the collection,
	// allowing for processing or analysis of all stored modules.
	// If the callback returns false, iteration stops early.
	//
	// Parameters:
	//   - fn: A callback function that receives a module and returns a boolean indicating
	//     whether to continue iteration (true) or stop (false)
	ModWalk(fn func(mod audmod.Module) bool)

	// ModKeys returns the identifiers of the modules of the collection.
	//
	// Returns:
	//   - []audids.ID: The identifiers of the stored modules.
	ModKeys() []audids.ID
	// ModList returns the paths of the modules of the collection.
	//
	// The modules are filtered on their vendor status, and the returned list
	// is sorted.
	//
	// Parameters:
	//   - vendor: true to list only the vendored modules, false to list only
	//     the modules that are not vendored.
	//
	// Returns:
	//   - []string: The sorted paths of the matching modules.
	ModList(vendor bool) []string

	// PkgNew creates a new Package associated with the specified module and path.
	//
	// This method initializes a new package instance associated with a specific module.
	// Packages represent logical units within a module that contain related code elements.
	// They are typically created when parsing Go source files to organize code by package boundaries.
	//
	// Parameters:
	//   - mod: The module ID to associate with the new package
	//   - path: The file system path for the package, usually representing the package directory
	//
	// Returns:
	//   - A newly created Package instance that represents a logical grouping of code elements
	PkgNew(mod audids.ID, path string) audpkg.Package

	// PkgNewOrMerge creates a new Package or merges with an existing one if it already exists.
	//
	// This method attempts to create a new package with the given module ID and path.
	// If a package with the same identity already exists in the collection, it merges
	// the new data into the existing package instead of creating a duplicate.
	//
	// Parameters:
	//   - mod: The module ID to associate with the new package
	//   - path: The file system path for the package
	//
	// Returns:
	//   - The newly created or merged Package instance
	PkgNewOrMerge(mod audids.ID, path string) audpkg.Package

	// PkgGet retrieves a Package by its unique identifier.
	//
	// This method fetches an existing package from the collection using its unique identifier.
	// It allows for retrieving previously created packages when needed for further processing.
	// If no package with the specified ID exists, this method returns nil.
	//
	// Parameters:
	//   - id: The unique identifier of the package to retrieve
	//
	// Returns:
	//   - The Package instance if found, otherwise nil
	PkgGet(id audids.ID) audpkg.Package

	// PkgAdd stores a Package in the collection for later retrieval and processing.
	//
	// This method persists a package in the collection for use during AST parsing operations.
	// It is used to maintain state of packages that have been created during AST parsing.
	//
	// Parameters:
	//   - pkg: The Package instance to add to the collection
	PkgAdd(pkg audpkg.Package)

	// PkgGetOrAdd retrieves an existing Package or adds a new one if it doesn't exist.
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
	PkgGetOrAdd(pkg audpkg.Package) audpkg.Package

	// PkgDel removes a Package from the collection by its unique identifier.
	//
	// This method deletes a package from the collection using its unique identifier.
	// It is typically called when cleaning up or removing packages that are no longer needed,
	// such as during resource cleanup or when processing changes in source code structure.
	//
	// Parameters:
	//   - id: The unique identifier of the package to delete
	PkgDel(id audids.ID)

	// PkgLen returns the number of packages currently stored in the collection.
	//
	// This method provides a count of all packages managed by this manager instance,
	// which can be useful for monitoring or debugging purposes.
	//
	// Returns:
	//   - The total number of packages in the collection
	PkgLen() int

	// PkgWalk iterates over all packages in the collection using the provided callback function.
	//
	// This method executes a callback function for each package in the collection,
	// allowing for processing or analysis of all stored packages.
	// If the callback returns false, iteration stops early.
	//
	// Parameters:
	//   - fn: A callback function that receives a package and returns a boolean indicating
	//     whether to continue iteration (true) or stop (false)
	PkgWalk(fn func(pkg audpkg.Package) bool)

	// PkgSearch performs an exact match search for a package by its import path.
	//
	// This method searches for a package that exactly matches the given import path.
	// It is useful when looking for a specific package by its precise import path,
	// such as during import resolution or dependency analysis.
	//
	// Parameters:
	//   - path: The full import path to search for (e.g., "github.com/example/project/pkg")
	//
	// Returns:
	//   - The Package instance if found, otherwise nil
	PkgSearch(path string) audpkg.Package

	// PkgSearchLike performs a partial match search for the packages by their
	// import path.
	//
	// This method searches for the packages whose full path is contained in the
	// given search string. It returns the package with the longest matching
	// path, which is particularly useful when dealing with the partial or the
	// ambiguous package references, such as during the resolution or the
	// completion of a package name.
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
	PkgSearchLike(path string) audpkg.Package

	// PkgKeys returns the identifiers of the packages of the collection.
	//
	// Returns:
	//   - []audids.ID: The identifiers of the stored packages.
	PkgKeys() []audids.ID

	// PkgModIds returns the identifiers of the modules that contain the
	// packages of the collection.
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
	PkgModIds(vendor bool) []audids.ID

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
	PkgList(mod audids.ID) []string

	// EntNew creates a new Entry within a package using the specified parameters.
	//
	// This method initializes a new entry (code element) within a package. Entries represent individual
	// code constructs such as functions, variables, or types that are part of a package.
	// The entry creation process is essential for capturing specific elements from the AST during parsing.
	//
	// Parameters:
	//   - pkg: The package ID to associate with the new entry (provides context for the entry)
	//   - name: The name of the source element (e.g., function name, variable name)
	//   - typ: The type of the source element (e.g., "func", "var", "type")
	//   - codeType: The code type identifier (defines the category of code element)
	//   - lang: The language identifier for the source content
	//   - src: The raw source content (byte slice containing the actual source code)
	//
	// Returns:
	//   - A newly created Entry instance that represents an individual code element
	EntNew(pkg audids.ID, name string, typ string, codeType audtps.CodeType, lang string, src []byte) audent.Entry

	// EntNewOrMerge creates a new Entry or merges with an existing one if it already exists.
	//
	// This method attempts to create a new entry with the given parameters.
	// If an entry with the same identity already exists in the collection, it merges
	// the new data into the existing entry instead of creating a duplicate.
	//
	// Parameters:
	//   - pkg: The package ID to associate with the new entry
	//   - name: The name of the source element
	//   - typ: The type of the source element
	//   - codeType: The code type identifier
	//   - lang: The language identifier for the source content
	//   - src: The raw source content
	//
	// Returns:
	//   - The newly created or merged Entry instance
	EntNewOrMerge(pkg audids.ID, name string, typ string, codeType audtps.CodeType, lang string, src []byte) audent.Entry

	// EntGet retrieves an Entry by its unique identifier.
	//
	// This method fetches an existing entry from the collection using its unique identifier.
	// It allows for retrieving previously created entries when needed for further processing or analysis.
	// If no entry with the specified ID exists, this method returns nil.
	//
	// Parameters:
	//   - id: The unique identifier of the entry to retrieve
	//
	// Returns:
	//   - The Entry instance if found, otherwise nil
	EntGet(id audids.ID) audent.Entry

	// EntAdd stores an Entry in the collection for later retrieval and processing.
	//
	// This method persists an entry in the collection for use during AST parsing operations.
	// It is used to maintain state of entries that have been created during AST parsing.
	//
	// Parameters:
	//   - ent: The Entry instance to add to the collection
	EntAdd(ent audent.Entry)

	// EntGetOrAdd retrieves an existing Entry or adds a new one if it doesn't exist.
	//
	// This method first attempts to find an entry matching the provided entry's identity.
	// If found, it returns the existing entry. If not found, it adds the provided entry
	// to the collection and returns it.
	//
	// Parameters:
	//   - ent: The Entry instance to retrieve or add
	//
	// Returns:
	//   - The existing or newly added Entry instance
	EntGetOrAdd(ent audent.Entry) audent.Entry

	// EntDel removes an Entry from the collection by its unique identifier.
	//
	// This method deletes an entry from the collection using its unique identifier.
	// It is typically called when cleaning up or removing entries that are no longer needed,
	// such as during resource cleanup or when processing changes in source code structure.
	//
	// Parameters:
	//   - id: The unique identifier of the entry to delete
	EntDel(id audids.ID)

	// EntLen returns the number of entries currently stored in the collection.
	//
	// This method provides a count of all entries managed by this manager instance,
	// which can be useful for monitoring or debugging purposes.
	//
	// Returns:
	//   - The total number of entries in the collection
	EntLen() int

	// EntWalk iterates over all entries in the collection using the provided callback function.
	//
	// This method executes a callback function for each entry in the collection,
	// allowing for processing or analysis of all stored entries.
	// If the callback returns false, iteration stops early.
	//
	// Parameters:
	//   - fn: A callback function that receives an entry and returns a boolean indicating
	//     whether to continue iteration (true) or stop (false)
	EntWalk(fn func(ent audent.Entry) bool)

	// EntKeys returns the identifiers of the entries of the collection.
	//
	// Returns:
	//   - []audids.ID: The identifiers of the stored entries.
	EntKeys() []audids.ID

	// EntPkgIds returns the identifiers of the packages that contain the
	// entries of the collection.
	//
	// The entries are filtered on their vendor status, and the duplicate
	// identifiers are removed.
	//
	// Parameters:
	//   - vendor: true to take the vendored entries into account, false to
	//     ignore them.
	//
	// Returns:
	//   - []audids.ID: The identifiers of the packages that hold the matching
	//     entries.
	EntPkgIds(vendor bool) []audids.ID

	// EntDepPkgIds returns the identifiers of the packages that hold the
	// dependencies of the entries of the collection.
	//
	// The entries are filtered on their vendor status, the dependencies are
	// resolved on demand, and the duplicate identifiers are removed.
	//
	// Parameters:
	//   - vendor: true to take the vendored entries into account, false to
	//     ignore them.
	//
	// Returns:
	//   - []audids.ID: The identifiers of the packages that hold the
	//     dependencies of the matching entries.
	EntDepPkgIds(vendor bool) []audids.ID

	// EntList returns the full paths of the entries of the collection.
	//
	// The entries are filtered on the identifier of their package, and the
	// returned list is sorted.
	//
	// Parameters:
	//   - pkg: The identifier of the package that has to be taken into account.
	//     The zero identifier matches every package.
	//
	// Returns:
	//   - []string: The sorted full paths of the matching entries.
	EntList(pkg audids.ID) []string

	// SetStatusNone clears all status bits for the model instance.
	//
	// This method resets the status to the default "None" state where no
	// processing phase is active. After this call, all other status bits
	// are cleared and the instance returns to its initial state.
	SetStatusNone(audids.ID)

	// SetStatusParsed marks the AST parse phase as completed.
	//
	// This method sets bit 0 to indicate that the AST parsing phase has
	// successfully completed for the model instance.
	SetStatusParsed(audids.ID)

	// SetStatusVendor marks the AST vendor resolution phase as completed.
	//
	// This method sets bit 1 to indicate that the AST vendor resolution
	// phase has successfully completed for the model instance.
	SetStatusVendor(audids.ID)

	// SetStatusAnalyzed marks the summary analysis phase as completed.
	//
	// This method sets bit 2 to indicate that the summary analysis phase
	// has successfully completed for the model instance.
	SetStatusAnalyzed(audids.ID)

	// SetStatusAudited1 marks the first audit report as generated.
	//
	// This method sets bit 3 to indicate that the first audit report
	// has been generated for the model instance.
	SetStatusAudited1(audids.ID)

	// SetStatusAudited2 marks the second audit report as generated.
	//
	// This method sets bit 4 to indicate that the second audit report
	// has been generated for the model instance.
	SetStatusAudited2(audids.ID)

	// SetStatusAudited3 marks the third audit report as generated.
	//
	// This method sets bit 5 to indicate that the third audit report
	// has been generated for the model instance.
	SetStatusAudited3(audids.ID)

	// SetStatusAudited4 marks the fourth audit report as generated.
	//
	// This method sets bit 6 to indicate that the fourth audit report
	// has been generated for the model instance.
	SetStatusAudited4(audids.ID)

	// SetStatusOutdated marks the element designated by the given identifier
	// as out of date.
	//
	// This method sets bit 7 to indicate that the model instance has
	// been flagged as out of date.
	SetStatusOutdated(audids.ID)

	// UnSetStatusParsed clears the AST parse phase status bit.
	//
	// This method clears bit 0 from the model instance status, removing
	// the indication that AST parsing has completed. Other status bits
	// remain unchanged.
	UnSetStatusParsed(audids.ID)

	// UnSetStatusVendor clears the AST vendor resolution phase status bit.
	//
	// This method clears bit 1 from the model instance status, removing
	// the indication that vendor resolution has completed. Other status
	// bits remain unchanged.
	UnSetStatusVendor(audids.ID)

	// UnSetStatusAnalyzed clears the summary analysis phase status bit.
	//
	// This method clears bit 2 from the model instance status, removing
	// the indication that summary analysis has completed. Other status
	// bits remain unchanged.
	UnSetStatusAnalyzed(audids.ID)

	// UnSetStatusAudited1 clears the first audit report status bit.
	//
	// This method clears bit 3 from the model instance status, removing
	// the indication that the first audit report was generated. Other
	// status bits remain unchanged.
	UnSetStatusAudited1(audids.ID)

	// UnSetStatusAudited2 clears the second audit report status bit.
	//
	// This method clears bit 4 from the model instance status, removing
	// the indication that the second audit report was generated. Other
	// status bits remain unchanged.
	UnSetStatusAudited2(audids.ID)

	// UnSetStatusAudited3 clears the third audit report status bit.
	//
	// This method clears bit 5 from the model instance status, removing
	// the indication that the third audit report was generated. Other
	// status bits remain unchanged.
	UnSetStatusAudited3(audids.ID)

	// UnSetStatusAudited4 clears the fourth audit report status bit.
	//
	// This method clears bit 6 from the model instance status, removing
	// the indication that the fourth audit report was generated. Other
	// status bits remain unchanged.
	UnSetStatusAudited4(audids.ID)

	// UnSetStatusOutdated clears the out of date status bit.
	//
	// This method clears bit 7 from the model instance status, removing
	// the out of date flag. Other status bits remain unchanged.
	UnSetStatusOutdated(audids.ID)
}

// New creates and returns a new Manager instance along with its Linker.
//
// This function initializes a manager with the provided user interface
// manager, and with the default implementations of the operations. It sets up
// the internal structure that manages the elements of the data layer, with the
// default callback functions.
//
// The returned Manager instance provides the management capabilities, and the
// returned Linker provides the same capabilities as function pointers, so that
// they can be injected into the other components of the application.
//
// Parameters:
//   - uim: The user interface manager to use for the user interactions, that
//     provides the context of the user facing operations and of the feedback
//     during the processing.
//
// Returns:
//   - Manager: A new Manager instance, ready to be used.
//   - *Linker: A new Linker instance, whose function pointers all reference
//     the methods of the manager.
func New(uim auduxi.Manager) (Manager, *Linker) {
	m := &mdl{
		sem: uim,
		mod: audcol.New[audmod.Module](audmod.Empty),
		pkg: audcol.New[audpkg.Package](audpkg.Empty),
		ent: audcol.New[audent.Entry](audent.Empty),
	}

	return newLinker(m)
}
