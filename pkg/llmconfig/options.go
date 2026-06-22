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
	audpkg "github.com/nabbar/auditor/pkg/generic"
)

// Analyze resolves the effective CfgEndpoint for a given language and
// code type by merging configuration layers in order of precedence:
//
//  1. Default global settings (Default.Host, Default.Auth, Default.Options).
//  2. Default analyze overrides (Default.Analyze).
//  3. Language-specific settings (Lang[lng]).
//  4. Language-specific analyze overrides (Lang[lng].Analyze).
//  5. Code-type-specific settings (Lang[lng].Types[code]).
//  6. Code-type-specific analyze overrides (Lang[lng].Types[code].Analyze).
//
// If the language is not found, it falls back to the audpkg.Unknown language
// entry. If the code type is not found, it falls back to the EntryNone
// type. Each layer overrides only the non-empty fields from the previous
// layer via the merge method.
//
// Parameters:
//   - lng: the language identifier (e.g. "go", "python").
//   - code: the code type identifier (e.g. "function", "struct").
//
// Returns a pointer to the resolved CfgEndpoint with all applicable
// overrides applied.
func (o Config) Analyze(lng string, code audtps.CodeType) *CfgEndpoint {
	// TODO: factorise to optimize it and reduce redundant code
	var cfg = CfgEndpoint{
		Type:  apitps.ApiOllama,
		Host:  nil,
		Auth:  nil,
		Limit: nil,
	}

	o.merge(&cfg, CfgEndpoint{
		Type:  o.Default.Type,
		Host:  o.Default.Host,
		Auth:  o.Default.Auth,
		Limit: o.Default.Limit,
	})

	if o.Default.Analyze != nil {
		o.merge(&cfg, CfgEndpoint{
			Type:  o.Default.Analyze.Type,
			Host:  o.Default.Analyze.Host,
			Auth:  o.Default.Analyze.Auth,
			Limit: o.Default.Analyze.Limit,
		})
	}

	if l, k := o.Lang[lng]; k {
		o.merge(&cfg, CfgEndpoint{
			Type:  l.Type,
			Host:  l.Host,
			Auth:  l.Auth,
			Limit: l.Limit,
		})

		if l.Analyze != nil {
			o.merge(&cfg, CfgEndpoint{
				Type:  l.Analyze.Type,
				Host:  l.Analyze.Host,
				Auth:  l.Analyze.Auth,
				Limit: l.Analyze.Limit,
			})
		}

		if c, k := l.Types[code]; k {
			o.merge(&cfg, CfgEndpoint{
				Type:  c.Type,
				Host:  c.Host,
				Auth:  c.Auth,
				Limit: c.Limit,
			})

			if c.Analyze != nil {
				o.merge(&cfg, CfgEndpoint{
					Type:  c.Analyze.Type,
					Host:  c.Analyze.Host,
					Auth:  c.Analyze.Auth,
					Limit: c.Analyze.Limit,
				})
			}
		} else if c, k = l.Types[audtps.EntryNone]; k {
			o.merge(&cfg, CfgEndpoint{
				Type:  c.Type,
				Host:  c.Host,
				Auth:  c.Auth,
				Limit: c.Limit,
			})

			if c.Analyze != nil {
				o.merge(&cfg, CfgEndpoint{
					Type:  c.Analyze.Type,
					Host:  c.Analyze.Host,
					Auth:  c.Analyze.Auth,
					Limit: c.Analyze.Limit,
				})
			}
		}
	} else if l, k = o.Lang[audpkg.Unknown]; k {
		o.merge(&cfg, CfgEndpoint{
			Type:  l.Type,
			Host:  l.Host,
			Auth:  l.Auth,
			Limit: l.Limit,
		})

		if l.Analyze != nil {
			o.merge(&cfg, CfgEndpoint{
				Type:  l.Analyze.Type,
				Host:  l.Analyze.Host,
				Auth:  l.Analyze.Auth,
				Limit: l.Analyze.Limit,
			})
		}

		if c, k := l.Types[code]; k {
			o.merge(&cfg, CfgEndpoint{
				Type:  c.Type,
				Host:  c.Host,
				Auth:  c.Auth,
				Limit: c.Limit,
			})

			if c.Analyze != nil {
				o.merge(&cfg, CfgEndpoint{
					Type:  c.Analyze.Type,
					Host:  c.Analyze.Host,
					Auth:  c.Analyze.Auth,
					Limit: c.Analyze.Limit,
				})
			}
		} else if c, k = l.Types[audtps.EntryNone]; k {
			o.merge(&cfg, CfgEndpoint{
				Type:  c.Type,
				Host:  c.Host,
				Auth:  c.Auth,
				Limit: c.Limit,
			})

			if c.Analyze != nil {
				o.merge(&cfg, CfgEndpoint{
					Type:  c.Analyze.Type,
					Host:  c.Analyze.Host,
					Auth:  c.Analyze.Auth,
					Limit: c.Analyze.Limit,
				})
			}
		}
	}

	return &cfg
}

