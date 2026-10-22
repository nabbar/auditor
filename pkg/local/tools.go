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

// Package local provides functionality for managing local auditor data including prompts and reports.
package local

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// readPath recursively reads files from the specified path and organizes them by language and file type.
// This function traverses the directory structure starting at the given path, reading all files
// and organizing them into a nested map structure based on their directory hierarchy.
//
// Parameters:
//   - p: The directory path to read files from. This is the base directory where the search begins.
//   - p1: The parent directory name for categorization. When processing subdirectories,
//     this parameter represents the language or category identifier.
//
// Returns:
//   - map[string]map[string][]byte: A nested map structure where the first key is the directory name
//     (language/category), the second key is the file type or name, and the value is the file content as bytes.
//   - error: An error if reading or processing fails during directory traversal or file operations.
func readPath(p, p1 string) (map[string]map[string][]byte, error) {
	var (
		e error
		l []os.DirEntry
		d *os.Root
		f []byte

		r = make(map[string]map[string][]byte)
	)

	defer func() {
		if d != nil {
			_ = d.Close()
		}
	}()

	// Read the directory entries at the specified path
	if l, e = os.ReadDir(p); e != nil {
		return nil, e
	}

	// Iterate through all directory entries
	for i := 0; i < len(l); i++ {
		// Handle subdirectories
		if l[i].IsDir() {
			// Skip if we're already processing a subdirectory (p1 is not empty)
			if len(p1) > 0 {
				continue
			}

			// Recursively process the subdirectory
			var s map[string]map[string][]byte
			s, e = readPath(filepath.Join(p, l[i].Name()), l[i].Name())
			if e != nil {
				return nil, e
			}

			// Merge results from recursive call into main result map
			for k1, m1 := range s {
				for k2, v := range m1 {
					if r[k1] == nil {
						r[k1] = make(map[string][]byte)
					}
					r[k1][k2] = v
				}
			}

			continue
		}

		// Skip if no parent directory name is specified (p1 is empty)
		if len(p1) < 1 {
			continue
		}

		// Open root filesystem handle if not already opened
		if d == nil {
			if d, e = os.OpenRoot(p); e != nil {
				return nil, e
			}
		}

		// Read the file content
		if f, e = d.ReadFile(l[i].Name()); e != nil {
			return nil, e
		}

		// Initialize map for this directory if not already done
		if r[p1] == nil {
			r[p1] = make(map[string][]byte)
		}

		// Store file content in the result map
		r[p1][l[i].Name()] = f
	}

	return r, nil
}

// checkExtract checks if a directory exists in the root filesystem and extracts embedded files if needed.
// This function determines whether the specified directory already exists in the filesystem.
// If it does not exist, it triggers the extraction of embedded files to populate the directory.
//
// Parameters:
//   - r: The root filesystem handle used for file system operations. This provides access to
//     the local storage directory where files are to be extracted.
//   - dir: The directory path to check and extract. This specifies which embedded directory
//     should be checked and potentially extracted to the local filesystem.
//
// Returns:
//   - error: An error if checking or extraction fails. This includes cases where the directory
//     cannot be accessed, or where file extraction operations fail.
func checkExtract(r *os.Root, dir string) error {
	// Check if the directory exists in the filesystem
	if _, e := r.Stat(dir); e != nil {
		// Directory doesn't exist, so extract embedded files
		return extractPath(r, dir)
	}

	// Directory already exists, no extraction needed
	return nil
}

// extractPath extracts embedded files to the specified directory in the root filesystem.
// This function copies embedded files from the embedded filesystem (efs) to the local filesystem
// using the provided root filesystem handle. It recursively processes all embedded files and
// directories, creating corresponding structures in the local filesystem.
//
// Parameters:
//   - r: The root filesystem handle used for file system operations. This provides access to
//     the local storage directory where embedded files will be extracted.
//   - dir: The directory path where embedded files should be extracted. This specifies the
//     target location in the local filesystem for extraction.
//
// Returns:
//   - error: An error if extraction fails during file operations or directory creation.
//     This includes cases where files cannot be read from the embedded filesystem,
//     or where local file operations fail.
func extractPath(r *os.Root, dir string) error {
	// Walk through all embedded files in the specified directory
	return fs.WalkDir(efs, path.Join(baseEmbed, dir), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		var (
			rer error
			cnt []byte
			lfs *os.File

			pfs = strings.TrimPrefix(p, baseEmbed+"/")
		)

		defer func() {
			if lfs != nil {
				_ = lfs.Close()
			}
		}()

		// Handle directory creation
		if d.IsDir() {
			return r.Mkdir(pfs, 0700)
		}

		// Read file content from embedded filesystem
		if cnt, rer = efs.ReadFile(p); rer != nil {
			return rer
		}

		// Create and open file in local filesystem
		if lfs, rer = r.OpenFile(pfs, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600); rer != nil {
			return rer
		}

		// Write content to local file
		if _, rer = lfs.Write(cnt); rer != nil {
			return rer
		}

		return nil
	})
}
