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

package llmconfig

import (
	apitps "github.com/nabbar/auditor/pkg/apitype"
	audtps "github.com/nabbar/auditor/pkg/data/types"
	libdur "github.com/nabbar/golib/duration"
)

// Config represents the top-level LLM configuration. It contains a default
// configuration block and a map of per-language overrides keyed by language
// name.
type Config struct {
	// Default holds the base configuration applied to all operations unless
	// overridden by language-specific or type-specific settings.
	Default CfgDefault `json:"default" yaml:"default" toml:"default"`

	// Lang maps language names (e.g. "go", "python") to their respective
	// per-language configuration overrides.
	Lang map[string]CfgLang `json:"lang" yaml:"lang" toml:"lang"`
}

// CfgDefault holds the default LLM configuration applied globally. It
// provides base host, auth, and options settings, along with optional
// per-operation overrides for analysis and reporting.
type CfgDefault struct {
	// Type defines the default LLM endpoint API type.
	Type apitps.ApiType `json:"type" yaml:"type" toml:"type"`

	// Host defines the default LLM endpoint API settings (hostname, model, timeout, etc.).
	Host *apitps.Config `json:"api" yaml:"api" toml:"api"`

	// Auth defines the default authentication method for LLM requests.
	Auth *CfgAuth `json:"auth" yaml:"auth" toml:"auth"`

	// Limit defines the default request parameters (token limits,
	// temperature, rate limits).
	Limit *CfgLimit `json:"limit" yaml:"limit" toml:"limit"`

	// Analyze provides optional overrides specific to the analyze operation.
	Analyze *CfgEndpoint `json:"analyze" yaml:"analyze" toml:"analyze"`

	// Report maps report names to their specific configuration overrides.
	Report map[string]CfgEndpoint `json:"report" yaml:"report" toml:"report"`
}

// CfgLang holds per-language LLM configuration overrides. It allows
// different languages to use distinct endpoints, models, or authentication
// methods.
type CfgLang struct {
	// Lang is the language identifier (e.g. "go", "python").
	Lang string `json:"lang" yaml:"lang" toml:"lang"`

	// Type defines the default LLM endpoint API type.
	Type apitps.ApiType `json:"type" yaml:"type" toml:"type"`

	// Host defines the default LLM endpoint API settings (hostname, model, timeout, etc.).
	Host *apitps.Config `json:"api" yaml:"api" toml:"api"`

	// Auth defines the authentication method for this language.
	Auth *CfgAuth `json:"auth" yaml:"auth" toml:"auth"`

	// Limit defines the request parameters for this language.
	Limit *CfgLimit `json:"limit" yaml:"limit" toml:"limit"`

	// Analyze provides optional overrides specific to the analyze operation
	// for this language.
	Analyze *CfgEndpoint `json:"analyze" yaml:"analyze" toml:"analyze"`

	// Report maps report names to their specific configuration overrides
	// for this language.
	Report map[string]CfgEndpoint `json:"report" yaml:"report" toml:"report"`

	// Types maps code types (e.g. "function", "struct") to their
	// per-type configuration overrides within this language.
	Types map[audtps.CodeType]CfgLangTypes `json:"types" yaml:"types" toml:"types"`
}

// CfgLangTypes holds per-code-type LLM configuration overrides within a
// specific language. It allows different code constructs to use distinct
// endpoints, models, or authentication methods.
type CfgLangTypes struct {
	// Lang is the language identifier this type belongs to.
	Lang string `json:"lang" yaml:"lang" toml:"lang"`

	// Code is the code type identifier (e.g. "function", "struct").
	Code audtps.CodeType `json:"code" yaml:"code" toml:"code"`

	// Type defines the default LLM endpoint API type.
	Type apitps.ApiType `json:"type" yaml:"type" toml:"type"`

	// Host defines the default LLM endpoint API settings (hostname, model, timeout, etc.).
	Host *apitps.Config `json:"api" yaml:"api" toml:"api"`

	// Auth defines the authentication method for this code type.
	Auth *CfgAuth `json:"auth" yaml:"auth" toml:"auth"`

	// Limit defines the request parameters for this code type.
	Limit *CfgLimit `json:"limit" yaml:"limit" toml:"limit"`

	// Analyze provides optional overrides specific to the analyze operation
	// for this code type.
	Analyze *CfgEndpoint `json:"analyze" yaml:"analyze" toml:"analyze"`

	// Report maps report names to their specific configuration overrides
	// for this code type.
	Report map[string]CfgEndpoint `json:"report" yaml:"report" toml:"report"`
}

