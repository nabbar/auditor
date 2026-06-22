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

// Package pkg/local:
// ==================
//
// Package local provides functionality for managing local auditor data including prompts and reports.
//
// TODO: Add details information about the package
//
// Key components include:
//   - TODO: add information about keys components
//
// The import path for this package is : "github.com/nabbar/auditor/pkg/local" and the alias is "audloc"
//
// Interface `type Manager interface`
// TODO: add information about this interface
//
//   - Extend `io.Closer`:
//     TODO: add information about the extends, why is defined, how it works
//
//   - Function `Prompt(audtps.CodeType, audast.Lang) []byte`:
//     TODO: add information about this function, what it works, why request parameters, how is define the return, ...
//     Parameters:
//       - `type`: sentences explain the paramters type, what is awaiting, why is requested, how it is used...
//     TODO: the line before is a template for parameters, redefine the parameters list to be realist
//     Result:
//       - `type`: sentences explain the result type, what it is, how is build, ...
//     TODO: the line before is a template for result, redefine the result list to be realist
//
//   - Function `Report(string, audast.Lang) []byte`:
//     TODO: add information about this function, what it works, why request parameters, how is define the return, ...
//     Parameters:
//       - `type`: sentences explain the paramters type, what is awaiting, why is requested, how it is used...
//     TODO: the line before is a template for parameters, redefine the parameters list to be realist
//     Result:
//       - `type`: sentences explain the result type, what it is, how is build, ...
//     TODO: the line before is a template for result, redefine the result list to be realist
//
//   - Function `CatalogTo(io.Writer, string) error`:
//     TODO: add information about this function, what it works, why request parameters, how is define the return, ...
//     Parameters:
//       - `type`: sentences explain the paramters type, what is awaiting, why is requested, how it is used...
//     TODO: the line before is a template for parameters, redefine the parameters list to be realist
//     Result:
//       - `type`: sentences explain the result type, what it is, how is build, ...
//     TODO: the line before is a template for result, redefine the result list to be realist
//
//   - Function `CatalogFrom(io.Reader, string) (auddbm.Manager, error)`:
//     TODO: add information about this function, what it works, why request parameters, how is define the return, ...
//     Parameters:
//       - `type`: sentences explain the paramters type, what is awaiting, why is requested, how it is used...
//     TODO: the line before is a template for parameters, redefine the parameters list to be realist
//     Result:
//       - `type`: sentences explain the result type, what it is, how is build, ...
//     TODO: the line before is a template for result, redefine the result list to be realist
//
// Function `func New(p string, u auduxi.Manager, l func() auddbm.Manager, s func(auddbm.Manager)) (Manager, error)`:
// TODO: add information about this function, what it works, why request parameters, how is define the return, ...
//
// TODO: add information about the internal process of this function
//
// Parameters:
//   - `type`: sentences explain the paramters type, what is awaiting, why is requested, how it is used...
//     TODO: the line before is a template for parameters, redefine the parameters list to be realist
//
// Returns:
//   - Manager[D, T]: A thread-safe manager implementation capable of managing model instances of type T.
//     TODO: the line before is a template for result, redefine the result list to be realist
//
//

package engine
