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
	"context"
	"io"

	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	loglvl "github.com/nabbar/golib/logger/level"
	libsem "github.com/nabbar/golib/semaphore"
	semtps "github.com/nabbar/golib/semaphore/types"
)

// Manager defines the formal architectural contract for managing unified asynchronous
// progress tracking vectors and conditional cascade application runtime terminations.
// It serves as the primary interface for developers to interact with the logging
// and progress-bar subsystems concurrently.
//
// The Manager interface extends io.Closer and context.Context, providing a comprehensive
// set of methods for structured logging, progress tracking, and graceful resource cleanup.
type Manager interface {
	// Closer satisfies the io.Closer interface to ensure all resources, such as log
	// streams and concurrency semaphores, are gracefully released or flushed upon shutdown.
	io.Closer

	// Context returns the context associated with this manager.
	context.Context

	// Info dispatches a formatted message string at the informational severity level.
	// This should be used for general operational telemetry that does not require
	// immediate attention but is useful for audit trails and monitoring purposes.
	Info(format string, args ...interface{})

	// Warning dispatches a formatted message string at the warning severity level.
	// Use this to log conditions that are non-critical but might indicate potential
	// issues or deviations from the expected happy path. These messages should be
	// investigated but do not necessarily require immediate action.
	Warning(format string, args ...interface{})

	// Error processes standard operational exceptions. If strict error enforcement
	// is enabled in the configuration, this method elevates the event to a fatal
	// state, resulting in an immediate application shutdown.
	Error(format string, args ...interface{})

	// ErrorStack chains multiple historical operational errors into a single structured
	// log sweep. This is useful for aggregating root cause analysis data. If strict
	// error enforcement is enabled, it automatically triggers a FatalStack event.
	ErrorStack(message string, errs ...error)

	// Fatal registers a critical panic-level event and triggers immediate program
	// execution termination. Use this only for unrecoverable errors where the
	// integrity of the application can no longer be guaranteed.
	Fatal(format string, args ...interface{})

	// FatalStack appends an array of execution tracking errors into an immediate
	// terminal log vector. This captures the state of multiple errors simultaneously
	// before freezing the process execution.
	FatalStack(message string, errs ...error)

	// NewBar initializes a standalone visualization tracker boundary mapping multi-threaded
	// progress tokens. The returned SemBar allows for granular control over the progress
	// representation of a specific asynchronous task.
	NewBar(taskName string, totalItems int) semtps.SemBar

	Logger() liblog.Logger
}

// New initializes and returns an instantiated concrete Manager implementation.
// It orchestrates the underlying logger and semaphore structures based on the provided parameters.
//
// Parameters:
//   - ctx: The context.Context which governs the lifecycle and cancellation of background telemetry loops.
//   - cfg: A logcfg.Options structure used to configure log rotation, storage targets, and formatting styles.
//   - lvl: The minimum loglvl.Level threshold; messages below this severity will be filtered out.
//   - enableProgress: A boolean flag indicating whether console-based progress visualization rendering should be active.
//   - concurrentWorkers: An integer defining the baseline size bound allocated to the thread semaphore,
//     limiting the number of concurrent operations. If set to <= 0, it defaults to 1.
//   - exitOnError: A boolean flag which, when true, instructs the Manager to treat every reported
//     Error as an unrecoverable Fatal system halt.
//
// Returns an error if the logger configuration fails or if the structural instantiation is invalid.
func New(ctx context.Context, cfg logcfg.Options, lvl loglvl.Level, enableProgress bool, concurrentWorkers int) (Manager, error) {
	l := liblog.New(ctx)
	if e := l.SetOptions(&cfg); e != nil {
		return nil, e
	}

	l.SetLevel(lvl)

	o := &mdl{
		log: l,
		sem: libsem.New(ctx, concurrentWorkers, enableProgress),
	}

	return o, nil
}
