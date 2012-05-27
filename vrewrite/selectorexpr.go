// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package vrewrite

import (
	"go/ast"
)

func ExistSelectorFor(file *ast.File, sel string) bool {
	v := &selectorVisitor{ X: sel }
	ast.Walk(v, file)
	return v.Exists
}

type selectorVisitor struct {
	X      string
	Exists bool
}

func (v *selectorVisitor) Visit(x ast.Node) ast.Visitor {
	se, ok := x.(*ast.SelectorExpr)
	if !ok || se.X == nil {
		return v
	}
	id, ok := se.X.(*ast.Ident)
	if !ok || id.Name != v.X {
		return v
	}
	v.Exists = true
	return nil
}
