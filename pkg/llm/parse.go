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

// Package llm provides an isolated, production-grade local engine interface for
// orchestrating multi-pass source code audits, structural cataloging, and automated
// security reviews leveraging high-performance local Language Models (LLMs) via Ollama.
//
// The package acts as a high-level abstraction layer, decoupling the business logic
// of code auditing from the underlying implementation details of prompt engineering,
// HTTP transport protocols, and JSON response unmarshaling.
package llm

import (
	"encoding/json"
	"errors"
	"fmt"

	audent "github.com/nabbar/auditor/pkg/data/entry"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audprm "github.com/nabbar/auditor/pkg/data/params"
	audsts "github.com/nabbar/auditor/pkg/data/status"
)

// DTOs (Data Transfer Objects) mapping the expected JSON responses from the LLM.

// llmParam represents a parameter or field description returned by the LLM.
type llmParam struct {
	// Name is the identifier of the parameter or field.
	Name string `json:"name,omitempty"`

	// Type is the data type of the parameter or field.
	Type string `json:"type,omitempty"`

	// Package is the package qualifier for the type, if applicable.
	Package string `json:"package,omitempty"`

	// Summary is the natural language description of the parameter or field.
	Summary string `json:"summary,omitempty"`
}

// llmMethod represents a method description returned by the LLM for an
// interface.
type llmMethod struct {
	// Name is the identifier of the method.
	Name string `json:"name"`

	// Summary is the natural language description of the method.
	Summary string `json:"summary"`

	// ParamsInput lists the input parameters of the method.
	ParamsInput []llmParam `json:"params_input,omitempty"`

	// ParamsOutput lists the output parameters of the method.
	ParamsOutput []llmParam `json:"params_output,omitempty"`
}

// llmCustom represents a custom or primitive type description returned
// by the LLM.
type llmCustom struct {
	// Summary is the natural language description of the custom type.
	Summary string `json:"summary"`
}

// llmStruct represents a struct description returned by the LLM.
type llmStruct struct {
	// Summary is the natural language description of the struct.
	Summary string `json:"summary"`

	// Fields lists the fields of the struct.
	Fields []llmParam `json:"fields,omitempty"`
}

// llmInterface represents an interface description returned by the LLM.
type llmInterface struct {
	// Summary is the natural language description of the interface.
	Summary string `json:"summary"`

	// Methods lists the methods of the interface.
	Methods []llmMethod `json:"methods,omitempty"`
}

// llmFunction represents a function description returned by the LLM.
type llmFunction struct {
	// Summary is the natural language description of the function.
	Summary string `json:"summary"`

	// ParamsInput lists the input parameters of the function.
	ParamsInput []llmParam `json:"params_input,omitempty"`

	// ParamsOutput lists the output parameters of the function.
	ParamsOutput []llmParam `json:"params_output,omitempty"`
}

// parseResult dispatches the LLM response to the appropriate parsing
// function based on the type of the analyzed code element (primitive,
// custom, struct, interface, or function). It identifies the response
// type by inspecting the top-level JSON keys and delegates to the
// corresponding parse method.
//
// Parameters:
//   - []byte: the raw LLM response as a JSON byte slice.
//   - audids.ID: the auditor entry ID associated with this LLM request.
//
// Returns an error if parsing fails.
func (o *mdl) parseResult(rsp []byte, id audids.ID) error {
	var (
		err error
		has bool
		hin bool
		hot bool
		key map[string]json.RawMessage
	)

	// Optimization: extract only top-level keys with json.RawMessage.
	// This avoids unmarshaling the entire JSON into a generic map interface
	// and allows us to inspect keys directly.
	if err = json.Unmarshal(rsp, &key); err != nil {
		return fmt.Errorf("llm payload inspection error: %w", err)
	}

	// If "fields" key exists, treat as struct
	if _, has = key["fields"]; has {
		return o.parseStruct(rsp, id)
	}

	// If "methods" key exists, treat as interface
	if _, has = key["methods"]; has {
		return o.parseInterface(rsp, id)
	}

	// If "params_input" or "params_output" keys exist, treat as function
	_, hin = key["params_input"]
	_, hot = key["params_output"]

	if hin || hot {
		return o.parseFunction(rsp, id)
	}

	// Default: treat as custom/primitive type
	return o.parseCustom(rsp, id)
}

