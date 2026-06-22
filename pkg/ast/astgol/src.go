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
	"context"
	"errors"
	"go/ast"
	"os"
	"path/filepath"
)

// getFileSrc reads the source code of a file from disk.
//
// This function serves as the primary mechanism for retrieving raw source code content
// from the filesystem for Go language processing within the auditor tool. It provides
// secure and efficient access to file contents while implementing proper error handling
// and resource management.
//
// The function implements robust error handling to ensure that file operations are performed safely,
// including checking for file existence and proper resource cleanup. It handles both absolute
// and relative path resolution appropriately, ensuring consistent behavior regardless of
// how the file paths are specified.
//
// Key features:
//   - Uses os.OpenRoot for secure file access operations with proper directory handling
//   - Handles absolute vs relative path resolution appropriately through filepath.Join
//   - Implements deferred cleanup of resources to prevent memory leaks and ensure proper
//     resource management
//   - Returns raw byte content of the file for further processing in AST parsing operations
//
// Parameters:
//   - path: The filename or relative path to the source file within the module directory,
//     which may be absolute or relative to the module directory.
//
// Returns:
//   - []byte: Raw byte content of the file, suitable for AST parsing operations and
//     subsequent analysis.
//   - error: Any error encountered during file reading operations, including file not found errors,
//     permission errors, or I/O errors that prevent successful file access.
func (o *mdl) getFileSrc(path string) ([]byte, error) {
	var (
		e error
		r *os.Root
	)

	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()

	// Convert relative path to absolute path for consistent handling
	if !filepath.IsAbs(path) {
		path, _ = filepath.Abs(path)
	}

	// Validate that the file exists before attempting to read it
	if _, e = os.Stat(path); e != nil {
		return nil, e
	}

	// Open the directory containing the file for secure access
	if r, e = os.OpenRoot(filepath.Dir(path)); e != nil {
		return nil, e
	}

	// Read the specific file from the opened directory
	return r.ReadFile(filepath.Base(path))
}

// getSrc retrieves source code content for a specific AST node.
//
// This function serves as the core mechanism for extracting precise source code segments
// from Go files based on AST node positions within the file. It provides efficient access
// to source code content by implementing intelligent caching strategies to minimize disk I/O
// operations during repeated accesses to the same files.
//
// The implementation employs a caching strategy to optimize performance by avoiding redundant
// file I/O operations when the same file is accessed multiple times during parsing. This approach
// significantly improves performance for complex AST processing scenarios where multiple nodes
// from the same file are analyzed.
//
// Key features:
//   - Implements intelligent caching of file content to reduce disk access and improve performance
//   - Calculates precise byte ranges for AST nodes using position information from the AST structure
//   - Handles both cached and uncached content retrieval seamlessly through efficient lookup mechanisms
//   - Provides accurate source code extraction for specific AST elements with proper error handling
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - path: The relative path to the source file containing the AST node, used to locate the file.
//   - afs: The AST File node representing the complete file structure, providing context for position calculations.
//   - adc: The AST Node for which source code content is required, specifying the exact location within the file.
//
// Returns:
//   - []byte: Extracted source code content for the specific AST node, containing exactly the bytes
//     corresponding to the node's position in the file.
//   - error: Any error encountered during content retrieval or position calculation, including file access errors
//     or invalid position information that prevents proper extraction.
func (o *mdl) getSrc(ctx context.Context, path string, afs *ast.File, adc ast.Node) ([]byte, error) {
	var (
		err error
		cnt []byte
		stt int
		ste int
	)

	// Calculate the start and end positions for the AST node within the file
	stt, ste = o.getPosition(afs, adc)
	if stt <= 0 || ste <= 0 {
		return nil, errors.New("cannot retrieve content of file " + path)
	} else if stt > ste {
		return nil, errors.New("invalid position in file " + path)
	}

	// Attempt to retrieve cached file content first to avoid redundant disk I/O
	if cnt, err = o.getCacheFile(ctx, path); err != nil {
		return nil, err
	}

	// Validate that the cached content is sufficient for the requested range
	if len(cnt) > 0 && len(cnt) < ste {
		return nil, errors.New("out of range for file " + path)
	}

	// Return cached content if available and within bounds
	if len(cnt) > 0 {
		return cnt[stt:ste], nil
	}

	// Read the file from disk if not cached
	if cnt, err = o.getFileSrc(path); err != nil {
		return nil, err
	}

	o.x.Lock()
	defer o.x.Unlock()

	// Check for context cancellation before proceeding with cache update
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Initialize the cache map if it's empty
	if len(o.c) < 1 {
		o.c = make(map[string][]byte)
	}

	// Store the file content in the cache for future access
	o.c[path] = cnt

	// Validate that the retrieved content is sufficient for the requested range
	if len(cnt) < ste {
		return nil, errors.New("out of range for file " + path)
	}

	// Return the extracted byte range from the file content
	return cnt[stt:ste], nil
}

