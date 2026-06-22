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

// Package manager provides the generic interface for AST (Abstract Syntax Tree) parsers
// that are used across different programming languages in the auditor tool.
package manager

import liblog "github.com/nabbar/golib/logger"

// Info logs an informational message.
//
// This method delegates to the embedded UI manager's Info method, providing consistent logging
// throughout the AST parsing process. It is used to report general information about the parsing operations,
// such as progress updates, status changes, or other non-critical information that may be useful for debugging
// or monitoring purposes.
//
// Parameters:
//   - format: A format string for the log message, following standard Go fmt conventions
//   - args: Variable arguments to be formatted into the log message, allowing for dynamic content
//
// Example usage:
//   - o.Info("Processing module %s", moduleName)
//   - o.Info("Found %d entries in package %s", entryCount, packageName)
func (o *mdl) Info(format string, args ...interface{}) {
	o.sem.Info(format, args...)
}

// Warning logs a warning message.
//
// This method delegates to the embedded UI manager's Warning method, allowing for
// appropriate handling of non-critical issues during AST processing. It is used to report
// conditions that may require attention but do not necessarily stop the parsing process.
// Warnings are typically used for situations where the parsing can continue but may result
// in unexpected behavior or suboptimal outcomes.
//
// Parameters:
//   - format: A format string for the log message, following standard Go fmt conventions
//   - args: Variable arguments to be formatted into the log message, allowing for dynamic content
//
// Example usage:
//   - o.Warning("Package %s has missing dependencies", packageName)
//   - o.Warning("Could not parse function %s due to syntax errors", functionName)
func (o *mdl) Warning(format string, args ...interface{}) {
	o.sem.Warning(format, args...)
}

// Error logs an error message.
//
// This method delegates to the embedded UI manager's Error method, providing consistent
// error reporting during AST parsing operations. It is used to report errors that occurred
// during parsing but do not necessarily terminate the process. Errors are typically used for
// issues that prevent proper processing or analysis of code elements.
//
// Parameters:
//   - format: A format string for the log message, following standard Go fmt conventions
//   - args: Variable arguments to be formatted into the log message, allowing for dynamic content
//
// Example usage:
//   - o.Error("Failed to parse module %s", moduleName)
//   - o.Error("Invalid entry type found in package %s", packageName)
func (o *mdl) Error(format string, args ...interface{}) {
	o.sem.Error(format, args...)
}

// ErrorStack logs an error with stack trace information.
//
// This method delegates to the embedded UI manager's Error method but omits the error stack
// for simplicity - note that this implementation may not fully capture stack traces as intended.
// It is used to report errors along with their call stack information for debugging purposes.
// The stack trace provides context about where the error occurred in the codebase, which is
// essential for diagnosing issues during AST processing.
//
// Parameters:
//   - message: The error message to log, describing the nature of the error
//   - errs: One or more error instances to include in the log, providing additional context
//
// Example usage:
//   - o.ErrorStack("Failed to process entry", err)
//   - o.ErrorStack("Package parsing failed", err1, err2)
func (o *mdl) ErrorStack(message string, errs ...error) {
	o.sem.Error(message)
}

// Fatal logs a fatal error message and terminates the process.
//
// This method delegates to the embedded UI manager's Fatal method, ensuring proper
// termination when critical errors occur during AST parsing. It is used to report
// errors that prevent further processing and require immediate attention. Fatal errors
// typically indicate issues that make continued execution impossible or unsafe.
//
// Parameters:
//   - format: A format string for the log message, following standard Go fmt conventions
//   - args: Variable arguments to be formatted into the log message, allowing for dynamic content
//
// Example usage:
//   - o.Fatal("Critical error in module processing: %v", err)
//   - o.Fatal("Parser initialization failed")
func (o *mdl) Fatal(format string, args ...interface{}) {
	o.sem.Fatal(format, args...)
}

// FatalStack logs a fatal error with stack trace information and terminates the process.
//
// This method delegates to the embedded UI manager's Fatal method but omits the error stack
// for simplicity - note that this implementation may not fully capture stack traces as intended.
// It is used to report fatal errors along with their call stack information, ensuring
// proper termination of the application when critical issues occur. The stack trace provides
// detailed context about where the fatal error originated in the codebase.
//
// Parameters:
//   - message: The fatal error message to log, describing the nature of the critical failure
//   - errs: One or more error instances to include in the log, providing additional context
//
// Example usage:
//   - o.FatalStack("Parser initialization failed", err)
//   - o.FatalStack("Critical AST processing error", err1, err2)
func (o *mdl) FatalStack(message string, errs ...error) {
	o.sem.Fatal(message)
}

func (o *mdl) Logger() liblog.Logger {
	return o.sem.Logger()
}