// CfgEndpoint holds configuration overrides specific to an operation
// (e.g., analyze, report). It can provide a distinct host, auth, or options
// set for that operation's requests.
type CfgEndpoint struct {
	// Type defines the default LLM endpoint API type.
	Type apitps.ApiType `json:"type" yaml:"type" toml:"type"`

	// Host defines the default LLM endpoint API settings (hostname, model, timeout, etc.).
	Host *apitps.Config `json:"api" yaml:"api" toml:"api"`

	// Auth defines the authentication method for the operation.
	Auth *CfgAuth `json:"auth" yaml:"auth" toml:"auth"`

	// Limit defines the request parameters for the operation.
	Limit *CfgLimit `json:"limit" yaml:"limit" toml:"limit"`
}

// CfgLimit defines the request parameters for LLM calls, including
// token limits, temperature, and rate limiting.
type CfgLimit struct {
	// MaxToken is the maximum number of tokens allowed in a single
	// LLM response.
	MaxToken uint16 `json:"maxToken" yaml:"maxToken" toml:"maxToken"`

	// Temperature controls the randomness of the LLM output. Higher
	// values produce more diverse results.
	Temperature float32 `json:"temperature" yaml:"temperature" toml:"temperature"`

	// MaxReqSecond is the maximum number of requests allowed per
	// second to the LLM endpoint.
	MaxReqSecond uint16 `json:"max-request-second" yaml:"max-request-second" toml:"max-request-second"`

	// MaxReqTokHour is the maximum number of tokens allowed per hour
	// across all LLM requests.
	MaxReqTokHour uint64 `json:"max-request-token-hour" yaml:"max-request-token-hour" toml:"max-request-token-hour"`

	// MaxRespTokHour is the maximum number of tokens allowed per hour
	// across all LLM responses.
	MaxRespTokHour uint64 `json:"max-response-token-hour" yaml:"max-response-token-hour" toml:"max-response-token-hour"`
}

// CfgAuth defines the authentication method for LLM requests. Exactly
// one of Basic, Api, or Oauth should be non-nil.
type CfgAuth struct {
	// Basic holds HTTP Basic authentication credentials.
	Basic *AuthBasic `json:"basic" yaml:"basic" toml:"basic"`

	// Api holds API key authentication credentials.
	Api *AuthAPI `json:"api" yaml:"api" toml:"api"`

	// Oauth holds OAuth 2.0 authentication credentials.
	Oauth *AuthOauth `json:"oauth" yaml:"oauth" toml:"oauth"`
}

// AuthBasic holds HTTP Basic authentication credentials with a
// username and password.
type AuthBasic struct {
	// Username is the authentication username.
	Username string `json:"username" yaml:"username" toml:"username"`

	// Password is the authentication password.
	Password string `json:"password" yaml:"password" toml:"password"`
}

// AuthAPI holds API key authentication credentials.
type AuthAPI struct {
	// ApiKey is the API key used for authentication.
	ApiKey string `json:"api-key" yaml:"api-key" toml:"api-key"`
}

// AuthOauth holds OAuth 2.0 authentication credentials.
type AuthOauth struct {
	// LoginUri is the OAuth 2.0 authorization endpoint URL.
	LoginUri string `json:"login-uri" yaml:"login-uri" toml:"login-uri"`

	// TimeoutLogin is the maximum duration allowed for the OAuth 2.0
	// token exchange request.
	TimeoutLogin libdur.Duration `json:"timeout-login" yaml:"timeout-login" toml:"timeout-login"`

	// ClientId is the OAuth 2.0 client identifier.
	ClientId string `json:"client-id" yaml:"client-id" toml:"client-id"`

	// SecretId is the OAuth 2.0 client secret.
	SecretId string `json:"secret-id" yaml:"secret-id" toml:"secret-id"`

	// GrantType is the OAuth 2.0 grant type (e.g. "client_credentials",
	// "authorization_code").
	GrantType string `json:"grant-type" yaml:"grant-type" toml:"grant-type"`
}

