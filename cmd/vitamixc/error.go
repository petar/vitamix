// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"go/token"
)

// Error is represents a semantic error in the source code
type Error struct {
	Position token.Position
	Msg      string
}

func NewError(position token.Position, msg string) *Error {
	return &Error{
		Position: position,
		Msg:      msg,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("(%s) %s", e.Position.String(), e.Msg)
}

// ErrorQueue is a list of errors accumulated during a pass through the AST of a file
type ErrorQueue struct {
	queue []error
}

func (x *ErrorQueue) Add(e error) {
	x.queue = append(x.queue, e)
}

func (x *ErrorQueue) String() string {
	var w bytes.Buffer
	for _, e := range x.queue {
		w.WriteString(e.Error())
		w.WriteByte('\n')
	}
	return string(w.Bytes())
}

func (x *ErrorQueue) Len() int {
	return len(x.queue)
}

func (x *ErrorQueue) Error() string {
	return x.String()
}
