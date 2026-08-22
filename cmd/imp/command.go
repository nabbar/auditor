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

package imp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	audpkg "github.com/nabbar/auditor/pkg"
	auddbm "github.com/nabbar/auditor/pkg/data/manager"
	spfcbr "github.com/spf13/cobra"
)

const (
	CommandPrintErrorName  = "import"
	CommandPrintErrorShort = "Import and merge serialized AST state files (JSON, YAML, TOML, CBOR) into the local catalog"
	CommandPrintErrorLong  = `
Import and merge serialized AST dataset files into active memory.

  The 'import' command allows loading externalized AST datasets into the active 
  in-memory data manager. The application automatically hydrates its working state 
  from the local catalog ('.auditor/catalog.bin') upon execution and automatically 
  persists any updated state back to disk upon process termination.

  When running 'import', the specified dataset is parsed and merged directly into 
  the pre-loaded in-memory state. This allows combining local analysis data with 
  external AST catalogs. Once the merge operation completes, the consolidated state 
  is saved automatically to '.auditor/catalog.bin' on exit.

Supported Formats
  The command automatically detects the format of the input file based on its file extension:
  - JSON (.json)
  - YAML (.yaml, .yml)
  - TOML (.toml)
  - CBOR (.cbor, .bin)

State Integration & Merging
  Once the input file is opened, the engine parses its content into a temporary 
  in-memory collection manager. It then invokes the 'Merge' routine on the primary 
  data manager, seamlessly blending existing modules, packages, and code entries 
  with the imported elements while resolving structural duplicates.

Pre-Flight Validation
  The command verifies that exactly one argument is provided, that the target 
  file path exists on disk, and that it resolves to a regular, readable file.

`
	CommandPrintErrorUsage   = "<path_to_file>"
	CommandPrintErrorExample = "auditor.json"
)

func InitCmd() *spfcbr.Command {
	cmd := audpkg.GetCobra().NewCommand(CommandPrintErrorName, CommandPrintErrorShort, CommandPrintErrorLong, CommandPrintErrorUsage, CommandPrintErrorExample)
	cmd.Args = spfcbr.ExactArgs(1)
	cmd.PreRunE = preRun
	cmd.RunE = run
	cmd.DisableFlagsInUseLine = true
	cmd.SilenceErrors = true

	return cmd
}

func preRun(_ *spfcbr.Command, args []string) error {
	if len(args) < 1 || len(args[0]) < 1 {
		return errors.New("no arguments specified")
	}

	if i, e := os.Stat(args[0]); e == nil && !i.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", args[0])
	} else if e != nil {
		return fmt.Errorf("cannot access to import file %s: %w", args[0], e)
	}

	return nil
}

func run(_ *spfcbr.Command, args []string) error {
	var (
		err error
		frt *os.Root
		ffs *os.File
		dbm auddbm.Manager
		dir = filepath.Dir(args[0])
		nam = filepath.Base(args[0])
		ext = filepath.Ext(args[0])
	)

	defer func() {
		if frt != nil {
			_ = frt.Close()
		}
	}()

	defer func() {
		if ffs != nil {
			_ = ffs.Close()
		}
	}()

	if frt, err = os.OpenRoot(dir); err != nil {
		return err
	}

	if ffs, err = frt.OpenFile(nam, os.O_RDONLY, 0644); err != nil {
		return err
	}

	if dbm, _, err = audpkg.GetLocal().CatalogFrom(ffs, ext); err != nil {
		return err
	}

	audpkg.GetDBManager().Merge(dbm)

	return nil
}
