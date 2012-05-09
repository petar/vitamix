// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
)

func fixCall(file *ast.File) {
	ast.Walk(VisitorNoReturnFunc(fixCallVisit), file)
}

func fixCallVisit(x ast.Node) {
	callexpr, ok := x.(*ast.CallExpr)
	if !ok || callexpr.Fun == nil {
		return
	}
	sexpr, ok := callexpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	sx, ok := sexpr.X.(*ast.Ident)
	if !ok {
		return
	}
	if sx.Name != "time" {
		return
	}
	if sexpr.Sel.Name == "Now" || sexpr.Sel.Name == "Sleep" {
		sx.Name = "vtime"
	}
}
