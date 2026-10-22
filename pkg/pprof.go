//go:build pprof
// +build pprof

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

package pkg

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	loglvl "github.com/nabbar/golib/logger/level"
)

var flgTracePort uint16

func PPRofFlags() {
	// Add & Mark flag pprof as hidden
	GetCobra().AddFlagUint16(true, &flgTracePort, "trace-listen-port", "", 0, "Enable a trace http server to the given port number (callable by http://localhost:<given port>/debug/pprof)")
	GetLogger().CheckError(loglvl.FatalLevel, loglvl.DebugLevel, "mark hidden flag trace", GetCobra().Cobra().PersistentFlags().MarkHidden("trace-listen-port"))
}

func PPRofStart() {
	if flgTracePort < 1 {
		return
	}

	GetLogger().Entry(loglvl.InfoLevel, "init pprof server").FieldAdd("Address", "http://localhost:"+strconv.Itoa(flgTracePort)+"/debug/pprof").Log()

	go func() {
		for {
			time.Sleep(5 * time.Second)

			if GetContext().Err() != nil {
				return
			}

			GetLogger().Entry(loglvl.ErrorLevel, "starting pprof server").FieldAdd("Address", "http://localhost:"+strconv.Itoa(flgTracePort)+"/debug/pprof").ErrorAdd(true, http.ListenAndServe(fmt.Sprintf("localhost:%d", flgTracePort), nil)).Check(loglvl.InfoLevel)
		}
	}()
}