// parseCustom parses the LLM response for a custom or primitive code element.
// It unmarshals the response into an llmCustom DTO, retrieves the entry
// from the database, updates its summary and status, and persists the
// changes.
//
// Parameters:
//   - []byte: the raw LLM response as a JSON byte slice.
//   - audids.ID: the auditor entry ID associated with this LLM request.
//
// Returns an error if unmarshaling fails or the entry is not found.
func (o *mdl) parseCustom(result []byte, id audids.ID) error {
	var (
		err error
		dto llmCustom
		ent audent.Entry
	)

	// Unmarshal the LLM response into the custom DTO
	if err = json.Unmarshal(result, &dto); err != nil {
		return fmt.Errorf("llm custom unmarshal error: %w", err)
	}

	// Retrieve the entry from the database
	if ent = o.dbm.EntGet(id); ent == nil || ent.IsEmpty() {
		return errors.New("llm entry not found for id")
	}

	// Update the entry with the LLM summary and mark as analyzed
	ent.SetSummary(dto.Summary)
	ent.SetStatus(audsts.Analyzed)

	// Persist the updated entry
	o.dbm.EntAdd(ent)
	return nil
}

// parseStruct parses the LLM response for a struct code element.
// It unmarshals the response into an llmStruct DTO, retrieves the entry
// from the database, updates its summary and field summaries, and
// persists the changes. If the LLM omits a field, an error is logged.
//
// Parameters:
//   - []byte: the raw LLM response as a JSON byte slice.
//   - audids.ID: the auditor entry ID associated with this LLM request.
//
// Returns an error if unmarshaling fails or the entry is not found.
func (o *mdl) parseStruct(result []byte, id audids.ID) error {
	var (
		err error
		dto llmStruct
		ent audent.Entry
		old []audprm.Params
		prm []audprm.Params
	)

	// Unmarshal the LLM response into the struct DTO
	if err = json.Unmarshal(result, &dto); err != nil {
		return fmt.Errorf("llm struct unmarshal error: %w", err)
	}

	// Retrieve the entry from the database
	if ent = o.dbm.EntGet(id); ent == nil || ent.IsEmpty() {
		return errors.New("llm entry not found for id")
	}

	// Update the entry with the LLM summary and mark as analyzed
	ent.SetSummary(dto.Summary)
	ent.SetStatus(audsts.Analyzed)

	// Copy existing input parameters to preserve them
	old = ent.GetInputs()
	prm = make([]audprm.Params, len(old))
	copy(prm, old)

	// Update each field's summary from the LLM response
	for i := range prm {
		if i < len(dto.Fields) {
			prm[i].SetSummary(dto.Fields[i].Summary)
		} else {
			// LLM did not provide a summary for this field
			o.uim.Error(fmt.Sprintf("LLM omitted field at index %d for entry %s", i, ent.GetFullPath()))
			break
		}
	}

	// Replace the entry's input parameters with the updated ones
	ent.DelInputs()
	ent.AddInputs(prm...)

	// Persist the updated entry
	o.dbm.EntAdd(ent)
	return nil
}

