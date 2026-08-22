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

// Package data provides interfaces and utilities for managing model instances.
//
// This package defines the core interfaces and data structures required to manage
// model instances within the auditor system. It includes definitions for:
//   - Data interface: Defines methods for managing model identity, status, summary, and reports
//   - ManData interface: Provides thread-safe operations for loading, storing, deleting, and iterating over model instances
//   - New function: Creates a new ManData instance with optional constructor function
//
// The design supports generic types to allow for different model implementations while
// maintaining consistent interfaces across all model types.
package data

import (
	"encoding/json"

	"github.com/fxamacker/cbor/v2"
	audids "github.com/nabbar/auditor/pkg/data/id"
	audrep "github.com/nabbar/auditor/pkg/data/reports"
	audsts "github.com/nabbar/auditor/pkg/data/status"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

// Data represents the interface that all model instances must implement.
//
// The Data interface defines a consistent set of methods for managing model instances,
// including identity management, status tracking, summary information, and report handling.
// All concrete implementations of model instances must satisfy this interface to ensure
// consistent behavior across the system.
//
// Key features:
//   - Identity management through unique identifiers (ID)
//   - Status tracking with comprehensive status management
//   - Summary text for concise model descriptions
//   - Report management for detailed analysis and results
//
// The generic type parameter T allows for different model types while maintaining
// consistent interfaces.
type Data[M any] interface {
	json.Marshaler
	json.Unmarshaler
	yaml.Marshaler
	yaml.Unmarshaler
	toml.Marshaler
	toml.Unmarshaler
	cbor.Marshaler
	cbor.Unmarshaler

	// GetID returns the unique identifier of the model instance.
	//
	// This method provides access to the unique identifier that distinguishes
	// this model instance from others in the system. The ID is typically generated
	// upon creation and remains constant throughout the instance's lifetime.
	GetID() audids.ID

	// IsEmpty checks if the model instance is empty or uninitialized.
	//
	// This method determines whether the model instance represents an empty or
	// uninitialized state. An empty instance typically indicates that it has not
	// been properly initialized or does not contain valid data.
	IsEmpty() bool

	// IsEqual compares this instance with another instance of the same type.
	//
	// This method performs a deep comparison between two model instances to determine
	// if they are equivalent in terms of their state and content. It is used for
	// validation purposes and equality checks within collections.
	IsEqual(M) bool

	// Merge compares and update this instance with another instance of the same type.
	//
	// This method attempts to merge the state of another instance into this one.
	// It performs a comparison between the current instance and the provided instance,
	// updating fields where appropriate. The return value indicates whether any changes
	// were made during the merge operation.
	Merge(any) bool

	// SetStatus updates the status of the model instance.
	//
	// This method allows for setting or changing the status of the model instance.
	// Status values typically indicate the current state of processing, such as
	// pending, processing, completed, failed, etc. The status is crucial for tracking
	// the lifecycle and progress of model instances.
	SetStatus(audsts.Status)

	// GetStatus retrieves the current status of the model instance.
	//
	// This method provides access to the current status of the model instance,
	// which reflects its processing state within the system. The status is used
	// for monitoring, reporting, and decision-making based on instance state.
	GetStatus() audsts.Status

	// SetSummary updates the summary text of the model instance.
	//
	// This method sets or replaces the summary text that provides a brief overview
	// of the model instance's content or purpose. The summary is typically used for
	// quick identification and understanding of the model's role within the system.
	SetSummary(string)

	// GetSummary retrieves the summary text of the model instance.
	//
	// This method returns the current summary text associated with the model instance.
	// The summary provides a concise description that helps in identifying and
	// understanding the purpose or content of the instance at a glance.
	GetSummary() string

	// SetReports adds or updates a report associated with the model instance.
	//
	// This method allows for associating or updating reports with the model instance.
	// Reports contain detailed analysis, results, or findings related to the model's
	// processing. Multiple reports can be maintained for comprehensive analysis.
	SetReports(audrep.Report)

	// DelReports removes a specific report by name from the model instance.
	//
	// This method deletes a report identified by its name from the model instance.
	// It is used when reports are no longer needed or when updating with new versions.
	// If the report does not exist, this operation has no effect.
	DelReports(string)

	// GetReports retrieves a specific report by name from the model instance.
	//
	// This method returns a report associated with the model instance based on
	// its name. It provides access to detailed analysis or results that were
	// generated during processing. If the report does not exist, it returns an
	// empty report.
	GetReports(string) audrep.Report

	// LstReports returns a slice of all reports associated with the model instance.
	//
	// This method provides a complete list of all reports currently associated
	// with the model instance. It enables iteration over all available reports
	// for display, processing, or analysis purposes.
	LstReports() []audrep.Report
}
