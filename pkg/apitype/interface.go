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
	"math"
	"strings"
)

// ApiType represents a specific AI/LLM API provider.
// Each constant corresponds to a distinct API endpoint or service.
type ApiType uint8

const (
	// ApiOllama represents the Ollama API (local/self-hosted LLM inference).
	ApiOllama ApiType = iota

	// ApiOpenAI represents the OpenAI API (GPT models).
	ApiOpenAI

	// ApiCopilot represents the GitHub Copilot API.
	ApiCopilot

	// ApiAnthropic represents the Anthropic API (Claude models).
	ApiAnthropic

	// ApiGemini represents the Google Gemini API.
	ApiGemini

	// ApiMistral represents the Mistral AI API.
	ApiMistral
)

// List returns a slice containing all defined ApiType values.
// This is useful for iteration, validation, or serialization purposes.
func List() []ApiType {
	return []ApiType{
		ApiOllama,
		ApiOpenAI,
		ApiCopilot,
		ApiAnthropic,
		ApiGemini,
		ApiMistral,
	}
}

// Parse converts a string representation of an API type into its corresponding
// ApiType constant. The comparison is case-insensitive.
//
// If the input string does not match any known API type, ApiOllama is returned
// as the default fallback.
func Parse(s string) ApiType {
	switch {
	case strings.EqualFold(s, ApiOpenAI.String()):
		return ApiOpenAI
	case strings.EqualFold(s, ApiCopilot.String()):
		return ApiCopilot
	case strings.EqualFold(s, ApiAnthropic.String()):
		return ApiAnthropic
	case strings.EqualFold(s, ApiGemini.String()):
		return ApiGemini
	case strings.EqualFold(s, ApiMistral.String()):
		return ApiMistral
	default:
		return ApiOllama
	}
}

// ParseUint8 converts a uint8 value into the corresponding ApiType constant.
//
// If the input value does not match any known API type, ApiOllama is returned
// as the default fallback.
func ParseUint8(i uint8) ApiType {
	switch i {
	case uint8(ApiOpenAI):
		return ApiOpenAI
	case uint8(ApiCopilot):
		return ApiCopilot
	case uint8(ApiAnthropic):
		return ApiAnthropic
	case uint8(ApiGemini):
		return ApiGemini
	case uint8(ApiMistral):
		return ApiMistral
	default:
		return ApiOllama
	}
}

// ParseInt64 converts an int64 value into the corresponding ApiType constant.
//
// It iterates over all known API types and returns the first match.
// If no match is found, ApiOllama is returned as the default fallback.
func ParseInt64(i int64) ApiType {
	for _, v := range List() {
		if int64(v) == i {
			return v
		}
	}

	return ApiOllama
}

// ParseUint64 converts a uint64 value into the corresponding ApiType constant.
//
// It iterates over all known API types and returns the first match.
// If no match is found, ApiOllama is returned as the default fallback.
func ParseUint64(i uint64) ApiType {
	for _, v := range List() {
		if uint64(v) == i {
			return v
		}
	}

	return ApiOllama
}

// EstimateToken estimates the number of tokens in a byte buffer using a
// heuristic based on character classification. It returns 0 for an empty buffer.
//
// The estimation algorithm works as follows:
//
//  1. Each syntactic symbol (e.g., '{', '}', '[', ']', ':', ',', '"', '(', ')',
//     '.', '*', '&', ';', '=') is counted as one full token, since these
//     characters often form individual tokens in most tokenizers.
//
//  2. The remaining bytes (total length minus syntax symbols) are divided by 4,
//     reflecting the approximate ratio of 1 token per 4 bytes for natural text.
//     This is a rough heuristic derived from typical English text tokenization.
//
//  3. Whitespace characters (' ', '\t', '\n', '\r') are counted separately and
//     divided by 2, contributing a small additional token count. This accounts
//     for the fact that whitespace can sometimes form separate tokens in certain
//     tokenizers.
//
// The final estimate is the sum of:
//   - syntax symbols (symb)
//   - text bytes / 4 (txtb / 4)
//   - whitespace / 2 (wrds / 2)
//
// If the result exceeds math.MaxUint16, it is capped at that value to prevent
// overflow in downstream consumers that expect a uint16.
//
// Note: This is a rough heuristic and should not be used for precise token
// counting. For accurate tokenization, use the actual tokenizer provided by
// the target API.
func EstimateToken(buf []byte) uint16 {
	// Return early for empty buffers
	if len(buf) == 0 {
		return 0
	}

	var (
		wrds int64 // count of whitespace characters
		symb int64 // count of syntactic symbols
		txtb int64 // count of text bytes (total - symbols)
		estk int64 // final estimated token count
	)

	// Iterate over each byte in the buffer to classify characters
	// and accumulate counts for whitespace and syntactic symbols.
	for _, b := range buf {
		switch b {
		case ' ', '\t', '\n', '\r':
			// Whitespace characters: spaces, tabs, newlines, carriage returns
			wrds++
		case '{', '}', '[', ']', ':', ',', '"', '(', ')', '.', '*', '&', ';', '=':
			// Syntactic symbols: braces, brackets, punctuation, operators
			symb++
		}
	}

	// Calculate the number of text bytes by subtracting syntactic symbols
	// from the total buffer length. Guard against negative values in case
	// the symbol count exceeds the buffer length (should not happen in practice).
	if txtb = int64(len(buf)) - symb; txtb < 0 {
		txtb = 0
	}

	// Compute the final token estimate:
	//   - Each syntactic symbol counts as 1 token
	//   - Text bytes are divided by 4 (approx. 1 token per 4 bytes)
	//   - Whitespace is divided by 2 (partial token contribution)
	if estk = symb + (txtb / 4) + (wrds / 2); estk > math.MaxUint16 {
		return math.MaxUint16
	}

	return uint16(estk)
}
