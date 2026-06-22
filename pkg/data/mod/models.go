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

// File models.go contains the concrete model of the Module interface.
//
// The mdl type stores the canonical module path, the path of the associated
// file, the status, the summary and the reports of a Go module. Every field is
// protected by a sync.RWMutex, which makes the model safe for concurrent
// use.

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
// It stores the canonical path of the module, the path of the associated
// file, the lifecycle status, the summary and the reports of a Go module.
//
// The vendor status of the module is not stored: it is computed by IsVendor
// from the path of the file.
//
// All the fields are protected by a read-write mutex to ensure a thread-safe
// concurrent access.
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
// The returned string is the path of the module, for example
// "github.com/user/repo". It is the path that has been supplied to the
// constructor New, where it has been normalized by CleanPath, or the path
// that has been supplied to SetMod, which stores it as it is.
//
// Returns:
//   - string: The path of the module.
func (o *mdl) GetMod() string {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return o.mod
}

// GetPkg resolves a path against this module and returns the directory of
// the package, the name of the file and a validity flag.
//
// The path is first normalized by CleanPath. When the resulting directory
// does not start with the path of the module, the path is considered outside
// of the module and the function reports an invalid result. Otherwise, the
// prefix of the module is stripped and the remainder is returned as the
// directory of the package.
//
// Parameters:
//   - path: a path to resolve against the module, for example
//     "subdir/file.go".
//
// Returns:
//   - string: the directory of the package derived from path, for example
//     "subdir", or an empty string when the path is invalid.
//   - string: the base name of the file derived from path, for example
//     "file.go", or an empty string when the file is not a Go source file and
//     is not the go.mod file.
//   - bool: true when the path has been resolved inside the module, false
//     otherwise.
func (o *mdl) GetPkg(path string) (string, string, bool) {
	var f string

	if path, f = CleanPath(path); len(path) < 1 {
		return "", "", false
	}

	if p := o.GetMod(); !strings.HasPrefix(path, p) {
		return "", "", false
	} else {
		path, _ = CleanPath(path[len(p):])
	}

	return path, f, true
}

// SetFile updates the path of the file associated with this module.
//
// The supplied path is stored as it is, and it is not validated by CleanPath.
// The vendor status of the module is not stored either: it is computed by
// IsVendor from the path of the file. The operation is thread-safe.
//
// Parameters:
//   - p: the new file path to store.
func (o *mdl) SetFile(p string) {
	o.mx.Lock()
	defer o.mx.Unlock()

	o.fgm = p
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
// A module is considered vendored when it has no associated file path, or
// when its file path contains a "/vendor/" element. Vendored modules are the
// third-party dependencies managed by the vendor mechanism of Go, and they
// may be treated differently from the direct dependencies during the analysis.
//
// Returns:
//   - bool: true when the module is vendored, false otherwise.
func (o *mdl) IsVendor() bool {
	o.mx.RLock()
	defer o.mx.RUnlock()

	return len(o.fgm) < 1 || strings.Contains(o.fgm, "/vendor/")
}
