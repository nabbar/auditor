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

// Package mod provides the model of the Go modules tracked by the auditor.
//
// A Module represents a Go module: it holds the canonical path of the module,
// the path of the file it has been discovered in, its lifecycle status, its
// summary and its audit reports.
//
// The Module interface embeds the generic auddat.Data[Module] contract, which
// provides the identity, the status, the summary and the report management
// shared by every model of the auditor data layer, and it adds the accessors
// and the mutators specific to a Go module.
//
// # Helpers
//
// The package exposes the following helpers:
//   - New: builds a module from a module path and a file path.
//   - Empty: builds an uninitialized module, used as a placeholder.
//   - CleanPath: normalizes a path into its directory and its file name
//     components.
//
// # Serialization
//
// The module values support the json, yaml, toml and cbor formats.
//
// # Thread safety
//
// All the state mutations are guarded by a read-write mutex, so that the
// module values are safe for concurrent use.
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
	// The returned string is the path of the module, for example
	// "github.com/user/repo". It is the path that has been supplied to the
	// constructor New, where it has been normalized by CleanPath, or the path
	// that has been supplied to SetMod, which stores it as it is.
	//
	// Returns:
	//   - string: The path of the module.
	GetMod() string

	// GetPkg resolves a path against this module and returns the directory of
	// the package, the name of the file and a validity flag.
	//
	// The path is first normalized by CleanPath. When the resulting directory
	// does not start with the path of the module, the path is considered
	// outside of the module and the function reports an invalid result.
	// Otherwise, the prefix of the module is stripped and the remainder is
	// returned as the directory of the package.
	//
	// Parameters:
	//   - path: a path to resolve against the module, for example
	//     "subdir/file.go".
	//
	// Returns:
	//   - pkg: the directory of the package derived from path, for example
	//     "subdir", or an empty string when the path is invalid.
	//   - mod: the base name of the file derived from path, for example
	//     "file.go", or an empty string when the file is not a Go source file
	//     and is not the go.mod file.
	//   - valid: true when the path has been resolved inside the module, false
	//     otherwise.
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
	// A module is considered vendored when it has no associated file path, or
	// when its file path contains a "/vendor/" element. Vendored modules are
	// the third-party dependencies managed by the vendor mechanism of Go, and
	// they may be treated differently from the direct dependencies during the
	// analysis.
	//
	// Returns:
	//   - bool: true when the module is vendored, false otherwise.
	IsVendor() bool

	// SetMod updates the module path.
	//
	// The supplied path is stored as it is: it is not normalized by CleanPath.
	// The caller is expected to provide an already normalized path. The
	// operation is thread-safe.
	//
	// Parameters:
	//   - mod: the new module path to store.
	SetMod(string)

	// SetFile updates the path of the file associated with this module.
	//
	// The supplied path is stored as it is, and it is not validated by
	// CleanPath. The vendor status of the module is not stored either: it is
	// computed by IsVendor from the path of the file. The operation is
	// thread-safe.
	//
	// Parameters:
	//   - file: the new file path to store.
	SetFile(string)
}

// New constructs a Module from a module path and an associated file path.
//
// The module path is normalized by CleanPath. When the resulting path is
// empty, New returns nil.
//
// The returned Module is initialized with:
//   - id:  a unique identifier derived from the cleaned module path via
//     audids.GenID.
//   - mod: the cleaned module path.
//   - fgm: the file path as it has been supplied, without any validation.
//   - sts: audsts.None, indicating that the module has not been processed
//     yet.
//   - sum: an empty checksum.
//   - rep: a nil map of reports.
//
// The vendor status of the module is not stored: it is computed by IsVendor
// from the path of the file.
//
// Parameters:
//   - mod:  the module path (directory structure).
//   - file: the file path associated with this module.
//
// Returns:
//   - Module: A new Module instance, or nil when the module path is invalid
//     or empty.
func New(mod, file string) Module {
	if mod, _ = CleanPath(mod); len(mod) < 1 {
		return nil
	}

	return &mdl{
		mx:  sync.RWMutex{},
		id:  audids.GenID(mod),
		mod: mod,
		fgm: file,
		sts: audsts.None,
		sum: "",
		rep: nil,
	}
}

// CleanPath normalizes a file path into its directory and its file name
// components.
//
// The function performs the following steps:
//  1. Extracts the base file name using filepath.Base.
//  2. Validates the file name: it must have a ".go" extension, or be exactly
//     "go.mod". If neither condition holds, the returned file name is "".
//  3. Strips the redundant components from the directory path: the leading
//     "/" or "." prefixes and the trailing "/" suffixes are removed
//     iteratively, until none of them remains.
//
// Parameters:
//   - path: the raw file path to normalize.
//
// Returns:
//   - string: the cleaned directory path.
//   - string: the validated base file name, or "" when the extension is not
//     ".go" and the name is not "go.mod".
func CleanPath(path string) (string, string) {
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

	return path, file
}

// Empty returns a zero-value Module with no associated module path or file.
//
// The returned Module is initialized with:
//   - id:  0, no identifier assigned.
//   - mod: "", an empty module path.
//   - fgm: "", no file associated.
//   - sts: audsts.None, indicating that the module is in an undefined or
//     unprocessed state.
//   - sum: "", an empty checksum.
//   - rep: nil, no report.
//
// This is useful as a placeholder or as a sentinel value when no concrete
// module information is available.
func Empty() Module {
	return &mdl{
		mx:  sync.RWMutex{},
		id:  0,
		mod: "",
		fgm: "",
		sts: audsts.None,
		sum: "",
		rep: nil,
	}
}
