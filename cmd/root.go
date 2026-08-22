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

package cmd

import (
	"os"

	libcol "github.com/fatih/color"
	cmdexp "github.com/nabbar/auditor/cmd/exp"
	cmdimp "github.com/nabbar/auditor/cmd/imp"
	cmdlst "github.com/nabbar/auditor/cmd/list"
	cmdscn "github.com/nabbar/auditor/cmd/step1"
	cmdanl "github.com/nabbar/auditor/cmd/step2"
	cmdaud "github.com/nabbar/auditor/cmd/step3"
	cmdrep "github.com/nabbar/auditor/cmd/step4"
	audpkg "github.com/nabbar/auditor/pkg"
	libcon "github.com/nabbar/golib/console"
	liberr "github.com/nabbar/golib/errors"
	loglvl "github.com/nabbar/golib/logger/level"
)

var (
	flgVerbose int
	flgNoColor bool
	flgSetFD   bool
	flgLogPth  string
	flgThread  int8
)

func init() {
	liberr.SetModeReturnError(liberr.ModeReturnCodeErrorFull)
	audpkg.InitCommon()
	audpkg.GetCobra().SetFuncInit(initConfig)
	audpkg.GetCobra().Init()
	audpkg.GetCobra().SetFlagVerbose(true, &flgVerbose)

	// add global flags
	audpkg.GetCobra().AddFlagBool(true, &flgNoColor, "no-color", "", false, "Disabled color for output message & error")
	audpkg.GetCobra().AddFlagBool(true, &flgSetFD, "fd", "", false, "Change FileDescriptor to max allowed")
	audpkg.GetCobra().AddFlagInt8(true, &flgThread, "threads", "t", 0, "Maximum concurrent thread boundary allocations assigned to execution pools")
	audpkg.GetCobra().AddFlagString(true, &flgLogPth, "logfile", "l", "", "Physical log file pathway; allocating this flag redirects text logging and unlocks visual progress trackers")

	// Add some generic command
	audpkg.GetCobra().AddCommandCompletion()
	audpkg.GetCobra().AddCommand(cmdscn.InitCmd())
	audpkg.GetCobra().AddCommand(cmdanl.InitCmd())
	audpkg.GetCobra().AddCommand(cmdaud.InitCmd())
	audpkg.GetCobra().AddCommand(cmdrep.InitCmd())
	audpkg.GetCobra().AddCommand(cmdlst.InitCmd())
	audpkg.GetCobra().AddCommand(cmdexp.InitCmd())
	audpkg.GetCobra().AddCommand(cmdimp.InitCmd())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	audpkg.UpdateConfig(flgLogPth, int(flgThread))

	// Redefine the root logger level with the verbose flag
	switch flgVerbose {
	case 0:
		audpkg.GetLogger().SetLevel(loglvl.NilLevel)
	case 1:
		audpkg.GetLogger().SetLevel(loglvl.WarnLevel)
	case 2:
		audpkg.GetLogger().SetLevel(loglvl.InfoLevel)
	default:
		audpkg.GetLogger().SetLevel(loglvl.DebugLevel)
	}

	if flgNoColor {
		libcon.SetColor(libcon.ColorPrompt, int(libcol.Reset))
	}

	audpkg.SetFD(flgSetFD)

	// Run debug server if available
	audpkg.PPRofStart()
}

// Execute is called by main.main() to parse and run the app.
func Execute() {
	if e := audpkg.GetCobra().Execute(); e != nil {
		// print error
		os.Exit(1)
	}

	os.Exit(0)
}
