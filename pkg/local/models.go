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
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	audast "github.com/nabbar/auditor/pkg/ast"
	auddbm "github.com/nabbar/auditor/pkg/data/manager"
	audtps "github.com/nabbar/auditor/pkg/data/types"
	llmcfg "github.com/nabbar/auditor/pkg/llmconfig"
	auduxi "github.com/nabbar/auditor/pkg/uxi"
	semtps "github.com/nabbar/golib/semaphore/types"
	"gopkg.in/yaml.v3"
)

// mdl represents the local manager implementation.
// It encapsulates the data structures and methods for managing prompts, reports, and catalog operations.
// This struct implements the Manager interface and provides thread-safe access to all managed data.
type mdl struct {
	// mux provides thread-safe access to the manager's data structures.
	// It ensures that concurrent access to prompt and report templates is safe.
	mux sync.RWMutex
	// uim holds the user interface manager instance.
	// This is used for creating new database manager instances when loading catalogs.
	uim auduxi.Manager
	// cat specifies the path to the catalog file.
	// This determines where the catalog data will be saved and loaded from.
	cat string
	// cfg specifies the path to the llm config file.
	cfg string
	// fdb is a function that returns a new database manager instance.
	// It provides a way to create fresh database manager instances for serialization.
	fdb func() auddbm.Manager
	// prt stores prompt templates organized by language and code type.
	// This map allows quick lookup of prompt templates based on language and code type.
	prt map[audast.Lang]map[audtps.CodeType][]byte
	// rep stores report templates organized by language and name.
	// This map allows quick lookup of report templates based on language and template name.
	rep map[audast.Lang]map[string][]byte
}

// Close closes the manager and saves the catalog data to disk.
// It releases all prompt and report templates from memory and ensures that any unsaved changes are persisted.
//
// Returns:
//   - error: An error if saving the catalog fails during the closing process.
func (o *mdl) Close() error {
	err := o.CatalogSave()

	o.mux.Lock()
	defer o.mux.Unlock()

	// Clear the prompt templates from memory
	o.prt = make(map[audast.Lang]map[audtps.CodeType][]byte)
	// Clear the report templates from memory
	o.rep = make(map[audast.Lang]map[string][]byte)

	// Save the current catalog data to disk
	return err
}

// Prompt retrieves a prompt template for the specified code type and language.
// It first attempts to find a matching template for the given language and code type.
// If not found, it tries to find a default template for the same code type in the unknown language.
// If still not found, it attempts to retrieve a default template for the specified code type in the unknown language.
//
// Parameters:
//   - code: The code type for which to retrieve the prompt.
//   - lang: The language of the prompt template.
//
// Returns:
//   - []byte: The prompt template as bytes, or nil if no matching template is found.
func (o *mdl) Prompt(code audtps.CodeType, lang audast.Lang) []byte {
	o.mux.RLock()
	defer o.mux.RUnlock()

	// Check if we have templates for the specified language
	if _, k := o.prt[lang]; k {
		// Check if we have a specific template for this code type and language
		if _, k = o.prt[lang][code]; k {
			return o.prt[lang][code]
		}
		// Check if we have a default template for this language (EntryNone)
		if _, k = o.prt[lang][audtps.EntryNone]; k {
			return o.prt[lang][audtps.EntryNone]
		}
	}

	// If no specific language template was found, check for unknown language templates
	if _, k := o.prt[audast.LangUnknown]; k {
		// Check if we have a specific template for this code type in unknown language
		if _, k = o.prt[audast.LangUnknown][code]; k {
			return o.prt[audast.LangUnknown][code]
		}
		// Check if we have a default template for unknown language (EntryNone)
		if _, k = o.prt[audast.LangUnknown][audtps.EntryNone]; k {
			return o.prt[audast.LangUnknown][audtps.EntryNone]
		}
	}

	// No matching template found
	return nil
}

