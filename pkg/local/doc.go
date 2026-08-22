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

// Package pkg/local:
// ==================
//
// Package local provides functionality for managing local auditor data including prompts and reports.
//
// The local package implements a comprehensive system for handling local storage of auditor data,
// including prompt templates, report templates, and catalog management. It provides thread-safe
// access to these resources and supports various serialization formats for catalog persistence.
//
// The architecture includes local storage management with embedded file extraction capabilities,
// thread-safe data structures for concurrent access, flexible catalog serialization support
// (JSON, YAML, TOML, CBOR), and a template-based prompt and report management system.
//
// Key design elements include thread-safe access patterns using RWMutex, hierarchical template
// organization by language and type, flexible serialization formats for catalog persistence,
// and embedded file system support for default templates.
//
// The import path is "github.com/nabbar/auditor/pkg/local" with the package alias "audloc".
//
// Interface `type Manager interface`:
//   - Extends `io.Closer`:
//     The Manager interface extends io.Closer, which means it provides a Close() method
//     for proper resource cleanup. This is necessary in this package because it manages
//     catalog data that must be persisted to disk when the manager is closed. The Close()
//     method ensures that any unsaved changes to the catalog are written to the file system,
//     preventing data loss and allowing the manager to be used safely in defer statements.
//
//   - Function `Prompt(audtps.CodeType, audast.Lang) []byte`:
//     Retrieves a prompt template for a specific code type and language.
//     Parameters:
//       - audtps.CodeType: The code type identifier used to select the appropriate prompt template
//       - audast.Lang: The language identifier used to select the appropriate prompt template
//     Result:
//       - []byte: The prompt template content as bytes, or nil if no matching template is found
//
//   - Function `Report(string, audast.Lang) []byte`:
//     Retrieves a report template for a given name and language.
//     Parameters:
//       - string: The name of the report template to retrieve
//       - audast.Lang: The language identifier used to select the appropriate report template
//     Result:
//       - []byte: The report template content as bytes, or nil if no matching template is found
//
//   - Function `CatalogTo(io.Writer, string) error`:
//     Serializes and writes catalog data to the provided writer with the specified filename.
//     Parameters:
//       - io.Writer: The destination writer where serialized catalog data will be written
//       - string: The filename or extension used to determine the serialization format
//     Result:
//       - error: An error if serialization fails or an invalid extension is provided
//
//   - Function `CatalogFrom(io.Reader, string) (auddbm.Manager, error)`:
//     Deserializes catalog data from the provided reader with the specified filename
//     and returns a new manager instance.
//     Parameters:
//       - io.Reader: The source reader from which serialized catalog data will be read
//       - string: The filename or extension used to determine the deserialization format
//     Result:
//       - auddbm.Manager: A new database manager instance populated with loaded catalog data
//       - error: An error if deserialization fails or an invalid extension is provided
//
// Function `func New(p string, u auduxi.Manager, l func() auddbm.Manager, s func(auddbm.Manager)) (Manager, error)`:
//   Creates a new Manager instance with the specified parameters.
//   This function initializes the local storage directory, extracts embedded files if needed,
//   loads prompt and report templates, and handles catalog loading.
//
//   - Parameters:
//     - string: The path to the local storage directory. If empty, it defaults to ".auditor" in the current working directory
//     - auduxi.Manager: An instance of auduxi.Manager for user interface operations
//     - func() auddbm.Manager: A function that returns a database manager instance
//     - func(auddbm.Manager): A function that accepts a database manager instance for update
//
//   - Returns:
//     - Manager: The initialized manager instance
//     - error: An error if any step fails during initialization
//
//

package local
