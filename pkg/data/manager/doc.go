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

// Package manager provides the central data manager for the auditor.
//
// The Manager is the root of the data layer and holds the three collections
// that represent the parsed source code:
//
//  -- Modules: top-level source code units. A module is identified by its
//     import path (e.g. "github.com/example/project"). Each module may
//     contain multiple packages.
//  -- Packages: logical groupings of related code within a module. A package
//     is identified by its module id and its file-system path.
//  -- Entries: individual code constructs such as functions, variables, types,
//     and declarations. An entry belongs to a package and carries the raw
//     source content, its type, the code type category, and the language
//     identifier.
//
// These three collections form the hierarchy: a module contains packages,
// a package contains entries, and each entry may reference its own list of
// dependencies (also entries in other packages).
//
// # New and Linker
//
// The exported New function is the primary constructor. It takes a
// auduxi.Manager (the user-interface manager) and returns two values:
//
//   -- A Manager interface, which exposes all public operations.
//   -- A *Linker, a struct whose fields are function pointers that
//     directly reference the corresponding methods of the manager.
//
// The Linker is designed for dependency injection: callers can pass
// the linker into other components so that those components can invoke
// manager operations through function calls instead of holding a direct
// reference to the Manager interface. This also makes it straightforward
// to replace or mock individual operations during tests. The Linker type
// is exported; callers can modify individual fields before handoff, for
// example to instrument a single operation or to override it with a stub.
//
// The Linker.IsEmpty method checks whether any of its function pointers
// is nil; it returns true when the linker is not fully configured.
//
// # Context
//
// The Manager implements context.Context by delegating to the embedded
// auduxi.Manager. The Deadline, Done, Err, and Value methods are
// forwarded to the UI manager, so that cancellation signals and
// deadlines propagate through every layer that holds a reference to the
// manager. This means a manager can be used anywhere a context.Context
// is expected, for instance to pass a deadline into a child analysis
// goroutine.
//
// # User Interface Manager
//
// The Manager also embeds the auduxi.Manager interface, which provides
// logging methods (Info, InfoData, Warning, WarningData, Error,
// ErrorData, ErrorStack, ErrorStackData, Fatal, FatalStack) and the
// ability to create progress bars (NewBar). All calls are delegated to
// the embedded UI manager so that the analysis messages and the UI
// messages share the same output channel.
//
// # Module Operations
//
// The following set of public, exported methods operate on modules:
//
//   ModNew(name, path string) audmod.Module
//     Creates a new module. name is the module path
//     (e.g. "github.com/foo/bar"), path is the associated file path.
//     Returns the new module, or nil if name is invalid.
//
//   ModNewOrMerge(name, path string) audmod.Module
//     Creates a new module and stores it. If a module with the same
//     identity already exists, the new data is merged into the stored
//     instance. The method returns the NEWLY-created module, NOT the
//     stored instance. When a merge occurred, the caller should use
//     ModGet to retrieve the merged instance from the collection.
//
//   ModGet(id audids.ID) audmod.Module
//     Retrieves a module by ID. Returns nil if not found.
//
//   ModAdd(mod audmod.Module)
//     Stores a module. If a module with the same ID already exists, the
//     data is merged.
//
//   ModGetOrAdd(mod audmod.Module) audmod.Module
//     Looks up the module by identity; returns the existing one if
//     found, otherwise stores and returns the provided one.
//
//   ModDel(id audids.ID)
//     Removes a module and cascades the deletion: all packages that
//     belong to the module are removed first, which in turn cascades
//     to the entries in those packages. This prevents orphaned
//     references.
//
//   ModLen() int
//     Returns the number of modules in the collection.
//
//   ModWalk(fn func(audmod.Module) bool)
//     Walks the module collection. The callback receives each module and
//     returns a boolean: true to continue, false to stop iteration.
//
//   ModKeys() []audids.ID
//     Returns the IDs of all modules.
//
//   ModList(vendor bool) []string
//     Returns the sorted paths of modules, filtered by vendor status.
//     vendor=true lists only vendored modules; vendor=false lists only
//     non-vendored modules.
//
// # Package Operations
//
// The following public methods operate on packages:
//
//   PkgNew(mod audids.ID, path string) audpkg.Package
//     Creates a new package linked to the given module ID. Returns the
//     new package.
//
//   PkgNewOrMerge(mod audids.ID, path string) audpkg.Package
//     Creates and stores a package. If a package with the same identity
//     already exists, the new data is merged. Returns the NEWLY-created
//     package, not the stored instance. Use PkgGet to retrieve the
//     merged instance.
//
//   PkgGet(id audids.ID) audpkg.Package
//     Retrieves a package by ID. Returns nil if not found.
//
//   PkgAdd(pkg audpkg.Package)
//     Stores a package, merging if the same ID already exists.
//
//   PkgGetOrAdd(pkg audpkg.Package) audpkg.Package
//     Looks up by identity; returns existing or stores and returns the
//     provided one.
//
//   PkgDel(id audids.ID)
//     Removes a package and cascades to entries: all entries that
//     belong to the package are removed first.
//
//   PkgLen() int
//     Returns the number of packages.
//
//   PkgWalk(fn func(audpkg.Package) bool)
//     Walks the package collection.
//
//   PkgSearch(path string) audpkg.Package
//     Exact-match search: returns the package whose full path equals
//     the given path.
//
//   PkgSearchLike(path string) audpkg.Package
//     Partial-match search: returns the package whose full path is
//     contained in the given search string and whose full path is
//     longest among the matches. Useful for resolving ambiguous or
//     partial package references.
//
//   PkgKeys() []audids.ID
//     Returns the IDs of all packages.
//
//   PkgModIds(vendor bool) []audids.ID
//     Returns the deduplicated module IDs of the modules that contain
//     the matching packages, filtered by vendor status.
//
//   PkgList(mod audids.ID) []string
//     Returns the sorted full paths of packages. If mod is zero, all
//     packages are returned; otherwise only the packages belonging to
//     that module are returned.
//
// # Entry Operations
//
// The following public methods operate on code entries:
//
//   EntNew(pkg audids.ID, name, typ string, codeType audtps.CodeType,
//           lang string, src []byte) audent.Entry
//     Creates a new entry. pkg is the package ID, name is the element
//     name, typ is its Go type string (e.g. "func", "var"), codeType is
//     the code type category, lang is the language identifier,
//     and src is the raw source content.
//
//   EntNewOrMerge(...) audent.Entry
//     Creates, stores, and possibly merges an entry. Same semantics as
//     PkgNewOrMerge: returns the newly created instance, not the stored
//     one.
//
//   EntGet(id audids.ID) audent.Entry
//     Retrieves an entry by ID.
//
//   EntAdd(ent audent.Entry)
//     Stores an entry, merging if the same ID exists.
//
//   EntGetOrAdd(ent audent.Entry) audent.Entry
//     Looks up by identity; returns existing or stores and returns.
//
//   EntDel(id audids.ID)
//     Removes an entry from the collection.
//
//   EntLen() int
//     Returns the number of entries.
//
//   EntWalk(fn func(audent.Entry) bool)
//     Walks the entry collection.
//
//   EntKeys() []audids.ID
//     Returns the IDs of all entries.
//
//   EntPkgIds(vendor bool) []audids.ID
//     Returns the deduplicated package IDs of packages that contain
//     entries, filtered by vendor status.
//
//   EntDepPkgIds(vendor bool) []audids.ID
//     Returns the deduplicated package IDs of packages that hold the
//     dependencies of the entries, resolved on demand. The dependencies
//     of each entry are walked, and each dependency's package ID is
//     collected.
//
//   EntList(pkg audids.ID) []string
//     Returns the sorted full paths of entries. If pkg is zero, all
//     entries are returned; otherwise only entries belonging to that
//     package are returned.
//
// # Status Mutators
//
// Each element (module, package, entry) carries a status bitmask that
// tracks the processing phases. The manager exposes methods that
// address an element of ANY kind by its identifier and set or clear
// individual bits. The method first looks up the element in the
// appropriate collection (module, package, or entry) and applies the
// change to whichever collection holds it.
//
// The eight status bits and their corresponding mutators are:
//
//   Bit 0 -- Parsed:      SetStatusParsed / UnSetStatusParsed
//              The AST parse phase completed successfully.
//
//   Bit 1 -- Vendor:      SetStatusVendor / UnSetStatusVendor
//              The AST vendor resolution phase completed.
//
//   Bit 2 -- Analyzed:    SetStatusAnalyzed / UnSetStatusAnalyzed
//              The summary analysis phase completed.
//
//   Bit 3 -- Audited1:    SetStatusAudited1 / UnSetStatusAudited1
//              The first audit report was generated.
//
//   Bit 4 -- Audited2:    SetStatusAudited2 / UnSetStatusAudited2
//              The second audit report was generated.
//
//   Bit 5 -- Audited3:    SetStatusAudited3 / UnSetStatusAudited3
//              The third audit report was generated.
//
//   Bit 6 -- Audited4:    SetStatusAudited4 / UnSetStatusAudited4
//              The fourth audit report was generated.
//
//   Bit 7 -- Outdated:    SetStatusOutdated / UnSetStatusOutdated
//              The element has been flagged as out of date.
//
// SetStatusNone clears all bits for an element, resetting it to the
// initial unprocessed state.
//
// # SearchID
//
// SearchID(s ...string) audids.ID computes an element identifier from
// a sequence of path components. The components are joined using the
// operating-system path separator (filepath.Join), and the resulting
// string is hashed by audids.GenID. The returned ID can then be used
// with the Get, Add, Del, or Set/UnSetStatus methods.
//
// # Merge
//
// Merge(Manager) combines all elements from another manager into the
// current one. It iterates over the other manager's modules, then
// packages, then entries, and adds each element via the corresponding
// Add method. Elements with equal identity in both managers will be
// merged by the underlying collections.
//
// # Clean
//
// Clean removes all elements from the three collections in this order:
// entries first, then packages, then modules. This ensures that
// cascading deletion invariants are respected.
//
// # Serialization
//
// The Manager interface declares marshaler and unmarshaler methods for
// four formats, implemented by the mdl concrete type:
//
//   -- JSON:   MarshalJSON / UnmarshalJSON
//   -- YAML:   MarshalYAML / UnmarshalYAML
//   -- TOML:   MarshalTOML / UnmarshalTOML
//   -- CBOR:   MarshalCBOR / UnmarshalCBOR
//
// The MarshalBinary and UnmarshalBinary methods delegate to the CBOR
// implementation, so binary serialization is equivalent to CBOR.
//
// Internally, marshaling creates an Export struct that holds the three
// collections (Mod, Pkg, Ent). The field tags are "lib" for modules,
// "pkg" for packages, and "ent" for entries in JSON, YAML, and TOML.
// Unmarshaling deserializes into a new Export, then applies the three
// collections to the current manager instance via unmarshallApply, which
// replaces the internal collections atomically.
//
// # Data Flow
//
// The typical processing pipeline follows this flow:
//
//  1. Source code is read and the AST is parsed.
//  2. Modules are created as top-level nodes via ModNew or ModNewOrMerge.
//  3. Packages are created under their containing module via PkgNew or
//     PkgNewOrMerge, linking each package to its module ID.
//  4. Entries (functions, types, variables, etc.) are created under their
//     containing package via EntNew or EntNewOrMerge.
//  5. As each element is added, its status bits are updated to reflect
//     the current processing phase (Parsed, Vendor, Analyzed, etc.).
//  6. Once parsing is complete, EntDepPkgIds can be used to resolve the
//     cross-package dependencies and identify all packages that are
//     referenced as dependencies by the entries.
//  7. The status bits drive subsequent analysis phases: each auditor
//     reads the status of an element to determine whether it needs
//     processing and writes the relevant bit when done.
//
// The Manager is thread-safe at the collection level: the underlying
// audcol.Collection provides safe access, and each element type
// (audmod.Module, audpkg.Package, audent.Entry) is independently
// thread-safe for its internal state mutators.
//
// # Example Usage
//
//   uim := uxi.New(...)              // create the UI manager
//   mngr, linker := manager.New(uim)
//
//   // Create and link a module, a package, and an entry
//   mod := mngr.ModNewOrMerge("github.com/foo/bar", "/path/to/go.mod")
//
//   pkg := mngr.PkgNewOrMerge(mod.GetID(), "/path/to/pkg")
//
//   ent := mngr.EntNewOrMerge(
//       pkg.GetID(),
//       "MyFunc",
//       "func",
//       audtps.CodeTypeFunction,
//       "go",
//       []byte("package pkg; func MyFunc() {}"),
//   )
//
//   mngr.SetStatusParsed(ent.GetID())
//
//   // Walk all entries in a package
//   mngr.EntWalk(func(e audent.Entry) bool {
//       fmt.Println(e.GetFullPath())
//       return true
//   })

package manager
