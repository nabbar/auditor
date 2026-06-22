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
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	audent "github.com/nabbar/auditor/pkg/data/entry"
	auddbm "github.com/nabbar/auditor/pkg/data/manager"
	audpkg "github.com/nabbar/auditor/pkg/data/pkg"
	audeng "github.com/nabbar/auditor/pkg/engine"
	audllm "github.com/nabbar/auditor/pkg/llm"
	audloc "github.com/nabbar/auditor/pkg/local"
	auduxi "github.com/nabbar/auditor/pkg/uxi"
	audver "github.com/nabbar/auditor/release"
	libtls "github.com/nabbar/golib/certificates"
	tlscas "github.com/nabbar/golib/certificates/ca"
	tlscpr "github.com/nabbar/golib/certificates/cipher"
	tlscrv "github.com/nabbar/golib/certificates/curves"
	tlsvrs "github.com/nabbar/golib/certificates/tlsversion"
	libcbr "github.com/nabbar/golib/cobra"
	libcfg "github.com/nabbar/golib/config"
	libdur "github.com/nabbar/golib/duration"
	libprm "github.com/nabbar/golib/file/perm"
	libhtc "github.com/nabbar/golib/httpcli"
	htcdns "github.com/nabbar/golib/httpcli/dns-mapper"
	libfds "github.com/nabbar/golib/ioutils/fileDescriptor"
	liblog "github.com/nabbar/golib/logger"
	logcfg "github.com/nabbar/golib/logger/config"
	loglvl "github.com/nabbar/golib/logger/level"
	libsiz "github.com/nabbar/golib/size"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)

const (
	minFileDescriptor = 1000000
	maxFileDescriptor = 1048576
)

var (
	cbr libcbr.Cobra
	cfg libcfg.Config
	vpr libvpr.Viper
	vrs libver.Version
	wrk string
	uim auduxi.Manager
	dbm auddbm.Manager
	dbl *auddbm.Linker
	llm audllm.Manager
	eng audeng.Engine
	loc audloc.Manager
	fdg = func() auddbm.Manager {
		return dbm
	}
	fds = func(m auddbm.Manager) {
		dbm.Merge(m)
	}
)

func ShutDown(code int) {
	cfg.Shutdown(code)
	libcfg.Shutdown()
}

func InitCommon() {
	vrs = libver.NewVersion(libver.License_MIT, audver.Package, audver.Description, audver.Date, audver.Build, audver.Release, audver.Author, audver.Prefix, audver.EmptyStruct{}, 1)
	cfg = libcfg.New(vrs)
	wrk = getWorkingPath()

	cfg.CancelAdd(func() {
		if eng != nil {
			_ = eng.Close()
		}
	})

	cfg.CancelAdd(func() {
		if loc != nil {
			_ = loc.Close()
		}
	})

	cfg.CancelAdd(func() {
		if dbm != nil {
			_ = dbm.Close()
		}
	})

	cfg.CancelAdd(func() {
		if uim != nil {
			_ = uim.Close()
		}
	})

	libhtc.SetDefaultDNSMapper(htcdns.New(GetContext(), &htcdns.Config{
		TimerClean: libdur.ParseDuration(5 * time.Minute),
		Transport: htcdns.TransportConfig{
			TLSConfig: &libtls.Config{
				CurveList:            tlscrv.List(),
				CipherList:           tlscpr.List(),
				RootCA:               []tlscas.Cert{audver.GetRootCACert()},
				ClientCA:             nil,
				Certs:                nil,
				VersionMin:           tlsvrs.VersionTLS12,
				VersionMax:           tlsvrs.VersionTLS13,
				AuthClient:           0,
				InheritDefault:       false,
				DynamicSizingDisable: false,
				SessionTicketDisable: false,
			},
		},
	}, audver.GetRootCACert, nil))

	go libcfg.WaitNotify()

	var err error

	if uim, err = auduxi.New(cfg.Context(), cfgLog(""), loglvl.InfoLevel, false, 0); err != nil {
		panic(err)
	}

	cbr = libcbr.New()
	cbr.SetVersion(vrs)
	cbr.SetLogger(GetLogger)
}

