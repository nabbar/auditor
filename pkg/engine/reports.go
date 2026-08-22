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
	"errors"
	"fmt"
	"os"
	"time"

	audent "github.com/nabbar/auditor/pkg/data/entry"
	audmod "github.com/nabbar/auditor/pkg/data/mod"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	librun "github.com/nabbar/golib/runner"
	semtps "github.com/nabbar/golib/semaphore/types"
)

// Report collects individual specialized documentation blocks from the database
// and compiles them into a unified, structured Markdown document for export.
// It iterates through all analyzed entries, retrieves their associated reports,
// and orchestrates the final file serialization process.
func (o *eng) Report(pth string) error {
	defer func() {
		if r := recover(); r != nil {
			librun.RecoveryCaller("auditor/engine/Report", r)
		}
	}()

	var (
		err error
		bar semtps.SemBar
		frt *os.Root
		ffs *os.File
	)

	defer func() {
		if bar != nil {
			bar.DeferMain()
		}
	}()

	defer func() {
		if frt != nil {
			_ = frt.Close()
		}
	}()

	defer func() {
		if ffs != nil {
			_ = ffs.Close()
		}
	}()

	if i, e := os.Stat(pth); e == nil && !i.IsDir() {
		return errors.New("'" + pth + "' is not a directory")
	} else if e != nil && !errors.Is(e, os.ErrNotExist) {
		return fmt.Errorf("error on accessing '%s': %w", pth, e)
	}

	if frt, err = os.OpenRoot(pth); err != nil {
		return fmt.Errorf("error on opening '%s': %w", pth, err)
	}

	o.u.Info("Create report files")
	bar = o.u.NewBar("Create report files", o.d.EntLen()+o.d.PkgLen()+o.d.ModLen())

	o.d.EntWalk(func(ent audent.Entry) bool {
		defer bar.Inc(1)

		if ent == nil || ent.IsEmpty() {
			bar.Inc(1)
			return true
		}

		if ent.GetID() == 0 {
			bar.Inc(1)
			return true
		}

		if ent.IsVendor() {
			bar.Inc(1)
			return true
		}

		// render report for entry
		// append report to report of package file

		return true
	})

	o.d.PkgWalk(func(pkg audpkg.Package) bool {
		defer bar.Inc(1)

		if pkg == nil || pkg.IsEmpty() {
			bar.Inc(1)
			return true
		}

		if pkg.GetID() == 0 {
			bar.Inc(1)
			return true
		}

		if pkg.IsVendor() {
			bar.Inc(1)
			return true
		}

		// open file for package
		// store content in buffer
		// truncate file content
		// render report for package
		// append report to report file
		// append buffer to report file

		// add package report into module file

		return true
	})

	o.d.ModWalk(func(mod audmod.Module) bool {
		defer bar.Inc(1)

		if mod == nil || mod.IsEmpty() {
			bar.Inc(1)
			return true
		}

		if mod.GetID() == 0 {
			bar.Inc(1)
			return true
		}

		if mod.IsVendor() {
			bar.Inc(1)
			return true
		}

		// open file for module
		// store content in buffer
		// truncate file content
		// render report for module
		// append report to report file
		// append buffer to report file

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