// Report retrieves a report template for the specified name and language.
// It first attempts to find a matching template for the given language and name.
// If not found, it tries to find a default template for the same name in the unknown language.
//
// Parameters:
//   - name: The name of the report template.
//   - lang: The language of the report template.
//
// Returns:
//   - []byte: The report template as bytes, or nil if no matching template is found.
func (o *mdl) Report(name string, lang audast.Lang) []byte {
	o.mux.RLock()
	defer o.mux.RUnlock()

	// Check if we have templates for the specified language
	if _, k := o.rep[lang]; k {
		// Check if we have a specific template for this name and language
		if _, k = o.rep[lang][name]; k {
			return o.rep[lang][name]
		}
	}

	// If no specific language template was found, check for unknown language templates
	if _, k := o.rep[audast.LangUnknown]; k {
		// Check if we have a specific template for this name in unknown language
		if _, k = o.rep[audast.LangUnknown][name]; k {
			return o.rep[audast.LangUnknown][name]
		}
	}

	// No matching template found
	return nil
}

func (o *mdl) LoadLLMConfig() (*llmcfg.Config, error) {
	// Check if the catalog file exists
	if _, e := os.Stat(o.cat); e != nil {
		return nil, nil
	}

	var (
		e error
		r *os.Root
		f *os.File
		p []byte
		c llmcfg.Config
		b semtps.SemBar

		d = filepath.Dir(o.cfg)
		n = filepath.Base(o.cfg)
	)

	o.mux.RLock()
	defer o.mux.RUnlock()

	defer func() {
		if b != nil {
			b.DeferMain()
		}
	}()

	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()

	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()

	// Open the directory containing the catalog file
	if r, e = os.OpenRoot(d); e != nil {
		o.uim.ErrorStack("failed to open LLM Config file", e)
		return nil, e
	}

	// Open the catalog file for reading
	if f, e = r.OpenFile(n, os.O_RDONLY, 0600); e != nil {
		o.uim.ErrorStack("failed to open LLM Config file", e)
		return nil, e
	}

	o.uim.Info("Loading LLM Config from " + n)
	b = o.uim.NewBar("Loading LLM Config from "+n, 1)
	if e = b.NewWorker(); e != nil {
		o.uim.ErrorStack("failed to initialize worker to load LLM Config from "+n, e)
		return nil, e
	}

	// Read all data from the reader
	if p, e = io.ReadAll(f); e != nil {
		o.uim.ErrorStack("failed to load LLM Config from "+n, e)
		return nil, e
	} else if e = json.Unmarshal(p, &c); e != nil {
		o.uim.ErrorStack("failed to parse loaded LLM Config from "+n, e)
		return nil, e
	}

	o.uim.Info("success LLM Config loaded from " + n)
	b.DeferWorker()
	time.Sleep(5 * time.Millisecond)

	if !b.Completed() {
		b.Inc(1)
		time.Sleep(5 * time.Millisecond)
	}

	return &c, nil
}

func (o *mdl) SaveLLMConfig(cfg *llmcfg.Config) error {
	var (
		e error
		b semtps.SemBar
		r *os.Root
		f *os.File
		p []byte
		d = filepath.Dir(o.cfg)
		n = filepath.Base(o.cfg)
	)

	o.mux.RLock()
	defer o.mux.RUnlock()

	defer func() {
		if b != nil {
			b.DeferMain()
		}
	}()

	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()

	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()

	// Open the directory containing the catalog file
	if r, e = os.OpenRoot(d); e != nil {
		o.uim.ErrorStack("failed to open LLM Config file", e)
		return e
	}

	// Create or truncate the catalog file with write permissions
	if f, e = r.OpenFile(n, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600); e != nil {
		o.uim.ErrorStack("failed to open/create LLM Config file", e)
		return e
	}

	// Determine the appropriate marshaling method based on file extension
	o.uim.Info("Saving LLM Config to " + n)
	b = o.uim.NewBar("Saving LLM Config to "+n, 1)
	if e = b.NewWorker(); e != nil {
		o.uim.ErrorStack("failed to initialize worker to save LLM Config to "+n, e)
		return e
	}

	// Serialize the catalog data
	if p, e = json.Marshal(cfg); e != nil {
		o.uim.ErrorStack("failed to marshall LLM Config from "+n, e)
		return e
	} else if _, e = f.Write(p); e != nil {
		o.uim.ErrorStack("failed to write LLM Config to file "+n, e)
		return e
	}

	o.uim.Info("success LLM Config saved to " + n)
	b.DeferWorker()
	time.Sleep(5 * time.Millisecond)

	if !b.Completed() {
		b.Inc(1)
		time.Sleep(5 * time.Millisecond)
	}

	return nil
}

