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

package main

import (
	"fmt"
	"strings"
	"time"
)

// incPgr updates the global progress counter and redraws the progress bar.
// This function serves as a wrapper around the drawPgr function to increment the analysis counter
// and update the visual progress display. It is designed to be called after each completed analysis step
// to maintain real-time feedback about operation progress.
//
// The function implements thread safety through mutex locking to prevent race conditions when
// multiple goroutines attempt to update the progress simultaneously. This ensures that
// concurrent access to the global progress counter remains synchronized and prevents
// inconsistent state during parallel execution.
//
// Before incrementing the counter, the function checks whether progress bar visualization
// is enabled (useBar flag). If disabled, the function returns early without performing any operations,
// which allows for conditional execution based on user preferences or operational requirements.
//
// The function performs the following steps:
// 1. Checks if progress bar visualization is enabled
// 2. Acquires a mutex lock to ensure thread safety during counter increment
// 3. Increments the global analysis counter (cntAlz)
// 4. Calls drawPgr with updated values to refresh the progress display
//
// This function is typically called from within a loop or parallel processing context where
// each iteration represents a completed analysis task. It provides immediate visual feedback
// to users about their operation's progress and completion status.
func incPgr() {
	// Early return if progress bar visualization is disabled
	if !useBar {
		return
	}

	// Acquire mutex lock to ensure thread safety during concurrent access to the progress counter
	muxPgr.Lock()
	defer muxPgr.Unlock()

	// Increment the global analysis counter by one unit (typically representing one completed task)
	cntAlz++

	// Call drawPgr with updated counter values to refresh the visual progress display
	drawPgr(cntAlz, totFct, tmsStr)
}

// drawPgr renders a visual progress bar that displays the completion status of an analysis operation.
// This function takes three parameters: the current number of completed items, the total number of items,
// and the start time of the operation to calculate performance metrics.
//
// The progress bar consists of two primary components:
// 1. A graphical representation showing the percentage complete using Unicode block characters
// 2. Performance statistics including completion percentage, item counts, and functions per second
//
// The function performs several calculations to determine the visual representation:
// - It calculates the percentage complete based on current vs total items
// - It determines the length of the filled portion of the progress bar based on the current completion
// - It computes the elapsed time since the operation started
// - It calculates the rate of processing (functions per second) based on the elapsed time and completed items
//
// This function is designed to be called repeatedly during an operation to provide real-time feedback
// to users about the progress of their analysis. The output is written to standard output using fmt.Printf
// with a carriage return to overwrite the previous progress display, creating a dynamic updating effect.
//
// The function includes several safety checks:
// - If total items is zero or negative, the function returns early to prevent division by zero errors
// - If current items exceed total items (which shouldn't happen in normal operation), the bar will still render correctly
// - Performance calculations are only performed when elapsed time is greater than zero to avoid division by zero
//
// Example usage:
//
//	drawPgr(5, 10, startTime) // Displays: [█████████-------] 50% (5/10) | 2.50 fn/sec
//
// When the operation completes (current equals total), a newline is printed to ensure proper formatting
// and separation from subsequent output.
func drawPgr(current, total int, startTime time.Time) {
	// Early exit if total items is zero or negative to prevent division by zero errors
	if total <= 0 {
		return
	}

	// Calculate the percentage complete: current / total * 100
	pct := (current * 100) / total

	// Define the fixed length of the progress bar in characters
	const barLength = 30

	// Determine how many characters should be filled in the progress bar based on current completion
	fln := (current * barLength) / total

	// Create the visual representation of the progress bar:
	// - Fill with Unicode block characters (█) for completed portion
	// - Fill with dash characters (-) for remaining portion
	bar := strings.Repeat("█", fln) + strings.Repeat("-", barLength-fln)

	// Calculate the total elapsed time since the operation started
	elt := time.Since(startTime).Seconds()

	// Initialize speed calculation variable to zero
	spd := 0.0

	// Only calculate speed if elapsed time is greater than zero to avoid division by zero
	if elt > 0 {
		// Calculate functions per second: total completed items divided by elapsed seconds
		spd = float64(current) / elt
	}

	// Print the formatted progress bar with statistics using carriage return to overwrite previous output
	fmt.Printf("\r[%s] %d%% (%d/%d) | %.2f fn/sec ", bar, pct, current, total, spd)

	// If the operation has completed (current equals total), print a newline for proper formatting
	if current == total {
		fmt.Println()
	}
}
