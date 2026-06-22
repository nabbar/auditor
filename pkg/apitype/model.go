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

	libdur "github.com/nabbar/golib/duration"
)

const (
	// optNumPredict is the key for the number of predicted tokens option
	// in the LLM request options map.
	optNumPredict = "num_predict"

	// optTemperature is the key for the temperature option in the LLM
	// request options map.
	optTemperature = "temperature"

	// pthOllama is the Ollama Model API endpoint path for generating awaiting results.
	pthOllama = "/api/generate"

	// pthOpenAI is the OpenAI Model API endpoint path for generating awaiting results.
	pthOpenAI = "/v1/chat/completions"

	// pthAnthropic is the Anthropic Model API endpoint path for generating awaiting results.
	pthAnthropic = "/v1/messages"

	// hrdContent is the HTTP header name for Content-Type.
	hrdContent = "Content-Type"

	// hrdAccept is the HTTP header name for Accept.
	hrdAccept = "Accept"

	// hrdJson is the MIME type for JSON content.
	hrdJson = "application/json"

	// hrdAntVersion is...
	hrdAntVersion = "anthropic-version"

	// valAntVersion is...
	valAntVersion = "2023-06-01"
)

// Config holds the configuration parameters required to connect to an LLM API
// endpoint. It includes the hostname, model identifier, request timeout, and
// optional parameters that can be passed to the API.
type Config struct {
	// Hostname is the base URL or hostname of the LLM endpoint.
	Hostname string `json:"hostname" yaml:"hostname" toml:"hostname"`

	// Model is the LLM model identifier to use for requests.
	Model string `json:"model" yaml:"model" toml:"model"`

	// Timeout is the maximum duration allowed for a single LLM request.
	Timeout libdur.Duration `json:"timeout" yaml:"timeout" toml:"timeout"`

	// Options is a map of optional parameters that can be passed to the API.
	// Common keys include "num_predict" (number of predicted tokens) and
	// "temperature" (sampling temperature).
	Options map[string]interface{} `json:"options" yaml:"options" toml:"options"`
}

// String returns the human-readable string representation of the ApiType.
// This implements the fmt.Stringer interface.
func (o ApiType) String() string {
	switch o {
	case ApiOpenAI:
		return "OpenAI"
	case ApiCopilot:
		return "Copilot"
	case ApiAnthropic:
		return "Anthropic"
	case ApiGemini:
		return "Gemini"
	case ApiMistral:
		return "Mistral"
	default:
		return "Ollama"
	}
}

// Uint8 returns the ApiType value as a uint8.
func (o ApiType) Uint8() uint8 {
	return uint8(o)
}

// Uint64 returns the ApiType value as a uint64.
func (o ApiType) Uint64() uint64 {
	return uint64(o)
}

// Int64 returns the ApiType value as an int64.
func (o ApiType) Int64() int64 {
	return int64(o)
}

// Request constructs and returns an HTTP request for the specified API type,
// along with an estimated token count for the payload.
//
// It validates the input parameters (Config and payload) and delegates to the
// appropriate API-specific request builder. Currently, only the Ollama API
// is supported; all other API types return an error indicating they are not
// yet implemented.
//
// Parameters:
//   - ctx: the context for the request, used for cancellation and timeouts.
//   - cfg: the configuration for the API endpoint.
//   - prt: the payload bytes to send to the API.
//
// Returns:
//   - *http.Request: the constructed HTTP request.
//   - uint16: the estimated number of tokens in the payload.
//   - error: any error that occurred during request construction.
func (o ApiType) Request(ctx context.Context, cfg *Config, prt []byte) (*http.Request, uint16, error) {
	// Validate that the configuration is not nil
	if cfg == nil {
		return nil, 0, errors.New("api backend: nil Config")
	}

	// Validate that the payload is not empty
	if len(prt) < 1 {
		return nil, 0, errors.New("api backend: empty payload")
	}

	// Dispatch to the appropriate API-specific request builder based on the
	// ApiType. Currently, only Ollama is implemented; all other types return
	// a "not supported yet" error.
	switch o {
	case ApiOpenAI:
		return oaiRequest(ctx, cfg, prt)
	case ApiCopilot:
		return cplRequest(ctx, cfg, prt)
	case ApiAnthropic:
		return antRequest(ctx, cfg, prt)
	case ApiGemini:
		return gemRequest(ctx, cfg, prt)
	case ApiMistral:
		return mstRequest(ctx, cfg, prt)
	default:
		return olaRequest(ctx, cfg, prt)
	}
}

// Response processes the HTTP response from the API and extracts the token
// count and response body.
//
// It validates the input parameters (Config and response) and delegates to the
// appropriate API-specific response handler. Currently, only the Ollama API
// is supported; all other API types return an error indicating they are not
// yet implemented.
//
// Parameters:
//   - ctx: the context for the response processing.
//   - cfg: the configuration for the API endpoint.
//   - rsp: the HTTP response from the API.
//
// Returns:
//   - uint16: the number of tokens in the response.
//   - []byte: the response body bytes.
//   - error: any error that occurred during response processing.
func (o ApiType) Response(ctx context.Context, cfg *Config, rsp *http.Response) (uint16, []byte, error) {
	// Validate that the configuration is not nil
	if cfg == nil {
		return 0, nil, errors.New("api backend: nil Config")
	}

	// Validate that the response is not nil
	if rsp == nil {
		return 0, nil, errors.New("api backend: nil response")
	}

	// Dispatch to the appropriate API-specific response handler based on the
	// ApiType. Currently, only Ollama is implemented; all other types return
	// a "not supported yet" error.
	switch o {
	case ApiOpenAI:
		return oaiResponse(ctx, cfg, rsp)
	case ApiCopilot:
		return cplResponse(ctx, cfg, rsp)
	case ApiAnthropic:
		return antResponse(ctx, cfg, rsp)
	case ApiGemini:
		return gemResponse(ctx, cfg, rsp)
	case ApiMistral:
		return mstResponse(ctx, cfg, rsp)
	default:
		return olaResponse(ctx, cfg, rsp)
	}
}