// CatalogSave saves the current catalog data to the file specified by the catalog path.
// It opens the directory containing the catalog file, creates or truncates the file,
// and writes the catalog data using the appropriate serialization method based on the file extension.
//
// Returns:
//   - error: An error if opening the file or writing the catalog fails.
func (o *mdl) CatalogSave() error {
	var (
		e error
		r *os.Root
		f *os.File
	)

	o.mux.RLock()
	defer o.mux.RUnlock()

	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()

	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()

	// Open the directory containing the catalog file
	if r, e = os.OpenRoot(filepath.Dir(o.cat)); e != nil {
		o.uim.ErrorStack("failed to open catalog file", e)
		return e
	}

	// Create or truncate the catalog file with write permissions
	if f, e = r.OpenFile(filepath.Base(o.cat), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600); e != nil {
		o.uim.ErrorStack("failed to open/create catalog file", e)
		return e
	}

	// Serialize and write the catalog data to the file
	if e = o.CatalogTo(f, filepath.Ext(o.cat)); e != nil {
		o.uim.ErrorStack("failed to save catalog to file "+o.cat, e)
		return e
	}

	return nil
}

// CatalogLoad loads catalog data from the file specified by the catalog path.
// It opens the directory containing the catalog file and reads its contents.
// The loaded data is then unmarshaled into a new database manager instance.
//
// Returns:
//   - auddbm.Manager: The loaded database manager instance, or nil if no catalog exists.
//   - error: An error if opening or reading the file fails.
func (o *mdl) CatalogLoad() (auddbm.Manager, error) {
	// Check if the catalog file exists
	if _, e := os.Stat(o.cat); e != nil {
		return nil, nil
	}

	var (
		e error
		r *os.Root
		f *os.File
		d auddbm.Manager
	)

	o.mux.RLock()
	defer o.mux.RUnlock()

	defer func() {
		if r != nil {
			_ = r.Close()
		}
	}()

	defer func() {
		if f != nil {
			_ = f.Close()
		}
	}()

	// Open the directory containing the catalog file
	if r, e = os.OpenRoot(filepath.Dir(o.cat)); e != nil {
		o.uim.ErrorStack("failed to open catalog file", e)
		return nil, e
	}

	// Open the catalog file for reading
	if f, e = r.OpenFile(filepath.Base(o.cat), os.O_RDONLY, 0600); e != nil {
		o.uim.ErrorStack("failed to open catalog file", e)
		return nil, e
	}

	// Deserialize and load the catalog data into a new database manager instance
	if d, _, e = o.CatalogFrom(f, filepath.Ext(o.cat)); e != nil {
		o.uim.ErrorStack("failed to save catalog to file "+o.cat, e)
		return d, e
	}

	return d, e
}

