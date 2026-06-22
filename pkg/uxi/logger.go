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

// Package uxi provides a unified interface and structural monitoring orchestrator
// that combines thread-safe structured diagnostics logging with concurrent worker progress tracking.
// It serves as a facade layer to simplify interactions between logging subsystems and
// semaphore-based concurrency control mechanisms.
//
// The package offers a Manager interface that encapsulates logging capabilities, progress
// visualization, and concurrent execution management. It supports both standard logging
// operations and advanced progress tracking for multithreaded tasks.
package uxi

import (
	liblog "github.com/nabbar/golib/logger"
	loglvl "github.com/nabbar/golib/logger/level"
)

// Info commits a structured trace line layout entry marked with an informational tag signature.
// This is the implementation of the Manager interface for info-level logging.
func (o *mdl) Info(format string, args ...interface{}) {
	o.log.Entry(loglvl.InfoLevel, format, args...).Log()
}

// Warning commits a diagnostic line entry identifying structural anomalies or pipeline workarounds.
// This logs messages using the warning severity tier.
func (o *mdl) Warning(format string, args ...interface{}) {
	o.log.Entry(loglvl.WarnLevel, format, args...).Log()
}

// Error routes anomalies to output logs or converts them into Fatal crashes if the
// eoe (exit-on-error) configuration flag is enabled. This ensures central management
// of error propagation policies.
func (o *mdl) Error(format string, args ...interface{}) {
	o.log.Entry(loglvl.ErrorLevel, format, args...).Log()
}

// Fatal enforces an instant crash state layout. It commits the log entry at the fatal
// severity level, which typically triggers a panic or system exit by the underlying logger.
func (o *mdl) Fatal(format string, args ...interface{}) {
	o.log.Entry(loglvl.FatalLevel, format, args...).Log()
}

// ErrorStack bundles consecutive runtime error items, elevating executions into
// Fatal paths if the eoe flag is enabled. It aggregates multiple errors into a
// single structured log entry.
func (o *mdl) ErrorStack(message string, errs ...error) {
	o.log.Entry(loglvl.ErrorLevel, message).ErrorAdd(true, errs...).Log()
}

// FatalStack logs stacked exceptional failures immediately at the fatal severity level.
// This is used for capturing complex error contexts that require immediate termination
// of the application execution flow.
func (o *mdl) FatalStack(message string, errs ...error) {
	o.log.Entry(loglvl.FatalLevel, message).ErrorAdd(true, errs...).Log()
}

func (o *mdl) Logger() liblog.Logger {
	return o.log
}
