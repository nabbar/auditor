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

// Package llm provides functionality for interacting with Ollama language models.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/nabbar/auditor/models"
)

// Client represents a client for making requests to the Ollama API.
// It encapsulates the base URL and model name needed for API communication.
type Client struct {
	// URL is the endpoint of the Ollama API server
	URL string
	// Model specifies which language model to use for inference
	Model string
}

// NewClient creates and returns a new Ollama client instance with the specified URL and model.
// The URL should point to the base endpoint of the Ollama service (e.g., "http://localhost:11434").
// The Model parameter specifies which model to use for inference (e.g., "mistral:7b").
func NewClient(url, model string) *Client {
	return &Client{URL: url, Model: model}
}

// Call sends a prompt to the Ollama API and returns the generated response.
// This method manages the complete lifecycle of an LLM request including:
//   - Request preparation with specified parameters
//   - Context-based timeout handling (120 seconds)
//   - HTTP request execution
//   - Response parsing and error handling
//
// Parameters:
//   - prompt: The text prompt to send to the language model
//   - jsonFormat: When true, instructs the API to return a valid JSON response
//   - maxTokens: Maximum number of tokens to generate in the response
//
// Returns:
//   - string: The generated response from the language model
//   - error: Any error that occurred during the request or processing
func (c *Client) Call(prompt string, jsonFormat bool, maxTokens int) (string, error) {
	// Construct the request body with model specification and prompt
	reqBody := models.OllamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]any{
			"temperature": 0.1,
			"num_predict": maxTokens,
		},
	}

	// Configure JSON formatting if requested
	// This ensures that the model will return a valid JSON structure when jsonFormat is true
	if jsonFormat {
		reqBody.Format = "json"
	}

	// Marshal the request body into JSON format for transmission
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		log.Printf("[ERROR] [llm] Failed to marshal JSON request: %v", err)
		return "", err
	}

	// Record the start time for performance monitoring
	start := time.Now()

	// Create a context with timeout to manage request lifecycle
	// Using a 120-second timeout ensures that requests don't hang indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Create HTTP request with the context-bound timeout
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.URL, bytes.NewBuffer(jsonBytes))
	if err != nil {
		log.Printf("[ERROR] [llm] Failed to create HTTP request: %v", err)
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request using a standard HTTP client
	httpClient := &http.Client{}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		// Check if the error was caused by context timeout
		if ctx.Err() == context.DeadlineExceeded {
			log.Printf("[CRITICAL] [llm] Ollama API request timed out after 120s")
			return "", ctx.Err()
		}
		log.Printf("[ERROR] [llm] Network error connecting to Ollama API: %v", err)
		return "", err
	}
	defer func() {
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()

	// Log warning if the response status code is not OK (200)
	if resp.StatusCode != http.StatusOK {
		log.Printf("[WARNING] [llm] Ollama responded with non-200 status code: %d", resp.StatusCode)
	}

	// Decode the API response into a structured format
	var ollamaResp models.OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		log.Printf("[ERROR] [llm] Failed to decode Ollama response: %v", err)
		return "", err
	}

	// Log the completion time for performance monitoring
	log.Printf("[INFO] [llm] Call completed in %v (Max tokens requested: %d)", time.Since(start), maxTokens)

	// Return the response text from the language model
	return ollamaResp.Response, nil
}
