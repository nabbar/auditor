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

// Package main provides the main entry point for the Auditor application.
//
// The Auditor application is a sophisticated code analysis tool that performs multi-stage
// analysis of source code using Large Language Models (LLMs) and SQLite databases.
// It implements a comprehensive workflow consisting of:
// 1. Project scanning to catalog all functions and their dependencies
// 2. Concurrent dependency tree analysis using multiple worker threads
// 3. Logical bug audit passes on extracted functions to identify potential issues
//
// The application supports configuration through command-line flags, enables silent mode
// via log files, and provides progress reporting through visual indicators when needed.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/nabbar/auditor/database"
	"github.com/nabbar/auditor/fileutils"
	"github.com/nabbar/auditor/llm"
	"github.com/nabbar/auditor/models"
)

// Global configuration variables defined via command-line flags.
var (
	// flgMod specifies the LLM model name to be used for code analysis
	flgMod string

	// flgDBN is the path to the SQLite database file used for storing analysis results
	flgDBN string

	// flgRep is the path to the markdown report file where anomalies will be documented
	flgRep string

	// flgLLM is the Ollama API URL endpoint for interacting with LLM services
	flgLLM string

	// flgPs1 defines the number of logical bug audit passes to perform (0 to disable)
	flgPs1 int

	// flgErr determines whether execution should stop if a function fails definitively after retries
	flgErr bool

	// flgLog specifies the path to a log file for writing application logs
	flgLog string

	// flgWrk sets the number of concurrent workers for dependency tree analysis
	flgWrk int

	// flgDpt controls the maximum recursion depth allowed during tree analysis
	flgDpt int

	// flgRst determines whether to reset audit step completely
	flgRst bool

	// totFct stores the total number of functions identified in the project
	totFct int

	// cntAlz tracks the count of analyzed functions for progress reporting
	cntAlz int

	// tmsStr captures the start time of the analysis process for performance metrics
	tmsStr time.Time

	// useBar determines whether to display progress bars during execution
	useBar bool

	// muxPgr is a mutex protecting concurrent access to the progress counter
	muxPgr sync.Mutex
)

// init initializes command-line flags for configuration of the Auditor application.
//
// This function sets up all available command-line arguments that can be used to configure
// the behavior of the auditor tool. These flags control aspects such as LLM model selection,
// database file paths, report generation, API endpoints, and execution parameters.
// Each flag is configured with appropriate default values and descriptive help text.
func init() {
	// Configure the LLM model flag with a default value for code analysis
	flag.StringVar(&flgMod, "model", "qwen3-code-expert", "LLM model name")

	// Specify the SQLite database file path for storing analysis results
	flag.StringVar(&flgDBN, "db", "audit_vault.db", "SQLite database file")

	// Define the markdown report file path for documenting identified anomalies
	flag.StringVar(&flgRep, "report", "audit_bugs_report.md", "Markdown report file")

	// Set the Ollama API URL endpoint for LLM interaction
	flag.StringVar(&flgLLM, "url", "http://localhost:11434/api/generate", "Ollama API URL")

	// Configure the number of logical bug audit passes (0 to disable)
	flag.IntVar(&flgPs1, "audit-passes", 1, "Number of logical bug audit passes (0 to disable)")

	// Determine whether execution should halt on function analysis failure
	flag.BoolVar(&flgErr, "suspend-on-error", false, "Stop execution if a function fails definitively after retries")

	// Specify the path to a log file for writing application logs
	flag.StringVar(&flgLog, "log-file", "", "Path to log file for writing logs")

	// Reset audit flag to completely reset audit step while preserving step 1 results
	flag.BoolVar(&flgRst, "reset-audit", false, "Reset audit step completely (make complete new report but don't rescan all function)")

	// Define the number of concurrent workers for dependency tree analysis
	flag.IntVar(&flgWrk, "workers", 4, "Number of concurrent workers for tree analysis")

	// Set the maximum recursion depth allowed during dependency tree analysis
	flag.IntVar(&flgDpt, "max-depth", 10, "Maximum recursion depth allowed for tree analysis")
}

