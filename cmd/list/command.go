// MIT License
//
// Copyright (c) 2026 Nicolas JUHEL
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package list implements the "list" CLI command that enumerates all parsed
// source entries and vendor dependencies stored in the in-memory data manager.
// It provides a human-readable, structured output organized by Module, Package,
// and Entry hierarchy, with each construct's unique ID, full path, type,
// and status displayed. This command is primarily intended for diagnostic and
// inspection purposes, allowing auditors and developers to verify the state of
// the internal data collection before proceeding to semantic analysis or
// report generation stages.
package list

import (
	"fmt"
	"os"

	audpkg "github.com/nabbar/auditor/pkg"
	audids "github.com/nabbar/auditor/pkg/data/id"
	spfcbr "github.com/spf13/cobra"
)

// CommandListName holds the display name of the CLI command as registered
// with the Cobra framework. This identifier is used to invoke the command from
// the command line (e.g. "auditor list").
//
// CommandListShort is the one-line summary displayed when users run
// "auditor --help". It gives a concise description of the "list" command's
// purpose.
//
// CommandListLong is the extended description displayed when users run
// "auditor help list" or "auditor list --help". It explains the command's
// behavior, output format, and pre-flight validation requirements in detail.
//
// CommandListUsage is the usage pattern string shown in the help output.
// An empty value means Cobra will auto-generate a standard usage line.
//
// CommandListExample holds example invocation strings displayed in the
// help output. An empty value means no usage examples are provided.
const (
	CommandListName  = "list"
	CommandListShort = "Enumerate all parsed source entries and vendor dependencies stored in memory"
	CommandListLong  = `
Enumerate all parsed source entries stored in the in-memory data manager.  

  The 'list' command is a diagnostic and inspection utility designed to output 
  the current state of the internal in-memory data collection. It provides a 
  clear, human-readable ledger of all structural code constructs (entries) 
  identified during the parsing phases (e.g., after running 'step1' and subsequent steps).

  The output categorizes and lists every successfully cataloged element, giving 
  auditors and developers immediate visibility into what the engine has mapped 
  before proceeding to complex semantic analysis or report generation.


Output Format
  The command iterates through the in-memory manager and displays the project 
  source entries structured hierarchically:
  
  - Module: The top-level project root container.
  - Package: The logical grouping of source files within a module.
  - Entry: Individual code constructs (e.g., functions, types, variables, constants).
  - Depend: Dependency references associated with each entry.

  Each listed element displays:
  - Internal ID: The unique identifier assigned by the in-memory collection.
  - Full Path / Name: The structural path or name of the source element.
  - Type & Status: The categorized code construct type and its current lifecycle status.

Pre-Flight Validation
  To execute successfully, the command requires a populated in-memory state. 
  It will verify that at least one Module, Package, and Entry exist before 
  attempting to iterate and display the contents.
`
	CommandListUsage   = ""
	CommandListExample = ""
)

// InitCmd creates and configures the Cobra command for the "list" CLI
// sub-command. It instantiates a new command using the predefined constant
// values for name, short description, long description, usage pattern, and
// example text, all sourced from the package-level const block.
//
// The resulting command is configured with the following behavior:
//   - It accepts no positional arguments (cobra.NoArgs).
//   - Its PreRunE hook is bound to the preRun function, which performs
//     pre-flight validation of the in-memory data manager state before
//     allowing the command to execute.
//   - Its RunE hook is bound to the run function, which performs the
//     actual enumeration and output of stored entries and dependencies.
//
// The returned *spfcbr.Command should be attached to the root CLI application
// by the caller, so that the "auditor list" invocation becomes available.
func InitCmd() *spfcbr.Command {
	// Create the Cobra command with the constant-defined metadata.
	cmd := audpkg.GetCobra().NewCommand(CommandListName, CommandListShort, CommandListLong, CommandListUsage, CommandListExample)
	// The command accepts no positional arguments since it operates on the
	// in-memory data store without requiring additional input from the user.
	cmd.Args = spfcbr.NoArgs
	// Bind the pre-flight validation hook that checks whether the in-memory
	// data manager contains at least one module, one package, and one entry.
	// If any of these conditions are not met, the command aborts with an
	// error before the run phase is reached.
	cmd.PreRunE = preRun
	// Bind the main execution hook that iterates over all stored modules,
	// packages, entries, and their dependencies, printing a structured
	// hierarchy to standard output.
	cmd.RunE = run

	return cmd
}

