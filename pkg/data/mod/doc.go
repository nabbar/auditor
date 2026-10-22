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

// Package mod provides the data model of Go modules tracked by the auditor.
//
// # Module
//
// The public interface of the package is Module, which embeds the generic
// auddat.Data[Module] contract. That contract provides identity management
// through a unique audids.ID, lifecycle status tracking as a bit field
// (see the status sub-package), summary text, and keyed audit report handling.
// On top of these shared facilities, the Module interface adds accessors and
// mutators specific to a Go module.
//
// A Module instance holds:
//
//   - The canonical module path (e.g. "github.com/user/repo").
//   - The local file path of the associated go.mod file or a .go source file.
//   - The lifecycle status as a bit field with eight phases.
//   - A brief summary description of the module.
//   - A collection of audit reports keyed by name.
//
// The vendor status is not stored as a field. It is computed on demand by the
// IsVendor method, which returns true when the file path is empty or contains
// "/vendor/". This keeps the model lightweight and avoids stale state after
// the file path is updated via SetFile.
//
// # Construction
//
// Modules are created through two exported constructors:
//
//   - New(mod, file string) Module: builds a Module from a module path and
//     an associated file path. The module path is normalized by CleanPath;
//     if the resulting path is empty, New returns nil. The ID is derived from
//     the cleaned module path via audids.GenID. The status is initialized to
//     audsts.None, the summary is empty, and the report map is nil.
//   - Empty() Module: builds an uninitialized Module with zero fields. This is
//     useful as a placeholder or as a sentinel value when no concrete module
//     information is available.
//
// A FuncMod is exported as a constructor function type (func() Module). It can
// be used in factory patterns to decouple callers from the concrete Module
// implementation.
//
// # Data contract methods (from auddat.Data[Module])
//
// Every Module instance provides the following methods inherited from the
// generic Data[M] contract, which is shared by all auditor models (modules,
// packages, and code entries):
//
//   - GetID(): returns the unique, immutable identifier derived at construction.
//   - IsEmpty(): reports whether the module carries enough data to be
//     processed. It checks the ID, the module path, and the file path.
//   - IsEqual(M): performs a deep content comparison with another Module.
//   - Merge(any): folds the data of another compatible Module into this
//     instance, preserving data that is already present in the receiver.
//
// # Lifecycle status
//
// The status is a bit field with eight independent bits that track the
// completion of processing phases:
//
//   Bit 0 - Parsed:  the AST parsing phase has completed.
//   Bit 1 - Vendor:  the AST vendor resolution phase has completed.
//   Bit 2 - Analyzed: the summary analysis phase has completed.
//   Bit 3 - Audited1: the first audit report has been generated.
//   Bit 4 - Audited2: the second audit report has been generated.
//   Bit 5 - Audited3: the third audit report has been generated.
//   Bit 6 - Audited4: the fourth audit report has been generated.
//   Bit 7 - Outdated: the module has been flagged as out of date.
//
// Each bit can be set with a SetStatus{Phase}() method, cleared with an
// UnSetStatus{Phase}() method, or checked with an IsStatus{Phase}() method.
// The SetStatusNone() method clears all bits at once. The generic SetStatus()
// method replaces the entire bit field; the dedicated phase methods are
// preferred because they update a single phase while leaving the others
// untouched.
//
// # Package path resolution
//
// GetPkg(path string) resolves a file path against the module and returns the
// package directory, the base file name, and a validity flag. The path is
// first normalized by CleanPath; when the resulting path does not start with
// the module's path, the path is considered outside the module and a zero
// result is reported. Otherwise, the module-prefix is stripped and the
// remainder is returned as the package directory.
//
// # Serialization
//
// Modules support full serialization and deserialization in four formats:
// JSON, YAML, TOML, and CBOR. The Module interface embeds the corresponding
// marshaler and unmarshaler contracts from the standard library and the
// fxamacker/cbor, pelletier/go-toml and gopkg.in/yaml.v3 packages.
//
// All serialization methods share a common Export struct that holds the
// complete representation of the module data, ensuring consistency across
// formats. The Export struct contains the ID, module path, file path, vendor
// status, lifecycle status, summary, and the keyed map of audit reports.
//
// Marshal operations read the module fields directly; the vendor status is
// computed under the read lock. A concurrent modification of the module
// during marshaling is therefore a data race, and callers must serialize
// modifications with marshaling on their own.
//
// Unmarshal operations apply the deserialized data to the current module
// instance under a write lock, making them thread-safe.
//
// BinaryMarshaler and BinaryUnmarshaler are also implemented by delegating
// to the CBOR format, providing a compact binary representation.
//
// # Helpers
//
// The following helper is exported:
//
//   - CleanPath(path string) (string, string): normalizes a file path into
//     its directory and file name components, validating that the file has a
//     ".go" extension or is named "go.mod". It also strips leading "/" or "."
//     prefixes and trailing "/" suffixes from the directory path.
//
// # Thread safety
//
// All state mutations in a Module are guarded by an internal read-write mutex
// (sync.RWMutex). Write operations (SetMod, SetFile, SetStatus, SetSummary,
// SetReports, DelReports, and all SetStatus{Phase}/UnSetStatus{Phase} methods)
// acquire an exclusive lock. Read operations (GetMod, GetFile, GetPkg,
// IsVendor, GetStatus, GetSummary, GetReports, LstReports, and all
// IsStatus{Phase} methods) acquire a shared lock. Merge acquires an exclusive
// lock for the entire duration of the update. This makes Module instances
// safe for concurrent access.
//
// # Data flow
//
// The module data follows this processing pipeline:
//
//   Discovery          -> New(mod, file) creates the Module with status None
//           |
//   AST Parsing        -> Parsed bit set after the module is scanned
//           |
//   Vendor Resolution  -> Vendor bit set after vendor status is resolved
//           |
//   Summary Analysis   -> Analyzed bit set after a summary is computed
//           |
//   Audit              -> Audited1/2/3/4 bits set as each report is generated
//           |
//   Reporting          -> Reports are stored in the module's keyed map
//
// The status allows the auditor to resume an interrupted analysis: it can
// inspect each bit to determine which phase remains to be completed.
//
//   Data sources                              Module state
//   ---------------                           --------------
//   module path ----------------------> mod   (canonical path, cleaned on New)
//   file path ------------------------> fgm   (local file path, no validation)
//   mod ------------------------------>  id   (deterministic, derived by GenID)
//   processing -----------------------> sts   (8-bit status field)
//   analysis -------------------------> sum   (summary text)
//   audit  ---------------------------> rep   (keyed map of audrep.Report)

package mod
