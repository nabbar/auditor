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

// Package render provides functionality for rendering audit entries into structured formats,
// including extracting headers from markdown content and processing templates.
//
// This package enables the transformation of audit data into various output formats by utilizing
// Go's text/template engine. It supports header extraction from markdown-style content and
// provides a flexible interface for custom rendering logic.
package render

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	audent "github.com/nabbar/auditor/pkg/data/entry"
)

// Header represents a structured header entry extracted from markdown content.
// It includes the title, normalized anchor identifier, and heading level.
type Header struct {
	Title  string
	Anchor string
	Level  int
}

// Render defines the interface for rendering audit entries into formatted output.
// The interface supports parsing audit data with templates and extracting headers from text content.
type Render interface {
	// Parse renders an audit entry using a pre-compiled template.
	// It returns the rendered string output or an error if rendering fails.
	Parse(ent audent.Entry) ([]byte, error)

	// Head extracts all markdown-style headers from a given string content.
	// It returns a slice of Header structs representing the extracted headers.
	Head(str string) []Header
}

// New initializes and returns a new Render instance based on the provided template string.
//
// Parameters:
//   - tpl: A string containing the Go template used for rendering audit entries.
//
// Returns:
//   - A Render interface implementation.
//   - An error if the template is invalid or empty.
//
// The function validates the template string, compiles it using Go's text/template engine,
// and returns a structured renderer that can be used to process audit entries.
func New(tpl []byte) (Render, error) {
	tpl = bytes.TrimSpace(tpl)

	if len(tpl) < 10 {
		return nil, fmt.Errorf("template is empty")
	}

	// Define the common formatting operations allowed within the reporting templates scope blocks.
	// These functions provide utility for transforming text within templates.
	fm := template.FuncMap{
		"codeQuote": func() string { return "`" },
		"toLower":   strings.ToLower,
		"toUpper":   strings.ToUpper,
	}

	// Initialize the abstract compiler matrix and append structural function mappings.
	// This step compiles the template into a usable structure for rendering.
	t, err := template.New("report_transparent_matrix").Funcs(fm).Parse(string(tpl))
	if err != nil {
		return nil, fmt.Errorf("failed compiling text layout matrix validation schema: %w", err)
	}

	return &mod{
		tpl: t,
	}, nil
}
