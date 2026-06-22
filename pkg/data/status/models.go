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

// File models.go contains the helper methods of the Status type.
//
// They provide the string representation, the raw numeric conversion, and the
// setters, un-setters and inspectors for each phase bit of the status.
package status

import "strings"

// String returns a comma-separated list of the active phase names, ordered
// from the lowest to the highest bit.
//
// If no phase is active, it returns "None". The format is compatible with
// Parse for the round tripping, for example:
//
//	Parsed,Analyzed,Outdated
//
// Returns:
//   - string: The comma-separated names of the active phases, or "None" when
//     the status is empty.
func (o Status) String() string {
	var r = make([]string, 0, 8)

	if o.IsParsed() {
		r = append(r, strParsed)
	}

	if o.IsVendor() {
		r = append(r, strVendor)
	}

	if o.IsAnalyzed() {
		r = append(r, strAnalyzed)
	}

	if o.IsAudited1() {
		r = append(r, strAudited1)
	}

	if o.IsAudited2() {
		r = append(r, strAudited2)
	}

	if o.IsAudited3() {
		r = append(r, strAudited3)
	}

	if o.IsAudited4() {
		r = append(r, strAudited4)
	}

	if o.IsOutdated() {
		r = append(r, strOutdated)
	}

	if len(r) == 0 {
		return strNone
	}

	return strings.Join(r, ",")
}

// Uint8 returns the raw uint8 representation of the status. Use this to
// pass the value to serialization formats or numeric APIs that expect an
// unsigned 8-bit integer.
func (o Status) Uint8() uint8 {
	return uint8(o)
}

// SetNone clears all the bits of the status.
//
// The Status type is a value type: the method does not modify its receiver,
// it returns the new status. The caller has to assign the result to use it:
//
//	status = status.SetNone()
//
// After the call the status represents a reset item where no phase is active.
//
// Returns:
//   - Status: A new status where every phase is unset.
func (o Status) SetNone() Status {
	return 0
}

// IsNone returns true when all the status bits are cleared, that is when the
// status equals zero.
//
// The method does not modify the status.
//
// Returns:
//   - bool: true when no phase is active, false otherwise.
func (o Status) IsNone() bool {
	return o == 0
}

// SetParsed sets bit 0 to mark the AST parse phase as completed.
//
// The Status type is a value type: the method does not modify its receiver,
// it returns the new status. The caller has to assign the result to use it.
//
// Returns:
//   - Status: A new status with the Parsed bit set.
func (o Status) SetParsed() Status {
	return o | Parsed
}

// UnSetParsed clears bit 0, which marks the AST parse phase as completed.
//
// The other bits of the status are left unchanged. The Status type is a value
// type: the method does not modify its receiver, it returns the new status.
//
// Returns:
//   - Status: A new status with the Parsed bit cleared.
func (o Status) UnSetParsed() Status {
	return o & ^Parsed
}

// IsParsed returns true if bit 0, which marks the AST parse phase, is set.
//
// The method does not modify the status.
//
// Returns:
//   - bool: true when the bit is set, false otherwise.
func (o Status) IsParsed() bool {
	return o&Parsed != 0
}

// SetVendor sets bit 1 to mark the AST vendor resolution phase as completed.
//
// The Status type is a value type: the method does not modify its receiver,
// it returns the new status. The caller has to assign the result to use it.
//
// Returns:
//   - Status: A new status with the Vendor bit set.
func (o Status) SetVendor() Status {
	return o | Vendor
}

// UnSetVendor clears bit 1, which marks the AST vendor resolution phase as
// completed.
//
// The other bits of the status are left unchanged. The Status type is a value
// type: the method does not modify its receiver, it returns the new status.
//
// Returns:
//   - Status: A new status with the Vendor bit cleared.
func (o Status) UnSetVendor() Status {
	return o & ^Vendor
}

// IsVendor returns true if bit 1, which marks the AST vendor resolution phase, is set.
//
// The method does not modify the status.
//
// Returns:
//   - bool: true when the bit is set, false otherwise.
func (o Status) IsVendor() bool {
	return o&Vendor != 0
}

// SetAnalyzed sets bit 2 to mark the summary analysis phase as completed.
//
// The Status type is a value type: the method does not modify its receiver,
// it returns the new status. The caller has to assign the result to use it.
//
// Returns:
//   - Status: A new status with the Analyzed bit set.
func (o Status) SetAnalyzed() Status {
	return o | Analyzed
}

// UnSetAnalyzed clears bit 2, which marks the summary analysis phase as
// completed.
//
// The other bits of the status are left unchanged. The Status type is a value
// type: the method does not modify its receiver, it returns the new status.
//
// Returns:
//   - Status: A new status with the Analyzed bit cleared.
func (o Status) UnSetAnalyzed() Status {
	return o & ^Analyzed
}

