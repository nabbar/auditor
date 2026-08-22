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

package engine

import (
	"fmt"
	"time"

	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	audsts "github.com/nabbar/auditor/pkg/data/status"
	librun "github.com/nabbar/golib/runner"
	semtps "github.com/nabbar/golib/semaphore/types"
)

// Analyze orchestrates the multi-pass linguistic auditing pipeline by dispatching
// parallel, multithreaded worker pools to process code elements extracted
// during the AST scan phase.
func (o *eng) Analyze() error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("auditor/engine/Analyze", r)
		}
	}()

	var bar semtps.SemBar

	defer func() {
		if bar != nil {
			bar.DeferMain()
		}
	}()

	o.u.Info("Analyze entries")
	bar = o.u.NewBar("Analyze entries", len(o.i))

	for i := 0; i < len(o.i); i++ {
		if err := bar.NewWorker(); err != nil {
			o.u.ErrorStack("cannot create new worker for analyze entries", err)
			return err
		}

		go func(n int) {
			defer bar.DeferWorker()

			if o.i[n] < 1 {
				// skip not valid entry
				return
			}

			if ent := o.d.EntGet(o.i[n]); ent == nil || ent.IsEmpty() {
				// skip not valid entry
				return
			} else if ent.GetStatus() == audsts.Analyzed {
				// skip still analyzed
				return
			} else if !ent.IsVendor() {
				// check dependencies are still analyzed
				var (
					rdy = true
					lst = ent.GetDepend()
				)
				if len(lst) > 1 {
					for _, m := range lst {
						if m.GetStatus() != audsts.Analyzed {
							rdy = false
							break
						}
					}
				}
				if !rdy {
					o.u.Error("entry %s (id %d) has depend not analyzed", ent.GetFullPath(), ent.GetID())
					return
				}
			}

			if e := o.l.Analyze(bar, o.i[n]); e != nil {
				o.u.ErrorStack(fmt.Sprintf("error on analyze entry id %d", o.i[n]), e)
				return
			}
		}(i)
	}

	if err := bar.WaitAll(); err != nil {
		o.u.ErrorStack("cannot wait all worker for analyze entries", err)
		return err
	}

	// force update bar
	for !bar.Completed() {
		bar.Inc(1)
		time.Sleep(time.Millisecond)
	}

	// timer to wait ui is updated
	time.Sleep(500 * time.Millisecond)
	return nil
}

func (o *eng) Review() error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("auditor/engine/Review", r)
		}
	}()

	if e := o.ReviewEntries(); e != nil {
		return e
	}

	if e := o.ReviewPackage(); e != nil {
		return e
	}

	if e := o.ReviewModule(); e != nil {
		return e
	}

	return nil
}

func (o *eng) ReviewEntries() error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("auditor/engine/ReviewEntries", r)
		}
	}()

	var bar semtps.SemBar

	defer func() {
		if bar != nil {
			bar.DeferMain()
		}
	}()

	o.u.Info("Review entries")
	bar = o.u.NewBar("Review entries", len(o.i))

	for i := 0; i < len(o.i); i++ {
		if err := bar.NewWorker(); err != nil {
			o.u.ErrorStack("cannot create new worker for analyze entries", err)
			return err
		}

		go func(n int) {
			defer bar.DeferWorker()

			ent := o.d.EntGet(o.i[n])

			if ent == nil || ent.IsEmpty() {
				// skip not valid entry
				return
			}

			if ent.GetID() == 0 {
				// skip not valid entry
				return
			}

			if ent.IsVendor() {
				// skip vendor entry
				return
			}

			// render prompt for language and type
			// call LLM for report
			// register report

		}(i)
	}

	if err := bar.WaitAll(); err != nil {
		o.u.ErrorStack("cannot wait all worker for analyze entries", err)
		return err
	}

	// force update bar
	for !bar.Completed() {
		bar.Inc(1)
		time.Sleep(time.Millisecond)
	}

	// timer to wait ui is updated
	time.Sleep(500 * time.Millisecond)
	return nil
}

func (o *eng) ReviewPackage() error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("auditor/engine/ReviewPackage", r)
		}
	}()

	var bar semtps.SemBar

	defer func() {
		if bar != nil {
			bar.DeferMain()
		}
	}()

	o.u.Info("Review packages")
	bar = o.u.NewBar("Review packages", o.d.PkgLen())

	o.d.PkgWalk(func(pkg audpkg.Package) bool {
		defer bar.Inc(1)

		if pkg == nil || pkg.IsEmpty() {
			// skip not valid entry
			return true
		}

		if pkg.GetID() == 0 {
			// skip not valid entry
			return true
		}

		if pkg.IsVendor() {
			// skip vendor entry
			return true
		}

		// collect report of all entries of the package
		// render prompt for package
		// call LLM for report
		// register report

		return true
	})

	// force update bar
	for !bar.Completed() {
		bar.Inc(1)
		time.Sleep(time.Millisecond)
	}

	// timer to wait ui is updated
	time.Sleep(500 * time.Millisecond)
	return nil
}

func (o *eng) ReviewModule() error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("auditor/engine/ReviewModule", r)
		}
	}()

	var bar semtps.SemBar

	defer func() {
		if bar != nil {
			bar.DeferMain()
		}
	}()

	o.u.Info("Review modules")
	bar = o.u.NewBar("Review modules", o.d.PkgLen())

	o.d.PkgWalk(func(pkg audpkg.Package) bool {
		defer bar.Inc(1)

		if pkg == nil || pkg.IsEmpty() {
			// skip not valid entry
			return true
		}

		if pkg.GetID() == 0 {
			// skip not valid entry
			return true
		}

		if pkg.IsVendor() {
			// skip vendor entry
			return true
		}

		// collect report of all packages of the module
		// render prompt for module
		// call LLM for module
		// register module

		return true
	})

	// force update bar
	for !bar.Completed() {
		bar.Inc(1)
		time.Sleep(time.Millisecond)
	}

	// timer to wait ui is updated
	time.Sleep(500 * time.Millisecond)
	return nil
}
