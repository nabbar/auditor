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

// Package status tracks the progress of the analysis of the elements handled
// by the auditor.
//
// A Status is a bit field where each bit records the completion of one
// processing phase. The phases are independent: any combination of them can
// be active at the same time, which allows a partially completed analysis to
// be resumed, and allows a caller to know exactly which step still has to be
// performed.
//
// The bit layout is:
//
//	Bit  Phase     Meaning
//	---  --------  --------------------------------------------------
//	 0   Parsed    the AST of the element has been parsed
//	 1   Vendor    the vendor of the element has been resolved
//	 2   Analyzed  the summary of the element has been computed
//	 3   Audited1  the first audit report has been produced
//	 4   Audited2  the second audit report has been produced
//	 5   Audited3  the third audit report has been produced
//	 6   Audited4  the fourth audit report has been produced
//	 7   Outdated  the element is known to be out of date
//
// The constants are powers of two, so a phase is set with the bitwise OR
// operator and tested with the bitwise AND operator:
//
//	status |= status.Parsed    // sets the Parsed bit
//	status&status.Parsed != 0 // tests the Parsed bit
//
// The Status type provides the named helpers for both operations, and the
// String method renders the active phases as a comma separated list.
//
// # Serialization
//
// The package provides the serialization of the Status type for the json, yaml,
// toml and cbor formats, plus the binary encoding interfaces. The bit field is
// always serialized as its raw uint8 value, so that the round trip is
// lossless.
package status

import "strings"

// The bit-field constants of the Status type.
//
// Each constant corresponds to a single bit and is the power of two at its
// bit position, so that the constants can be combined with the bitwise
// operators to build any valid status value.
//
// The constants are grouped in a single block because they belong to the same
// enumeration and share the same type.
//
// The first constant, None, is the default value of a Status and the zero
// value of the type.
const (
	// None indicates that no processing phase is active. It is the
	// default, zero value.
	None Status = 0

	// Parsed indicates that the AST parsing phase has completed. Bit 0.
	Parsed Status = 1 << 0

	// Vendor indicates that the AST vendor resolution phase has completed.
	// Bit 1.
	Vendor Status = 1 << 1

	// Analyzed indicates that the summary analysis phase has completed.
	// Bit 2.
	Analyzed Status = 1 << 2

	// Audited1 indicates that the first audit report has been generated.
	// Bit 3.
	Audited1 Status = 1 << 3

	// Audited2 indicates that the second audit report has been generated.
	// Bit 4.
	Audited2 Status = 1 << 4

	// Audited3 indicates that the third audit report has been generated.
	// Bit 5.
	Audited3 Status = 1 << 5

	// Audited4 indicates that the fourth audit report has been generated.
	// Bit 6.
	Audited4 Status = 1 << 6

	// Outdated indicates that the item has been marked as out of date.
	// Bit 7.
	Outdated Status = 1 << 7
)

// The textual representations of the status phases.
//
// They are the tokens accepted by Parse and produced by String, and they
// match the names of the exported constants.
const (
	strNone     = "None"
	strParsed   = "Parsed"
	strVendor   = "Vendor"
	strAnalyzed = "Analyzed"
	strAudited1 = "Audited1"
	strAudited2 = "Audited2"
	strAudited3 = "Audited3"
	strAudited4 = "Audited4"
	strOutdated = "Outdated"
)

// Status is a bit-field type that tracks which processing phases have been
// completed for an audit item.
//
// Each bit corresponds to a distinct phase (see the package constants).
// Because it is a bit field, multiple phases can be active at the same time,
// and each of them can be set or cleared without disturbing the others.
//
// The type is an uint8, so it stays small and can be stored as a single byte
// in the serialized documents.
type Status uint8

// New returns a zero-valued Status with all the phases unset.
//
// The returned value is the same as the None constant: no processing phase is
// active, which is the state of an element that has not been processed yet.
//
// Returns:
//   - Status: A new status where every phase is unset.
func New() Status {
	return Status(0)
}

// Get constructs a Status from the given raw uint8 value. The caller is
// responsible for ensuring the value represents a valid combination of
// the defined phase bits.
func Get(i uint8) Status {
	return Status(i)
}

// Parse returns a Status by interpreting a comma-separated list of phase
// names (for example "Parsed,Analyzed").
//
// Each recognized name sets the corresponding bit. The names are compared
// after a trimming of the surrounding spaces, so "Parsed, Analyzed" is
// accepted as well. The unrecognized names and the empty names are silently
// ignored, and the occurrence of the "None" name clears every bit.
//
// The parsing is case insensitive: the tokens are the exact names of the
// exported constants.
//
// Parameters:
//   - s: The comma-separated list of the phase names to decode.
//
// Returns:
//   - Status: The status built from the recognized phase names.
func Parse(s string) Status {
	var (
		r = None
		p = strings.Split(s, ",")
	)

	for _, v := range p {
		switch strings.ToLower(strings.TrimSpace(v)) {
		case strings.ToLower(strNone):
			r = r.SetNone()
		case strings.ToLower(strParsed):
			r = r.SetParsed()
		case strings.ToLower(strVendor):
			r = r.SetVendor()
		case strings.ToLower(strAnalyzed):
			r = r.SetAnalyzed()
		case strings.ToLower(strAudited1):
			r = r.SetAudited1()
		case strings.ToLower(strAudited2):
			r = r.SetAudited2()
		case strings.ToLower(strAudited3):
			r = r.SetAudited3()
		case strings.ToLower(strAudited4):
			r = r.SetAudited4()
		case strings.ToLower(strOutdated):
			r = r.SetOutdated()
		}
	}

	return r
}
