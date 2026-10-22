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

// mod represents a renderer module that handles template execution and header extraction.
type mod struct {
	tpl *template.Template
}

// Parse executes the pre-compiled template with the provided audit entry data,
// returning the rendered output as a string or an error if template execution fails.
//
// Parameters:
//   - ent: The audit entry containing data to be rendered.
//
// Returns:
//   - A string representation of the rendered content.
//   - An error if template execution fails.
func (r *mod) Parse(ent audent.Entry) ([]byte, error) {
	var buf bytes.Buffer

	// Bind data directly into the pre-compiled layout structure.
	if err := r.tpl.Execute(&buf, ent); err != nil {
		return nil, fmt.Errorf("failed executing analytical formatting composition matrix for target %s: %w", ent.GetName(), err)
	}

	return buf.Bytes(), nil
}

// Head extracts all header entries from a given content string,
// identifying markdown-style headers and normalizing them into structured data.
//
// Parameters:
//   - content: The input text content potentially containing markdown headers.
//
// Returns:
//   - A slice of Header structs representing the extracted headers.
func (r *mod) Head(content string) []Header {
	var (
		res []Header
		lns = strings.Split(content, "\n")
	)

	for _, l := range lns {
		trm := strings.TrimSpace(l)
		if !strings.HasPrefix(trm, "#") {
			continue
		}

		// Count the consecutive occurrence of hashes to compute the heading priority scope level bounds.
		lvl := 0
		for lvl < len(trm) && trm[lvl] == '#' {
			lvl++
		}

		// Ensure there is a mandatory spacing block separating the hashes from the literal string token.
		if lvl >= len(trm) || trm[lvl] != ' ' {
			continue
		}

		ttl := strings.TrimSpace(trm[lvl:])
		if ttl == "" {
			continue
		}

		// Normalize the literal header string layout into a valid HTML/Markdown compatible anchor identifier slug.
		anc := strings.ToLower(ttl)
		anc = strings.ReplaceAll(anc, " ", "-")
		anc = strings.ReplaceAll(anc, ":", "")
		anc = strings.ReplaceAll(anc, "(", "")
		anc = strings.ReplaceAll(anc, ")", "")
		anc = strings.ReplaceAll(anc, "*", "")
		anc = strings.ReplaceAll(anc, "`", "")
		anc = strings.ReplaceAll(anc, ".", "")
		anc = strings.ReplaceAll(anc, ",", "")

		res = append(res, Header{
			Title:  ttl,
			Anchor: anc,
			Level:  lvl,
		})
	}

	return res
}
