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

// Package mod provides interfaces and utilities for managing Go module information.
//
// The package defines the Module interface, which abstracts a Go module by its
// canonical path, associated source file, and vendor status. It also exposes
// constructor and path-normalization helpers used to build Module values.
//
// Module values support serialization to JSON, YAML, TOML, and CBOR, and all
// state mutations are guarded by a read-write mutex for safe concurrent use.
package mod

import (
	"strings"
	"sync"

	audids "github.com/nabbar/auditor/pkg/data/id"
	audrep "github.com/nabbar/auditor/pkg/data/reports"
	audsts "github.com/nabbar/auditor/pkg/data/status"
)

// mdl is the concrete implementation of the Module interface.
//
// It stores the canonical module path, an associated file path (typically the
// go.mod file), and a vendor flag that indicates whether the module is a
// vendored dependency. All fields are protected by a read-write mutex to
// ensure thread-safe concurrent access.
type mdl struct {
	// mx guards all mutable state in this struct.
	// Write operations (SetMod, SetFile) acquire an exclusive lock; read
	// operations (GetMod, GetFile, IsVendor, GetPkg) acquire a shared lock.
	mx sync.RWMutex

	// id is the unique identifier for this module.
	// It is derived from the module path via audids.GenID and remains
	// constant for the lifetime of the module.
	id audids.ID

	// mod is the canonical Go module path (e.g. "github.com/user/repo").
	// It is used for import resolution and package management.
	mod string

	// fgm is the local file path to the module's dependencies catalog,
	// typically a go.mod file. An empty string means no file is associated.
	fgm string

	// vdr indicates whether this module is a vendored dependency.
	// When true, the module is located under a vendor directory and is
	// treated as a third-party dependency rather than local source code.
	vdr bool

	// sts tracks the lifecycle status of the module (e.g. Pending, Active).
	// It is managed through the embedded auddat.Data[Module] contract.
	sts audsts.Status

	// sum is a brief summary or description of the module.
	sum string

	// rep holds audit reports keyed by report name.
	// Values are audrep.Report instances that may contain security findings,
	// compliance status, or other audit results.
	rep map[string]audrep.Report
}

// SetMod updates the module path.
//
// The supplied path is stored as-is; callers are expected to have already
// normalized it (for example via CleanPath). The operation is thread-safe.
//
// Parameters:
//   - p: the new module path to store.
func (o *mdl) SetMod(p string) {
	o.mx.Lock()
	defer o.mx.Unlock()

	o.mod = p
}

// GetMod returns the canonical module path.
//
// The returned string is the cleaned module path (for example
// "github.com/user/repo"). It is normalized: leading and trailing
// separators and redundant dot components have been removed.
func (o *mdl) GetMod() string {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.mod
}

// GetPkg resolves a relative path against this module and returns the
// package name, the owning module path, and a validity flag.
//
// The function first normalizes the input path via CleanPath. If the
// resulting directory path does not start with the receiver's module path,
// the path is considered outside the module and the function returns false.
// Otherwise, the module prefix is stripped and the remainder is returned as
// the package name.
//
// Parameters:
//   - path: a path relative to the module root (for example
//     "subdir/file.go").
//
// Returns:
//   - pkg:   the package name derived from path (for example "subdir").
//   - mod:   the module path that owns the package. This may differ from
//     the receiver's own module path when path crosses a module boundary.
//   - valid: true if path could be resolved to a known package within
//     the module hierarchy; false otherwise.
func (o *mdl) GetPkg(path string) (string, string, bool) {
	var (
		k bool
		f string
	)

	if path, f, k = CleanPath(path); !k || len(path) < 1 {
		return "", "", false
	}

	if p := o.GetMod(); !strings.HasPrefix(path, p) {
		return "", "", false
	} else {
		path, _, _ = CleanPath(path[len(p):])
	}

	return path, f, true
}

// SetFile updates the associated file path and recomputes the vendor
// flag.
//
// If the supplied path is empty, fgm is set to "" and vdr is set to true
// (treated as vendored). If the path is non-empty, fgm is set to the path
// and vdr is set to false. The operation is thread-safe.
//
// Parameters:
//   - p: the new file path to store.
func (o *mdl) SetFile(p string) {
	o.mx.Lock()
	defer o.mx.Unlock()

	if len(p) < 1 {
		o.fgm = ""
		o.vdr = true
	} else {
		o.fgm = p
		o.vdr = false
	}
}

// GetFile returns the full path of the Go source file associated with
// this module.
//
// The path identifies the specific .go file (or go.mod) that this Module
// value was created from. An empty string indicates that no file is
// associated (see Empty).
func (o *mdl) GetFile() string {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.fgm
}

// IsVendor reports whether this module is a vendored dependency.
//
// A module is considered vendored when its associated file path lies
// under a vendor directory. Vendored modules are third-party
// dependencies managed by Go's vendor mechanism and may be treated
// differently from direct dependencies during analysis.
func (o *mdl) IsVendor() bool {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.vdr
}