// Report resolves the effective CfgEndpoint for a given language,
// code type, and report name by merging configuration layers in order
// of precedence:
//
//  1. Default global settings (Default.Host, Default.Auth, Default.Options).
//  2. Default report overrides for the given name (Default.Report[name]).
//  3. Language-specific settings (Lang[lng]).
//  4. Language-specific report overrides (Lang[lng].Report[name]).
//  5. Code-type-specific settings (Lang[lng].Types[code]).
//  6. Code-type-specific report overrides (Lang[lng].Types[code].Report[name]).
//
// If the language is not found, it falls back to the audpkg.Unknown language
// entry. If the code type is not found, it falls back to the EntryNone
// type. If the report name is not found, it falls back to the audpkg.Unknown
// report entry. Each layer overrides only the non-empty fields from the
// previous layer via the merge method.
//
// Parameters:
//   - lng: the language identifier (e.g. "go", "python").
//   - code: the code type identifier (e.g. "function", "struct").
//   - name: the report name identifier.
//
// Returns a pointer to the resolved CfgEndpoint with all applicable
// overrides applied.
func (o Config) Report(lng string, code audtps.CodeType, name string) *CfgEndpoint {
	// TODO: factorise to optimize it and reduce redundant code
	var cfg = CfgEndpoint{
		Type:  apitps.ApiOllama,
		Host:  nil,
		Auth:  nil,
		Limit: nil,
	}

	o.merge(&cfg, CfgEndpoint{
		Type:  o.Default.Type,
		Host:  o.Default.Host,
		Auth:  o.Default.Auth,
		Limit: o.Default.Limit,
	})

	if len(o.Default.Report) > 0 {
		if v, k := o.Default.Report[name]; k {
			o.merge(&cfg, CfgEndpoint{
				Type:  v.Type,
				Host:  v.Host,
				Auth:  v.Auth,
				Limit: v.Limit,
			})
		} else if v, k = o.Default.Report[audpkg.Unknown]; k {
			o.merge(&cfg, CfgEndpoint{
				Type:  v.Type,
				Host:  v.Host,
				Auth:  v.Auth,
				Limit: v.Limit,
			})
		}
	}

	if l, k := o.Lang[lng]; k {
		o.merge(&cfg, CfgEndpoint{
			Type:  l.Type,
			Host:  l.Host,
			Auth:  l.Auth,
			Limit: l.Limit,
		})

		if len(l.Report) > 0 {
			if v, k := l.Report[name]; k {
				o.merge(&cfg, CfgEndpoint{
					Type:  v.Type,
					Host:  v.Host,
					Auth:  v.Auth,
					Limit: v.Limit,
				})
			} else if v, k = l.Report[audpkg.Unknown]; k {
				o.merge(&cfg, CfgEndpoint{
					Type:  v.Type,
					Host:  v.Host,
					Auth:  v.Auth,
					Limit: v.Limit,
				})
			}
		}

		if c, k := l.Types[code]; k {
			o.merge(&cfg, CfgEndpoint{
				Type:  c.Type,
				Host:  c.Host,
				Auth:  c.Auth,
				Limit: c.Limit,
			})

			if len(c.Report) > 0 {
				if v, k := c.Report[name]; k {
					o.merge(&cfg, CfgEndpoint{
						Type:  v.Type,
						Host:  v.Host,
						Auth:  v.Auth,
						Limit: v.Limit,
					})
				} else if v, k = c.Report[audpkg.Unknown]; k {
					o.merge(&cfg, CfgEndpoint{
						Type:  v.Type,
						Host:  v.Host,
						Auth:  v.Auth,
						Limit: v.Limit,
					})
				}
			}
		} else if c, k = l.Types[audtps.EntryNone]; k {
			o.merge(&cfg, CfgEndpoint{
				Type:  c.Type,
				Host:  c.Host,
				Auth:  c.Auth,
				Limit: c.Limit,
			})

			if len(c.Report) > 0 {
				if v, k := c.Report[name]; k {
					o.merge(&cfg, CfgEndpoint{
						Type:  v.Type,
						Host:  v.Host,
						Auth:  v.Auth,
						Limit: v.Limit,
					})
				} else if v, k = c.Report[audpkg.Unknown]; k {
					o.merge(&cfg, CfgEndpoint{
						Type:  v.Type,
						Host:  v.Host,
						Auth:  v.Auth,
						Limit: v.Limit,
					})
				}
			}
		}
	} else if l, k = o.Lang[audpkg.Unknown]; k {
		o.merge(&cfg, CfgEndpoint{
			Type:  l.Type,
			Host:  l.Host,
			Auth:  l.Auth,
			Limit: l.Limit,
		})

		if len(l.Report) > 0 {
			if v, k := l.Report[name]; k {
				o.merge(&cfg, CfgEndpoint{
					Type:  v.Type,
					Host:  v.Host,
					Auth:  v.Auth,
					Limit: v.Limit,
				})
			} else if v, k = l.Report[audpkg.Unknown]; k {
				o.merge(&cfg, CfgEndpoint{
					Type:  v.Type,
					Host:  v.Host,
					Auth:  v.Auth,
					Limit: v.Limit,
				})
			}
		}

		if c, k := l.Types[code]; k {
			o.merge(&cfg, CfgEndpoint{
				Type:  c.Type,
				Host:  c.Host,
				Auth:  c.Auth,
				Limit: c.Limit,
			})

			if len(c.Report) > 0 {
				if v, k := c.Report[name]; k {
					o.merge(&cfg, CfgEndpoint{
						Type:  v.Type,
						Host:  v.Host,
						Auth:  v.Auth,
						Limit: v.Limit,
					})
				} else if v, k = c.Report[audpkg.Unknown]; k {
					o.merge(&cfg, CfgEndpoint{
						Type:  v.Type,
						Host:  v.Host,
						Auth:  v.Auth,
						Limit: v.Limit,
					})
				}
			}
		} else if c, k = l.Types[audtps.EntryNone]; k {
			o.merge(&cfg, CfgEndpoint{
				Type:  c.Type,
				Host:  c.Host,
				Auth:  c.Auth,
				Limit: c.Limit,
			})

			if len(c.Report) > 0 {
				if v, k := c.Report[name]; k {
					o.merge(&cfg, CfgEndpoint{
						Type:  v.Type,
						Host:  v.Host,
						Auth:  v.Auth,
						Limit: v.Limit,
					})
				} else if v, k = c.Report[audpkg.Unknown]; k {
					o.merge(&cfg, CfgEndpoint{
						Type:  v.Type,
						Host:  v.Host,
						Auth:  v.Auth,
						Limit: v.Limit,
					})
				}
			}
		}
	}

	return &cfg
}
