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

// Package pkg/uxi:
// ================
// this package provides a unified interface and structural monitoring orchestrator
// that combines thread-safe structured diagnostics logging with concurrent worker progress tracking.
// It serves as a facade layer to simplify interactions between logging subsystems and
// semaphore-based concurrency control mechanisms.
//
// The package offers a Manager interface that encapsulates logging capabilities, progress
// visualization, and concurrent execution management. It supports both standard logging
// operations and advanced progress tracking for multithreaded tasks.
//
// It includes:
// - Thread-safe structured logging with multiple severity levels
// - Semaphore-based concurrency control for limiting concurrent operations
// - Progress visualization for asynchronous tasks
// - Graceful resource cleanup via io.Closer interface
// - Context-aware operations through context.Context integration
//
// `type Manager interface`:
//   - Extends `io.Closer` : Provides graceful resource cleanup and shutdown capabilities
//   - Extends `context.Context` : Enables cancellation and timeout handling for background operations
//   - `Info(format string, args ...interface{})` : Logs informational messages for operational telemetry
//   - `Warning(format string, args ...interface{})` : Logs warning conditions that are non-critical but may indicate potential issues
//   - `Error(format string, args ...interface{})` : Logs error conditions and handles fatal termination based on configuration
//   - `ErrorStack(message string, errs ...error)` : Aggregates multiple errors into a single structured log entry
//   - `Fatal(format string, args ...interface{})` : Logs critical errors that trigger immediate program termination
//   - `FatalStack(message string, errs ...error)` : Logs multiple fatal errors with complete stack trace information
//   - `NewBar(taskName string, totalItems int) semtps.SemBar` : Creates progress tracking visualization for asynchronous tasks
//
// `func New(context.Context, logcfg.Options, loglvl.Level, bool, int, bool) (Manager, error)`:
//   - Creates and initializes a Manager instance with specified configuration parameters
//   - context.Context: Governs the lifecycle and cancellation of background telemetry loops
//   - logcfg.Options: Configures log rotation, storage targets, and formatting styles
//   - loglvl.Level: Sets minimum log level threshold for message filtering
//   - enableProgress: Controls whether console-based progress visualization rendering is active
//   - concurrentWorkers: Defines baseline size of thread semaphore limiting concurrent operations
//   - exitOnError: Determines if reported errors should trigger immediate fatal termination
//   - Returns Manager interface and error if instantiation fails
//

package uxi
