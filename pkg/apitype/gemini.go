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

// gemReq represents the request payload sent to the Gemini API.
// It maps to the JSON schema expected by Google's Generative Language API.
type gemReq struct {
	// Contents is the array of conversation turns (user/assistant messages).
	Contents []gemCnt `json:"contents"`

	// GenerationConfig contains optional parameters controlling the model's
	// output behavior (temperature, response format, etc.).
	GenerationConfig *gemGenCfg `json:"generationConfig,omitempty"`
}

// gemCnt represents a single conversation turn in the Gemini API request.
type gemCnt struct {
	// Parts is the array of content parts within a turn (text, images, etc.).
	Parts []gemPrt `json:"parts"`
}

// gemPrt represents a single content part within a conversation turn.
type gemPrt struct {
	// Text is the text content of this part.
	Text string `json:"text"`
}

// gemGenCfg contains optional generation parameters for the Gemini model.
type gemGenCfg struct {
	// Temperature controls the randomness of the model output.
	// Higher values increase diversity; lower values make output more deterministic.
	Temperature *float64 `json:"temperature,omitempty"`

	// ResponseMimeType specifies the expected output format.
	// Set to "application/json" to force JSON output from the model.
	ResponseMimeType string `json:"responseMimeType,omitempty"`
}

// gemRsp represents the response structure from the Gemini API.
// It contains one or more candidate responses, each with content parts.
type gemRsp struct {
	// Candidates is the array of possible model responses.
	Candidates []struct {
		Content struct {
			// Parts is the array of content parts in the response.
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// gemRequest constructs and returns an HTTP POST request targeting the Gemini
// Generative Language API. It marshals the request payload, applies
// configuration options (model, temperature, response format), and sets the
// required headers.
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
func gemRequest(ctx context.Context, cfg *Config, prt []byte) (*http.Request, uint16, error) {
	// Validate required configuration fields.
	if cfg == nil || len(cfg.Hostname) < 1 || len(cfg.Model) < 1 {
		return nil, 0, errors.New("gemini: config is nil or invalid")
	}

	// Validate that the prompt is non-empty.
	if len(prt) < 1 {
		return nil, 0, errors.New("gemini: prompt is empty")
	}

	var (
		err error
		buf []byte
		mod gemReq
		nbr uint16
		pth string
		req *http.Request
	)

	// Construct the Gemini API endpoint URL.
	// The Gemini API embeds the model name in the URL path.
	// Example: https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent
	if pth, err = url.JoinPath(cfg.Hostname, "/v1beta/models/", cfg.Model+":generateContent"); err != nil {
		return nil, 0, fmt.Errorf("gemini: failed to generate url path: %w", err)
	}

	// Build the request payload with the prompt and generation config.
	// ResponseMimeType is set to "application/json" to force JSON output.
	mod = gemReq{
		Contents: []gemCnt{
			{
				Parts: []gemPrt{
					{
						Text: string(prt),
					},
				},
			},
		},
		GenerationConfig: &gemGenCfg{
			ResponseMimeType: hrdJson, // Force JSON output format
		},
	}

	// Apply temperature override if provided in config.Options.
	if v, k := cfg.Options[optTemperature]; k {
		switch i := v.(type) {
		case float32:
			a := float64(i)
			mod.GenerationConfig.Temperature = &a
		case float64:
			mod.GenerationConfig.Temperature = &i
		}
	}

	// Marshal the request struct to JSON.
	if buf, err = json.Marshal(mod); err != nil {
		return nil, 0, fmt.Errorf("gemini: marshal error: %w", err)
	}

	// Estimate the token count for the request payload.
	nbr = EstimateToken(buf)

	// Create the HTTP POST request with the marshaled JSON body.
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, pth, bytes.NewBuffer(buf)); err != nil {
		return nil, 0, fmt.Errorf("gemini: http request error: %w", err)
	}

	// Set the Content-Type header for JSON.
	req.Header.Set(hrdContent, hrdJson)

	return req, nbr, nil
}

// gemResponse processes the HTTP response from the Gemini API.
// It validates the status code, reads the response body, unmarshals the JSON,
// and extracts the text content from the first candidate's first part.
//
// Parameters:
//   - context.Context: not used.
//   - *Config: not used.
//   - *http.Response: The HTTP response from the Gemini API.
//
// Returns:
//   - uint16: An estimated token count for the response body.
//   - []byte: The raw text output from the model.
//   - error: Any error encountered during processing.
func gemResponse(_ context.Context, _ *Config, rsp *http.Response) (uint16, []byte, error) {
	// Validate the response object.
	if rsp == nil {
		return 0, nil, errors.New("gemini: response is nil")
	}

	// Validate the HTTP status code; only 200 OK is accepted.
	if rsp.StatusCode != http.StatusOK {
		return 0, nil, fmt.Errorf("gemini: http status code: %d", rsp.StatusCode)
	}

	var (
		err error
		buf []byte
		nbr uint16
		mod gemRsp
	)

	// Read the response body.
	if rsp.Body != nil {
		if buf, err = io.ReadAll(rsp.Body); err != nil {
			return 0, nil, fmt.Errorf("gemini: http body read error: %w", err)
		}
	}

	// Estimate the token count for the response body.
	nbr = EstimateToken(buf)

	// Unmarshal the JSON response into the gemRsp struct.
	if err = json.Unmarshal(buf, &mod); err != nil {
		return nbr, nil, fmt.Errorf("gemini: decode error (token estimated: %d): %w", nbr, err)
	}

	// Validate that the response contains at least one candidate with content.
	if len(mod.Candidates) == 0 || len(mod.Candidates[0].Content.Parts) == 0 {
		return nbr, nil, errors.New("gemini: empty response candidates")
	}

	// Extract the text from the first candidate's first part.
	buf = []byte(mod.Candidates[0].Content.Parts[0].Text)

	return nbr, buf, nil
}
