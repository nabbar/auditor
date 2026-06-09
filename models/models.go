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

// Package models defines the data structures used throughout the auditing system.
// These structures represent various aspects of the audit process including functions,
// dependencies, and API interactions.
package models

// Function represents a function under audit (Pass 1 & 2).
// This structure stores comprehensive information about functions being analyzed during the auditing process.
// It provides metadata that allows tracking and managing the audit lifecycle of each function.
type Function struct {
	// ID uniquely identifies the function within the system.
	// This identifier is essential for maintaining references to specific functions
	// across different stages of the audit process.
	ID string

	// Package contains the name of the Go package where the function is defined.
	// This field enables grouping and categorization of functions by their package context,
	// which is crucial for understanding dependencies and scope.
	Package string

	// FuncName holds the name of the function.
	// This identifier allows for direct referencing of the function within its package
	// and provides human-readable identification during audit processes.
	FuncName string

	// FilePath stores the path to the source file containing the function.
	// This field enables locating the exact source code location,
	// which is necessary for detailed analysis and reporting.
	FilePath string

	// Summary provides a brief description of the function's purpose and behavior.
	// This summary serves as a quick reference for auditors and helps in understanding
	// the core functionality without requiring deep code inspection.
	Summary string

	// Status indicates the current audit status of the function.
	// Possible values include pending, analyzed, reviewed, etc.
	// This field enables tracking progress through different audit stages
	// and helps in managing workflow efficiently.
	Status string
}

// CatalogEntry represents a raw entry in the global catalog.
// This structure stores basic information about functions that are catalogued for auditing purposes.
// It serves as a foundational data structure for maintaining an inventory of all functions
// that require audit attention, enabling systematic analysis and tracking.
type CatalogEntry struct {
	// Package contains the name of the Go package where the function is defined.
	// This field allows for organizing entries by their source packages,
	// which facilitates hierarchical analysis and dependency mapping.
	Package string

	// FuncName holds the name of the function.
	// This identifier enables quick lookup and identification of functions
	// within the catalog without requiring additional metadata retrieval.
	FuncName string

	// FilePath stores the path to the source file containing the function.
	// This field ensures that each catalog entry can be linked back to its source code,
	// which is essential for verification and detailed examination during audits.
	FilePath string
}

// Dependency represents an external element identified in Pass 1.
// This structure tracks external dependencies that functions rely on during analysis.
// It captures information about external components that may affect function behavior,
// security implications, or integration requirements.
type Dependency struct {
	// FunctionID references the ID of the function that contains this dependency.
	// This field establishes a direct relationship between dependencies and their parent functions,
	// enabling comprehensive tracking and analysis of how external elements impact specific functions.
	FunctionID string

	// DepName specifies the name of the external dependency.
	// This identifier allows for categorizing and searching dependencies by their names,
	// which is crucial for understanding the scope and nature of external integrations.
	DepName string

	// DepType indicates the type of external dependency (e.g., package, interface, etc.).
	// This field provides classification information that helps in assessing
	// the complexity and potential impact of different types of dependencies.
	DepType string
}

// OllamaRequest structures the request to the Ollama API.
// This structure defines the format for sending prompts to the language model service.
// It encapsulates all necessary parameters required to make effective API calls to the LLM system,
// ensuring consistency and proper formatting of requests across different audit processes.
type OllamaRequest struct {
	// Model specifies the name of the language model to be used for processing.
	// This parameter determines which AI model will analyze the function or prompt content,
	// allowing for selection of appropriate models based on specific requirements.
	Model string `json:"model"`

	// Prompt contains the text prompt to be sent to the language model.
	// This field holds the actual content that will be processed by the LLM,
	// containing instructions, context, or questions related to the audit task.
	Prompt string `json:"prompt"`

	// Stream indicates whether the response should be streamed or returned as a complete response.
	// When set to true, the API returns partial results incrementally, which can improve
	// user experience and reduce waiting times for large responses.
	Stream bool `json:"stream"`

	// Format specifies the output format for the response (e.g., json, text).
	// This parameter determines how the model's response will be formatted,
	// enabling consistent parsing and processing of results across different systems.
	Format string `json:"format,omitempty"`

	// Options contains additional configuration parameters for the model request.
	// These optional settings allow fine-tuning of model behavior such as temperature,
	// max tokens, or other hyperparameters that influence response quality and characteristics.
	Options map[string]any `json:"options"`
}

// OllamaResponse captures the raw response from Ollama.
// This structure holds the unprocessed output from the language model API.
// It serves as a container for the complete response from the LLM system,
// preserving all data before it's parsed or processed into more structured formats.
type OllamaResponse struct {
	// Response contains the text output from the language model.
	// This field holds the primary content generated by the AI model,
	// which may contain analysis results, summaries, or other insights relevant to auditing.
	Response string `json:"response"`
}

// P1Response decodes the strict JSON structure expected in Pass 1.
// This structure defines the expected format of responses during the first pass of analysis.
// It ensures that all responses from the LLM conform to a standardized schema,
// enabling consistent processing and integration into the audit workflow.
type P1Response struct {
	// Summary provides a concise overview of the function's behavior and purpose.
	// This field contains a high-level description generated by the AI model
	// that captures essential aspects of the function's functionality and intent.
	Summary string `json:"summary"`

	// External contains a list of external dependencies identified during analysis.
	// This slice holds information about all external components referenced by the function,
	// enabling comprehensive dependency tracking and risk assessment.
	External []struct {
		// Name specifies the name of an external dependency.
		// This field identifies the specific external element that the function depends on,
		// allowing for targeted analysis of integration points and potential vulnerabilities.
		Name string `json:"name"`

		// Type indicates the type of the external dependency.
		// This field classifies the nature of the external dependency (e.g., package, interface),
		// which helps in understanding the scope and characteristics of each dependency.
		Type string `json:"type"`
	} `json:"external"`
}
