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

// Package status provides enumerations and utilities for managing audit status states.
//
// This package defines the different lifecycle states an audit item can be in during processing.
// Each status represents a distinct phase in the audit workflow, from initial discovery to final review.
// The status values are designed to be used as flags in state machines or process tracking systems.
package status

import "strings"

// TODO : review status based on bit operation
//   each bit 0 : not done, 1 done
//   bit 0: ast parse
//   bit 1: ast vendor
//   bit 2: analyzed (summaries)
//   bit 3: audited report 1
//   bit 4: audited report 2
//   bit 5: audited report 3
//   bit 6: audited report 4
//   bit 7: outdated

// Status represents the different states an audit item can be in during its lifecycle.
// It provides a standardized enumeration for tracking the progress and current state
// of each audit item through various phases of processing.
//
// This type is used throughout the auditing system to consistently represent audit item
// progression from initial discovery to final review.
type Status uint8

// Constants represent the predefined lifecycle states for audit items.
// Each constant corresponds to a distinct phase in the audit workflow:
// - None: Uninitialized or corrupt state
// - Pending: Newly discovered item awaiting processing
// - Progress: Item currently being processed
// - Failed: Processing failed with unrecoverable error
// - Retry: Item marked for reprocessing after transient failure
// - Extracted: Item successfully processed by AST decoupler
// - Analyzed: Semantic signatures and summaries generated
// - Audited: Deep-dive multi-pass reports complete
// - Reviewed: Human or architectural validation completed
// - Outdated: Item invalidated due to upstream changes
const (
	// None declares an uninitialized or corrupt state matrix baseline.
	// This represents the default zero value for Status and indicates that
	// an audit item is in an invalid or uninitialized state.
	//
	// It should typically not appear in normal operation as items should
	// always be in a defined state. However, it can be useful for initial
	// state detection or state validation in certain scenarios.
	//
	// Note: The zero value for Status is None (0), which means any unassigned
	// Status variable will have a value of None by default.
	//
	// Example usage:
	// status := Status(0) // Status{None}
	// status := Status(0) // Status{None}
	//
	// fmt.Println(status) // Output: none
	//
	// To safely handle unassigned Status variables, always initialize them
	// explicitly:
	// status := Status(None)
	// status := Status(0) // Same effect as above
	//
	// If you're using a Status variable without initialization, consider
	// using a safe default value to prevent accidental None state propagation:
	// status := Status(Pending) // Safer default
	//
	// In most cases, avoid using the zero value directly. Instead, initialize
	// Status variables explicitly with a valid state to maintain consistent
	// auditing state tracking:
	// status := Status(0) // Do not do this
	// status := Status(None) // Correct initialization
	None Status = iota

	// Pending identifies an item newly discovered but not yet subjected to any processing.
	//
	// When an audit item is first discovered, it enters the Pending state where it awaits
	// processing by the system. This is the initial state for all new audit items.
	//
	// Example:
	// // Create a new audit item
	// item := AuditItem{...}
	//
	// // Initialize the status
	// item.Status = Pending
	//
	// // Item is now pending processing
	// fmt.Println(item.Status) // Output: pending
	//
	// Once the item is processed, its status will transition to Progress (1).
	Pending

	// Progress flags an item currently undergoing active execution or network tasks.
	//
	// Items in this state are actively being processed, either through local execution
	// or network operations. This indicates that processing has started but is not yet complete.
	//
	// Example:
	// // Item is being processed
	// item.Status = Progress
	//
	// // Item is still processing, do not interrupt
	// fmt.Println(item.Status) // Output: progress
	//
	// During processing, the item may transition to Failed (2) if a fatal error occurs.
	Progress

	// Failed marks an execution loop termination caused by unrecoverable processing errors.
	//
	// When an audit item encounters a fatal error during processing that cannot be recovered,
	// it transitions to the Failed state. This typically requires manual intervention or
	// system-level remediation before reprocessing can occur.
	//
	// Example:
	// // Item failed processing
	// item.Status = Failed
	//
	// // Item cannot continue, requires manual review
	// fmt.Println(item.Status) // Output: failed
	//
	// Before reprocessing, you must resolve the underlying issue causing the failure.
	Failed

	// Retry identifies an item tagged for automated reprocessing after transient failure.
	//
	// Items in this state have experienced a temporary failure but are marked for automatic
	// retry processing. This allows for handling of intermittent issues without requiring
	// manual intervention, assuming the underlying cause is transient.
	//
	// Example:
	// // Item experienced transient failure
	// item.Status = Retry
	//
	// // Item will automatically retry
	// fmt.Println(item.Status) // Output: retry
	//
	// For transient errors, automatic retry is preferred over manual intervention.
	Retry

	// Extracted identifies an element successfully processed by the local AST syntax decoupler.
	//
	// After initial parsing and extraction, audit items reach this state where they have been
	// processed by the Abstract Syntax Tree (AST) decoupling component. This represents a
	// successful first phase of processing.
	//
	// Example:
	// // Item was successfully extracted
	// item.Status = Extracted
	//
	// // Item is ready for further analysis
	// fmt.Println(item.Status) // Output: extracted
	Extracted

	// Analyzed marks an item whose semantic signatures and summaries have been generated.
	//
	// Items in this state have undergone semantic analysis where their meaning and structure
	// have been extracted and summarized. This is typically the point where detailed analysis
	// begins, generating insights about code quality, security implications, etc.
	//
	// Example:
	// // Item was analyzed and generated summaries
	// item.Status = Analyzed
	//
	// // Item is ready for review
	// fmt.Println(item.Status) // Output: analyzed
	Analyzed

	// Audited registers the final state when all deep-dive multi-pass reports are complete.
	//
	// This represents the completion of comprehensive auditing processes including multiple
	// analysis passes and report generation. Items in this state have been thoroughly examined
	// and documented according to audit standards.
	//
	// Example:
	// // Item was thoroughly audited
	// item.Status = Audited
	//
	// // Item is ready for distribution
	// fmt.Println(item.Status) // Output: audited
	//
	// Once audited, an item remains in this state until it is manually reviewed.
	Audited

	// Reviewed flags items validated by human or architectural confirmation steps.
	//
	// After automated auditing is complete, items may enter the Reviewed state where they
	// undergo human review or architectural validation. This ensures that the automated
	// analysis results are correct and align with system requirements.
	//
	// Example:
	// // Item was manually reviewed
	// item.Status = Reviewed
	//
	// // Item is ready for distribution
	// fmt.Println(item.Status) // Output: reviewed
	Reviewed

	// Outdated marks downstream consumer elements invalidated due to upstream signature changes.
	//
	// When changes occur in upstream dependencies or source code, items that depend on these
	// elements may become outdated. This state indicates that they need to be reprocessed
	// to reflect the latest changes and maintain consistency.
	//
	// Example:
	// // Item's dependencies were updated
	// item.Status = Outdated
	//
	// // Item needs reprocessing
	// fmt.Println(item.Status) // Output: outdated
	Outdated
)