// IsAnalyzed returns true if bit 2, which marks the summary analysis phase, is set.
//
// The method does not modify the status.
//
// Returns:
//   - bool: true when the bit is set, false otherwise.
func (o Status) IsAnalyzed() bool {
	return o&Analyzed != 0
}

// SetAudited1 sets bit 3 to mark the first audit report as generated.
//
// The Status type is a value type: the method does not modify its receiver,
// it returns the new status. The caller has to assign the result to use it.
//
// Returns:
//   - Status: A new status with the Audited1 bit set.
func (o Status) SetAudited1() Status {
	return o | Audited1
}

// UnSetAudited1 clears bit 3, which marks the first audit report as
// generated.
//
// The other bits of the status are left unchanged. The Status type is a value
// type: the method does not modify its receiver, it returns the new status.
//
// Returns:
//   - Status: A new status with the Audited1 bit cleared.
func (o Status) UnSetAudited1() Status {
	return o & ^Audited1
}

// IsAudited1 returns true if bit 3, which marks the first audit report, is set.
//
// The method does not modify the status.
//
// Returns:
//   - bool: true when the bit is set, false otherwise.
func (o Status) IsAudited1() bool {
	return o&Audited1 != 0
}

// SetAudited2 sets bit 4 to mark the second audit report as generated.
//
// The Status type is a value type: the method does not modify its receiver,
// it returns the new status. The caller has to assign the result to use it.
//
// Returns:
//   - Status: A new status with the Audited2 bit set.
func (o Status) SetAudited2() Status {
	return o | Audited2
}

// UnSetAudited2 clears bit 4, which marks the second audit report as
// generated.
//
// The other bits of the status are left unchanged. The Status type is a value
// type: the method does not modify its receiver, it returns the new status.
//
// Returns:
//   - Status: A new status with the Audited2 bit cleared.
func (o Status) UnSetAudited2() Status {
	return o & ^Audited2
}

// IsAudited2 returns true if bit 4, which marks the second audit report, is set.
//
// The method does not modify the status.
//
// Returns:
//   - bool: true when the bit is set, false otherwise.
func (o Status) IsAudited2() bool {
	return o&Audited2 != 0
}

// SetAudited3 sets bit 5 to mark the third audit report as generated.
//
// The Status type is a value type: the method does not modify its receiver,
// it returns the new status. The caller has to assign the result to use it.
//
// Returns:
//   - Status: A new status with the Audited3 bit set.
func (o Status) SetAudited3() Status {
	return o | Audited3
}

// UnSetAudited3 clears bit 5, which marks the third audit report as
// generated.
//
// The other bits of the status are left unchanged. The Status type is a value
// type: the method does not modify its receiver, it returns the new status.
//
// Returns:
//   - Status: A new status with the Audited3 bit cleared.
func (o Status) UnSetAudited3() Status {
	return o & ^Audited3
}

// IsAudited3 returns true if bit 5, which marks the third audit report, is set.
//
// The method does not modify the status.
//
// Returns:
//   - bool: true when the bit is set, false otherwise.
func (o Status) IsAudited3() bool {
	return o&Audited3 != 0
}

// SetAudited4 sets bit 6 to mark the fourth audit report as generated.
//
// The Status type is a value type: the method does not modify its receiver,
// it returns the new status. The caller has to assign the result to use it.
//
// Returns:
//   - Status: A new status with the Audited4 bit set.
func (o Status) SetAudited4() Status {
	return o | Audited4
}

// UnSetAudited4 clears bit 6, which marks the fourth audit report as
// generated.
//
// The other bits of the status are left unchanged. The Status type is a value
// type: the method does not modify its receiver, it returns the new status.
//
// Returns:
//   - Status: A new status with the Audited4 bit cleared.
func (o Status) UnSetAudited4() Status {
	return o & ^Audited4
}

// IsAudited4 returns true if bit 6, which marks the fourth audit report, is set.
//
// The method does not modify the status.
//
// Returns:
//   - bool: true when the bit is set, false otherwise.
func (o Status) IsAudited4() bool {
	return o&Audited4 != 0
}

// SetOutdated sets bit 7 to mark the item as out of date.
//
// The Status type is a value type: the method does not modify its receiver,
// it returns the new status. The caller has to assign the result to use it.
//
// Returns:
//   - Status: A new status with the Outdated bit set.
func (o Status) SetOutdated() Status {
	return o | Outdated
}

// UnSetOutdated clears bit 7, which marks the item as out of date.
//
// The other bits of the status are left unchanged. The Status type is a value
// type: the method does not modify its receiver, it returns the new status.
//
// Returns:
//   - Status: A new status with the Outdated bit cleared.
func (o Status) UnSetOutdated() Status {
	return o & ^Outdated
}

// IsOutdated returns true if bit 7, which marks the out of date flag, is set.
//
// The method does not modify the status.
//
// Returns:
//   - bool: true when the bit is set, false otherwise.
func (o Status) IsOutdated() bool {
	return o&Outdated != 0
}
