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
// A Status is a bit-field stored in a single uint8 byte. Each bit records the
// completion of one processing phase. The phases are independent: any
// combination can be active at the same time, which allows a partially completed
// analysis to be resumed, and allows a caller to know exactly which step still
// has to be performed.
//
// # Type definition
//
// The Status type is an alias of uint8. Because it is a value type, all Set*
// and UnSet* methods take a value receiver and return a new Status; they do not
// modify the receiver. The caller must assign the result:
//
//	s = s.SetParsed()
//
// # Bit layout
//
// The bit-field constants are exported and can be combined with the standard
// bitwise operators. Each constant corresponds to a single bit position:
//
//	 0  Parsed    the AST of the element has been parsed
//	 1  Vendor    the vendor of the element has been resolved
//	 2  Analyzed  the summary of the element has been computed
//	 3  Audited1  the first audit report has been generated
//	 4  Audited2  the second audit report has been generated
//	 5  Audited3  the third audit report has been generated
//	 6  Audited4  the fourth audit report has been generated
//	 7  Outdated  the element is known to be out of date
//
// The zero value of Status (None) means that no phase is active.
//
// # Creation
//
// New returns a zero-valued Status with all phases unset. Get constructs a
// Status from a raw uint8 value, allowing the caller to build a status from
// external data at the cost of losing safety -- the caller is responsible for
// ensuring the value represents a valid combination of the defined phase bits.
//
// # Phase helpers
//
// Each phase has three helpers: Set*, UnSet*, and Is*. They encapsulate the
// bitwise OR, AND-NOT, and AND operations respectively.
//
//	status = status.SetParsed()      // sets bit 0
//	status = status.UnSetParsed()    // clears bit 0
//	status.IsParsed()                // true if bit 0 is set
//
// In addition, SetNone clears all bits, and IsNone returns true when no
// phase is active. The raw uint8 value of a status can be retrieved with
// Uint8().
//
// # Textual representation
//
// The Parse function decodes a comma-separated list of phase names into a
// Status. It is case-insensitive and ignores unknown tokens. The "None" name
// clears all bits. Spaces around tokens are trimmed. Parse and the String
// method are symmetric: String(s) can be fed back to Parse(s) without loss.
//
//	s := status.Parse("Parsed,Analyzed,Outdated")
//	s.String() // "Parsed,Analyzed,Outdated"
//
// # Serialization
//
// The bit field is serialized as its raw uint8 value, preserving the bit field
// losslessly across formats.
//
// Serialization support per format:
//
//   - json: serialized as a JSON number (uint8). A JSON string is rejected
//     with a type error on unmarshal.
//   - yaml: serialized as a YAML scalar holding the uint8 value.
//   - toml: serialized as a TOML integer via the go-toml/v2 library. The
//     status must be encoded as a field within a structure.
//   - cbor: serialized as a canonical CBOR unsigned integer. This is also
//     the canonical binary representation used by MarshalBinary and
//     UnmarshalBinary.
//
// # Data flow
//
// The typical lifecycle of a Status is the following:
//
//	 1. New() or status.None            --- create an empty status
//	 2. Get(raw)                        --- restore from a raw uint8 value
//	 3. Parse("Parsed,Vendor,...")     --- restore from a textual representation
//	 4. .SetPhase() / .UnSetPhase()    --- mark a phase as done or undone
//	 5. .IsPhase()                     --- check whether a phase is done
//	 6. .String()                      --- produce a comma-separated list of
//	                                    active phases (round-trips with Parse)
//	 7. .Uint8()                       --- extract the raw value
//	 8. Marshal*/Unmarshal*            --- serialize / deserialize in one of
//	                                    the supported formats (json, yaml, toml,
//	                                    cbor, or binary)
//
// The status is not thread-safe by design. If multiple goroutines mutate the
// same Status, external synchronization is the responsibility of the caller.
//
// # Example usage
//
//	var s = status.New()
//	s = s.SetParsed()
//	s = s.SetVendor()
//	if s.IsParsed() { // use s }
//
//	// Round-trip through text
//	text := s.String()               // "Parsed,Vendor"
//	recovered := status.Parse(text)  // same bits as s
//
//	// Round-trip through JSON
//	data, _ := json.Marshal(s)
//	json.Unmarshal(data, &s)         // restores the bits
//

package status
