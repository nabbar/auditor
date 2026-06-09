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

// Package main contains the core logic for the auditor tool that analyzes Go functions
// and identifies potential bugs through LLM-based code review techniques.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nabbar/auditor/database"
	"github.com/nabbar/auditor/fileutils"
	"github.com/nabbar/auditor/llm"
	"github.com/nabbar/auditor/models"
)

// lazyTreeAnalyze performs a recursive tree analysis of a function's dependencies.
//
// This function implements a depth-first search approach to analyze dependencies,
// ensuring that each function is analyzed only once and avoiding circular dependencies
// by tracking the analysis status in the database. It recursively traverses the
// dependency tree, analyzing functions in a top-down manner.
//
// The algorithm works as follows:
// 1. Checks if maximum recursion depth has been exceeded to prevent stack overflow
// 2. Verifies if function is already analyzed or currently being processed
// 3. Marks function as "IN_PROGRESS" to prevent concurrent processing
// 4. Executes LLM-based analysis to extract function code and dependencies
// 5. Recursively analyzes dependent functions (if they are internal)
// 6. Updates the final summary after dependency resolution
//
// Parameters:
//   - db: Database manager for persisting analysis results and managing state
//   - ai: LLM client for code analysis using Ollama services
//   - f: Function model containing information about the function to analyze
//   - depth: Current recursion depth level to prevent stack overflow
//
// Returns:
//   - bool: Indicates whether the analysis was successful or not
func lazyTreeAnalyze(db *database.DBManager, ai *llm.Client, f models.Function, depth int) bool {
	// Strictly limit the recursion depth to prevent stack overflow.
	// This safeguard protects against infinite recursion during dependency tree traversal.
	// The maximum depth is controlled by flgDpt flag which defines the allowed recursion level
	if depth > flgDpt {
		log.Printf("[WARNING] [main] Maximum recursion depth (%d) reached at %s. Stopping descent.", flgDpt, f.ID)
		return false
	}

	// Retrieve current status of the function from database
	stsCur := db.GetStatus(f.ID)

	// If function is currently being processed, it indicates a circular dependency
	// In such cases, we return true to avoid processing the same function again
	if stsCur == "IN_PROGRESS" {
		log.Printf("[WARNING] [main] Circular dependency detected on %s. Using stub.", f.ID)
		return true
	}

	// If function has already been extracted or failed parsing, no need to reprocess
	if stsCur == "EXTRACTED" || stsCur == "FAILED_PARSE" {
		return true
	}

	// Mark the function as being analyzed.
	// This prevents concurrent workers from processing the same function simultaneously.
	// It ensures that each function is processed exactly once at a time
	_ = db.UpdateStatus(f.ID, "IN_PROGRESS")

	// Execute Pass 1 analysis which extracts code and dependencies using LLM
	pssScs, pssSum, pssDep := runStep1(db, ai, f, "")

	// If initial analysis fails, attempt retries up to 3 times
	if !pssScs {
		for attempt := 1; attempt <= 3; attempt++ {
			log.Printf("[INFO] [main] Retry %d/3 for discovery of %s", attempt, f.ID)
			pssScs, pssSum, pssDep = runStep1(db, ai, f, "⚠️ ATTENTION: Strict JSON format required.")
			if pssScs {
				break
			}
		}
	}

	// If after retries the analysis still fails, mark as FAILED_PARSE and optionally exit
	if !pssScs {
		log.Printf("[ERROR] [main] Definitive analysis failure for structure of %s", f.ID)
		_ = db.UpdateStatus(f.ID, "FAILED_PARSE")
		if flgErr {
			log.Fatalf("[CRITICAL] Stopping on error (-suspend-on-error)")
		}
		return false
	}

	// Save the dependencies identified during Pass 1 analysis to database
	db.SaveDependencies(f.ID, pssDep)

	// Initialize flag for determining if any dependency needed evaluation
	nedEvl := false

	// Iterate through all dependencies identified in the function
	for _, dep := range pssDep {
		// Only process dependencies that are functions (not external libraries)
		if dep.DepType == "function" {
			// Find the dependent function in catalog and check if it's internal
			childFunc, estInterne := db.FindInCatalog(dep.DepName)

			// If dependency is internal to the project
			if estInterne {
				// Check if the dependency has already been analyzed (has summary)
				if !db.HasSummary(dep.DepName) {
					log.Printf("[TREE] [%d] Descending from %s -> to child: %s", depth, f.ID, childFunc.ID)
					// Recursive call (depth + 1) - continue analysis of dependent functions
					descendanceReussie := lazyTreeAnalyze(db, ai, childFunc, depth+1)
					if descendanceReussie {
						nedEvl = true
					}
				} else {
					// If dependency already has summary, mark that evaluation is needed
					nedEvl = true
				}
			}
		}
	}

	// If any dependency required further evaluation, update the summary with enriched context
	if nedEvl {
		log.Printf("[TREE] [%d] Enriching summary of %s using its resolved dependencies.", depth, f.ID)
		_, pssSum, _ = runStep1(db, ai, f, "")
	}

	// Update the final summary in database for this function
	_ = db.UpdatePasse1Final(f.ID, pssSum)

	// Update progress only for root functions at level 1.
	// This ensures that the progress bar accurately reflects the completion of analysis
	// rather than just counting individual dependencies.
	// Root functions are those that don't have any dependency in the current context
	if depth == 1 {
		incPgr()
	}

	return true
}

