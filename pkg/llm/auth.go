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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	llmcfg "github.com/nabbar/auditor/pkg/llmconfig"
	libhtc "github.com/nabbar/golib/httpcli"
)

// authToken represents a cached authentication token with its expiration time.
type authToken struct {
	token  string
	expire time.Time
}

// authCache is a thread-safe in-memory cache for authentication tokens.
// It uses a read-write mutex to allow concurrent reads while serializing writes.
type authCache struct {
	mux sync.RWMutex
	tok map[string]authToken
}

// Get retrieves a valid token from the cache for the given key.
// A token is considered valid only if it exists and has not expired.
// A 30-second safety margin is applied before the actual expiration time
// to avoid using a token that is about to expire.
func (o *authCache) Get(key string) (string, bool) {
	o.mux.RLock()
	defer o.mux.RUnlock()

	// Return early if the cache is empty
	if len(o.tok) < 1 {
		return "", false
	}

	tk, ok := o.tok[key]
	if !ok {
		return "", false
	}

	// Check if the token is about to expire (within 30 seconds)
	if time.Now().Add(30 * time.Second).After(tk.expire) {
		return "", false
	}

	return tk.token, ok
}

// Set stores a token in the cache with the given key and expiration time.
// If the cache map is nil, it is initialized on first use.
func (o *authCache) Set(key string, token string, expire time.Time) {
	o.mux.Lock()
	defer o.mux.Unlock()

	if len(o.tok) < 1 {
		o.tok = make(map[string]authToken)
	}

	o.tok[key] = authToken{token, expire}
}

// authApply dispatches the authentication logic based on the configured strategy.
// It supports three authentication methods:
//   - Basic Auth (username/password)
//   - API Key (Bearer token)
//   - OAuth2 (token exchange with an authorization server)
//
// If no authentication configuration is provided, the request is left unchanged.
func (o *mdl) authApply(ctx context.Context, cfg *llmcfg.CfgAuth, req *http.Request) error {
	if cfg == nil {
		return nil
	}

	if cfg.Basic != nil {
		return o.authBasic(ctx, cfg.Basic, req)
	} else if cfg.Api != nil {
		return o.authApi(ctx, cfg.Api, req)
	} else if cfg.Oauth != nil {
		return o.authOAuth2(ctx, cfg.Oauth, req)
	}

	return nil
}

// authBasic applies HTTP Basic Authentication to the request.
// If either the username or password is non-empty, the Authorization header
// is set using the standard Basic scheme.
func (o *mdl) authBasic(_ context.Context, cfg *llmcfg.AuthBasic, req *http.Request) error {
	if cfg == nil {
		return nil
	}

	if len(cfg.Username) > 0 || len(cfg.Password) > 0 {
		req.SetBasicAuth(cfg.Username, cfg.Password)
	}

	return nil
}

// authApi applies API Key authentication by setting the Authorization header
// with a Bearer token containing the provided API key.
func (o *mdl) authApi(_ context.Context, cfg *llmcfg.AuthAPI, req *http.Request) error {
	if cfg == nil {
		return nil
	}

	if len(cfg.ApiKey) > 0 {
		req.Header.Set(hdrAuthorization, fmt.Sprintf(hdrBearer, cfg.ApiKey))
	}

	return nil
}

// authOAuth2 handles OAuth2 authentication by retrieving or regenerating
// an access token. It first checks the cache for a valid token; if none
// exists or the token has expired, it triggers a token regeneration flow.
// The resulting token is then attached to the request as a Bearer header.
func (o *mdl) authOAuth2(ctx context.Context, cfg *llmcfg.AuthOauth, req *http.Request) error {
	if cfg == nil {
		return nil
	}

	var (
		ok bool
		tk string

		err error
		// The cache key is composed of the login URI and client ID to ensure
		// tokens are scoped per OAuth2 client configuration.
		key = fmt.Sprintf("%s|%s", cfg.LoginUri, cfg.ClientId)
	)

	// Attempt to retrieve a valid token from the cache
	if tk, ok = o.auth.Get(key); !ok {
		// No valid token found; regenerate it
		if err = o.authOAuth2Regen(ctx, cfg, key); err != nil {
			return err
		}
		// After regeneration, verify the token is now available and valid
		if tk, ok = o.auth.Get(key); !ok {
			o.uim.Error("auth oauth2: request new token has success, but token still not exits or valid")
			return errors.New("auth oauth2: request new token has success, but token still not exits or valid")
		}
	}

	req.Header.Set(hdrAuthorization, fmt.Sprintf(hdrBearer, tk))
	return nil
}

