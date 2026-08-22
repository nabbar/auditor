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
)

// oaiReq represents the request payload sent to the OpenAI Chat Completions API.
// It maps directly to the JSON schema expected by OpenAI's /v1/chat/completions endpoint.
type oaiReq struct {
	// Model specifies the OpenAI model to use (e.g., "gpt-4o", "gpt-3.5-turbo").
	Model string `json:"model"`

	// Messages is the array of conversation turns (user/assistant messages).
	Messages []oaiMsg `json:"messages"`

	// Stream indicates whether to use Server-Sent Events (SSE) for streaming.
	// Currently set to false for non-streaming (blocking) requests.
	Stream bool `json:"stream"`

	// ResponseFormat forces the model to output valid JSON.
	// Set to {"type": "json_object"} to ensure structured output.
	ResponseFormat map[string]string `json:"response_format,omitempty"`

	// Temperature controls the randomness of the model output.
	// Higher values increase diversity; lower values make output more deterministic.
	Temperature *float64 `json:"temperature,omitempty"`
}

// oaiMsg represents a single message in the conversation turn, used in the
// Messages array of oaiReq.
type oaiMsg struct {
	// Role is the sender of the message: "user", "assistant", or "system".
	Role string `json:"role"`

	// Content is the text content of the message.
	Content string `json:"content"`
}

// oaiRsp represents the simplified response structure from OpenAI.
// Only the first choice's message content is extracted; richer structures
// (e.g., tool_calls, logprobs) are ignored in this implementation.
type oaiRsp struct {
	// Choices is the array of possible model responses.
	Choices []struct {
		Message struct {
			// Content is the text output from the model.
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// oaiRequest constructs and returns an HTTP POST request targeting the OpenAI
// Chat Completions API. It marshals the request payload, applies configuration
// options (model, temperature, response format), and sets the required headers.
//
// Parameters:
//   - context.Context: The context for the request lifecycle.
//   - *Config: API configuration containing hostname, model name, and options.
//   - []byte: The raw prompt bytes to be sent as the user message.
//
// Returns:
//   - *http.Request: The prepared request ready to be executed.
//   - uint16: An estimated token count for the request payload.
//   - error: Any error encountered during construction.
func oaiRequest(ctx context.Context, cfg *Config, prt []byte) (*http.Request, uint16, error) {
	// Validate required configuration fields.
	if cfg == nil || len(cfg.Hostname) < 1 || len(cfg.Model) < 1 {
		return nil, 0, errors.New("openai: config is nil or invalid")
	}

	// Validate that the prompt is non-empty.
	if len(prt) < 1 {
		return nil, 0, errors.New("openai: prompt is empty")
	}

	var (
		err error
		buf []byte
		hst = cfg.Hostname
		nbr uint16
		mod oaiReq
		pth string
		req *http.Request
	)

	// Construct the full API endpoint URL.
	if pth, err = url.JoinPath(hst, pthOpenAI); err != nil {
		return nil, 0, fmt.Errorf("openai: invalid path: %w", err)
	}

	// Build the request payload with the prompt and generation config.
	// ResponseFormat is set to {"type": "json_object"} to force JSON output.
	mod = oaiReq{
		Model: cfg.Model,
		Messages: []oaiMsg{
			{
				Role:    "user",
				Content: string(prt), // The full prompt as a direct user message
			},
		},
		Stream: false,
		ResponseFormat: map[string]string{
			"type": "json_object", // Force JSON output format
		},
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
		return nil, 0, fmt.Errorf("openai: marshal error: %w", err)
	}

	// Estimate the token count for the request payload.
	nbr = EstimateToken(buf)

	// Create the HTTP POST request with the marshaled JSON body.
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, pth, bytes.NewBuffer(buf)); err != nil {
		return nil, 0, fmt.Errorf("openai: http request error: %w", err)
	}

	// Set required HTTP headers for the OpenAI API.
	req.Header.Set(hrdContent, hrdJson) // Content-Type: application/json
	req.Header.Set(hrdAccept, hrdJson)  // Accept: application/json

	return req, nbr, nil
}

// oaiResponse processes the HTTP response from the OpenAI Chat Completions API.
// It validates the status code, reads the response body, unmarshals the JSON,
// and extracts the text content from the first choice's message.
//
// Parameters:
//   - context.Context: not used.
//   - *Config: not used.
//   - *http.Response: The HTTP response from the OpenAI API.
//
// Returns:
//   - uint16: An estimated token count for the response body.
//   - []byte: The raw text output from the model.
//   - error: Any error encountered during processing.
func oaiResponse(_ context.Context, _ *Config, rsp *http.Response) (uint16, []byte, error) {
	// Validate the response object.
	if rsp == nil {
		return 0, nil, errors.New("openai: response is nil")
	}

	// Validate the HTTP status code; only 200 OK is accepted.
	if rsp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("openai: http status code: %d", rsp.StatusCode)
	}

	var (
		err error
		buf []byte
		mod oaiRsp
		nbr uint16
	)

	// Read the response body.
	if rsp.Body != nil {
		if buf, err = io.ReadAll(rsp.Body); err != nil {
			return 0, nil, fmt.Errorf("openai: http body read error: %w", err)
		}
	}

	// Estimate the token count for the response body.
	nbr = EstimateToken(buf)

	// Unmarshal the JSON response into the oaiRsp struct.
	if err = json.Unmarshal(buf, &mod); err != nil {
		return nbr, nil, fmt.Errorf("openai: decode error (token estimated: %d): %w", nbr, err)
	}

	// Validate that the response contains at least one choice.
	if len(mod.Choices) == 0 {
		return nbr, nil, errors.New("openai: empty response choices")
	}

	// Extract the text from the first choice's message.
	return nbr, []byte(mod.Choices[0].Message.Content), nil
}
