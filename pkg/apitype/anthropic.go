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
	"math"
	"net/http"
	"net/url"
)

// antReq represents the request payload sent to the Anthropic Messages API.
// It maps directly to the JSON schema expected by Anthropic's /v1/messages endpoint.
type antReq struct {
	Model     string   `json:"model"`
	Messages  []antMsg `json:"messages"`
	MaxTokens int      `json:"max_tokens"` // Required by Anthropic; caps the output length

	// Stream indicates whether to use Server-Sent Events (SSE) for streaming.
	// Currently set to false for non-streaming (blocking) requests.
	Stream bool `json:"stream"`

	// System is the system-level instruction passed at the root level.
	// Anthropic uses this field for the system prompt, rather than embedding it
	// inside the messages array.
	System string `json:"system,omitempty"`

	// Temperature controls the randomness of the model output.
	// Higher values increase diversity; lower values make output more deterministic.
	Temperature *float64 `json:"temperature,omitempty"`
}

// antMsg represents a single message in the conversation turn, used in the
// Messages array of antReq.
type antMsg struct {
	Role    string `json:"role"`    // "user" or "assistant"
	Content string `json:"content"` // The text content of the message
}

// antRsp represents the simplified response structure from Anthropic.
// Only the first content block (text) is extracted; richer structures (e.g.,
// tool_use, citations) are ignored in this implementation.
type antRsp struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

// antRequest constructs and returns an HTTP POST request targeting the Anthropic
// Messages API. It marshals the request payload, applies configuration options
// (model, max_tokens, temperature), and sets the required headers.
//
// Parameters:
//   - context.Context: ...
//   - *Config: API configuration containing hostname, model name, and optional overrides.
//   - []byte: The raw prompt bytes to be sent as the user message.
//
// Returns:
//   - *http.Request: The prepared request ready to be executed.
//   - uint16: An estimated token count for the request payload.
//   - error: Any error encountered during construction.
func antRequest(ctx context.Context, cfg *Config, prt []byte) (*http.Request, uint16, error) {
	// Validate required configuration fields.
	if cfg == nil || len(cfg.Hostname) < 1 || len(cfg.Model) < 1 {
		return nil, 0, errors.New("anthropic: config is nil or invalid")
	}

	// Validate that the prompt is non-empty.
	if len(prt) < 1 {
		return nil, 0, errors.New("anthropic: prompt is empty")
	}

	var (
		err error
		buf []byte
		mod antReq
		nbr uint16
		pth string
		req *http.Request

		// Default configuration values
		hst = cfg.Hostname
		tok = 4096 // Default max_tokens; can be overridden via config.Options
	)

	// Construct the full API endpoint URL.
	if pth, err = url.JoinPath(hst, pthAnthropic); err != nil {
		return nil, 0, fmt.Errorf("anthropic: invalid path: %w", err)
	}

	// Override max_tokens if a custom value is provided in config.Options.
	// Supports all numeric types (int, uint, float) with safe casting.
	if v, k := cfg.Options[optNumPredict]; k {
		switch i := v.(type) {
		case int:
			tok = i
		case int8:
			tok = int(i)
		case int16:
			tok = int(i)
		case int32:
			tok = int(i)
		case int64:
			if i < math.MaxInt {
				tok = int(i)
			} else {
				tok = math.MaxInt
			}
		case uint:
			if i < math.MaxInt {
				tok = int(i)
			} else {
				tok = math.MaxInt
			}
		case uint8:
			tok = int(i)
		case uint16:
			tok = int(i)
		case uint32:
			tok = int(i)
		case uint64:
			if i < math.MaxInt {
				tok = int(i)
			} else {
				tok = math.MaxInt
			}
		case float32:
			if i < math.MaxInt64 {
				tok = int(i)
			} else {
				tok = math.MaxInt
			}
		case float64:
			if i < math.MaxInt64 {
				tok = int(i)
			} else {
				tok = math.MaxInt
			}
		}
	}

	// Build the request payload.
	// The "prefill" trick: the assistant's first message is seeded with "{" to
	// force the model to start its output with an opening brace. This eliminates
	// introductory filler text and markdown backticks, ensuring clean JSON output.
	mod = antReq{
		Model: cfg.Model,
		Messages: []antMsg{
			{
				Role:    "user",
				Content: string(prt), // The full prompt, including instructions and source code
			},
			{
				Role:    "assistant",
				Content: "{",
				// Prefill trick: forces Claude to begin with "{" so that no
				// introductory text or markdown backticks are generated.
			},
		},
		MaxTokens: tok,
		Stream:    false,
	}

	// Apply temperature override if provided in config.Options.
	if v, k := cfg.Options[optTemperature]; k {
		switch i := v.(type) {
		case float32:
			a := float64(i)
			mod.Temperature = &a
		case float64:
			mod.Temperature = &i
		}
	}

	// Marshal the request struct to JSON.
	if buf, err = json.Marshal(mod); err != nil {
		return nil, 0, fmt.Errorf("anthropic: marshal error: %w", err)
	}

	// Estimate the token count for the request payload.
	nbr = EstimateToken(buf)

	// Create the HTTP POST request with the marshaled JSON body.
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, pth, bytes.NewBuffer(buf)); err != nil {
		return nil, 0, fmt.Errorf("anthropic: http request error: %w", err)
	}

	// Set required HTTP headers for the Anthropic API.
	req.Header.Set(hrdContent, hrdJson)          // Content-Type: application/json
	req.Header.Set(hrdAccept, hrdJson)           // Accept: application/json
	req.Header.Set(hrdAntVersion, valAntVersion) // X-Api-Key / beta header required by Anthropic

	return req, nbr, nil
}

// antResponse processes the HTTP response from the Anthropic Messages API.
// It validates the status code, reads the response body, unmarshals the JSON,
// and reconstructs the output by prepending the prefill "{" that Anthropic
// omits from its response.
//
// Parameters:
//   - context.Context: not used
//   - *Config: not used
//   - *http.Response: The HTTP response from the Anthropic API.
//
// Returns:
//   - uint16: An estimated token count for the response body.
//   - []byte: The raw JSON output from the model (with prefill brace restored).
//   - error: Any error encountered during processing.
func antResponse(_ context.Context, _ *Config, rsp *http.Response) (uint16, []byte, error) {
	// Validate the response object.
	if rsp == nil {
		return 0, nil, errors.New("anthropic: response is nil")
	}

	// Validate the HTTP status code; only 200 OK is accepted.
	if rsp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("anthropic: http status code: %d", rsp.StatusCode)
	}

	var (
		err error
		buf []byte
		nbr uint16
		mod antRsp
	)

	// Read the response body.
	if rsp.Body != nil {
		if buf, err = io.ReadAll(rsp.Body); err != nil {
			return 0, nil, fmt.Errorf("anthropic: http body read error: %w", err)
		}
	}

	// Estimate the token count for the response body.
	nbr = EstimateToken(buf)

	// Unmarshal the JSON response into the antRsp struct.
	if err = json.Unmarshal(buf, &mod); err != nil {
		return nbr, nil, fmt.Errorf("anthropic: decode error (token estimated: %d): %w", nbr, err)
	}

	// Validate that the response contains at least one content block.
	if len(mod.Content) == 0 {
		return nbr, nil, errors.New("anthropic: empty response content")
	}

	// IMPORTANT: Because we seeded the assistant message with "{", Anthropic does
	// not include that opening brace in its response. We must prepend it manually
	// to reconstruct a valid JSON output.
	buf = []byte("{" + mod.Content[0].Text)

	return nbr, buf, nil
}
