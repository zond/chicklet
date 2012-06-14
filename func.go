// Copyright 2009 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package chicklet

import (
	"fmt"
)

/*
 * Virtual machine
 */

type Thread struct {
	abort chan error
	pc    uint
	// The execution frame of this function.  This remains the
	// same throughout a function invocation.
	f *Frame
}

type code []func(*Thread)

func (i code) exec(t *Thread) {
	opc := t.pc
	t.pc = 0
	l := uint(len(i))
	for t.pc < l {
		pc := t.pc
		t.pc++
		i[pc](t)
	}
	t.pc = opc
}

/*
 * Code buffer
 */

type codeBuf struct {
	instrs code
}

func newCodeBuf() *codeBuf { return &codeBuf{make(code, 0, 16)} }

func (b *codeBuf) push(instr func(*Thread)) {
	b.instrs = append(b.instrs, instr)
}

func (b *codeBuf) nextPC() uint { return uint(len(b.instrs)) }

func (b *codeBuf) get() code {
	// Freeze this buffer into an array of exactly the right size
	a := make(code, len(b.instrs))
	copy(a, b.instrs)
	return code(a)
}

/*
 * User-defined functions
 */

type evalFunc struct {
	outer     *Frame
	frameSize int
	inTypes   []Type
        outTypes  []Type
	code      code
}

func (f *evalFunc) Execute(things... Thing) ([]Thing, error) {
	if len(things) != len(f.inTypes) {
		return nil, &CallError{fmt.Sprint("Wrong number of arguments. Wanted ", len(f.inTypes), " but got ", len(things))}
	}
	frame := f.NewFrame()
	thread := &Thread{}
	for index, thing := range things {
		frame.Vars[index] = ValueFromNative(thing, thread)
	}
	for index, t := range f.outTypes {
		frame.Vars[len(f.inTypes) + index] = t.(Type).Zero()
	}
	thread.f = frame
	f.Call(thread)
	var rval []Thing
	for index, _ := range f.outTypes {
		rval = append(rval, frame.Vars[len(f.inTypes) + index].GetNative(thread))
	}
	return rval, nil
}

func (f *evalFunc) NewFrame() *Frame { return f.outer.child(f.frameSize) }

func (f *evalFunc) Call(t *Thread) { f.code.exec(t) }