// preRun performs pre-flight validation of the in-memory data manager before
// the command is allowed to execute. It verifies that the data store contains
// a minimum viable state by checking three conditions:
//   - At least one module exists (ModLen >= 1).
//   - At least one package exists (PkgLen >= 1).
//   - At least one entry exists (EntLen >= 1).
//
// If any of these checks fails, an error is returned with a descriptive
// message indicating which structural element is missing from the parsed
// data. Returning a non-nil error from PreRunE prevents the RunE hook
// from executing, effectively aborting the command before any output is
// generated. This is necessary because the run function expects a populated
// data store and would produce empty or misleading output otherwise.
func preRun(_ *spfcbr.Command, _ []string) error {
	// Verify that at least one module was discovered during the scanning
	// phase. Without any modules, there is nothing to enumerate.
	if audpkg.GetDBManager().ModLen() < 1 {
		return fmt.Errorf("scanning files result no modules")
	}

	// Verify that at least one package was discovered. Packages are the
	// organizational units within modules, and their absence indicates
	// an incomplete or empty parse result.
	if audpkg.GetDBManager().PkgLen() < 1 {
		return fmt.Errorf("scanning files result no packages")
	}

	// Verify that at least one entry (code construct) was discovered.
	// Entries represent the individual source-level elements such as
	// functions, types, variables, constants, etc. Without entries there
	// is no substantive data to display.
	if audpkg.GetDBManager().EntLen() < 1 {
		return fmt.Errorf("scanning files result no entries")
	}

	return nil
}

// run executes the primary enumeration logic of the "list" command. It
// iterates over the in-memory data manager's hierarchy starting from modules,
// then descending into packages, then into individual entries, and finally
// into the dependencies of each entry. For each element, it prints a
// structured line containing the element type, full path, unique ID, and
// type/status string, using indentation markers to visually represent the
// hierarchical relationship between parent and child elements.
//
// The output hierarchy follows this structure:
//
//	Module (top-level project root)
//	  Package (logical grouping within a module)
//	    Entry (individual code construct: function, type, variable, etc.)
//	      Depend (dependency references from the entry)
//
// If any element retrieved from the manager is nil or reports itself as
// empty, it is skipped to avoid spurious output from stale or corrupted
// data references. The function always returns nil on success since output
// is written directly to os.Stdout via fmt.Fprintf.
func run(_ *spfcbr.Command, _ []string) error {
	// Print the header section for stored (non-vendor) entries.
	_, _ = fmt.Fprintf(os.Stdout, "📁 Stored Entries...\n\n")

	// Iterate over all modules tracked by the in-memory data manager.
	// The false argument to ModList indicates that only non-vendor modules
	// are requested, as vendor modules are reported separately.
	for _, m := range audpkg.GetDBManager().ModList(false) {
		// Retrieve the full module object by its ID. A unique ID is
		// generated from the module key using GenID to ensure consistent
		// lookup semantics across the manager.
		mod := audpkg.GetDBManager().ModGet(audids.GenID(m))

		// Skip any module that is nil or reports as empty; this can
		// happen when a module reference exists in the list but the
		// actual data was removed or failed to initialize.
		if mod == nil || mod.IsEmpty() {
			continue
		}

		// Print the module descriptor with its name, unique ID, and
		// the source file path or location it represents.
		_, _ = fmt.Fprintf(os.Stdout, "> Module : %s (id: %22d), file: %s\n", mod.GetMod(), mod.GetID(), mod.GetFile())

		// Iterate over all packages that belong to this module. Each
		// package is identified by a unique package ID that was
		// generated during the parsing phase.
		for _, p := range audpkg.GetDBManager().PkgList(mod.GetID()) {
			// Retrieve the full package object by its ID, using the
			// same GenID-based lookup pattern as for modules.
			pkg := audpkg.GetDBManager().PkgGet(audids.GenID(p))

			// Skip any package that is nil or empty to avoid
			// displaying incomplete or corrupted data.
			if pkg == nil || pkg.IsEmpty() {
				continue
			}

			// Print the package descriptor with its full structural
			// path and unique ID, indented to show its subordination
			// to the parent module.
			_, _ = fmt.Fprintf(os.Stdout, "|--> Package : %s (id: %22d)\n", pkg.GetFullPath(), pkg.GetID())

			// Iterate over all entries (code constructs) contained
			// within this package. An entry represents a single
			// named element in the source code such as a function,
			// type definition, variable declaration, constant, etc.
			for _, e := range audpkg.GetDBManager().EntList(pkg.GetID()) {
				// Retrieve the full entry object by its ID.
				ent := audpkg.GetDBManager().EntGet(audids.GenID(e))

				// Skip any entry that is nil or empty.
				if ent == nil || ent.IsEmpty() {
					continue
				}

				// Print the entry descriptor with its full path, unique
				// ID, construct type string (e.g., "func", "type"), and
				// current status string, indented to show its subordination
				// to the parent package.
				_, _ = fmt.Fprintf(os.Stdout, "|  |--> Entry : %s (id: %22d), %s - %s\n", ent.GetFullPath(), ent.GetID(), ent.GetTypeString(), ent.GetStatus())

				// Iterate over all dependencies associated with this
				// entry. Dependencies represent references or imports
				// that this entry depends on, which may resolve to
				// other entries or to external code constructs.
				for _, d := range ent.GetDepend() {
					// Check that the dependency object is valid and
					// non-empty before attempting to access its fields.
					if d == nil || d.IsEmpty() {
						continue
					}
					// Print the dependency descriptor with its full path,
					// unique ID, type string, and status string, indented
					// further to show its subordination to the parent entry.
					_, _ = fmt.Fprintf(os.Stdout, "|  |  |--> Depend : %s (id: %22d), %s - %s\n", d.GetFullPath(), d.GetID(), d.GetTypeString(), d.GetStatus())
				}
			}
		}
	}

	return nil
}
