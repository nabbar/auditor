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

// Package render:
// ==================
//
// Package render provides functionality for rendering audit entries into structured formats,
// including extracting headers from markdown content and processing templates.
//
// This package enables the transformation of audit data into various output formats by utilizing
// Go's text/template engine. It supports header extraction from markdown-style content and
// provides a flexible interface for custom rendering logic.
//
// The package is designed to be used in conjunction with audit entry data structures to produce
// formatted reports, documentation, or other textual outputs based on predefined templates.
//
// Key components include:
//   - Template-based rendering engine using Go's text/template package
//   - Header extraction functionality for markdown-style content
//   - Flexible interface for custom rendering implementations
//
// The import path for this package is: "github.com/nabbar/auditor/pkg/render" and the alias is "audrdr"
//
// Interface `type Render interface`:
//   - Defines the contract for rendering audit entries into formatted output.
//   - Provides methods for template-based parsing and header extraction.
//
//   - Function `Parse(ent audent.Entry) (string, error)`:
//     This function renders an audit entry using a pre-compiled template.
//     It takes an audit entry as input and returns the rendered string output or an error if rendering fails.
//     Parameters:
//       - `audent.Entry`: The audit entry containing data to be rendered.
//     Result:
//       - `string`: The rendered content based on the provided template and audit entry data.
//       - `error`: An error if template execution fails or if there are issues with rendering.
//
//   - Function `Head(str string) []Header`:
//     This function extracts all markdown-style headers from a given string content.
//     It parses the input text to identify header entries and returns them as structured data.
//     Parameters:
//       - `string`: The input text content potentially containing markdown headers.
//     Result:
//       - `[]Header`: A slice of Header structs representing the extracted headers.
//
// Struct `type Header struct`:
//   Represents a structured header entry extracted from markdown content.
//   Contains the title, normalized anchor identifier, and heading level.
//   - Field `Title string`: The original title text of the header.
//   - Field `Anchor string`: A normalized version of the title suitable for use as an HTML/Markdown anchor identifier.
//   - Field `Level int`: The heading level (1-6) indicating the hierarchy of the header in the document structure.
//
// Function `func New(tpl string) (Render, error)`:
//   Initializes and returns a new Render instance based on the provided template string.
//   This function compiles the template using Go's text/template engine and prepares it for rendering.
//
//   The function validates the template string, compiles it with Go's text/template engine,
//   and returns a structured renderer that can be used to process audit entries.
//
//   - Parameters:
//     - `string`: A string containing the Go template used for rendering audit entries.
//
//   - Returns:
//     - `Render`: An implementation of the Render interface capable of processing audit entries.
//     - `error`: An error if the template is invalid, empty, or compilation fails.
//

package render
