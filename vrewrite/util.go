// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package vrewrite

import (
	"go/ast"
	"go/token"
)

func makeSimpleCallStmt(pkgAlias, funcName string, pos token.Pos) ast.Stmt {
	return &ast.ExprStmt{
		X: &ast.CallExpr{
			Fun: &ast.SelectorExpr{
				X:   &ast.Ident{ Name: pkgAlias },
				Sel: &ast.Ident{ Name: funcName },
			},
			Lparen: pos,
		},
	}
}

type VisitorNoReturnFunc func(ast.Node)

func (v VisitorNoReturnFunc) Visit(x ast.Node) ast.Visitor {
	v(x)
	return v
}