func UpdateConfig(logfile string, thread int) {
	var err error

	_ = uim.Close()

	if uim, err = auduxi.New(cfg.Context(), cfgLog(logfile), loglvl.InfoLevel, len(logfile) > 0, thread); err != nil {
		panic(err)
	}

	uim.Info("Initializing Auditor Engine targeting local workspace scope: %s", wrk)

	if thread > 0 {
		uim.Info("Active Parallel Thread Boundary Allocations: %d Concurrent Workers", thread)
	} else {
		uim.Info("Active Parallel Thread Boundary Allocations: System Auto-Configured Workers")
	}

	if dbm, dbl = auddbm.New(uim); dbm == nil {
		panic("dataManager is nil")
	}

	if dbl == nil {
		panic("dataLinker is nil")
	}

	audpkg.RegisterGetModule(dbm.ModGet)
	audent.RegisterGetPackage(dbm.PkgGet)
	audent.RegisterGetEntry(dbm.EntGet)

	if loc, err = audloc.New(wrk, uim, fdg, fds); err != nil {
		panic(err)
	}

	if c, e := loc.LoadLLMConfig(); e != nil {
		panic(e)
	} else if c == nil {
		panic("no llm config loaded")
	} else if llm, err = audllm.New(*c, uim, loc, dbl); err != nil {
		panic(err)
	}

	// Instantiate the core auditing engine with the provided configuration.
	eng = audeng.New(dbl, llm, uim, audeng.Options{
		RepoPath: wrk,
	})
}

func SetFD(flag bool) {
	if !flag {
		return
	}

	GetUIManager().Info("check number of file descriptors")
	o, m, err := libfds.SystemFileDescriptor(0)
	if err != nil {
		GetUIManager().ErrorStack(fmt.Sprintf("retrieve limit of file descriptors (current %d | max %d)", o, m), err)
	}

	if m < minFileDescriptor {
		m = maxFileDescriptor
	}

	if o < m {
		var c int
		c, _, err = libfds.SystemFileDescriptor(m)
		if err != nil {
			GetUIManager().ErrorStack(fmt.Sprintf("error while update limit of file descriptors (new value %d | old value %d | awaiting value %d)", c, o, m), err)
		} else {
			GetUIManager().Info("update limit of file descriptors (new value %d | old value %d)", c, o)
		}
	}
}

func GetViper() libvpr.Viper {
	return vpr
}

func GetCobra() libcbr.Cobra {
	return cbr
}

func GetHttpCli() *http.Client {
	return libhtc.GetClient()
}

func GetContext() context.Context {
	return cfg.Context()
}

func GetVersion() libver.Version {
	return vrs
}

func GetUIManager() auduxi.Manager {
	return uim
}

func GetLogger() liblog.Logger {
	return GetUIManager().Logger()
}

func GetDBManager() auddbm.Manager {
	return dbm
}

func GetDBLinker() *auddbm.Linker {
	return dbl
}

func GetLLManager() audllm.Manager {
	return llm
}

func GetEngine() audeng.Engine {
	return eng
}

func GetLocal() audloc.Manager {
	return loc
}

func cfgLog(logfile string) logcfg.Options {
	var c = logcfg.Options{
		InheritDefault: false,
		TraceFilter:    GetVersion().GetRootPackagePath(),
		Stdout: &logcfg.OptionsStd{
			DisableStandard:  false,
			DisableStack:     false,
			DisableTimestamp: false,
			EnableTrace:      true,
			DisableColor:     false,
			EnableAccessLog:  false,
		},
		LogFileExtend:   false,
		LogFile:         nil,
		LogSyslogExtend: false,
		LogSyslog:       nil,
	}

	if len(logfile) > 0 {
		c.Stdout = nil
		c.LogFile = []logcfg.OptionsFile{
			{
				LogLevel:         nil,
				Filepath:         logfile,
				Create:           true,
				CreatePath:       false,
				FileMode:         libprm.ParseFileMode(0644),
				PathMode:         libprm.ParseFileMode(0755),
				DisableStack:     false,
				DisableTimestamp: false,
				EnableTrace:      true,
				EnableAccessLog:  false,
				MessageMaxSize:   libsiz.SizeKilo * 8,
			},
		}
	}

	return c
}

func getWorkingPath() string {
	var pth string

	// Determine the absolute path of the current directory to establish the audit root context.
	cur, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	pth, err = filepath.Abs(cur)
	if err != nil {
		panic(err)
	}

	return pth
}
