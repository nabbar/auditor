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
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	apitps "github.com/nabbar/auditor/pkg/apitype"
	audids "github.com/nabbar/auditor/pkg/data/id"
	llmcfg "github.com/nabbar/auditor/pkg/llmconfig"
	libhtc "github.com/nabbar/golib/httpcli"
)

// call orchestrates the full lifecycle of an LLM request: validation, HTTP
// construction, rate-limiting, execution, and response parsing. It returns an
// error if any step fails.
//
// Parameters:
//   - ctx: the context for cancellation and timeout control.
//   - cfg: the analysis configuration containing host, model, and options.
//   - id: the unique identifier for the audit entry being processed.
//   - prompt: the raw byte slice containing the prompt text to send to the LLM.
//
// The method performs the following steps:
//  1. Validates the configuration and prompt.
//  2. Constructs the HTTP request via callRequest, which also handles rate limiting.
//  3. Executes the HTTP request using the shared HTTP client.
//  4. Parses the response via callResponse, which updates the audit entry.
func (o *mdl) call(ctx context.Context, cfg *llmcfg.CfgEndpoint, id audids.ID, prompt []byte) error {
	if cfg == nil {
		return errors.New("llm config is nil")
	}

	if cfg.Host == nil || len(cfg.Host.Hostname) < 1 || len(cfg.Host.Model) < 1 {
		return errors.New("llm config is nil")
	}

	if id < 1 {
		return fmt.Errorf("id is invalid to perform llm request")
	}

	if len(prompt) < 1 {
		return fmt.Errorf("llm request body length is too short")
	}

	var (
		err error
		cxr context.Context
		cnl context.CancelFunc
		req *http.Request
		rsp *http.Response
		buf []byte

		tok uint16
		res uint16 = 16384

		htc = libhtc.GetClient()
		api = &apitps.Config{
			Hostname: cfg.Host.Hostname,
			Model:    cfg.Host.Model,
			Timeout:  cfg.Host.Timeout,
			Options:  cfg.Host.Options,
		}
	)

	// Ensure the cancel function is called on exit to release resources
	defer func() {
		if cnl != nil {
			cnl() // defer cancelFunc if defined
		}
	}()

	// Ensure the request body is closed on exit
	defer func() {
		if req != nil && req.Body != nil {
			_ = req.Body.Close()
		}
	}()

	// Ensure the response body is closed on exit
	defer func() {
		if rsp != nil && rsp.Body != nil {
			_ = rsp.Body.Close()
		}
	}()

	// if max tokens configured, keep as cap for result token
	// and if not defined into api options, set it
	if cfg.Limit != nil && cfg.Limit.MaxToken > 0 {
		res = cfg.Limit.MaxToken
		if _, k := api.Options[optNumPredict]; !k {
			api.Options[optNumPredict] = cfg.Limit.MaxToken
		}
	}

	// if temperature configured and if not defined into api options, set it
	if cfg.Limit != nil && cfg.Limit.Temperature > 0 {
		if _, k := api.Options[optTemperature]; !k {
			api.Options[optTemperature] = cfg.Limit.Temperature
		}
	}

	// Construct the HTTP request, applying rate limiting and timeout
	if req, tok, err = cfg.Type.Request(ctx, api, prompt); err != nil {
		o.uim.ErrorStack("llm request error", err)
		return err
	}

	// apply auth if needed
	if err = o.authApply(ctx, cfg.Auth, req); err != nil {
		return err
	}

	// Wait until the rate limiter allows a new request
	for ctx.Err() == nil {
		if o.lim.Inc(cfg.Host.Hostname, tok, res) {
			break
		}
		// waiting to send new request
		time.Sleep(100 * time.Millisecond)
	}

	// Check if the context was canceled while waiting
	if err = ctx.Err(); err != nil {
		o.uim.ErrorStack("llm context error", err)
		return err
	}

	// if timeout is defined in option used as context timeout,
	// don't update client timeout to allow different timeout for same http client
	if cfg.Host.Timeout > 0 {
		cxr, cnl = context.WithTimeout(ctx, cfg.Host.Timeout.Time())
		req = req.WithContext(cxr)
	} else {
		cxr, cnl = context.WithCancel(ctx)
		req = req.WithContext(cxr)
	}

	// Execute the HTTP request
	if rsp, err = htc.Do(req); err != nil {
		o.uim.ErrorStack("llm request do error", err)
		return err
	}

	// Process the response and parse the result
	if res, buf, err = cfg.Type.Response(ctx, api, rsp); err != nil {
		o.lim.Dec(cfg.Host.Hostname, res)
		return err
	}

	// send result []byte buf and id to dedicated model parsing
	// audids.ID given is necessary to update the target record into dbm
	return o.parseResult(buf, id)
}
