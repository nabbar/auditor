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

package step4

import (
	"errors"
	"fmt"
	"os"

	audpkg "github.com/nabbar/auditor/pkg"
	spfcbr "github.com/spf13/cobra"
)

const (
	cmdName  = "step4"
	cmdShort = "Final ledger aggregation and exporting to the specified directory path"
	cmdLong  = `
Final ledger aggregation and structured Markdown report export.

  The 'step4' command (also accessible via the 'report' alias) represents the 
  final phase of the auditing toolchain. It is responsible for extracting the 
  audited elements from the in-memory matrices, consolidating the individual 
  evaluation blocks into a coherent narrative, and persisting a unified Markdown 
  document hierarchy to the specified output pathway.

Pre-Flight Validation:
  Before generating the final ledger, the command enforces environmental integrity:
  - Validates that a target output folder is provided as an argument and exists 
    as a valid directory.
  - Verifies the internal in-memory state to ensure at least one Module, Package, 
    and Entry have been successfully parsed and analyzed during prior phases.

Dependency Graph Resolution:
  Prior to file generation, the engine executes a 'ReOrder' synchronization pass. 
  This phase resolves semantic dependency graphs and inter-package reference maps, 
  ordering all entries to ensure that foundational dependencies are evaluated and 
  placed logically before the components that rely on them.

Report Aggregation Hierarchy
  The generation engine walks through the in-memory state, systematically ignoring 
  empty structures, invalid IDs, and external vendor files to produce a clean audit. 
  The compilation occurs in a cascading architecture:
  
  1. Entries: Individual specialized documentation blocks are retrieved for each 
     code construct.
  2. Packages: Entry reports are buffered and dynamically appended to their 
     corresponding parent package files.
  3. Modules: Package reports are aggregated into top-level module documentation, 
     resulting in a comprehensive, fully structured Markdown document for export.

`
	cmdUsage   = "<path_to_directory>"
	cmdExample = "./auditor_reports/"
)

func InitCmd() *spfcbr.Command {
	cmd := audpkg.GetCobra().NewCommand(cmdName, cmdShort, cmdLong, cmdUsage, cmdExample)
	cmd.Args = spfcbr.ExactArgs(1)
	cmd.PreRunE = preRun
	cmd.RunE = run
	cmd.DisableFlagsInUseLine = true
	cmd.SilenceErrors = true

	cmd.Aliases = append(cmd.Aliases, "report")

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
		return errors.New("output folder is not defined in arguments")
	}

	if i, e := os.Stat(args[0]); e == nil && !i.Mode().IsDir() {
		return fmt.Errorf("%s is not a directory", args[0])
	}

	return nil
}

func run(_ *spfcbr.Command, args []string) error {
	audpkg.GetUIManager().Info("⛓️ Resolving Semantic Dependency Graphs & Inter-Package Reference Maps...")
	if e := audpkg.GetEngine().ReOrder(); e != nil {
		return fmt.Errorf("error trigger on Resolving Semantic Dependency Graphs: %w", e)
	}

	audpkg.GetDBManager().Info("📝 Consolidating multi-pass analytical documentation segments into final report markdown...")
	if e := audpkg.GetEngine().Report(args[0]); e != nil {
		return fmt.Errorf("error trigger on Reviewing Audit: %w", e)
	}

	return nil
}
