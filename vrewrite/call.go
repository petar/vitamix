// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package vrewrite

import (
	"go/ast"
)

func rewriteTimeCalls(file *ast.File) (needVtime, needTime bool) {
	v := &callVisitor{}
	ast.Walk(v, file)
	return v.NeedPkgVtime, v.NeedPkgTime
}

type callVisitor struct {
	NeedPkgVtime bool
	NeedPkgTime  bool  // True if after rewriting invokations to time.Sleep and time.Now, other references to pkg "time" remain
}

func (v *callVisitor) Visit(x ast.Node) ast.Visitor {
	callexpr, ok := x.(*ast.CallExpr)
	if !ok || callexpr.Fun == nil {
		return v
	}
	sexpr, ok := callexpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return v
	}
	sx, ok := sexpr.X.(*ast.Ident)
	if !ok {
		return v
	}
	if sx.Name != "time" {
		return v
	}
	if sexpr.Sel.Name == "Now" || sexpr.Sel.Name == "Sleep" {
		sx.Name = "vtime"
		v.NeedPkgVtime = true
	} else {
		v.NeedPkgTime = true
	}
	return v
}