// CatalogTo serializes the catalog data to the provided writer using the specified file extension.
// It determines the appropriate marshaling method based on the file extension and writes the serialized data.
//
// Parameters:
//   - wrt: The writer to which the catalog data will be written.
//   - ext: The file extension used to determine the serialization format.
//
// Returns:
//   - error: An error if serialization fails or an invalid extension is provided.
func (o *mdl) CatalogTo(wrt io.Writer, ext string) error {
	var (
		b semtps.SemBar
		f func() ([]byte, error)
		d = o.fdb()
	)

	defer func() {
		if b != nil {
			b.DeferMain()
		}
	}()

	// Determine the appropriate marshaling method based on file extension
	switch ext {
	case ".bin", ".cbor":
		f = d.MarshalCBOR
	case ".json":
		f = d.MarshalJSON
	case ".toml":
		f = d.MarshalTOML
	case ".yaml", ".yml":
		f = func() ([]byte, error) {
			i, e := d.MarshalYAML()
			if e != nil {
				return nil, e
			}
			if p, k := i.([]byte); k {
				return p, e
			}
			return nil, errors.New("invalid return of YAML Marshaller")
		}
	default:
		return errors.New("invalid extension")
	}

	o.uim.Info("Saving catalog to " + ext[1:] + " writer...")
	b = o.uim.NewBar("Saving Catalog to "+ext[1:]+" writer", 1)
	if e := b.NewWorker(); e != nil {
		o.uim.ErrorStack("failed to initialize worker to save catalog to "+ext[1:]+" writer", e)
		return e
	}

	// Serialize the catalog data
	if p, e := f(); e != nil {
		o.uim.ErrorStack("failed to marshall "+ext[1:]+" data", e)
		return e
	} else if _, e = wrt.Write(p); e != nil {
		o.uim.ErrorStack("failed to write "+ext[1:]+" data to writer", e)
		return e
	}

	o.uim.Info("success catalog saved to " + ext[1:] + " writer...")
	b.DeferWorker()
	time.Sleep(5 * time.Millisecond)

	if !b.Completed() {
		b.Inc(1)
		time.Sleep(5 * time.Millisecond)
	}

	return nil
}

// CatalogFrom deserializes catalog data from the provided reader using the specified file extension.
// It determines the appropriate unmarshaling method based on the file extension and loads the data into a new database manager instance.
//
// Parameters:
//   - rdr: The reader from which the catalog data will be read.
//   - ext: The file extension used to determine the deserialization format.
//
// Returns:
//   - auddbm.Manager: The loaded database manager instance.
//   - error: An error if deserialization fails or an invalid extension is provided.
func (o *mdl) CatalogFrom(rdr io.Reader, ext string) (auddbm.Manager, *auddbm.Linker, error) {
	var (
		f    func([]byte) error
		b    semtps.SemBar
		d, l = auddbm.New(o.uim)
	)

	defer func() {
		if b != nil {
			b.DeferMain()
		}
	}()

	// Determine the appropriate unmarshaling method based on file extension
	switch ext {
	case ".bin", ".cbor":
		f = d.UnmarshalCBOR
	case ".json":
		f = d.UnmarshalJSON
	case ".toml":
		f = func(i []byte) error {
			return d.UnmarshalTOML(i)
		}

	case ".yaml", ".yml":
		f = func(i []byte) error {
			return yaml.Unmarshal(i, d)
		}
	default:
		return nil, nil, errors.New("invalid extension")
	}

	o.uim.Info("Loading catalog from " + ext[1:] + " reader...")
	b = o.uim.NewBar("Loading Catalog from "+ext[1:]+" reader", 1)
	if e := b.NewWorker(); e != nil {
		o.uim.ErrorStack("failed to initialize worker to load catalog from "+ext[1:]+" reader", e)
		return nil, nil, e
	}

	// Read all data from the reader
	if p, e := io.ReadAll(rdr); e != nil {
		o.uim.ErrorStack("failed to load buffer from "+ext[1:]+" reader", e)
		return nil, nil, e
	} else if e = f(p); e != nil {
		o.uim.ErrorStack("failed to parse loaded buffer from "+ext[1:]+" reader", e)
		return nil, nil, e
	}

	o.uim.Info("success catalog loaded from " + ext[1:] + " reader")
	b.DeferWorker()
	time.Sleep(5 * time.Millisecond)

	if !b.Completed() {
		b.Inc(1)
		time.Sleep(5 * time.Millisecond)
	}

	return d, l, nil
}
