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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// olaReq represents the JSON payload sent to the Ollama API to initiate a
// generation request. It encapsulates the model identifier, the prompt text,
// streaming preference, output format constraints, and runtime generation options.
type olaReq struct {
	// Model declares the textual identifier string tag of the target execution model.
	Model string `json:"model"`

	// Prompt stores the complete aggregated instructions, role assignments, and source blocks.
	Prompt string `json:"prompt"`

	// Stream enforces immediate atomic batch rendering responses instead of progressive chunking.
	// Set to false to wait for the complete generation before receiving the response.
	Stream bool `json:"stream"`

	// Format targets structured constraints, forcing the LLM engine to respond in valid raw JSON.
	// Used primarily during the Scan pass to ensure parseable output.
	Format string `json:"format,omitempty"`

	// Options injects localized runtime generation parameter controls directly into the LLM context.
	// Used to tune factors like temperature (creativity) and max tokens.
	Options map[string]interface{} `json:"options,omitempty"`
}

// olaRsp represents the JSON response returned by the Ollama API after a
// generation request completes. It contains the generated text, metadata about
// the model used, and timing information.
type olaRsp struct {
	// Model defines the precise model execution target utilized during the generation process.
	Model string `json:"model"`

	// CreatedAt registers the synchronized network timestamp clock tracking generation completion.
	CreatedAt time.Time `json:"created_at"`

	// Response preserves the raw text or stringified JSON layout generated within the engine context.
	Response string `json:"response"`

	// Done confirms if the token evaluation sequence has reached its deterministic terminal state.
	Done bool `json:"done"`

	// TotalDuration tracks the elapsed nanosecond execution runtime within the upstream engine.
	TotalDuration time.Duration `json:"total_duration,omitempty"`
}

// olaRequest constructs and prepares the HTTP POST request for the Ollama API.
// It serializes the prompt into a JSON payload, applies rate limiting, and
// configures the request headers and timeout.
//
// Parameters:
//   - context.Context: the context for cancellation and timeout control.
//   - *Config: the analysis configuration containing host, model, and options.
//   - []byte: the raw byte slice containing the prompt text.
//
// Returns:
//   - *http.Request: the prepared HTTP request.
//   - uint16: the estimated token count of the request.
//   - error: any error encountered during preparation.
//
// The method performs the following steps:
//  1. Validates the configuration and prompt.
//  2. Normalizes the host URL and constructs the endpoint path.
//  3. Serializes the request payload to JSON.
//  4. Estimates the token count of the serialized payload.
//  5. Creates the HTTP POST request with JSON headers.
func olaRequest(ctx context.Context, cfg *Config, prt []byte) (*http.Request, uint16, error) {
	// Validate that the configuration is not nil and contains required fields
	if cfg == nil || len(cfg.Hostname) < 1 || len(cfg.Model) < 1 {
		return nil, 0, errors.New("ollama: given config is nil")
	}

	// Validate that the prompt is not empty
	if len(prt) < 1 {
		return nil, 0, errors.New("ollama: given prompt is nil")
	}

	var (
		err error
		buf []byte
		nbr uint16
		pth string
		req *http.Request

		hst = cfg.Hostname
		mod = olaReq{
			Model:   cfg.Model,
			Prompt:  string(prt),
			Stream:  false,
			Format:  "json",
			Options: cfg.Options,
		}
	)

	// Normalize the URL with Path
	if pth, err = url.JoinPath(hst, pthOllama); err != nil {
		return nil, 0, fmt.Errorf("ollama: normalize url path error: %w", err)
	}

	// Serialize the request payload to JSON
	if buf, err = json.Marshal(mod); err != nil {
		return nil, 0, fmt.Errorf("ollama: llm marshall error: %w", err)
	}

	// Estimate the token count of the serialized JSON payload
	nbr = EstimateToken(buf)

	// Create the HTTP POST request with the serialized payload
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, pth, bytes.NewBuffer(buf)); err != nil {
		return nil, 0, fmt.Errorf("ollama: llm http request error: %w", err)
	}

	// Set JSON content type and accept headers
	req.Header.Set(hrdContent, hrdJson)
	req.Header.Set(hrdAccept, hrdJson)

	return req, nbr, nil
}

// olaResponse processes the HTTP response from the Ollama API. It reads the
// response body, estimates the actual token count, adjusts the rate limiter if
// necessary, and parses the JSON response to update the audit entry.
//
// Parameters:
//   - context.Context: the context for cancellation and timeout control.
//   - *Config: the analysis configuration containing host, model, and options.
//   - *http.Response: the HTTP response from the Ollama API.
//
// Returns:
//   - uint16: the estimated token count of the response.
//   - []byte: the JSON-formatted response body awaiting further processing.
//   - error: any error encountered during processing.
//
// The method performs the following steps:
//  1. Validates the response.
//  2. If the status code is not OK, returns an error with the status code.
//  3. Reads the response body into a buffer.
//  4. Estimates the actual response token count.
//  5. Decodes the JSON response into an olaRsp struct.
//  6. Returns the estimated token count and the raw response text.
func olaResponse(_ context.Context, _ *Config, rsp *http.Response) (uint16, []byte, error) {
	var (
		err error
		buf []byte
		mod olaRsp
		nbr uint16
	)

	// Validate that the response is not nil
	if rsp == nil {
		return 0, nil, errors.New("ollama: llm response is nil")
	}

	// If the response is not OK, return an error with the status code
	if rsp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("ollama: llm http status code: %d", rsp.StatusCode)
	}

	// Read the body contents into a local buffer
	if rsp.Body != nil {
		if buf, err = io.ReadAll(rsp.Body); err != nil {
			return 0, nil, fmt.Errorf("ollama: llm http body read error: %w", err)
		}
	}

	// Estimate the token count of the response body
	nbr = EstimateToken(buf)

	// Decode the JSON response into the olaRsp struct
	if err = json.Unmarshal(buf, &mod); err != nil {
		// Return the estimated token count along with the error, so that
		// downstream consumers can still use the estimate for rate limiting
		// or contestation purposes if needed.
		return nbr, nil, fmt.Errorf("ollama: llm decode error (token estimated: %d): %w", nbr, err)
	}

	// Return the estimated token count and the raw response text
	return nbr, []byte(mod.Response), nil
}
