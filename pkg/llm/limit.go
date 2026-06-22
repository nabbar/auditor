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
	"sync"
	"time"
)

// mrq represents a per-host rate-limiting tracker that enforces three independent
// quotas:
//   - Requests per second (cntReq / maxReq)
//   - Request tokens per hour (cntTkq / maxTkq)
//   - Response tokens per hour (cntTks / maxTks)
//
// Each quota is tracked in its own time window (1 second or 1 hour). When a window
// elapses, the corresponding counter is reset to zero. A call to Inc succeeds
// (returns true) only if all three quotas would remain within their limits after
// the increment; otherwise it returns false and no counters are modified.
type mrq struct {
	// tss is the timestamp of the last second-window reset. It is updated whenever
	// a full second has elapsed since the previous reset.
	tss time.Time

	// tsh is the timestamp of the last hour-window reset. It is updated whenever
	// a full hour has elapsed since the previous reset.
	tsh time.Time

	// cntReq is the current request count within the active second window.
	cntReq uint64

	// cntTkq is the accumulated request-token count within the active hour window.
	cntTkq uint64

	// cntTks is the accumulated response-token count within the active hour window.
	cntTks uint64

	// maxReq is the maximum allowed requests per second. A value of 0 disables
	// the per-second request limit.
	maxReq uint64

	// maxTkq is the maximum allowed request tokens per hour. A value of 0 disables
	// the per-hour request-token limit.
	maxTkq uint64

	// maxTks is the maximum allowed response tokens per hour. A value of 0 disables
	// the per-hour response-token limit.
	maxTks uint64
}

// Inc attempts to increment the rate-limiting counters by treq request tokens and
// trsp response tokens. It returns true if all three quotas (per-second requests,
// per-hour request tokens, per-hour response tokens) would remain within their
// respective limits after the increment; in that case the counters are updated.
// It returns false if any quota would be exceeded, and no counters are modified.
//
// The method performs the following checks in order:
//  1. If a full second has elapsed since tss, the second window is reset.
//  2. If a full hour has elapsed since tsh, the hour window is reset.
//  3. The per-second request limit is checked against cntReq.
//  4. The per-hour request-token limit is checked against cntTkq + treq.
//  5. The per-hour response-token limit is checked against cntTks + trsp.
//
// If all checks pass, the counters are incremented and true is returned.
func (o *mrq) Inc(treq, trsp uint16) bool {
	var (
		req = true
		tkq = true
		tks = true
	)

	// Initialize the second window timestamp if not set
	if o.tss.IsZero() {
		o.tss = time.Now()
		o.cntReq = 0
	}

	// Initialize the hour window timestamp if not set
	if o.tsh.IsZero() {
		o.tsh = time.Now()
		o.cntTkq = 0
	}

	// Reset the second window if a full second has elapsed
	if time.Since(o.tss) >= time.Second {
		o.tss = time.Now()
		o.cntReq = 0
	}

	// Reset the hour window if a full hour has elapsed
	if time.Since(o.tsh) >= time.Hour {
		o.tsh = time.Now()
		o.cntTkq = 0
		o.cntTks = 0
	}

	// Check if the request count limit would be exceeded
	if o.maxReq > 0 {
		req = false
		if o.maxReq-1 > o.cntReq {
			req = true
		}
	}

	// Check if the request token limit would be exceeded
	if o.maxTkq > 0 && treq > 0 {
		tkq = false
		if o.maxTkq-uint64(treq) > o.cntTkq {
			tkq = true
		}
	}

	// Check if the response token limit would be exceeded
	if o.maxTks > 0 && trsp > 0 {
		tks = false
		if o.maxTks-uint64(trsp) > o.cntTks {
			tks = true
		}
	}

	// If all limits are satisfied, increment the counters
	if req && tkq && tks {
		if o.maxReq > 0 {
			o.cntReq++
		}
		if o.maxTkq > 0 {
			o.cntTkq += uint64(treq)
		}
		if o.maxTks > 0 {
			o.cntTks += uint64(trsp)
		}
		return true
	}

	return false
}

// Dec decrements the response-token counter (cntTks) by t tokens. This is intended
// to release capacity when a previously counted response is no longer valid (e.g.,
// after a failed or cancelled request). The decrement is only applied if both the
// limit (maxTks) and the current counter (cntTks) are greater than t, preventing
// underflow.
func (o *mrq) Dec(t uint16) {
	if o.maxTks > 0 && t > 0 {
		if o.maxTks >= uint64(t) && o.cntTks >= uint64(t) {
			o.cntTks -= uint64(t)
			return
		}
	}
}

// Reset clears all rate-limiting state by resetting both time windows to the
// current time and zeroing all counters. This is useful for testing or when
// a host's quota should be fully refreshed.
func (o *mrq) Reset() {
	o.tss = time.Now()
	o.tsh = time.Now()
	o.cntReq = 0
	o.cntTkq = 0
	o.cntTks = 0
}

// lmt is a thread-safe registry of per-host rate-limiting trackers. It maps host
// names to their respective mrq instances and provides methods to add hosts,
// increment counters, and decrement response-token counters.
type lmt struct {
	// mux protects concurrent access to the lst map.
	mux sync.Mutex

	// lst maps host names to their respective rate-limiting trackers.
	lst map[string]*mrq
}

// Add registers a new rate-limiting tracker for the given host with the specified
// limits. If the host is already registered, the existing tracker is left unchanged.
//
// Parameters:
//   - host: the host name to register.
//   - maxReq: maximum allowed requests per second (0 to disable).
//   - maxTokReq: maximum allowed request tokens per hour (0 to disable).
//   - maxTokResp: maximum allowed response tokens per hour (0 to disable).
func (o *lmt) Add(host string, maxReq, maxTokReq, maxTokResp uint64) {
	o.mux.Lock()
	defer o.mux.Unlock()

	if o.lst == nil {
		o.lst = make(map[string]*mrq)
	}

	// Only create a new tracker if the host is not already registered
	if _, k := o.lst[host]; !k {
		o.lst[host] = &mrq{
			tss:    time.Now(),
			tsh:    time.Now(),
			cntReq: 0,
			cntTkq: 0,
			cntTks: 0,
			maxReq: maxReq,
			maxTkq: maxTokReq,
			maxTks: maxTokResp,
		}
	}
}

// Inc attempts to increment the rate-limiting counters for the given host by
// tokReq request tokens and tokResp response tokens. It returns true if the
// increment succeeds (all quotas are within limits); false otherwise.
//
// If the host is not registered in the registry, the call always returns true
// (no rate limiting is applied for unknown hosts).
func (o *lmt) Inc(host string, tokReq, tokResp uint16) bool {
	o.mux.Lock()
	defer o.mux.Unlock()

	if _, k := o.lst[host]; k {
		if !o.lst[host].Inc(tokReq, tokResp) {
			return false
		}
	}

	return true
}

// Dec decrements the response-token counter for the given host by tok tokens.
// If the host is not registered, the call is a no-op.
func (o *lmt) Dec(host string, tok uint16) {
	o.mux.Lock()
	defer o.mux.Unlock()

	if _, k := o.lst[host]; k {
		o.lst[host].Dec(tok)
	}
}
