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

package step1

import (
	"fmt"

	audpkg "github.com/nabbar/auditor/pkg"
	spfcbr "github.com/spf13/cobra"
)

const (
	CmdName  = "step1"
	CmdShort = "Execute repository discovery, polyglot language identification, and AST parsing"
	CmdLong  = `
Execute local repository discovery, polyglot language identification, and initial Abstract 
Syntax Tree (AST) parsing.

  The 'step1' command acts as the foundational initialization phase of the auditor 
  toolchain. Before any deep semantic dependency analysis ('step2') or LLM-powered 
  vulnerability assessment ('step3') can occur, the tool must fully map the physical 
  filesystem and extract structural boundaries from the source code. 

  Unlike subsequent steps, 'step1' operates entirely locally. It requires no 
  external LLM configuration, consumes no API tokens, and makes no network requests. 
  It focuses purely on high-speed, multi-threaded parsing of source files to catalog 
  modules, packages, and individual code entries into the internal in-memory data 
  manager.

Discovery and Parsing phases:
  The command orchestrates a highly concurrent, multi-pass engine to analyze the codebase:

  1. Topological Discovery: The engine walks the configured repository path. 
     It intentionally skips hidden directories and files (dotfiles), utilizing 
     language-specific glob patterns to efficiently filter and map source files to 
     their respective programming languages prior to expensive parsing operations.
  2. Linguistic Identification: For each discovered file, the engine runs heuristic 
     checks to confirm the matched language and instantiates the appropriate polyglot 
     AST parser.
  3. Module Extraction: The AST engine scans the source files to construct 'Module' 
     entities, capturing high-level metadata, version constraints, and top-level 
     dependency declarations.
  4. Package Processing: The engine parses package-level structures, extracting 
     functions, type definitions (structs, interfaces), variable declarations, and 
     import relationships.
  5. Synchronization: A final update pass is executed to refresh and finalize the 
     package structures, ensuring the parsed in-memory representations perfectly align 
     with the extracted structural hierarchy.

In-Memory State Population & Integrity:
  The primary output of this command is a fully populated in-memory data collection. 
  This centralized manager structures code elements independently of any traditional 
  database and supports native serialization of the extracted AST mapping into JSON, 
  YAML, TOML, or CBOR formats.
  
  At the end of execution, the system runs a strict state check against the memory 
  collections. The execution will instantly fail if the state does not contain:
  
  - At least one Module (top-level library/project boundary)
  - At least one Package (sub-path or sub-library)
  - At least one distinct Entry (individual functions, types, AST nodes)

  This guarantees that subsequent analytical steps ('step2', 'step3') have 
  mathematically valid structural data to operate on.
`
	CmdUsage   = ""
	CmdExample = ""
)

func InitCmd() *spfcbr.Command {
	cmd := audpkg.GetCobra().NewCommand(CmdName, CmdShort, CmdLong, CmdUsage, CmdExample)
	cmd.Args = spfcbr.NoArgs
	cmd.PreRunE = preRun
	cmd.RunE = run
	cmd.DisableFlagsInUseLine = true
	cmd.SilenceErrors = true

	cmd.Aliases = []string{"scan"}

	return cmd
}

func preRun(_ *spfcbr.Command, _ []string) error {
	return nil
}

func run(_ *spfcbr.Command, _ []string) error {
	audpkg.GetUIManager().Info("📁 Local AST Discovery & Structural Signature Cataloging...")
	if e := audpkg.GetEngine().ScanFiles(); e != nil {
		return fmt.Errorf("error trigger on Local AST Discovery: %w", e)
	}

	if audpkg.GetDBManager().ModLen() < 1 {
		return fmt.Errorf("scanning files result no modules")
	} else if audpkg.GetDBManager().PkgLen() < 1 {
		return fmt.Errorf("scanning files result no packages")
	} else if audpkg.GetDBManager().EntLen() < 1 {
		return fmt.Errorf("scanning files result no entries")
	}

	return nil
}