// main is the entry point of the Auditor application.
//
// The main function orchestrates the entire code analysis workflow through a multi-stage process:
// 1. Parses command-line flags to configure application behavior
// 2. Initializes logging and database connections for persistent storage
// 3. Scans the project to catalog all functions and their dependencies
// 4. Performs concurrent dependency tree analysis using multiple workers in parallel
// 5. Executes logical bug audit passes on extracted functions to identify potential issues
// 6. Generates a markdown report of identified anomalies and bugs
//
// The application implements a sophisticated approach where:
// - Functions are scanned and cataloged for comprehensive code inventory
// - Dependency trees are analyzed concurrently using worker pools for efficient processing
// - Logical audits are performed on extracted functions to detect bugs and potential issues
// - Results are stored in database and reported in markdown format for human-readable output
//
// The workflow includes two primary passes:
// PASS 1: Concurrent dependency tree analysis with configurable workers and depth limits
// PASS 2: Logical audit passes that can be repeated multiple times for thorough bug detection
func main() {
	// Parse command-line flags to configure application behavior
	// This step reads all configured flags from the command line arguments
	flag.Parse()

	// Configure logging with date and time stamps for better traceability
	// Logging includes timestamps to help with debugging and monitoring execution
	log.SetFlags(log.Ldate | log.Ltime)

	// Configure logging output based on whether a log file is specified.
	// If a log file path is provided, all logs will be written to that file instead of stdout.
	// This enables silent mode execution where progress is shown through progress bars.
	// When using silent mode, the application operates without console output but still
	// writes detailed logs to the specified file for later review.
	if flgLog != "" {
		frt, err := os.OpenRoot(filepath.Dir(flgLog))
		if err != nil {
			log.Fatalf("[ERROR] [main] Unable to create log file: %v", err)
		}
		defer func() {
			_ = frt.Close()
		}()

		fLog, err := frt.OpenFile(filepath.Base(flgLog), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("[ERROR] [main] Unable to create log file: %v", err)
		}
		defer func() {
			_ = fLog.Close()
		}()

		log.SetOutput(fLog)
		useBar = true
		fmt.Printf("📝 Silent mode enabled. Concurrent tree analysis in progress (%d workers). Logs : %s\n\n", flgWrk, flgLog)
	} else {
		log.SetOutput(os.Stdout)
	}

	// Initialize the database manager for storing analysis results and function metadata
	// This establishes a connection to the SQLite database that will persist all analysis data
	db, err := database.NewDBManager(flgDBN)
	if err != nil {
		log.Fatalf("[ERROR] [main] Fatal SQLite error: %v", err)
	}
	defer db.Close()

	// Create an LLM client for interaction with Ollama API services
	// This client handles communication with the Large Language Model for code analysis tasks
	ai := llm.NewClient(flgLLM, flgMod)

	// Scan the project to identify all functions and save them to the catalog.
	// This step creates a comprehensive inventory of all functions within the codebase,
	// including their dependencies and metadata for later analysis.
	if useBar {
		fmt.Print("🔍 Project scan... ")
	}
	cat, err := fileutils.ScanProject()
	if err != nil {
		log.Fatalf("[ERROR] [main] Unable to scan project: %v", err)
	}
	if useBar {
		fmt.Printf("OK (%d functions detected)\n\n", len(cat))
	}

	// Save the catalog entries to the database and sync them into the pending queue.
	// This ensures that all discovered functions are persisted for further analysis,
	// making them available for dependency tree analysis in subsequent passes.
	_ = db.SaveCatalog(cat)
	_ = db.SyncCatalogToPendingQueue()

	// CRASH RECOVERY: Reset any stuck 'IN_PROGRESS' functions from a previous interrupted run back to 'PENDING'
	if err = db.ResetQueueStatus("IN_PROGRESS", "PENDING"); err != nil {
		msg := fmt.Sprintf("[WARNING] [main] Failed to recover stuck IN_PROGRESS functions: %v", err)
		if useBar {
			fmt.Println(msg)
		} else {
			log.Printf("%s", msg)
		}
	}

	var fctSts []models.Function

	// Retrieve all functions that are currently pending analysis.
	// These functions have been identified but not yet processed for dependency analysis.
	if fctSts, err = db.GetFunctionsByStatus("PENDING"); err != nil {
		msg := fmt.Sprintf("[FATAL] [main] Failed to list PENDING functions: %v", err)
		if useBar {
			fmt.Println(msg)
			os.Exit(1)
		} else {
			log.Fatalf("%s", msg)
		}
	}

	totFct = len(fctSts)
	tmsStr = time.Now()

	// Display information about the first pass (dependency tree analysis) in verbose mode
	if useBar {
		fmt.Printf("🌲 PASS 1 - Concurrent Dependency Tree Analysis (%d workers, max depth: %d)...\n", flgWrk, flgDpt)
	} else {
		log.Println("[INFO] [main] Starting tree mapping...")
	}

	// Create a channel for job distribution among workers
	// This channel facilitates the communication between the main process and worker goroutines
	var (
		sncWtg sync.WaitGroup

		chnJbs = make(chan models.Function, totFct)
	)

	// Start the pool of workers to process functions concurrently.
	// Each worker processes functions from the channel until it's closed.
	// The worker pool approach enables parallel processing of function dependencies,
	// significantly improving performance for large codebases.
	for w := 1; w <= flgWrk; w++ {
		sncWtg.Add(1)
		go func(workerID int) {
			defer sncWtg.Done()
			for f := range chnJbs {
				// Before processing, verify its status as it might have been resolved
				// by a transitive dependency during another node's analysis.
				// This check ensures no redundant work is performed on already processed functions.
				stsCur := db.GetStatus(f.ID)
				switch stsCur {
				case "PENDING":
					lazyTreeAnalyze(db, ai, f, 1)
				case "EXTRACTED":
					// Already processed by a subtree; still update the global progress bar.
					incPgr()
				}
			}
		}(w)
	}

	// Fill the job channel with functions to analyze.
	// This distributes all pending functions among the available workers for concurrent processing.
	for _, f := range fctSts {
		chnJbs <- f
	}
	close(chnJbs) // Signal that no more root tasks are available.
	sncWtg.Wait() // Wait for all analysis jobs to complete.

	// Check if need to reset status for audit
	if flgRst {
		if err = db.ResetQueueStatus("AUDITED", "EXTRACTED"); err != nil {
			msg := fmt.Sprintf("[WARNING] [main] Error on updating status: %v", err)
			if useBar {
				fmt.Printf("%s\n", msg)
			} else {
				log.Printf("%s", msg)
			}
		}
	}

	// Execute logical audit passes if specified (default is 1 pass)
	if flgPs1 > 0 {
		for i := 1; i <= flgPs1; i++ {
			if useBar {
				fmt.Printf("\n🛡️ PASS 2 - Logical Audit (Iteration %d/%d)\n", i, flgPs1)
			} else {
				log.Printf("[INFO] [main] >>> Launching PASS 2 (Audit) - Iteration %d/%d", i, flgPs1)
			}
			runStep2(db, ai, useBar)
			if i < flgPs1 {
				// Reset queue status to prepare for next audit iteration
				_ = db.ResetQueueStatus("AUDITED", "EXTRACTED")
			}
		}
	}

	if useBar {
		fmt.Println("\n🎉 All done! Check your anomaly report.")
	}
}
