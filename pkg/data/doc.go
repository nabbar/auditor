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

// Package data defines the types and the contracts that support the data layer
// of the auditor.
//
// It provides the interfaces and the helper types used to represent the elements
// of an analysis -- the modules, the packages and the code entries -- along with
// the generic collection that stores them, the identifiers that key the
// collections, and the status flags that track the progress of the processing.
//
// The elements of the data layer share a common contract: they embed the
// auddat.Data[M] interface, which declares GetID, IsEmpty, IsEqual, Merge,
// GetParentId, and the management of the status, the summary and the reports.
// That contract is implemented by the concrete models of each sub-package (mod,
// pkg, entry, params, src) and is enforced by the collections (via the generic
// Collection[M] type) and by the manager (via the status mutator methods).
//
// # Package hierarchy
//
// The modules are at the root of the hierarchy. Each module contains packages,
// and each package contains entries (functions, methods, types, variables, ...).
// The parent reference from a child to its parent is always an audids.ID, never
// a full object: the parent is resolved on demand through the resolver function
// registered with RegisterGetModule, RegisterGetPackage, or RegisterGetEntry.
//
// # Serialization
//
// Every type that belongs to the data layer supports json, yaml, toml and cbor
// serialization and deserialization, so that the data can be persisted or
// exchanged in any of the supported formats.
//
// # Thread safety
//
// All the mutable types (mdl) protect their state with a sync.RWMutex, so that
// the instances are safe for concurrent use across multiple goroutines. The
// Source type is an exception: it does not use a mutex because its internal
// state (a string) is immutable once constructed.
//
// # Sub-packages
//
//   - collection: generic, thread-safe map-backed collection with serialization.
//   - entry: the Entry interface, the concrete model of a code element
//     (function, type, variable, ...), and the FuncEnt constructor.
//   - id: the deterministic identifier type (ID) and the GenID constructor.
//   - manager: the Manager interface, the Linker, and the concrete model that
//     holds the three collections (modules, packages, entries).
//   - mod: the Module interface, the concrete model of a Go module, FuncMod,
//     and CleanPath.
//   - params: the Params interface, the concrete model of a parameter set, and
//     FuncParams.
//   - pkg: the Package interface, the concrete model of a Go package, and
//     FuncPkg.
//   - reports: the Report struct that holds the findings of an audit.
//   - src: the Source interface and the concrete model that wraps source code
//     as a hex-encoded string.
//   - status: the Status bit-field type that tracks the processing phases.
//   - types: the CodeType enumeration that categorizes the code elements.

package data