// ... existing code ...

// runStep1 executes a single call to Ollama for Pass 1 (optimized model).
//
// This function extracts function code and dependencies using LLM analysis.
// It performs a structured analysis of Go functions by:
// 1. Extracting the source code from file
// 2. Retrieving context about previously analyzed dependencies
// 3. Formulating a prompt for LLM with specific format requirements
// 4. Parsing the JSON response to extract function summary and dependencies
//
// Parameters:
//   - db: Database manager for storing extracted information and retrieving context
//   - ai: LLM client for code analysis using Ollama services
//   - f: Function model containing the target function information
//   - extraNote: Additional instructions to guide the LLM response format
//
// Returns:
//   - bool: Indicates whether the extraction was successful
//   - string: Summary of the function's purpose in English (max 45 words)
//   - []models.Dependency: List of dependencies identified in the function
func runStep1(db *database.DBManager, ai *llm.Client, f models.Function, extraNote string) (bool, string, []models.Dependency) {
	// Extract the code for the specified function from its file location
	// This utility function reads and extracts a specific function's source code from a Go file
	code, err := fileutils.ExtractFunctionCode(f.FilePath, f.FuncName)
	if err != nil || code == "" {
		return false, "", nil
	}

	// Retrieve context of dependencies if they exist in DB.
	// This provides previously analyzed information about related functions to guide LLM analysis
	// The context helps the LLM understand how dependencies have been previously interpreted
	ctxDep, _ := db.GetDependenciesContext(f.ID)

	// Prepare prompt for LLM analysis with structured format requirements
	// The prompt includes:
	// - Known context of child elements (previous analysis results)
	// - Code to analyze (the target function)
	// - Strict JSON format requirement (ensuring predictable output)
	pssPrt := fmt.Sprintf(`Analyze this Go function to understand its internal structure.

--- KNOWN CONTEXT OF CHILD ELEMENTS (PREVIOUS TURNS) ---
%s

--- CODE TO ANALYZE ---
%s
%s

--- FORMAT REQUIREMENT (STRICT JSON) ---
You must respond ONLY with a strict JSON block respecting the exact order of following keys:
1. "summary": An ultra-concise summary in English explaining globally the function.
   ⚠️ STRICT SIZE CONSTRAINT: Maximum 45 words. Get straight to the point to save tokens.
2. "external": An array of objects listing ONLY the most important direct and unique dependencies (functions or structures). No duplicates.

Valid format example:
{
  "summary": "This function validates session accesses.",
  "external": [{"name": "db.Query", "type": "function"}]
}

Respond ONLY with the JSON block, no explanatory text before or after.`, ctxDep, code, extraNote)

	// Call the LLM API with the prepared prompt
	rsp, err := ai.Call(pssPrt, true, 2048)
	if err != nil {
		return false, "", nil
	}

	// Parse the JSON response into a structured format
	var p1 models.P1Response
	if err := json.Unmarshal([]byte(rsp), &p1); err != nil {
		log.Printf("[WARNING] JSON error on %s: %v | Raw: %s", f.ID, err, strings.ReplaceAll(rsp, "\n", " "))
		return false, "", nil
	}

	// Convert the LLM response into dependency objects for storage
	var dep []models.Dependency
	for _, ext := range p1.External {
		dep = append(dep, models.Dependency{
			FunctionID: f.ID,
			DepName:    ext.Name,
			DepType:    ext.Type,
		})
	}

	return true, p1.Summary, dep
}

