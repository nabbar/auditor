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
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"

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

// getModFct returns a function that retrieves the current module information.
//
// This method provides a functional interface to access module data, creating a closure
// that encapsulates the module retrieval logic. It ensures thread-safe access and maintains
// consistency in how module information is accessed throughout the AST parsing process.
//
// Returns:
//   - audmod.FuncMod: A function that retrieves module information, or nil if no module exists.
func (o *mdl) getModFct() audmod.FuncMod {
	if m := o.getMod(); m != nil {
		i := m.GetID()
		return func() audmod.Module {
			return o.a.ModGet(i)
		}
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

// getPkgList generates a list of package paths for Go source files in the specified directory.
//
// This function performs a recursive walk through the directory structure to identify all
// Go source files and constructs package path lists based on their directory locations.
// It filters out test files and vendor directories when appropriate, ensuring that only
// relevant packages are included in the analysis. The function is crucial for determining
// which packages need to be parsed during AST processing.
//
// Key features:
//   - Recursively walks through the directory structure using filepath.Walk
//   - Filters out directories and test files to avoid unnecessary processing
//   - Excludes vendor directories when vendor is false, or includes them when vendor is true
//   - Matches only .go files to ensure proper file filtering
//   - Constructs package paths by joining the module name with directory prefixes
//   - Sorts package paths for consistent ordering and predictable processing
//
// Parameters:
//   - mod: The module name used as a base for constructing package paths
//   - pth: The directory path to walk through for finding Go source files
//   - vendor: Boolean flag indicating whether to include or exclude vendor directories
//
// Returns:
//   - []string: A sorted slice of package paths that represent the Go packages to be analyzed
//   - error: Any error encountered during directory traversal or file matching operations
func (o *mdl) getPkgList(mod, pth string, vendor bool) ([]string, error) {
	var (
		err error
		ptn = make([]string, 0)
	)

	err = filepath.Walk(pth, func(p string, i os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		// Skip directories as they are not source files
		if i.IsDir() {
			return nil
		}

		// Skip test files to avoid processing them during normal analysis
		if strings.Contains(p, "test") {
			return nil
		}

		// Filter vendor directories based on the vendor flag
		if vendor {
			// Include only vendor directories when vendor is true
			if !strings.Contains(p, "vendor") {
				return nil
			}
		} else {
			// Exclude vendor directories when vendor is false
			if strings.Contains(p, "vendor") {
				return nil
			}
		}

		// Match only .go files to ensure we're processing source code
		if ok, er := filepath.Match("*.go", i.Name()); er != nil {
			return er
		} else if !ok {
			return nil
		}

		// Construct package path by joining module name with directory prefix
		if pk := path.Join(mod, strings.TrimPrefix(filepath.Dir(p), pth)); !slices.Contains(ptn, pk) {
			ptn = append(ptn, pk)
		}

		return nil
	})

	// Sort the package paths for consistent ordering and predictable processing
	slices.Sort(ptn)

	return ptn, err
}