// Parse converts a string representation of a status into its corresponding Status enumeration value.
//
// This function performs case-insensitive matching and returns None for unrecognized input strings,
// making it robust against variations in input formatting.
//
// Example usage:
// status := Parse("pending")  // Returns Pending
// status := Parse("UNKNOWN")  // Returns None
//
// Note: This function is primarily used for human-readable string inputs from external sources,
// such as configuration files, user input, or logs. It allows for flexible string-based
// state tracking while maintaining strict enumeration integrity.
func Parse(s string) Status {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case Pending.String():
		return Pending
	case Progress.String():
		return Progress
	case Failed.String():
		return Failed
	case Retry.String():
		return Retry
	case Extracted.String():
		return Extracted
	case Analyzed.String():
		return Analyzed
	case Audited.String():
		return Audited
	case Reviewed.String():
		return Reviewed
	case Outdated.String():
		return Outdated
	default:
		return None
	}
}

// Get converts a uint8 value to a Status enumeration.
//
// This utility function provides safe numeric to enumeration conversion,
// returning None for invalid input values that do not match any defined status.
//
// Example usage:
// status := Get(2)  // Returns Progress (assuming 2 maps to Progress)
// status := Get(99) // Returns None (invalid value)
//
// Note: This function is useful when deserializing status information from numeric sources,
// such as database storage or JSON data. It helps prevent unexpected status values.
func Get(i uint8) Status {
	switch Status(i) {
	case Pending:
		return Pending
	case Progress:
		return Progress
	case Failed:
		return Failed
	case Retry:
		return Retry
	case Extracted:
		return Extracted
	case Analyzed:
		return Analyzed
	case Audited:
		return Audited
	case Reviewed:
		return Reviewed
	case Outdated:
		return Outdated
	default:
		return None
	}
}

// List returns a slice of all valid status string representations.
//
// This function provides a consistent list of string values that matches the order
// of the defined constants in this package. It allows for validation of status inputs
// against the full set of supported states.
//
// Example usage:
// statuses := List()  // Returns []string{"none", "pending", "progress", ...}
//
// Note: This function is useful for configuration validation, UI rendering, or
// reporting purposes where a human-readable status representation is needed.
func List() []string {
	return []string{
		Pending.String(),
		Progress.String(),
		Failed.String(),
		Retry.String(),
		Extracted.String(),
		Analyzed.String(),
		Audited.String(),
		Reviewed.String(),
		Outdated.String(),
	}
}