// getCacheFile retrieves cached file content for a given path.
//
// This function provides access to previously cached source code content to avoid redundant disk I/O operations.
// It implements thread-safe access to the cache through read locks and returns either cached content or nil
// if no cached content exists for the specified path. The function ensures that context cancellation is checked
// before accessing cached data, maintaining proper error handling throughout the process.
//
// Key features:
//   - Implements thread-safe access to the cache using read locks for concurrent access safety
//   - Checks for context cancellation to prevent processing when cancelled
//   - Returns cached content if available or nil if no cache entry exists
//   - Maintains consistency in cache access patterns and error handling
//
// Parameters:
//   - ctx: Context for managing cancellation and timeouts during processing.
//   - pth: The path to the file for which cached content is requested.
//
// Returns:
//   - []byte: Cached byte content if available, or nil if no cache entry exists.
//   - error: Any error encountered during context validation or cache access operations.
func (o *mdl) getCacheFile(ctx context.Context, pth string) ([]byte, error) {
	o.x.RLock()
	defer o.x.RUnlock()

	// Check for context cancellation before accessing cached data
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	// Retrieve cached content if it exists for the specified path
	if cnt, fnd := o.c[pth]; fnd {
		return cnt, nil
	}

	return nil, nil
}

// getPosition calculates the byte position of an AST node within a file.
//
// This function serves as the foundation for precise source code extraction by determining
// the exact byte range that corresponds to an AST node's location within its source file.
// It provides accurate positioning information that enables precise extraction of source code
// segments for specific AST elements.
//
// The implementation handles special cases where documentation comments exist, ensuring
// that the calculated positions include relevant documentation when present. This is particularly
// important for function and general declaration nodes where documentation may be associated
// with the node's definition.
//
// Key features:
//   - Calculates accurate start and end positions for AST nodes based on their position information
//   - Handles documentation comments by adjusting position boundaries when documentation exists
//   - Provides consistent positioning regardless of whether documentation is present or not
//   - Returns byte offsets relative to the file's starting position for precise extraction
//
// Parameters:
//   - afs: The AST File node representing the complete file structure, providing context for position calculations.
//   - adc: The AST Node for which position information is required, specifying the exact location within the file.
//
// Returns:
//   - int: Start byte position of the AST node within the file, relative to the file's beginning.
//   - int: End byte position of the AST node within the file, relative to the file's beginning.
func (o *mdl) getPosition(afs *ast.File, adc ast.Node) (int, int) {
	// Validate that both AST nodes are not nil
	if afs == nil || adc == nil {
		return -1, -1
	}

	var (
		pst = adc.Pos()
		pse = adc.End()
	)

	// Handle function declaration nodes to include documentation comments in positioning
	if add, ok := adc.(*ast.FuncDecl); ok {
		if add.Doc != nil {
			// Adjust start position if documentation exists and starts before the function
			if add.Doc.Pos() < adc.Pos() {
				pst = add.Doc.Pos()
			}
			// Adjust end position if documentation exists and ends after the function
			if add.Doc.End() > adc.End() {
				pse = add.Doc.End()
			}
		}
	}

	// Handle general declaration nodes to include documentation comments in positioning
	if add, ok := adc.(*ast.GenDecl); ok {
		if add.Doc != nil {
			// Adjust start position if documentation exists and starts before the declaration
			if add.Doc.Pos() < adc.Pos() {
				pst = add.Doc.Pos()
			}
			// Adjust end position if documentation exists and ends after the declaration
			if add.Doc.End() > adc.End() {
				pse = add.Doc.End()
			}
		}
	}

	// Return byte offsets relative to the file's starting position for precise extraction
	return int(pst) - int(afs.FileStart), int(pse) - int(afs.FileStart)
}
