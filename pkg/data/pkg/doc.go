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

// Package pkg provides the model of the Go packages tracked by the auditor.
//
// # Overview
//
// A Package represents a Go package: it holds the path of the package relative
// to its module, the reference to the containing module, the alternative names
// of the package, its lifecycle status, its summary, and its audit reports.
// The package is the main unit of the auditor: every source file is associated
// with a package, and every audit finding is attached to a package.
//
// The package exposes the Package interface, the New and Empty constructors,
// and the RegisterGetModule / FuncPkg helpers. The concrete model is the
// unexported mdl type, which implements the complete Package contract.
//
// # Public API
//
// The following symbols are exported and constitute the public interface of
// the package:
//
//   - Package: the interface that every package model must implement. It
//     embeds the generic auddat.Data[Package] contract and adds accessor
//     and mutator methods specific to Go packages.
//
//   - New(id, path): creates a new Package instance from a module identifier
//     and a package path. Returns nil on validation failure.
//
//   - Empty(): creates a zero-valued Package instance for use as a
//     placeholder or for initialization.
//
//   - RegisterGetModule(f): registers the module resolver function that is
//     required for module lookup during deserialization, merging, and
//     constructor calls.
//
//   - FuncPkg: a function type that creates and returns a new Package. Used
//     as a factory in higher-level components.
//
//   - Export: the intermediate struct used during serialization and
//     deserialization across all formats (json, yaml, toml, cbor).
//
// # Module Resolution
//
// The containing module is not stored as a full object in the package model
// but only as an identifier (audids.ID). The module itself is resolved on
// demand through the resolver function registered with RegisterGetModule.
// This design keeps the package layer decoupled from the module layer and
// avoids circular dependencies.
//
// RegisterGetModule must be called once, during application initialization,
// before any package operations that require module resolution (New, Merge,
// deserialization). If the resolver is not registered, these operations will
// panic on a nil function pointer.
//
// # Lifecycle Status
//
// The status of a package is tracked using bit flags defined in the
// pkg/data/status package. The following flags can be set and queried:
//
//   - Parsed (bit 0): the AST parsing phase has completed for the package.
//   - Vendor (bit 1): the vendor resolution phase has completed.
//   - Analyzed (bit 2): the summary analysis phase has completed.
//   - Audited1 (bit 3): the first audit report has been generated.
//   - Audited2 (bit 4): the second audit report has been generated.
//   - Audited3 (bit 5): the third audit report has been generated.
//   - Audited4 (bit 6): the fourth audit report has been generated.
//   - Outdated (bit 7): the package has been flagged as out of date.
//
// Status flags are managed through the SetStatus/UnSetStatus/IsStatus
// family of methods on the Package interface. SetStatusNone clears all
// flags at once. Multiple flags can coexist, representing the current
// progress through the audit pipeline.
//
// # Reports
//
// A package carries an unbounded collection of audit reports, stored as a
// map keyed on the report name. Reports are managed through SetReports,
// DelReports, GetReports, and LstReports. A report consists of a Name and
// Content field.
//
// # Alias Management
//
// A package can carry alternative names via the AddAlias and DelAlias
// methods. Aliases are checked against existing entries to prevent
// duplicates. A zero-length argument to DelAlias clears all aliases.
// GetAlias returns the internal slice directly; the caller must not modify
// it, and use at concurrency with another write will result in a data race.
//
// # Serialization
//
// Package values can be serialized to and deserialized from json, yaml,
// toml, and cbor. The Export struct is the intermediate representation
// used by all formats. The module itself is not serialized; only the
// module identifier is stored. During deserialization the module will be
// resolved on demand.
//
// Marshalling (Marshal*) does not acquire the mutex, meaning that the
// caller must ensure no concurrent write is in progress; otherwise a data
// race may occur. Unmarshalling (Unmarshal*) acquires the write lock and
// is therefore safe for concurrent use.
//
// Binary marshalling (MarshalBinary / UnmarshalBinary) delegates to the
// cbor implementation.
//
// # Merge Semantics
//
// When Merge(f) is called with another package instance of the same type
// and the same identifier, the following rules apply:
//
//   - If the receiver module is a vendor module and the source is not,
//     the module ID and the package path are overwritten by the source.
//   - Aliases that exist in the source but not in the receiver are appended.
//   - A non-empty summary that differs from the current one overwrites it.
//   - A non-None status that differs from the current one overwrites it.
//   - All source reports with non-empty content are copied into the
//     receiver.
//
// Merge returns false if the types do not match, the identifiers differ,
// or the underlying modules cannot be resolved.
//
// # Thread Safety
//
// All state mutations on the mdl type are protected by a sync.RWMutex.
// Read operations (IsEqual, GetID, GetModId, GetPackage, GetAlias,
// GetStatus, GetSummary, GetReports, LstReports, etc.) acquire a read
// lock. Write operations (AddAlias, DelAlias, SetStatus, SetSummary,
// SetReports, DelReports, Merge, SetStatus*, UnSetStatus*, etc.) acquire
// a write lock. Only the Marshal* family does not hold a lock, by design.
//
// # Data Flow
//
// The data within a Package flows through the following stages, each
// corresponding to a lifecycle status flag:
//
//     +---------------+       +-----------+      +-------------+
//     | Module (id)  ==*----> | Package   |----> | GetFullPath |
//     | identifier    |  ref  | (mdl)     |      |(module+pkg) |
//     +---------------+       +-----------+      +-------------+
//                                   |
//          +------------------------+-------------------+
//          |                        |                   |
//     +----------+             +---------+         +---------+
//     | Aliases  |             | Summary |         | Reports |
//     | []string |             | string  |         | map     |
//     +----------+             +---------+         +---------+
//          |
//      +-------+
//      | Status|
//      | bits  |
//      +-------+
//
// The lifecycle pipeline processes a package as:
//
//   Parse AST --> Vendor Resolution --> Summary Analysis --> Audits
//       [P]           [V]                    [A]         [A1/A2/A3/A4]
//
// Each phase sets its corresponding status bit. The Outdated flag
// is set independently to mark stale packages.
//
// The overall pipeline is:
//
//   New(id, path)
//       |
//       | New validates id != 0, resolves module via fmg, and checks
//       | that the path exists within the module through GetPkg.
//       | On failure New returns nil.
//       v
//   Parsed
//       |
//       |
//   Vendor Resolution
//       |
//       |
//   Summary Analysis
//       |
//       |
//   Audit Report(s) -- SetReports, merged via Merge(f), reported
//   via GetReports / LstReports
//       |
//       |
//   Outdated? -- optional, set by the auditor when a package may be
//   stale or superseded

package pkg