// runStep2 executes the logical audit pass on all extracted functions.
//
// This function identifies potential bugs by analyzing function logic and dependencies
// using LLM-based code review techniques to detect issues like nil pointers, resource leaks,
// context blocking, and mutex deadlocks. The audit process:
// 1. Retrieves all functions that have been successfully extracted
// 2. Opens or creates the report file for outputting findings
// 3. Iterates through each function to perform code review
// 4. Uses LLM to analyze logic and detect potential bugs
// 5. Records any identified anomalies in the report
//
// Parameters:
//   - db: Database manager for retrieving function data and storing results
//   - ai: LLM client for code analysis using Ollama services
//   - useProgressBar: Flag indicating whether to display progress information
func runStep2(db *database.DBManager, ai *llm.Client, useBar bool) {
	tmsStr = time.Now()

	// Retrieve all functions that have been successfully processed in Pass 1
	fctSts, err := db.GetFunctionsByStatus("EXTRACTED")
	if err != nil {
		log.Printf("[ERROR] [run] Failed to fetch functions for Pass 2: %v", err)
		return
	}

	var tot = len(fctSts)
	if tot == 0 {
		return
	}

	// Open the markdown report file in append mode (or create it if it doesn't exist)
	frp, err := os.OpenRoot(filepath.Dir(flgRep))
	if err != nil {
		log.Fatalf("[ERROR] [run] Failed to open report file %s: %v", flgRep, err)
	}
	defer func() {
		_ = frp.Close()
	}()

	rep, err := frp.OpenFile(filepath.Base(flgRep), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("[ERROR] [run] Failed to open report file %s: %v", flgRep, err)
	}
	defer func() {
		_ = rep.Close()
	}()

	// Process each function sequentially for detailed logical analysis
	for idx, f := range fctSts {
		if useBar {
			drawPgr(idx+1, tot, tmsStr)
		} else {
			log.Printf("[INFO] [run] Auditing function [%d/%d]: %s", idx+1, tot, f.ID)
		}

		// Extract the full source code body of the current function
		cod, err := fileutils.ExtractFunctionCode(f.FilePath, f.FuncName)
		if err != nil {
			log.Printf("[WARNING] [run] Failed to extract code for %s: %v", f.ID, err)
			continue
		} else if cod == "" {
			log.Printf("[WARNING] [run] Failed to extract code for %s: (empty)", f.ID)
			continue
		}

		// Retrieve the structured contextual information about its dependencies
		ctxDep, err := db.GetDependenciesContext(f.ID)
		if err != nil {
			ctxDep = "None (No internal project dependencies mapped)."
		}

		// Construct a highly descriptive prompt for logical verification
		prt := fmt.Sprintf(`You are a senior Go code auditor. You must validate the internal logic of the provided function and detect external usage errors.

--- KNOWN AND WELL-DEFINED BEHAVIOR OF EXTERNAL ELEMENTS ---
%s
(For any other element outside the project or not listed here, consider it safe and valid).

--- CODE TO AUDIT ---
%s

--- INSTRUCTIONS ---
- Look for critical bugs: Nil Pointers, resource leaks/context blocking, Mutex deadlocks.
- If everything is correct, respond exclusively: OK
- If a bug is present, briefly describe it in English (Issue Type + Root Cause). Do not provide code or correction suggestions.`, ctxDep, cod)

		// Call the LLM API to perform audit analysis on the function
		rsp, err := ai.Call(prt, false, 2048)
		if err != nil {
			log.Printf("[ERROR] [run] LLM call failed for %s: %v", f.ID, err)
			continue
		}

		// Format and normalize the response for status evaluation
		clnRsp := strings.ToUpper(strings.ReplaceAll(rsp, " ", ""))
		var bugStatus string

		if strings.Contains(clnRsp, "OK") && len(clnRsp) <= 10 {
			bugStatus = "🍏 **OK** - No problem detected."
		} else {
			// Enforce markdown formatting block for the AI response if it's an anomaly description
			bugStatus = fmt.Sprintf("🔴 **Issue detected :**\n\n%s", strings.TrimSpace(rsp))
		}

		// Prepare the external dependencies list block for the report
		if ctxDep == "None (No internal project dependencies mapped)." {
			ctxDep = "*No external dependencies found.*"
		}

		// Systematically compile and write the comprehensive report item
		reportItem := fmt.Sprintf(`## 🔍 Audit : %s
- **File :** %s
- **Summary :** %s

### 📦 Externals Items & Dependencies :
%s

### 🛡️ Analyzing Result :
%s

---
`, f.ID, "`"+f.FilePath+"`", f.Summary, ctxDep, bugStatus)

		_, err = rep.WriteString(reportItem)
		if err != nil {
			log.Printf("[ERROR] [run] Failed writing to markdown report: %v", err)
		}

		// Update function state to AUDITED to prevent double-processing on recovery
		_ = db.UpdateStatus(f.ID, "AUDITED")
	}
}
