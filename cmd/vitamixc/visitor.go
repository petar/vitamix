// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
)

// visitor is an AST visitor that supports some shared recursion functionality
// perused by rewriteVisitor and prohibitVisitor
type visitor struct {
	fileSet   *token.FileSet
	recursion int
	errs      ErrorQueue
}

// Init initializes a root-level visitor
func (t *visitor) Init(fset *token.FileSet) {
	t.fileSet = fset
}

// InitRecurse initializes the visitor from the calling visitor
func (t *visitor) InitRecurse(caller *visitor) {
	t.fileSet = caller.fileSet
	t.recursion = caller.recursion+1
}

// Error returns any errors that have been accumulated by this visitor
func (t *visitor) Error() error {
	if t.errs.Len() > 0 {
		return &t.errs
	}
	return nil
}

// Printf prints to standard error while formating in a way reflecting the visitor's recursive level
func (t *visitor) Printf(fmt_ string, args_ ...interface{}) {
	for i := 0; i < t.recursion; i++ {
		os.Stderr.WriteString("··")
	}
	fmt.Fprintf(os.Stderr, fmt_, args_...)
}

// AddError accumulates an error on the error stack of this visitor
func (t *visitor) AddError(pos token.Pos, msg string) {
	err := NewError(t.fileSet.Position(pos), msg)
	t.errs.Add(err)
	t.printf("%s\n", err)
}
