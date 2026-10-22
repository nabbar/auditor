//
// MIT License
//
// Copyright (c) 2026 Nicolas JUHEL
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//

// Package data defines the contract shared by every model of the auditor data
// layer.
//
// A model of the auditor (a module, a package, a code entry, ...) is a
// self-contained, identifiable and serializable object. The Data[M] interface
// describes that contract: it exposes the identity of the model, its
// lifecycle status, its summary and the reports produced by the audit.
//
// # Identity
//
// Every model owns a unique audids.ID, computed deterministically from the
// coordinates of the element it represents (module path, package path, entry
// name, ...). The identifier is stable: the same coordinates always produce
// the same identifier, in the same process and in any other process, which
// makes it usable as a map key, as a file name or as a database key.
//
// The IsEmpty method reports whether a model carries enough information to be
// processed. It is used by the collection sub-package to filter out the
// partially initialized models before they are stored.
//
// # Status
//
// The status of a model is a bit field (see the status sub-package) where each
// bit records the completion of one processing phase: AST parsing, vendor
// resolution, summary analysis and the four audit reports. The bits are set
// and cleared one at a time, so an interrupted analysis can be resumed where
// it stopped, and so a caller always knows which step remains to be done.
//
// # Merge
//
// Merge folds the data of another model of the same kind into the receiver.
// It is the mechanism used to refresh a model that has already been stored
// with a newly parsed version of the same element. Merge returns false when
// the two models do not designate the same element.
//
// # Equality
//
// IsEqual compares two models of the same kind. The comparison is a deep one:
// it is performed on the content of the models and not on their identity
// alone, so that two models built from the same coordinates but holding
// different data are reported as different.
//
// # Serialization
//
// The Data interface embeds the json, yaml, toml and cbor marshaler and
// unmarshaler contracts, so that any model can be persisted to, and restored
// from, any of those formats.
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

