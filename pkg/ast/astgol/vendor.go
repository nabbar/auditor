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
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"

	audids "github.com/nabbar/auditor/pkg/data/id"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
)

var sep = string(filepath.Separator) // sep holds the platform-specific file path separator character, derived from filepath.Separator, used to distinguish directory levels in file system paths on the current operating system; it evaluates to a forward slash (/) on Unix-like systems and a backslash (\") on Windows, enabling the code to correctly parse and manipulate file system paths in an OS-agnostic manner when resolving vendor directory prefixes, internal package detection, and package path construction during module and package discovery operations.

// initVendor prepares the vendor directory for the current Go module by
// sequentially executing three Go module commands: go mod tidy, go mod download
// all, and go mod vendor. These commands ensure that the module's dependencies
// are resolved, downloaded to the local module cache, and copied into a
// vendor/ directory under the module root, enabling offline and reproducible
// builds. The function uses exec.CommandContext to run each command in the
// module's working directory as returned by o.getDirWork(). Combined output
// from each command is captured and logged on failure. If any command exits
// with a non-zero status or the supplied context is canceled, the function
// returns immediately with the encountered error.
func (o *mdl) initVendor(ctx context.Context) error {
	var od = o.getDirWork()

	o.u.Info("[Path: %s] Preparing vendor folder", od)

	lst := [][]string{
		{"mod", "tidy"},
		{"mod", "download", "all"},
		{"mod", "vendor"},
	}

	for _, args := range lst {
		cmd := exec.CommandContext(ctx, "go", args...) // #nosec

		cmd.Dir = od // changing working directory to go.mod root folder

		// catching stdout to add information
		if out, err := cmd.CombinedOutput(); err != nil {
			o.u.ErrorStack(fmt.Sprintf("go %s failed: %s", strings.Join(args, " "), string(out)), err)
			return err
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}

	o.u.Info("[Path: %s] Successfully vendoring folder", od)
	return nil
}

// parseGoMod parses the contents of a go.mod file represented as a byte slice
// and extracts the module name declared by the module directive and the list of
// module dependencies declared in require blocks. The function processes the
// input line by line, skipping blank lines and comments. When it encounters a
// module directive, it extracts the module name from the second field. For
// require blocks spanning multiple lines, it collects the first field of each
// line as a dependency until the closing parenthesis is reached. Single-line
// require directives contribute their second field as a dependency. The parsed
// module name and the ordered list of dependency names are returned together.
func (o *mdl) parseGoMod(c []byte) (string, []string) {
	var (
		mod string
		req bool

		vdr = make([]string, 0)
	)

	for _, line := range bytes.Split(c, []byte("\n")) {
		line = bytes.TrimSpace(line)
		if len(line) == 0 || bytes.HasPrefix(line, []byte("//")) {
			continue
		}

		if bytes.HasPrefix(line, []byte("module ")) {
			fields := bytes.Fields(line)
			if len(fields) >= 2 {
				mod = string(fields[1])
			}
		} else if bytes.HasPrefix(line, []byte("require (")) {
			req = true
			continue
		} else if req && bytes.HasPrefix(line, []byte(")")) {
			req = false
			continue
		} else if req {
			fields := bytes.Fields(line)
			if len(fields) >= 1 {
				vdr = append(vdr, string(fields[0]))
			}
		} else if bytes.HasPrefix(line, []byte("require ")) {
			fields := bytes.Fields(line)
			if len(fields) >= 2 {
				vdr = append(vdr, string(fields[1]))
			}
		}
	}

	return mod, vdr
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
			if !strings.Contains(p, "vendor"+sep) {
				return nil
			}
			// Exclude all 'internal' folders or subfolders within the vendors directories
			// as they are strictly inaccessible outside of their own tier module.

			if strings.Contains(p, sep+"internal"+sep) || strings.HasSuffix(p, sep+"internal") {
				return nil
			}
		} else {
			// Exclude vendor directories when vendor is false
			if strings.Contains(p, "vendor"+sep) {
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
	o.u.InfoData("[getPkgList][Path: %s][Vendor %v] Packages list for module %s", ptn, pth, vendor, mod)

	return ptn, err
}

// findMod resolves which Go module a given package path belongs to by matching
// the path against the root module and any vendored or indirect modules known
// to the auditor index. When the path contains /vendor/, the vendored suffix
// is stripped to obtain the canonical import path. The cleaned path, produced by
// audmod.CleanPath, is then compared against the root module path: if it shares
// the root module prefix, the root module ID and path are returned immediately.
// Otherwise the function queries the auditor index by searching for an ID via
// o.a.SearchID; if a matching module exists in the index, its ID is returned;
// if only a matching package exists, the package's module ID is returned instead.
// As a last resort, findMod walks all known modules in the index looking for
// one whose declared module name is a prefix of the path, and finally falls
// back to scanning an alternative list mdp for a matching prefix. If no module
// is found, the zero ID and the cleaned path are returned.
func (o *mdl) findMod(pth string) (audids.ID, string) {
	if strings.Contains(pth, "/vendor/") {
		pth = pth[strings.Index(pth, "/vendor/")+8:]
	}

	pth, _ = audmod.CleanPath(pth)

	if strings.HasPrefix(pth, o.getMod().GetMod()) {
		return o.getModId(), pth
	}

	var (
		ids audids.ID
		mod audmod.Module
		pkg audpkg.Package
	)

	ids = o.a.SearchID(pth)

	if mod = o.a.ModGet(ids); mod != nil && !mod.IsEmpty() {
		return mod.GetID(), pth
	}

	if pkg = o.a.PkgGet(ids); pkg != nil && !pkg.IsEmpty() {
		return pkg.GetModId(), pth
	}

	ids = 0

	o.a.ModWalk(func(mod audmod.Module) bool {
		if mod == nil || mod.IsEmpty() {
			return true
		}
		if strings.HasPrefix(pth, mod.GetMod()) {
			ids = mod.GetID()
			return false
		}
		return true
	})

	if ids > 0 {
		return ids, pth
	}

	for _, mdp := range o.l {
		if strings.HasPrefix(pth, mdp) {
			if mod = o.a.ModGet(o.a.SearchID(mdp)); mod != nil || !mod.IsEmpty() {
				return mod.GetID(), pth
			}
		}
	}

	return 0, pth
}
