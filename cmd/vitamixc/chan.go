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

func fixChan(fset *token.FileSet, file *ast.File) {
	v := &fixChanVisitor{ fileSet: fset }
	ast.Walk(v, file)
	if err := v.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "Rewrite channel operations · errors parsing '%s':\n%s\n", file.Name.Name, err)
	}
}

// fixChanVisitor is an AST visitor that rewrites channel operations
type fixChanVisitor struct {
	fileSet   *token.FileSet
	errs      ErrorQueue
}

func (t *fixChanVisitor) Error() error {
	if t.errs.Len() > 0 {
		return &t.errs
	}
	return nil
}

// If node is a channel operation (send, receive, select), 
// filterChanStmt is the identity, otherwise it returns nil
func filterChanStmt(node ast.Node) ast.Node {
	switch node.(type) {
	case *ast.SelectStmt:
		return node
	case *ast.SendStmt:
		return node
	}
	return filterRecvStmt(node)
}

func filterRecvStmt(node ast.Node) ast.Node {
	switch q := node.(type) {
	case *ast.AssignStmt:
		for _, expr := range q.Rhs {
			if expr != nil && filterRecvExpr(expr) != nil {
				return node
			}
		}
	case *ast.ExprStmt:
		if q.X != nil && filterRecvExpr(q.X) != nil {
			return node
		}
	}
	return nil
}

func filterRecvExpr(e ast.Expr) ast.Expr {
	ue, ok := e.(*ast.UnaryExpr)
	if !ok || ue.Op.String() != "<-" {
		return nil
	}
	return e
}

func (t *fixChanVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return t
	}
	bstmt, ok := node.(*ast.BlockStmt)
	// If node is not a block statement, 
	if !ok {
		// Check that it is not a channel operation
		if filterChanStmt(node) != nil {
			t.errs.Add(NewError(t.fileSet.Position(node.Pos()), "Channel operation out of context"))
			return nil
		}
		// Continue the walk recursively
		return t
	}
	for i, stmt := range bstmt.List {
		switch q := stmt.(type) {
		case *ast.SelectStmt:
			t.fixSelect(bstmt, i, q)
			// Continue walking recursively from the comm clauses
			for _, commclause := range q.Body.List {
				ast.Walk(t, commclause)
			}
		default:
			// Continue the walk recursively, skipping the block statement
			ast.Walk(t, stmt)
		}
	}
	// Do not continue the parent walk recursively
	return nil
}

func (t *fixChanVisitor) fixSelect(bstmt *ast.BlockStmt, i int, selstmt *ast.SelectStmt) {
	// Place a call to Unblock immediately after each case and default
	for _, clause := range selstmt.Body.List {
		comm := clause.(*ast.CommClause)
		body := comm.Body
		comm.Body = append(
			[]ast.Stmt{ makeSimpleCallStmt("vtime", "Unblock") },
			body...,
		)
	}
	// Surround the select by a block statement and prefix it with a call to vtime.Block
	bstmt.List[i] = &ast.BlockStmt{
		List: []ast.Stmt{
			makeSimpleCallStmt("vtime", "Block"),
			selstmt,
		},
	}
}
