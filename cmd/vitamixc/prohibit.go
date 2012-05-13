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

// Prohibit creates a new prohibiting visitor
func Prohibit(fset *token.FileSet, node ast.Node) error {
	v := &prohibitVisitor{}
	v.visitor.Init(fset)
	ast.Walk(v, node)
	return v.Error()
}


// RecurseProhibit creates a new prohibiting visitor as a callee from the visitor caller
func RecurseProhibit(caller *visitor, node ast.Node) error {
	v := &prohibitVisitor{}
	v.visitor.InitRecurse(caller)
	ast.Walk(v, node)
	return v.Error()
}

// prohibitVisitor is an AST visitor that rewrites channel operations
type prohibitVisitor struct {
	fileSet   *token.FileSet
	recursion int
	errs      ErrorQueue
}

// Visit implements ast.Visistor's Visit method
func (t *prohibitVisitor) Visit(node ast.Node) ast.Visitor {
	?
	if node == nil {
		return t
	}
	bstmt, ok := node.(*ast.BlockStmt)
	// If node is not a block statement, 
	if !ok {
		// Check that we haven't reached a channel operation (statement or expression) out of context
		/*
		if filterChanStmtOrExpr(node) != nil {
			t.AddError(node.Pos(), "Channel operation out of context")
			return nil
		}
		*/
		// Continue the walk recursively
		return t
	}

	// Rewrite each statement of a block statement
	var list []ast.Stmt
	for _, stmt := range bstmt.List {
		t.printf("stmt@ %s ···· %s\n", t.fileSet.Position(stmt.Pos()), t.fileSet.Position(stmt.End()))
		switch q := stmt.(type) {
		case *ast.SelectStmt:
			list = append(list, t.rewriteSelectStmt(q)...)
		case *ast.SendStmt:
			list = append(list, t.rewriteSendStmt(q)...)
		case *ast.GoStmt:
			list = append(list, t.rewriteGoStmt(q)...)
		default:
			if filterRecvStmt(stmt) != nil {
				list = append(list, t.rewriteRecvStmt(stmt)...)
			} else {
				// Continue the walk recursively below this stmt
				t.Rewrite(stmt)
				list = append(list, stmt)
			}
		}
	}
	bstmt.List = list

	// Do not continue the parent walk recursively
	return nil
}

// If node is a channel operation (send, receive, select) statement or expression, 
// filterChanStmt is the identity, otherwise it returns nil
func filterChanStmtOrExpr(node ast.Node) ast.Node {
	switch q := node.(type) {
	case ast.Stmt:
		return filterChanStmt(q)
	case ast.Expr:
		return filterRecvExpr(q)
	}
	return nil
}

// If node is a channel operation (send, receive, select) statement, 
// filterChanStmt is the identity, otherwise it returns nil
func filterChanStmt(stmt ast.Stmt) ast.Stmt {
	switch q := stmt.(type) {
	case *ast.SelectStmt:
		return stmt
	case *ast.SendStmt:
		return stmt
	case ast.Stmt:
		return filterRecvStmt(q)
	}
	return nil
}

// If node is a receive operation statement, 
// filterRecvStmt is the identity, otherwise it returns nil
func filterRecvStmt(stmt ast.Stmt) ast.Stmt {
	switch q := stmt.(type) {
	case *ast.AssignStmt:
		for _, expr := range q.Rhs {
			if expr != nil && filterRecvExpr(expr) != nil {
				return stmt
			}
		}
	case *ast.ExprStmt:
		if q.X != nil && filterRecvExpr(q.X) != nil {
			return stmt
		}
	}
	return nil
}

// If node is a receive operation expression, 
// filterRecvExpr is the identity, otherwise it returns nil
func filterRecvExpr(e ast.Expr) ast.Expr {
	ue, ok := e.(*ast.UnaryExpr)
	if !ok || ue.Op.String() != "<-" {
		return nil
	}
	return e
}
