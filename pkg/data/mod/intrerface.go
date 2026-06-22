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
	"encoding/json"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fxamacker/cbor/v2"
	auddat "github.com/nabbar/auditor/pkg/data"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audsts "github.com/nabbar/auditor/pkg/data/status"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// FuncMod is a constructor function type that returns a new Module instance.
//
// It is used in factory patterns to decouple callers from the concrete Module
// implementation. A FuncMod value is invoked with no arguments and yields a
// fully initialized Module.
//
// Example:
//
//	var factory mod.FuncMod = func() mod.Module {
//	    return mod.New("example.com/repo", "/path/to/file.go")
//	}
//	m := factory()
type FuncMod func() Module

// Module is the interface describing a Go module tracked by the auditor.
//
// A Module embeds the generic auddat.Data[Module] interface, which provides
// identity, status, checksum, and representation management, and adds
// module-specific accessors and mutators. It also embeds the standard
// marshaler/unmarshaler interfaces for JSON, YAML, TOML, and CBOR so that
// Module values can be serialized in any of those formats.
//
// Implementations must be safe for concurrent use: all state reads and writes
// are protected by an internal read-write mutex.
type Module interface {
	json.Marshaler
	json.Unmarshaler
	yaml.Marshaler
	yaml.Unmarshaler
	toml.Marshaler
	toml.Unmarshaler
	cbor.Marshaler
	cbor.Unmarshaler

	// Data [Module] provides the common data-management contract:
	// unique identifier, lifecycle status, checksum, and representation
	// accessors shared by all auditor data types.
	auddat.Data[Module]

	// GetMod returns the canonical module path.
	//
	// The returned string is the cleaned module path (for example
	// "github.com/user/repo"). It is normalized: leading and trailing
	// separators and redundant dot components have been removed.
	GetMod() string

	// GetPkg resolves a relative path against this module and returns the
	// package name, the owning module path, and a validity flag.
	//
	// Parameters:
	//   - relPath: a path relative to the module root (for example
	//     "subdir/file.go").
	//
	// Returns:
	//   - pkg:   the package name derived from relPath (for example "subdir").
	//   - mod:   the module path that owns the package. This may differ from
	//           the receiver's own module path when relPath crosses a module boundary.
	//   - valid: true if relPath could be resolved to a known package within
	//           the module hierarchy; false otherwise.
	GetPkg(string) (pkg string, mod string, valid bool)

	// GetFile returns the full path of the Go source file associated with
	// this module.
	//
	// The path identifies the specific .go file (or go.mod) that this Module
	// value was created from. An empty string indicates that no file is
	// associated (see Empty).
	GetFile() string

	// IsVendor reports whether this module is a vendored dependency.
	//
	// A module is considered vendored when its associated file path lies
	// under a vendor directory. Vendored modules are third-party
	// dependencies managed by Go's vendor mechanism and may be treated
	// differently from direct dependencies during analysis.
	IsVendor() bool

	// SetMod updates the module path.
	//
	// The supplied path is normalized via CleanPath before being stored.
	// The operation is thread-safe.
	//
	// Parameters:
	//   - mod: the new module path to store.
	SetMod(string)

	// SetFile updates the associated file path and recomputes the vendor
	// flag.
	//
	// The vendor flag is set to true when the file path is empty or when it
	// points inside a vendor directory. The operation is thread-safe.
	//
	// Parameters:
	//   - file: the new file path to store.
	SetFile(string)
}

// New constructs a Module from a module path and an associated file path.
//
// The module path is normalized with CleanPath. If the resulting path is
// empty or CleanPath reports failure, New returns nil.
//
// The returned Module is initialized with:
//   - id:  a unique identifier derived from the cleaned module path via audids.GenID.
//   - mod: the cleaned module path.
//   - fgm: the file path as supplied.
//   - vdr: true if the file path is empty (no file associated), false otherwise.
//   - sts: audsts.Pending, indicating the module has not yet been processed.
//   - sum: an empty checksum.
//   - rep: a nil representation.
//
// Parameters:
//   - mod:  the module path (directory structure).
//   - file: the file path associated with this module.
//
// Returns:
//   - A new Module instance, or nil if the module path is invalid or empty.
func New(mod, file string) Module {
	var ok bool

	if mod, _, ok = CleanPath(mod); !ok || len(mod) < 1 {
		return nil
	}

	return &mdl{
		mx:  sync.RWMutex{},
		id:  audids.GenID(mod),
		mod: mod,
		fgm: file,
		vdr: len(file) < 1,
		sts: audsts.Pending,
		sum: "",
		rep: nil,
	}
}

// CleanPath normalizes a file path into its directory and filename
// components.
//
// The function performs the following steps:
//  1. Extracts the base filename using filepath.Base.
//  2. Validates the filename: it must have a ".go" extension or be exactly
//     "go.mod". If neither condition holds, the returned file is "".
//  3. Strips redundant components from the directory path: leading "/" or
//     "." prefixes and trailing "/" suffixes are removed iteratively until
//     none remain.
//
// Parameters:
//   - path: the raw file path to normalize.
//
// Returns:
//   - dir:   the cleaned directory path.
//   - file:  the validated base filename, or "" if the extension is not
//     ".go" and the name is not "go.mod".
//   - valid: always true; the function does not fail under normal operation.
func CleanPath(path string) (string, string, bool) {
	var file = filepath.Base(path)

	if filepath.Ext(file) != ".go" && file != "go.mod" {
		file = ""
	}

	for {
		if strings.HasPrefix(path, "/") {
			path = path[1:]
			continue
		}

		if strings.HasPrefix(path, ".") {
			path = path[1:]
			continue
		}

		if strings.HasSuffix(path, "/") {
			path = path[:len(path)-1]
			continue
		}

		break
	}

	return path, file, true
}

// Empty returns a zero-value Module with no associated module path or file.
//
// The returned Module is initialized with:
//   - id:  0 (no identifier assigned).
//   - mod: "" (empty module path).
//   - fgm: "" (no file associated).
//   - vdr: true (treated as vendored by default).
//   - sts: audsts.None, indicating the module is in an undefined or
//     unprocessed state.
//   - sum: "" (empty checksum).
//   - rep: nil (no representation).
//
// This is useful as a placeholder or sentinel value when no concrete module
// information is available.
func Empty() Module {
	return &mdl{
		mx:  sync.RWMutex{},
		id:  0,
		mod: "",
		fgm: "",
		vdr: true,
		sts: audsts.None,
		sum: "",
		rep: nil,
	}
}