// authOAuth2Regen performs the OAuth2 token exchange with the authorization server.
// It sends a POST request with form-encoded credentials (client_id, client_secret,
// grant_type) to the configured login URI, parses the JSON response, and caches
// the obtained access token with its expiration time.
//
// The function uses deferred cleanup for context cancellation, request body,
// and response body to prevent resource leaks.
func (o *mdl) authOAuth2Regen(ctx context.Context, cfg *llmcfg.AuthOauth, key string) error {
	if cfg == nil {
		return errors.New("invalid config OAuth2")
	}

	// Mod represents the expected JSON structure of the OAuth2 token response.
	type Mod struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"` // Token validity duration in seconds
	}

	var (
		err error
		buf []byte
		cnl context.CancelFunc
		dat = url.Values{}
		htc = libhtc.GetClient()
		req *http.Request
		rsp *http.Response
		mod Mod
	)

	// Ensure context cancellation is always called to prevent leaks
	defer func() {
		if cnl != nil {
			cnl()
		}
	}()

	// Ensure the request body is closed after use
	defer func() {
		if req != nil && req.Body != nil {
			_ = req.Body.Close()
		}
	}()

	// Ensure the response body is closed after use
	defer func() {
		if rsp != nil && rsp.Body != nil {
			_ = rsp.Body.Close()
		}
	}()

	// Build the x-www-form-urlencoded payload for the OAuth2 token request
	dat.Set("grant_type", cfg.GrantType)
	dat.Set("client_id", cfg.ClientId)
	dat.Set("client_secret", cfg.SecretId)

	// Create the HTTP POST request to the OAuth2 token endpoint
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, cfg.LoginUri, strings.NewReader(dat.Encode())); err != nil {
		o.uim.ErrorStack("oauth login request error: %w", err)
		return err
	}

	// Set required headers for form-encoded POST request
	req.Header.Set(hrdContent, hdrForm)
	req.Header.Set(hrdAccept, hrdJson)

	// Apply a timeout to the login request if configured; otherwise use a cancellable context
	if cfg.TimeoutLogin > 0 {
		ctx, cnl = context.WithTimeout(ctx, cfg.TimeoutLogin.Time())
		req = req.WithContext(ctx)
	} else {
		ctx, cnl = context.WithCancel(ctx)
		req = req.WithContext(ctx)
	}

	// Execute the HTTP request using the module's HTTP client
	if rsp, err = htc.Do(req); err != nil {
		o.uim.ErrorStack("oauth login request error: %w", err)
		return err
	}

	// Validate the HTTP response status code (expect 2xx success)
	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		o.uim.Error("oauth login request status: %s", rsp.Status)
		return fmt.Errorf("auth server returned status %d", rsp.StatusCode)
	}

	// Read the response body into a buffer
	if rsp.Body != nil {
		if buf, err = io.ReadAll(rsp.Body); err != nil {
			o.uim.ErrorStack("oauth reading response login error: %w", err)
			return err
		}
	}

	// Parse the JSON response into the Mod struct
	if err = json.Unmarshal(buf, &mod); err != nil {
		o.uim.ErrorStack("oauth unmarshal response login error: %w", err)
		return err
	}

	// Validate that the access token is present in the response
	if len(mod.AccessToken) < 1 {
		o.uim.Error("oauth login request access token is empty")
		return fmt.Errorf("auth response contained empty access token")
	}

	// Cache the token with its expiration time derived from the ExpiresIn field
	o.auth.Set(key, mod.AccessToken, time.Now().Add(time.Duration(mod.ExpiresIn)*time.Second))
	return nil
}
