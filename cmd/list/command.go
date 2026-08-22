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

package list

import (
	"fmt"
	"os"
	"slices"

	audpkg "github.com/nabbar/auditor/pkg"
	audent "github.com/nabbar/auditor/pkg/data/entry"
	spfcbr "github.com/spf13/cobra"
)

const (
	CommandPrintErrorName  = "list"
	CommandPrintErrorShort = "Enumerate all parsed source entries and vendor dependencies stored in memory"
	CommandPrintErrorLong  = `
Enumerate all parsed source entries and vendor dependencies.

  The 'list' command is a diagnostic and inspection utility designed to output 
  the current state of the internal in-memory data collection. It provides a 
  clear, human-readable ledger of all structural code constructs (entries) 
  identified during the parsing phase (e.g., after running 'step1').

  The output categorizes and lists every successfully cataloged element, giving 
  auditors and developers immediate visibility into what the engine has mapped 
  before proceeding to complex semantic analysis or report generation.


Output Format
  The command iterates through the in-memory manager and separates the output 
  into two distinct sorted categories:
  
  - Stored Entries: Standard project source code elements.
  - Stored Vendor Entries: Third-party dependencies or external vendor code.

  Each listed entry displays:
  - Internal ID: The unique identifier assigned by the in-memory collection.
  - Full Path: The structural or filesystem path to the source element.
  - Type: The categorized code construct type (e.g., func, type, var).

Pre-Flight Validation
  To execute successfully, the command requires a populated in-memory state. 
  It will verify that at least one Module, Package, and Entry exist before 
  attempting to iterate and display the contents.

`
	CommandPrintErrorUsage   = ""
	CommandPrintErrorExample = ""
)

func InitCmd() *spfcbr.Command {
	cmd := audpkg.GetCobra().NewCommand(CommandPrintErrorName, CommandPrintErrorShort, CommandPrintErrorLong, CommandPrintErrorUsage, CommandPrintErrorExample)
	cmd.Args = spfcbr.NoArgs
	cmd.PreRunE = preRun
	cmd.RunE = run
	cmd.DisableFlagsInUseLine = true
	cmd.SilenceErrors = true

	return cmd
}

func preRun(_ *spfcbr.Command, _ []string) error {
	if audpkg.GetDBManager().ModLen() < 1 {
		return fmt.Errorf("scanning files result no modules")
	} else if audpkg.GetDBManager().PkgLen() < 1 {
		return fmt.Errorf("scanning files result no packages")
	} else if audpkg.GetDBManager().EntLen() < 1 {
		return fmt.Errorf("scanning files result no entries")
	}

	return nil
}

func run(_ *spfcbr.Command, _ []string) error {
	var (
		buf = make([]string, 0, audpkg.GetDBManager().EntLen())
		vdr = make([]string, 0, audpkg.GetDBManager().EntLen())
	)

	audpkg.GetDBManager().EntWalk(func(ent audent.Entry) bool {
		if ent.IsVendor() {
			return true
		}
		buf = append(buf, fmt.Sprintf("  - %d : %s (%s)", ent.GetID(), ent.GetFullPath(), ent.GetTypeString()))
		return true
	})

	audpkg.GetDBManager().EntWalk(func(ent audent.Entry) bool {
		if !ent.IsVendor() {
			return true
		}
		buf = append(buf, fmt.Sprintf("  - %d : %s (%s)", ent.GetID(), ent.GetFullPath(), ent.GetTypeString()))
		return true
	})

	slices.Sort(buf)
	slices.Sort(vdr)

	_, _ = fmt.Fprintf(os.Stdout, "📁 Stored Entries...\n")
	for _, s := range buf {
		_, _ = fmt.Fprintf(os.Stdout, "%s\n", s)
	}

	_, _ = fmt.Fprintf(os.Stdout, "📁 Stored Vendor Entries...\n")
	for _, s := range vdr {
		_, _ = fmt.Fprintf(os.Stdout, "%s\n", s)
	}

	return nil
}
