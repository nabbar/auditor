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

// Package fileutils provides utilities for scanning and parsing Go source files
// to extract function information and code content.
package fileutils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/nabbar/auditor/models"
)

// ScanProject recursively scans the current directory and extracts all functions found.
// This function replaces the catalog.sh script functionality by traversing the entire
// project structure to identify Go source files and extract their function declarations.
//
// The scanning process excludes vendor directories, test files, and example files
// as per the exclusion filters that mirror those used in the original catalog.sh script.
//
// Returns a slice of CatalogEntry representing all discovered functions with their package names,
// file paths, and function names. If any error occurs during the scanning process,
// it returns the error to allow for proper handling by the caller.
func ScanProject() ([]models.CatalogEntry, error) {
	var entries []models.CatalogEntry

	// Get the absolute path of the current working directory (equivalent to $PWD)
	// This ensures we have a consistent base path for our recursive scanning
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// filepath.Walk recursively traverses the directory structure starting from
	// the current working directory, applying the provided function to each file
	// and directory encountered during traversal
	err = filepath.Walk(workingDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process files that end with .go extension
		// Directories are skipped as we're only interested in source code files
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		// --- EXCLUSION FILTERS (Identical to catalog.sh) ---
		// These filters ensure we exclude vendor directories, test files,
		// and example-related files from our function extraction process
		pathLower := strings.ToLower(path)
		if strings.Contains(pathLower, filepath.FromSlash("/vendor/")) ||
			strings.HasSuffix(pathLower, "_test.go") ||
			strings.Contains(pathLower, "exampl") ||
			strings.Contains(pathLower, "exempl") {
			return nil
		}

		// Extract package and function names from the file
		// This step processes each Go file to identify all functions and their metadata
		fileEntries, err := parseGoFile(path)
		if err == nil && len(fileEntries) > 0 {
			entries = append(entries, fileEntries...)
		}

		return nil
	})

	return entries, err
}

// parseGoFile reads a .go file to extract the package name and function names.
//
// This function processes each line of a Go source file to identify:
// 1. The package declaration (only extracted once per file)
// 2. All function and method declarations
//
// Returns a slice of CatalogEntry containing all functions found in the specified file,
// along with their package information and file path. If there are no functions or
// errors occur during parsing, it returns an empty slice or appropriate error.
func parseGoFile(filePath string) ([]models.CatalogEntry, error) {
	frt, err := os.OpenRoot(filepath.Dir(filePath))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = frt.Close()
	}()

	file, err := frt.Open(filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var (
		entries []models.CatalogEntry

		scanner     = bufio.NewScanner(file)
		packageName = ""
	)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// 1. Extract package name (only on the first encounter)
		// This logic ensures we capture the package declaration only once per file
		// to avoid multiple entries for the same package
		if packageName == "" && strings.HasPrefix(trimmed, "package ") {
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 {
				packageName = fields[1]
			}
			continue
		}

		// 2. Extract functions/methods
		// This identifies all function declarations in the file format:
		// - Method: "func (r *Receiver) MyMethod(ctx context.Context) error {"
		// - Function: "func MyFunction() {"
		if strings.HasPrefix(trimmed, "func ") {
			funcName := extractFuncNameFromLine(trimmed)
			if funcName != "" && packageName != "" {
				entries = append(entries, models.CatalogEntry{
					Package:  packageName,
					FuncName: funcName,
					FilePath: filePath,
				})
			}
		}
	}

	return entries, scanner.Err()
}

// extractFuncNameFromLine cleans the line to isolate only the function name.
//
// This utility function processes function declaration lines to extract just
// the function or method name by parsing the various formats:
// - Method format: "func (r *Receiver) MyMethod(ctx context.Context) error {"
// - Function format: "func MyFunction() {"
//
// Returns the extracted function name or an empty string if parsing fails.
// This ensures that only valid function names are returned for cataloging purposes.
func extractFuncNameFromLine(line string) string {
	// Step A: Remove "func "
	// This strips the keyword "func" from the line to isolate the function signature
	line = strings.TrimPrefix(line, "func ")
	line = strings.TrimSpace(line)

	// Step B: If line starts with '(', it's a method with receiver -> skip the receiver
	// For methods, we need to remove the receiver part (e.g., "(r *Receiver)")
	// before extracting the actual function name
	if strings.HasPrefix(line, "(") {
		endReceiver := strings.Index(line, ")")
		if endReceiver == -1 {
			return ""
		}
		line = line[endReceiver+1:]
		line = strings.TrimSpace(line)
	}

	// Step C: Function name ends when encountering the argument opening parenthesis '('
	// We identify where the function name ends by finding the first opening parenthesis
	// which marks the start of the function parameters
	endFuncName := strings.Index(line, "(")
	if endFuncName == -1 {
		return ""
	}

	// Isolate the name and clean remaining whitespace
	// Extract just the function name part before the parentheses
	name := line[:endFuncName]
	return strings.TrimSpace(name)
}

// ExtractFunctionCode extracts the source code of a function by counting braces.
//
// This implementation identifies the function start and counts opening and closing braces
// to determine when the function body ends. It returns the complete source code block
// for the specified function name.
//
// The algorithm works by:
// 1. Finding the line that starts with "func " and contains the target function name
// 2. Tracking brace counts to identify when the function body is complete
// 3. Collecting all lines from the function declaration through its closing brace
//
// Returns the source code of the function or an error if it cannot be read.
// This allows for retrieving the complete implementation of a specific function.
func ExtractFunctionCode(filePath, funcName string) (string, error) {
	frt, err := os.OpenRoot(filepath.Dir(filePath))
	if err != nil {
		return "", err
	}
	defer func() {
		_ = frt.Close()
	}()

	file, err := frt.Open(filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	defer func() {
		_ = file.Close()
	}()

	var (
		buffer strings.Builder

		scanner    = bufio.NewScanner(file)
		found      = false
		openBraces = 0
		started    = false
	)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check if we've found the target function declaration
		// This condition identifies the line that starts with "func " and contains our desired function name
		if !found && strings.HasPrefix(trimmed, "func ") && strings.Contains(line, funcName) {
			found = true
		}

		if found {
			// Add the current line to our buffer regardless of whether it's part of the function
			buffer.WriteString(line)
			buffer.WriteString("\n")

			// Count opening and closing braces in the current line
			opened := strings.Count(line, "{")
			closed := strings.Count(line, "}")

			// Mark that we've started processing braces if there are opening braces
			if opened > 0 {
				started = true
			}

			// Update the brace counter by adding new open braces and subtracting closed braces
			openBraces += (opened - closed)

			// If we've started processing and all braces have been closed, we're done
			if started && openBraces <= 0 {
				break
			}
		}
	}
	return buffer.String(), scanner.Err()
}