// parseInterface parses the LLM response for an interface code element.
// It unmarshals the response into an llmInterface DTO, retrieves the entry
// from the database, updates its summary and method/parameter summaries,
// and persists the changes. If the LLM omits a method or parameter, an
// error is logged.
//
// Parameters:
//   - []byte: the raw LLM response as a JSON byte slice.
//   - audids.ID: the auditor entry ID associated with this LLM request.
//
// Returns an error if unmarshaling fails or the entry is not found.
func (o *mdl) parseInterface(result []byte, id audids.ID) error {
	var (
		err error
		dto llmInterface
		ent audent.Entry
		old []audprm.Params
		prm []audprm.Params
	)

	// Unmarshal the LLM response into the interface DTO
	if err = json.Unmarshal(result, &dto); err != nil {
		return fmt.Errorf("llm interface unmarshal error: %w", err)
	}

	// Retrieve the entry from the database
	if ent = o.dbm.EntGet(id); ent == nil || ent.IsEmpty() {
		return errors.New("llm entry not found for id")
	}

	// Update the entry with the LLM summary and mark as analyzed
	ent.SetSummary(dto.Summary)
	ent.SetStatus(audsts.Analyzed)

	// Copy existing input parameters (methods) to preserve them
	old = ent.GetInputs()
	prm = make([]audprm.Params, len(old))
	copy(prm, old)

	// Update each method's summary and its input/output parameters
	for i := range prm {
		var (
			osb []audprm.Params
			lsb []audprm.Params
		)

		// Check if the LLM provided a summary for this method
		if i < len(dto.Methods) {
			prm[i].SetSummary(dto.Methods[i].Summary)
		} else {
			// LLM did not provide a summary for this method
			o.uim.Error(fmt.Sprintf("LLM omitted method at index %d for entry %s", i, ent.GetFullPath()))
			break
		}

		// Update input parameters for this method
		osb = prm[i].GetInput()
		lsb = make([]audprm.Params, len(osb))
		copy(lsb, osb)

		for j := range lsb {
			if j < len(dto.Methods[i].ParamsInput) {
				lsb[j].SetSummary(dto.Methods[i].ParamsInput[j].Summary)
			} else {
				// LLM did not provide a summary for this input parameter
				o.uim.Error(fmt.Sprintf("LLM omitted input parameter at index %d for method %d for entry %s", j, i, ent.GetFullPath()))
				break
			}
		}

		prm[i].DelInput()
		prm[i].SetInput(lsb...)

		// Update output parameters for this method
		osb = prm[i].GetOutput()
		lsb = make([]audprm.Params, len(osb))
		copy(lsb, osb)

		for j := range lsb {
			if i < len(dto.Methods[i].ParamsOutput) {
				lsb[j].SetSummary(dto.Methods[i].ParamsOutput[j].Summary)
			} else {
				// LLM did not provide a summary for this output parameter
				o.uim.Error(fmt.Sprintf("LLM omitted output parameter at index %d for method %d for entry %s", j, i, ent.GetFullPath()))
				break
			}
		}

		prm[i].DelOutput()
		prm[i].SetOutput(lsb...)
	}

	// Replace the entry's input parameters with the updated ones
	ent.DelInputs()
	ent.AddInputs(prm...)

	// Persist the updated entry
	o.dbm.EntAdd(ent)
	return nil
}

// parseFunction parses the LLM response for a function code element.
// It unmarshals the response into an llmFunction DTO, retrieves the entry
// from the database, updates its summary and input/output parameter
// summaries, and persists the changes. If the LLM omits a parameter,
// an error is logged.
//
// Parameters:
//   - []byte: the raw LLM response as a JSON byte slice.
//   - audids.ID: the auditor entry ID associated with this LLM request.
//
// Returns an error if unmarshaling fails or the entry is not found.
func (o *mdl) parseFunction(result []byte, id audids.ID) error {
	var (
		err error
		dto llmFunction
		ent audent.Entry
		old []audprm.Params
		prm []audprm.Params
	)

	// Unmarshal the LLM response into the function DTO
	if err = json.Unmarshal(result, &dto); err != nil {
		return fmt.Errorf("llm function unmarshal error: %w", err)
	}

	// Retrieve the entry from the database
	if ent = o.dbm.EntGet(id); ent == nil || ent.IsEmpty() {
		return errors.New("llm entry not found for id")
	}

	// Update the entry with the LLM summary and mark as analyzed
	ent.SetSummary(dto.Summary)
	ent.SetStatus(audsts.Analyzed)

	// Update input parameters
	old = ent.GetInputs()
	prm = make([]audprm.Params, len(old))
	copy(prm, old)

	for i := range prm {
		// Check if the LLM provided a summary for this input parameter
		if i < len(dto.ParamsInput) {
			prm[i].SetSummary(dto.ParamsInput[i].Summary)
		} else {
			// LLM did not provide a summary for this input parameter
			o.uim.Error(fmt.Sprintf("LLM omitted input parameter at index %d for entry %s", i, ent.GetFullPath()))
			break
		}
	}

	// Replace the entry's input parameters with the updated ones
	ent.DelInputs()
	ent.AddInputs(prm...)

	// Update output parameters
	old = ent.GetOutputs()
	prm = make([]audprm.Params, len(old))
	copy(prm, old)

	for i := range prm {
		// Check if the LLM provided a summary for this output parameter
		if i < len(dto.ParamsOutput) {
			prm[i].SetSummary(dto.ParamsOutput[i].Summary)
		} else {
			// LLM did not provide a summary for this output parameter
			o.uim.Error(fmt.Sprintf("LLM omitted output parameter at index %d for entry %s", i, ent.GetFullPath()))
			break
		}
	}

	// Replace the entry's output parameters with the updated ones
	ent.DelOutputs()
	ent.AddOutputs(prm...)

	// Persist the updated entry
	o.dbm.EntAdd(ent)
	return nil
}
