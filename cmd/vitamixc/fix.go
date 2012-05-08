// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package main

import (
	"go/ast"
)

func fixGo(file *ast.File) {
	ast.Walk(VisitorNoReturnFunc(fixGoVisit), file)
}

func fixGoVisit(x ast.Node) {
	gostmt, ok := x.(*ast.GoStmt)
	if !ok {
		return
	}
	origcall := gostmt.Call
	gostmt.Call = &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{ X: origcall },
					makeSimpleCallStmt("vtime", "Die"),
				},
			},
		},
	}
}

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
	// TODO: We are assuming that pkg 'time' is imported as 'time'
	if sx.Name != "time" {
		return
	}
	// TODO: We only catch direct calls of the form 'time.Now()'.
	// We would not catch indirect calls as in 'f := time.Now; f()'
	if sexpr.Sel.Name == "Now" || sexpr.Sel.Name == "Sleep" {
		sx.Name = "vtime"
	}
}
