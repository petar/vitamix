// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package vrewrite

import (
	"fmt"
	"go/token"
	"os"
)

type framed interface {
	Frame() *frame
}

// frame implements the shared recursion functionality
// used by rewriteVisitor and prohibitVisitor
type frame struct {
	fileSet   *token.FileSet
	errs      *ErrorQueue
	recursion int
}

// Init initializes a root-level frame
func (t *frame) Init(fset *token.FileSet) {
	t.fileSet = fset
	t.errs = NewErrorQueue()
}

// InitRecurse initializes the frame from the calling frame
func (t *frame) InitRecurse(caller framed) {
	t.fileSet = caller.Frame().fileSet
	t.errs = caller.Frame().errs
	t.recursion = caller.Frame().recursion+1
}

// Error returns any errors that have been accumulated by this frame
func (t *frame) Error() error {
	if t.errs.Len() > 0 {
		return t.errs
	}
	return nil
}

// Printf prints to standard error while formating in a way reflecting the frame's recursive level
func (t *frame) Printf(fmt_ string, args_ ...interface{}) {
	for i := 0; i < t.recursion; i++ {
		os.Stderr.WriteString("··")
	}
	os.Stderr.WriteString(" ")
	fmt.Fprintf(os.Stderr, fmt_, args_...)
}

// AddError accumulates an error on the error stack of this frame
func (t *frame) AddError(pos token.Pos, msg string) {
	err := NewError(t.fileSet.Position(pos), msg)
	t.errs.Add(err)
	t.Printf("[¢] %s\n", err)
}
