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

package exp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	audpkg "github.com/nabbar/auditor/pkg"
	spfcbr "github.com/spf13/cobra"
)

const (
	CommandPrintErrorName  = "export"
	CommandPrintErrorShort = "Export active in-memory AST state into a serialized file (JSON, YAML, TOML, CBOR)"
	CommandPrintErrorLong  = `
Export active in-memory AST state to a serialized dataset file.

  The 'export' command serializes and dumps the current in-memory AST dataset 
  to a specified destination file. The exported file contains all parsed 
  modules, packages, and code entries currently managed by the active state.

  This utility is designed for saving intermediate analysis snapshots, sharing 
  cataloged codebases across different machines or environments, and exporting 
  data for external audit processing.

Supported Formats
  The command automatically detects the format of the input file based on its file extension:
  - JSON (.json)
  - YAML (.yaml, .yml)
  - TOML (.toml)
  - CBOR (.cbor, .bin)

Pre-Flight Validation
  Before attempting to open and write to the output file, the command ensures system integrity by verifying that:
  - The in-memory state contains valid data (at least one Module, Package, and Entry).
  - Exactly one argument (destination path) is supplied.
  - The target path does not point to an existing directory.

`
	CommandPrintErrorUsage   = "<path to file exported>"
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
	if audpkg.GetDBManager().ModLen() < 1 {
		return fmt.Errorf("scanning files result no modules")
	} else if audpkg.GetDBManager().PkgLen() < 1 {
		return fmt.Errorf("scanning files result no packages")
	} else if audpkg.GetDBManager().EntLen() < 1 {
		return fmt.Errorf("scanning files result no entries")
	}

	if len(args) < 1 || len(args[0]) < 1 {
		return errors.New("no arguments specified")
	}

	if i, e := os.Stat(args[0]); e == nil && !i.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", args[0])
	}

	return nil
}

func run(_ *spfcbr.Command, args []string) error {
	var (
		err error
		frt *os.Root
		ffs *os.File
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

	if ffs, err = frt.OpenFile(nam, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
		return err
	}

	if err = audpkg.GetLocal().CatalogTo(ffs, ext); err != nil {
		return err
	}

	return nil
}
