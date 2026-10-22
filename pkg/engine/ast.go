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
	"maps"
	"slices"
	"sync"

	audast "github.com/nabbar/auditor/pkg/ast"
	astgen "github.com/nabbar/auditor/pkg/ast/generic"
)

type ColAst interface {
	AddAst(l audast.Lang, a astgen.AST)
	SetAst(l audast.Lang, a []astgen.AST)
	GetAst(l audast.Lang) []astgen.AST
	LenAst(l audast.Lang) int

	GetLang() []audast.Lang
}

type astcol struct {
	mux sync.Mutex
	lst map[audast.Lang][]astgen.AST
}

func (o *astcol) AddAst(l audast.Lang, a astgen.AST) {
	if l == audast.LangUnknown {
		return
	}

	if a == nil {
		return
	}

	o.mux.Lock()
	defer o.mux.Unlock()

	if len(o.lst) < 1 {
		o.lst = make(map[audast.Lang][]astgen.AST)
	}

	if len(o.lst[l]) < 1 {
		o.lst[l] = make([]astgen.AST, 0)
	}

	o.lst[l] = append(o.lst[l], a)
}

func (o *astcol) SetAst(l audast.Lang, a []astgen.AST) {
	if l == audast.LangUnknown {
		return
	}

	o.mux.Lock()
	defer o.mux.Unlock()

	if len(o.lst) < 1 {
		o.lst = make(map[audast.Lang][]astgen.AST)
	}

	o.lst[l] = make([]astgen.AST, 0)

	if len(a) > 0 {
		o.lst[l] = a
	}
}

func (o *astcol) GetAst(l audast.Lang) []astgen.AST {
	o.mux.Lock()
	defer o.mux.Unlock()

	if len(o.lst) < 1 {
		return nil
	}

	return o.lst[l]
}

func (o *astcol) LenAst(l audast.Lang) int {
	o.mux.Lock()
	defer o.mux.Unlock()

	if len(o.lst) < 1 {
		return 0
	}

	return len(o.lst[l])
}

func (o *astcol) GetLang() []audast.Lang {
	o.mux.Lock()
	defer o.mux.Unlock()

	if len(o.lst) < 1 {
		return nil
	}

	return slices.Collect(maps.Keys(o.lst))
}
