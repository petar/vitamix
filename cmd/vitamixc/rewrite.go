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
	v := &rewriteVisitor{ fileSet: fset }
	ast.Walk(v, file)
	/*
	if err := v.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "Rewrite channel operations · errors parsing '%s':\n%s\n", file.Name.Name, err)
	}
	*/
}

// Rewrite creates a new rewriting visitor
func Rewrite(fset *token.FileSet, node ast.Node) error {
	rwv := &rewriteVisitor{}
	rwv.visitor.Init(fset)
	ast.Walk(rwv, node)
	return rwv.Error()
}


// RecurseRewrite creates a new rewriting visitor as a callee from the visitor caller
func RecurseRewrite(caller *visitor, node ast.Node) error {
	rwv := &rewriteVisitor{}
	rwv.visitor.InitRecurse(caller)
	ast.Walk(rwv, node)
	return rwv.Error()
}

// rewriteVisitor is an AST visitor that rewrites channel operations
type rewriteVisitor struct {
	visitor
}

// Visit implements ast.Visistor's Visit method
func (t *rewriteVisitor) Visit(node ast.Node) ast.Visitor {
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
		t.Printf("stmt@ %s ···· %s\n", t.fileSet.Position(stmt.Pos()), t.fileSet.Position(stmt.End()))
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
				RecurseRewrite(t, stmt)
				list = append(list, stmt)
			}
		}
	}
	bstmt.List = list

	// Do not continue the parent walk recursively
	return nil
}

func (t *rewriteVisitor) rewriteGoStmt(gostmt *ast.GoStmt) []ast.Stmt {
	// Rewrite lower level nodes
	RecurseRewrite(t, gostmt.Call.Fun)
	for _, arg := range gostmt.Call.Args {
		// XXX: What if an argument contains chan operations
		RecurseRewrite(t, arg)
	}
	// Rewrite go statement itself
	gostmt.Call = &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params: &ast.FieldList{},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{ X: gostmt.Call },
					makeSimpleCallStmt("vtime", "Die"),
				},
			},
		},
	}
	return []ast.Stmt{
		makeSimpleCallStmt("vtime", "Go"),
		gostmt,
	}
}

func (t *rewriteVisitor) rewriteRecvStmt(stmt ast.Stmt) []ast.Stmt {
	// Rewrite lower level nodes
	/* XXX
	switch q := stmt.(type) {
	case *ast.AssignStmt:
		for _, expr := range q.Lhs {
			RecurseRewrite(t, expr)
		}
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
	*/
	// Rewrite receive statement itself
	return []ast.Stmt{
		makeSimpleCallStmt("vtime", "Block"),
		stmt,
		makeSimpleCallStmt("vtime", "Unblock"),
	}
}

func (t *rewriteVisitor) rewriteSendStmt(sendstmt *ast.SendStmt) []ast.Stmt {
	// Rewrite lower level nodes
	RecurseRewrite(t, sendstmt.Chan)
	RecurseRewrite(t, sendstmt.Value)
	// Rewrite send statement itself
	return []ast.Stmt{
		makeSimpleCallStmt("vtime", "Block"),
		sendstmt,
		makeSimpleCallStmt("vtime", "Unblock"),
	}
}

func (t *rewriteVisitor) rewriteSelectStmt(selstmt *ast.SelectStmt) []ast.Stmt {
	// Rewrite the comm clauses
	for _, commclause := range selstmt.Body.List {
		if err := RecurseRewrite(t, commclause); err != nil {
			t.errs.Add(err)
		}
	}

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
	return []ast.Stmt{
		makeSimpleCallStmt("vtime", "Block"),
		selstmt,
	}
}