// merge combines the non-empty fields from the source CfgEndpoint into
// the destination CfgEndpoint. It performs a deep merge where only
// non-zero, non-nil, and non-empty values from the source override the
// corresponding fields in the destination. If a field in the destination
// is nil and the source provides a non-nil value, a new instance is
// created before copying the source values.
//
// This merge strategy allows partial overrides: if the source Host is
// non-nil, the destination Host is reset to a zero-value CfgHost, and
// only the non-empty fields from the source are copied. This enables
// callers to explicitly clear specific sub-fields (e.g., removing a
// timeout or model) by providing a non-nil source with empty values.
//
// Parameters:
//   - res: the destination CfgEndpoint to be updated with merged values.
//   - cfg: the source CfgEndpoint providing override values.
func (o Config) merge(res *CfgEndpoint, cfg CfgEndpoint) {
	if cfg.Host != nil {
		res.Type = cfg.Type

		// If cfg.Host is defined, reset result to allow setting some option
		// to an empty value (e.g., remove timeout or model).
		res.Host = &apitps.Config{}

		if cfg.Host.Hostname != "" {
			res.Host.Hostname = cfg.Host.Hostname
		}

		if cfg.Host.Model != "" {
			res.Host.Model = cfg.Host.Model
		}

		if cfg.Host.Timeout > 0 {
			res.Host.Timeout = cfg.Host.Timeout
		}

		if len(cfg.Host.Options) > 0 {
			res.Host.Options = make(map[string]interface{})
			for k, v := range cfg.Host.Options {
				res.Host.Options[k] = v
			}
		}
	}

	if cfg.Auth != nil {
		// If cfg.Auth is defined, reset result to allow setting some option
		// to an empty value (e.g., changing auth type, or no auth for some
		// request).
		res.Auth = &CfgAuth{}

		if cfg.Auth.Api != nil {
			if cfg.Auth.Api.ApiKey != "" {
				res.Auth.Api = &AuthAPI{
					ApiKey: cfg.Auth.Api.ApiKey,
				}
			}
		} else if cfg.Auth.Basic != nil {
			if cfg.Auth.Basic.Username != "" || cfg.Auth.Basic.Password != "" {
				res.Auth.Basic = &AuthBasic{}
				if cfg.Auth.Basic.Username != "" {
					res.Auth.Basic.Username = cfg.Auth.Basic.Username
				}
				if cfg.Auth.Basic.Password != "" {
					res.Auth.Basic.Password = cfg.Auth.Basic.Password
				}
			}
		} else if cfg.Auth.Oauth != nil {
			if cfg.Auth.Oauth.LoginUri != "" || cfg.Auth.Oauth.ClientId != "" || cfg.Auth.Oauth.SecretId != "" {
				res.Auth.Oauth = &AuthOauth{}
				if cfg.Auth.Oauth.LoginUri != "" {
					res.Auth.Oauth.LoginUri = cfg.Auth.Oauth.LoginUri
				}
				if cfg.Auth.Oauth.ClientId != "" {
					res.Auth.Oauth.ClientId = cfg.Auth.Oauth.ClientId
				}
				if cfg.Auth.Oauth.SecretId != "" {
					res.Auth.Oauth.SecretId = cfg.Auth.Oauth.SecretId
				}
				if cfg.Auth.Oauth.GrantType != "" {
					res.Auth.Oauth.GrantType = cfg.Auth.Oauth.GrantType
				}
				if cfg.Auth.Oauth.TimeoutLogin > 0 {
					res.Auth.Oauth.TimeoutLogin = cfg.Auth.Oauth.TimeoutLogin
				}
			}
		}
	}

	if cfg.Limit != nil {
		// If cfg.Options is defined, reset result to allow setting some
		// option to an empty value (e.g., no max token for local endpoint).
		res.Limit = &CfgLimit{}

		if cfg.Limit.MaxToken != 0 {
			res.Limit.MaxToken = cfg.Limit.MaxToken
		}
		if cfg.Limit.MaxReqSecond != 0 {
			res.Limit.MaxReqSecond = cfg.Limit.MaxReqSecond
		}
		if cfg.Limit.MaxReqTokHour != 0 {
			res.Limit.MaxReqTokHour = cfg.Limit.MaxReqTokHour
		}
		if cfg.Limit.MaxRespTokHour != 0 {
			res.Limit.MaxRespTokHour = cfg.Limit.MaxRespTokHour
		}
		if cfg.Limit.Temperature != 0 {
			res.Limit.Temperature = cfg.Limit.Temperature
		}
	}
}
