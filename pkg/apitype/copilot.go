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

// Package apitype defines the API type enumeration and related utilities for
// identifying and converting between different AI/LLM API providers.
package apitype

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

// cplRequest constructs an HTTP request for the GitHub Copilot API.
//
// Copilot uses the same OpenAI-compatible API format, so this function delegates
// to oaiRequest and rebrands the error prefix from "openai:" to "copilot:" for
// clearer error attribution in logs and diagnostics.
//
// Parameters:
//   - context.Context: The context for the request lifecycle.
//   - *Config: API configuration containing hostname, model name, and options.
//   - []byte: The raw prompt bytes to be sent as the user message.
//
// Returns:
//   - *http.Request: The prepared request ready to be executed.
//   - uint16: An estimated token count for the request payload.
//   - error: Any error encountered during construction, with the prefix
//     rewritten to "copilot:" for clarity.
func cplRequest(ctx context.Context, cfg *Config, prt []byte) (*http.Request, uint16, error) {
	var (
		err error
		nbr uint16
		req *http.Request
	)

	// Delegate to the OpenAI-compatible request builder, since Copilot uses
	// the same API schema.
	if req, nbr, err = oaiRequest(ctx, cfg, prt); err != nil {
		// Rewrite the error prefix from "openai:" to "copilot:" so that
		// error messages clearly indicate the Copilot provider.
		return req, nbr, errors.New(strings.Replace(err.Error(), "openai: ", "copilot: ", -1))
	}

	return req, nbr, nil
}

// cplResponse processes the HTTP response from the GitHub Copilot API.
//
// Copilot uses the same OpenAI-compatible response format, so this function
// delegates to oaiResponse and rebrands the error prefix from "openai:" to
// "copilot:" for clearer error attribution in logs and diagnostics.
//
// Parameters:
//   - context.Context: The context for the response processing.
//   - *Config: API configuration (unused in this function, passed for consistency).
//   - *http.Response: The HTTP response from the Copilot API.
//
// Returns:
//   - uint16: An estimated token count for the response body.
//   - []byte: The raw output from the model.
//   - error: Any error encountered during processing, with the prefix
//     rewritten to "copilot:" for clarity.
func cplResponse(ctx context.Context, cfg *Config, rsp *http.Response) (uint16, []byte, error) {
	var (
		err error
		buf []byte
		nbr uint16
	)

	// Delegate to the OpenAI-compatible response parser, since Copilot uses
	// the same response schema.
	if nbr, buf, err = oaiResponse(ctx, cfg, rsp); err != nil {
		// Rewrite the error prefix from "openai:" to "copilot:" so that
		// error messages clearly indicate the Copilot provider.
		return nbr, buf, errors.New(strings.Replace(err.Error(), "openai: ", "copilot: ", -1))
	}

	return nbr, buf, nil
}
