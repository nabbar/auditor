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
	"embed"
	"io"
	"os"
	"path/filepath"
	"sync"

	audast "github.com/nabbar/auditor/pkg/ast"
	auddbm "github.com/nabbar/auditor/pkg/data/manager"
	audtps "github.com/nabbar/auditor/pkg/data/types"
	llmcfg "github.com/nabbar/auditor/pkg/llmconfig"
	auduxi "github.com/nabbar/auditor/pkg/uxi"
)

//go:embed embed
var efs embed.FS

const (
	// baseEmbed represents the base path for embedded files.
	baseEmbed = "embed"
	// promptPath defines the directory path where prompt templates are stored.
	promptPath = "analyse"
	// reportPath defines the directory path where report templates are stored.
	reportPath = "reports"
	// catalogFile specifies the filename of the catalog binary file.
	catalogFile   = "catalog.bin"
	llmConfigFile = "llmconfig.json"
)

// Manager interface defines the contract for managing local auditor data.
// It provides methods to access prompts and reports, as well as catalog operations.
type Manager interface {
	io.Closer

	// Prompt retrieves a prompt template for a specific code type and language.
	// It attempts to find a matching template for the given language and code type.
	// If not found, it tries to find a default template for the same code type in the unknown language.
	// If still not found, it attempts to retrieve a default template for the specified code type in the unknown language.
	//
	// Parameters:
	//   - code: The code type for which to retrieve the prompt.
	//   - lang: The language of the prompt template.
	//
	// Returns:
	//   - []byte: The prompt template as bytes, or nil if no matching template is found.
	Prompt(audtps.CodeType, audast.Lang) []byte

	// Report retrieves a report template for a given name and language.
	// It first attempts to find a matching template for the given language and name.
	// If not found, it tries to find a default template for the same name in the unknown language.
	//
	// Parameters:
	//   - name: The name of the report template.
	//   - lang: The language of the report template.
	//
	// Returns:
	//   - []byte: The report template as bytes, or nil if no matching template is found.
	Report(string, audast.Lang) []byte

	// CatalogTo writes the catalog data to the provided writer with the specified filename.
	// It serializes the catalog data using the appropriate format based on the file extension.
	//
	// Parameters:
	//   - wrt: The writer to which the catalog data will be written.
	//   - name: The filename used to determine the serialization format.
	//
	// Returns:
	//   - error: An error if serialization fails or an invalid extension is provided.
	CatalogTo(io.Writer, string) error

	// CatalogFrom reads catalog data from the provided reader with the specified filename
	// and returns a new manager instance.
	// It deserializes the catalog data using the appropriate format based on the file extension.
	//
	// Parameters:
	//   - rdr: The reader from which the catalog data will be read.
	//   - name: The filename used to determine the deserialization format.
	//
	// Returns:
	//   - auddbm.Manager: The loaded database manager instance.
	//   - error: An error if deserialization fails or an invalid extension is provided.
	CatalogFrom(io.Reader, string) (auddbm.Manager, *auddbm.Linker, error)

	LoadLLMConfig() (*llmcfg.Config, error)
	SaveLLMConfig(cfg *llmcfg.Config) error
}

// New creates a new Manager instance with the specified parameters.
// It initializes the local storage directory, extracts embedded files if needed,
// loads prompt and report templates, and handles catalog loading.
//
// Parameters:
//   - p: The path to the local storage directory. If empty, it defaults to ".auditor" in the current working directory.
//   - u: An instance of auduxi.Manager for user interface operations.
//   - l: A function that returns a database manager instance.
//   - s: A function that accepts a database manager instance for update.
//
// Returns:
//   - Manager: The initialized manager instance.
//   - error: An error if any step fails during initialization.
func New(p string, u auduxi.Manager, l func() auddbm.Manager, s func(auddbm.Manager)) (Manager, error) {
	var (
		e error
		r *os.Root
		b auddbm.Manager
		c map[string]map[string][]byte
	)

	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()

	if len(p) == 0 {
		if p, e = os.Getwd(); e != nil {
			return nil, e
		}
	}

	if filepath.Base(p) != ".auditor" {
		p = filepath.Join(p, ".auditor")
	}

	if p, e = filepath.Abs(p); e != nil {
		return nil, e
	}

	// Ensure the local storage directory exists
	if _, e = os.Stat(p); e != nil {
		if e = os.MkdirAll(p, 0700); e != nil {
			return nil, e
		}
	}

	// Open the root filesystem for the local storage directory
	if r, e = os.OpenRoot(p); e != nil {
		return nil, e
	}

	// Extract embedded prompt templates if they don't exist
	if e = checkExtract(r, promptPath); e != nil {
		return nil, e
	}

	// Extract embedded report templates if they don't exist
	if e = checkExtract(r, reportPath); e != nil {
		return nil, e
	}

	// Close the root filesystem handle after extraction
	_ = r.Close()
	r = nil

	// Initialize the manager with default values
	m := &mdl{
		mux: sync.RWMutex{},
		uim: u,
		fdb: l,
		cat: filepath.Join(p, catalogFile),
		cfg: filepath.Join(p, llmConfigFile),
		prt: make(map[audast.Lang]map[audtps.CodeType][]byte),
		rep: make(map[audast.Lang]map[string][]byte),
	}

	// Load prompt templates from the local storage directory
	if c, e = readPath(filepath.Join(p, promptPath), ""); e != nil {
		return nil, e
	}

	// parse result to apply into cache map
	for k1, v1 := range c {
		// parse language found to normed lang definition
		lg := audast.Parse(k1)

		for k2, v2 := range v1 {
			// parse entry's types found to normed type definition
			ct := audtps.Parse(k2)

			if m.prt[lg] == nil {
				m.prt[lg] = make(map[audtps.CodeType][]byte)
			}

			// update cache with calculated value
			m.prt[lg][ct] = v2
		}
	}

	// Load report templates from the local storage directory
	if c, e = readPath(filepath.Join(p, reportPath), ""); e != nil {
		return nil, e
	}

	// parse result to apply into cache map
	for k1, v1 := range c {
		// parse language found to normed lang definition
		lg := audast.Parse(k1)

		for k2, v2 := range v1 {
			// no parsing of template name
			// need a report name definition to do it

			if m.rep[lg] == nil {
				m.rep[lg] = make(map[string][]byte)
			}

			// update cache with calculated value
			m.rep[lg][k2] = v2
		}
	}

	// Load catalog data from the local storage directory
	if b, e = m.CatalogLoad(); e != nil {
		return nil, e
	} else if b != nil {
		// apply result database manager to registered instance
		s(b)

		// clean temporary database manager
		b.Clean()
		_ = b.Close()
	}

	return m, nil
}
