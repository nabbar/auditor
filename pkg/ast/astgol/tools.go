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

// Package astgol provides Go-specific AST parsing capabilities for the auditor tool.
//
// This package implements specialized Abstract Syntax Tree (AST) parsing functionality
// specifically designed for Go programming language files. It handles the identification
// and parsing of Go module files (go.mod) and Go source code files (.go). The package
// provides a structured approach to analyzing Go language constructs through AST-based
// parsing techniques.
//
// Key features include:
//   - Language identification for Go module files (go.mod)
//   - AST creation for Go source code files
//   - Pattern matching for Go file extensions and naming conventions
//   - Integration with the general AST framework for consistent language handling
package astgol

import (
	"go/ast"
	"path/filepath"

	audids "github.com/nabbar/auditor/pkg/data/id"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
)

// getFileName returns the filename of the go.mod file.
//
// This method retrieves the filename from the internal state, ensuring thread-safe access through the read mutex.
// It provides a consistent way to access the name of the Go module file being processed by this AST parser instance.
//
// Returns:
//   - string: The name of the go.mod file being processed.
func (o *mdl) getFileName() string {
	o.x.RLock()
	defer o.x.RUnlock()
	return o.f
}

// getDirWork returns the working directory path where Go module files are located.
//
// This method retrieves the directory from the internal state, ensuring thread-safe access through the read mutex.
// It provides a consistent way to access the directory path for the Go module files being processed by this AST parser instance.
//
// Returns:
//   - string: The directory path where the go.mod file is located.
func (o *mdl) getDirWork() string {
	o.x.RLock()
	defer o.x.RUnlock()
	return o.d
}

// getFilePath returns the full path to the go.mod file.
//
// This method constructs and returns the complete file path by joining the directory path with the filename,
// ensuring thread-safe access through the read mutex. It provides a consistent way to retrieve the absolute path
// of the Go module file being processed.
//
// Returns:
//   - string: The full path to the go.mod file.
func (o *mdl) getFilePath() string {
	return filepath.Join(o.getDirWork(), o.getFileName())
}

// setMod sets the module function for retrieving module information.
//
// This method stores a function that retrieves module information from the manager,
// allowing for delayed access to module data. It ensures thread-safe access to the module
// retrieval function and maintains consistency in accessing module information throughout
// the AST parsing lifecycle.
//
// Parameters:
//   - id: The identifier of the module to retrieve from the manager.
func (o *mdl) setMod(id audids.ID) {
	o.x.Lock()
	defer o.x.Unlock()

	o.m = func() audmod.Module {
		return o.a.ModGet(id)
	}
}

// getMod retrieves the current module information.
//
// This method accesses the stored module function to retrieve the module information,
// ensuring thread-safe access through the read mutex. It provides a consistent way to
// access the parsed module data after initialization, allowing for proper module state management.
//
// Returns:
//   - audmod.Module: The module object if available, or nil if no module has been initialized.
func (o *mdl) getMod() audmod.Module {
	o.x.RLock()
	defer o.x.RUnlock()

	if o.m != nil {
		return o.m()
	}

	return nil
}

// getModId retrieves the identifier of the current module.
//
// This method accesses the module information and returns its unique identifier,
// ensuring thread-safe access through the read mutex. It provides a consistent way
// to obtain the module's ID for use in package and entry management operations.
//
// Returns:
//   - audids.ID: The identifier of the module, or 0 if no module has been initialized.
func (o *mdl) getModId() audids.ID {
	if m := o.getMod(); m != nil {
		return m.GetID()
	}

	return 0
}

// addAST stores a parsed AST file in the per-module cache indexed by source path.
//
// This method persists the Abstract Syntax Tree of a Go source file together with
// its package identifier so that subsequent lookups can retrieve the file without
// re-parsing. It provides thread-safe write access to the internal map through the
// write mutex and lazily initializes the map on the first insertion to avoid
// unnecessary allocation.
//
// Parameters:
//   - path: The relative or absolute path of the Go source file, used as the cache key.
//   - pkg:  The audids.ID that identifies the Go package to which the source file belongs.
//   - afs:  The parsed AST File produced by go/parser for the given source file.
func (o *mdl) addAST(path string, pkg audids.ID, afs *ast.File) {
	o.x.Lock()
	defer o.x.Unlock()

	if o.s == nil {
		o.s = make(map[string]astCache)
	}

	o.s[path] = astCache{
		p: pkg,
		a: afs,
	}
}

// listAST returns a shallow copy of all cached AST files indexed by source path.
//
// This method creates a defensive copy of the entire AST cache map so that callers
// receive a snapshot that is not affected by concurrent writes to the original map.
// The returned map is suitable for iteration over all parsed source files without
// holding a lock on the module instance. Thread-safe read access is guaranteed
// through the write mutex used exclusively during the copy operation.
//
// Returns:
//   - map[string]astCache: A new map with the same keys and astCache values as the
//     internal cache, where each key is a source-file path and each value holds the
//     package ID and the parsed AST File for that file.
func (o *mdl) listAST() map[string]astCache {
	o.x.Lock()
	defer o.x.Unlock()

	var r = make(map[string]astCache)

	for k, v := range o.s {
		r[k] = v
	}

	return r
}