// Data is the interface that all the model instances of the auditor must
// implement.
//
// The Data interface defines a consistent set of methods for managing a model
// instance: identity management, status tracking, summary information and
// report handling. All the concrete implementations of a model instance must
// satisfy this interface, so that the collections and the managers of the
// auditor can handle modules, packages and code entries in the same way.
//
// Key features:
//   - Identity management through unique identifiers (GetID)
//   - Emptiness detection used to discard the incomplete models (IsEmpty)
//   - Deep comparison between two models of the same kind (IsEqual)
//   - Incremental update of a model from a freshly parsed one (Merge)
//   - Status tracking with a complete set of bit manipulation helpers
//   - Summary text for concise model descriptions
//   - Report management for detailed analysis and results
//
// The generic type parameter M is the concrete type of the model that
// implements the interface. It is used by the Merge method, which receives
// the model to merge as an untyped value, and by the IsEqual method, which
// receives the model to compare with. It allows several kinds of model to
// share the same contract while keeping the type safety of the operations.
//
// The interface also embeds the json, yaml, toml and cbor marshaler and
// unmarshaler contracts, so that any model can be persisted to, and restored
// from, any of those formats.
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
	// this model instance from the others in the system. The identifier is
	// derived from the coordinates of the element the model represents (its
	// module path, its package path, its name, ...) and is therefore
	// reproducible: two runs of the analysis over the same sources produce the
	// same identifier.
	//
	// The identifier is computed once, when the model is built, and remains
	// constant for the whole lifetime of the instance, even when the content
	// of the model is updated by Merge. It is the key used by the collections
	// to store and to retrieve the models.
	//
	// Returns:
	//   - audids.ID: The unique identifier of this model instance. The zero
	//     value is returned by an uninitialized model.
	GetID() audids.ID

	// IsEmpty checks if the model instance is empty or uninitialized.
	//
	// This method determines whether the model instance represents an empty or
	// uninitialized state. An empty instance typically indicates that it has
	// not been properly initialized, or that it does not carry the minimum
	// set of data required to be processed.
	//
	// The exact criteria depend on the concrete model: the modules, the
	// packages and the code entries each define the conditions that make them
	// meaningful. A model that reports itself as empty is discarded by the
	// collections instead of being stored.
	//
	// Returns:
	//   - bool: true when the instance carries no usable data, false when it
	//     is fully initialized.
	IsEmpty() bool

	// IsEqual compares this instance with another instance of the same type.
	//
	// This method performs a deep comparison between two model instances to
	// determine if they are equivalent in terms of their state and content. It
	// is used for the validation purposes and for the equality checks within
	// the collections.
	//
	// The comparison is performed on the content of the models and not on
	// their identity alone: two instances that share the same identifier but
	// hold different data are reported as different. The concrete criteria
	// depend on the model, but they always cover the identity and the
	// meaningful payload of the instance (its name, its type, its parameters,
	// its dependencies, ...).
	//
	// Parameters:
	//   - M: The model instance to compare with.
	//
	// Returns:
	//   - bool: true when both instances describe the same data, false
	//     otherwise.
	IsEqual(M) bool

	// Merge compares and updates this instance with another instance of the
	// same type.
	//
	// This method attempts to merge the state of another instance into this
	// one. It performs a comparison between the current instance and the
	// provided instance, and updates the fields where appropriate.
	//
	// The merge is a partial update: the empty or unset fields of the source
	// instance never overwrite the data already held by the receiver. Only the
	// fields that carry a meaningful value are transferred.
	//
	// The merge is refused when the two instances do not designate the same
	// element: the implementations compare their identifiers and their
	// coordinates, and return false without modifying the receiver when they
	// differ. Merging two different elements would silently corrupt the data
	// of the receiver, so it must never be attempted.
	//
	// Parameters:
	//   - any: The instance to merge into the receiver. The implementations
	//     expect the concrete type of the model; any other type is rejected.
	//
	// Returns:
	//   - bool: true when the merge was accepted, false when the source does
	//     not designate the same element as the receiver.
	Merge(any) bool

	// SetStatus updates the status of the model instance.
	//
	// This method allows for setting or changing the status of the model
	// instance. The status is a bit field (see the status sub-package) where
	// each bit tracks the completion of one processing phase: AST parsing,
	// vendor resolution, summary analysis and the four audit reports.
	//
	// This setter replaces the whole bit field at once. The dedicated
	// SetStatusParsed, SetStatusVendor, ... helpers are usually preferred,
	// because they update a single phase and leave the other ones untouched.
	//
	// Parameters:
	//   - audsts.Status: The new status of the model instance.
	SetStatus(audsts.Status)

	// GetStatus retrieves the current status of the model instance.
	//
	// This method provides access to the current status of the model instance,
	// which reflects its processing state within the system. The status is used
	// for the monitoring, the reporting and the decision making based on the
	// state of the instance.
	//
	// Returns:
	//   - audsts.Status: The current status of the model instance. The zero
	//     value (audsts.None) means that no processing phase is active.
	GetStatus() audsts.Status

	// SetStatusNone clears all the status bits of the model instance.
	//
	// This method resets the status to the default "None" state, where no
	// processing phase is active. After this call, every other status bit is
	// cleared and the instance returns to its initial state.
	//
	// It is typically used to restart the processing of an element from
	// scratch, or to mark an element that has to be analyzed again.
	SetStatusNone()

	// SetStatusParsed marks the AST parse phase as completed.
	//
	// This method sets bit 0 of the status to indicate that the AST parsing
	// phase has successfully completed for the model instance. The other bits
	// of the status are left unchanged.
	SetStatusParsed()

	// SetStatusVendor marks the AST vendor resolution phase as completed.
	//
	// This method sets bit 1 of the status to indicate that the AST vendor
	// resolution phase has successfully completed for the model instance. The
	// other bits of the status are left unchanged.
	SetStatusVendor()

	// SetStatusAnalyzed marks the summary analysis phase as completed.
	//
	// This method sets bit 2 of the status to indicate that the summary
	// analysis phase has successfully completed for the model instance. The
	// other bits of the status are left unchanged.
	SetStatusAnalyzed()

	// SetStatusAudited1 marks the first audit report as generated.
	//
	// This method sets bit 3 of the status to indicate that the first audit
	// report has been generated for the model instance. The other bits of the
	// status are left unchanged.
	SetStatusAudited1()

	// SetStatusAudited2 marks the second audit report as generated.
	//
	// This method sets bit 4 of the status to indicate that the second audit
	// report has been generated for the model instance. The other bits of the
	// status are left unchanged.
	SetStatusAudited2()

	// SetStatusAudited3 marks the third audit report as generated.
	//
	// This method sets bit 5 of the status to indicate that the third audit
	// report has been generated for the model instance. The other bits of the
	// status are left unchanged.
	SetStatusAudited3()

	// SetStatusAudited4 marks the fourth audit report as generated.
	//
	// This method sets bit 6 of the status to indicate that the fourth audit
	// report has been generated for the model instance. The other bits of the
	// status are left unchanged.
	SetStatusAudited4()

	// SetStatusOutdated marks the item as out of date.
	//
	// This method sets bit 7 of the status to indicate that the model
	// instance has been flagged as out of date, which means that the data it
	// holds no longer matches the sources it was built from. The other bits of
	// the status are left unchanged.
	SetStatusOutdated()

	// UnSetStatusParsed clears the AST parse phase status bit.
	//
	// This method clears bit 0 from the model instance status, removing the
	// indication that the AST parsing has completed. The other status bits
	// remain unchanged.
	UnSetStatusParsed()

	// UnSetStatusVendor clears the AST vendor resolution phase status bit.
	//
	// This method clears bit 1 from the model instance status, removing the
	// indication that the vendor resolution has completed. The other status
	// bits remain unchanged.
	UnSetStatusVendor()

	// UnSetStatusAnalyzed clears the summary analysis phase status bit.
	//
	// This method clears bit 2 from the model instance status, removing the
	// indication that the summary analysis has completed. The other status
	// bits remain unchanged.
	UnSetStatusAnalyzed()

	// UnSetStatusAudited1 clears the first audit report status bit.
	//
	// This method clears bit 3 from the model instance status, removing the
	// indication that the first audit report was generated. The other status
	// bits remain unchanged.
	UnSetStatusAudited1()

	// UnSetStatusAudited2 clears the second audit report status bit.
	//
	// This method clears bit 4 from the model instance status, removing the
	// indication that the second audit report was generated. The other status
	// bits remain unchanged.
	UnSetStatusAudited2()

	// UnSetStatusAudited3 clears the third audit report status bit.
	//
	// This method clears bit 5 from the model instance status, removing the
	// indication that the third audit report was generated. The other status
	// bits remain unchanged.
	UnSetStatusAudited3()

	// UnSetStatusAudited4 clears the fourth audit report status bit.
	//
	// This method clears bit 6 from the model instance status, removing the
	// indication that the fourth audit report was generated. The other status
	// bits remain unchanged.
	UnSetStatusAudited4()

	// UnSetStatusOutdated clears the out of date status bit.
	//
	// This method clears bit 7 from the model instance status, removing the
	// out of date flag. The other status bits remain unchanged.
	UnSetStatusOutdated()

	// IsStatusNone returns true when all status bits are cleared.
	//
	// This method checks whether all status bits are zero, indicating
	// that no processing phase is currently active on the model instance.
	//
	// Returns:
	//   - bool: true if the status equals zero (no active phases), false otherwise
	IsStatusNone() bool

	// IsStatusParsed returns true if the AST parse phase bit is set.
	//
	// This method checks whether bit 0 (Parsed) is set, indicating that
	// the AST parsing phase has completed for the model instance.
	//
	// Returns:
	//   - bool: true if the Parsed bit is set, false otherwise
	IsStatusParsed() bool

	// IsStatusVendor returns true if the AST vendor resolution bit is set.
	//
	// This method checks whether bit 1 (Vendor) is set, indicating that
	// the AST vendor resolution phase has completed for the model instance.
	//
	// Returns:
	//   - bool: true if the Vendor bit is set, false otherwise
	IsStatusVendor() bool

	// IsStatusAnalyzed returns true if the summary analysis bit is set.
	//
	// This method checks whether bit 2 (Analyzed) is set, indicating that
	// the summary analysis phase has completed for the model instance.
	//
	// Returns:
	//   - bool: true if the Analyzed bit is set, false otherwise
	IsStatusAnalyzed() bool

	// IsStatusAudited1 returns true if the first audit report bit is set.
	//
	// This method checks whether bit 3 (Audited1) is set, indicating that
	// the first audit report has been generated for the model instance.
	//
	// Returns:
	//   - bool: true if the Audited1 bit is set, false otherwise
	IsStatusAudited1() bool

	// IsStatusAudited2 returns true if the second audit report bit is set.
	//
	// This method checks whether bit 4 (Audited2) is set, indicating that
	// the second audit report has been generated for the model instance.
	//
	// Returns:
	//   - bool: true if the Audited2 bit is set, false otherwise
	IsStatusAudited2() bool

	// IsStatusAudited3 returns true if the third audit report bit is set.
	//
	// This method checks whether bit 5 (Audited3) is set, indicating that
	// the third audit report has been generated for the model instance.
	//
	// Returns:
	//   - bool: true if the Audited3 bit is set, false otherwise
	IsStatusAudited3() bool

	// IsStatusAudited4 returns true if the fourth audit report bit is set.
	//
	// This method checks whether bit 6 (Audited4) is set, indicating that
	// the fourth audit report has been generated for the model instance.
	//
	// Returns:
	//   - bool: true if the Audited4 bit is set, false otherwise
	IsStatusAudited4() bool

	// IsStatusOutdated returns true if the out of date bit is set.
	//
	// This method checks whether bit 7 (Outdated) is set, indicating that
	// the model instance has been flagged as out of date.
	//
	// Returns:
	//   - bool: true if the Outdated bit is set, false otherwise
	IsStatusOutdated() bool

	// SetSummary updates the summary text of the model instance.
	//
	// This method sets or replaces the summary text that provides a brief
	// overview of the content or of the purpose of the model instance. The
	// summary is typically used for the quick identification and the
	// understanding of the role of the model within the system.
	//
	// Parameters:
	//   - string: The new summary of the model instance.
	SetSummary(string)

	// GetSummary retrieves the summary text of the model instance.
	//
	// This method returns the current summary text associated with the model
	// instance. The summary provides a concise description that helps in
	// identifying and in understanding the purpose or the content of the
	// instance at a glance.
	//
	// Returns:
	//   - string: The summary of the model instance. An empty string is
	//     returned when no summary has been set.
	GetSummary() string

	// SetReports adds or updates a report associated with the model instance.
	//
	// This method allows for associating or updating a report with the model
	// instance. The reports contain the detailed analysis, the results or the
	// findings related to the processing of the model. Multiple reports can be
	// maintained, they are keyed by their own name.
	//
	// Adding a report whose name is already known replaces the previous
	// version of that report.
	//
	// Parameters:
	//   - audrep.Report: The report to add or to update. Its Name field is
	//     used as the key of the report collection.
	SetReports(audrep.Report)

	// DelReports removes a specific report by name from the model instance.
	//
	// This method deletes the report identified by its name from the model
	// instance. It is used when the reports are no longer needed, or when
	// they have to be replaced by newer versions. If the report does not
	// exist, this operation has no effect.
	//
	// Parameters:
	//   - string: The name of the report to remove.
	DelReports(string)

	// GetReports retrieves a specific report by name from the model instance.
	//
	// This method returns the report associated with the model instance that
	// matches the given name. It provides an access to the detailed analysis
	// or to the results that were generated during the processing. If the
	// report does not exist, an empty report is returned.
	//
	// Parameters:
	//   - string: The name of the report to retrieve.
	//
	// Returns:
	//   - audrep.Report: The requested report, or the zero value when no
	//     report carries that name.
	GetReports(string) audrep.Report

	// LstReports returns a slice of all the reports associated with the model
	// instance.
	//
	// This method provides a complete list of all the reports currently
	// associated with the model instance. It enables the iteration over all
	// the available reports for the display, the processing or the analysis
	// purposes.
	//
	// The order of the reports is not specified and may change between two
	// calls, because the reports are held in a map.
	//
	// Returns:
	//   - []audrep.Report: The reports of the model instance. An empty slice
	//     is returned when the model holds no report.
	LstReports() []audrep.Report

	// GetParentId returns the identifier of the parent of the model instance.
	//
	// The parent of a model is the element that contains it: the module for a
	// package, the package for a code entry. The identifier is resolved
	// lazily, by the way of the resolver functions registered by the
	// application, so that a model never has to hold a reference to its
	// parent.
	//
	// A model that has no parent, such as a module, returns the zero
	// identifier.
	//
	// Returns:
	//   - audids.ID: The identifier of the parent of the model instance, or
	//     the zero value when the model has no parent.
	GetParentId() audids.ID
}
