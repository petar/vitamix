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

func Rewrite(fset *token.FileSet, node ast.Node) error {
	rwv := &rewriteVisitor{ fileSet: fset }
	ast.Walk(rwv, node)
	return rwv.Error()
}

// rewriteVisitor is an AST visitor that rewrites channel operations
type rewriteVisitor struct {
	fileSet   *token.FileSet
	recursion int
	errs      ErrorQueue
}

func (t *rewriteVisitor) Rewrite(node ast.Node) error {
	rwv := &rewriteVisitor{ fileSet: t.fileSet, recursion: t.recursion+1 }
	ast.Walk(rwv, node)
	return rwv.Error()
}

func (t *rewriteVisitor) Error() error {
	if t.errs.Len() > 0 {
		return &t.errs
	}
	return nil
}

func (t *rewriteVisitor) printf(fmt_ string, args_ ...interface{}) {
	for i := 0; i < t.recursion; i++ {
		os.Stderr.WriteString("··")
	}
	fmt.Fprintf(os.Stderr, fmt_, args_...)
}

func (t *rewriteVisitor) saveError(pos token.Pos, msg string) {
	err := NewError(t.fileSet.Position(pos), msg)
	t.errs.Add(err)
	t.printf("%s\n", err)
}

func filterChanStmtOrExpr(node ast.Node) ast.Node {
	switch q := node.(type) {
	case ast.Stmt:
		return filterChanStmt(q)
	case ast.Expr:
		return filterRecvExpr(q)
	}
	return nil
}

// If node is a channel operation (send, receive, select), 
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

func filterRecvExpr(e ast.Expr) ast.Expr {
	ue, ok := e.(*ast.UnaryExpr)
	if !ok || ue.Op.String() != "<-" {
		return nil
	}
	return e
}

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
			t.saveError(node.Pos(), "Channel operation out of context")
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

func (t *rewriteVisitor) rewriteGoStmt(gostmt *ast.GoStmt) []ast.Stmt {
	// Rewrite lower level nodes
	t.Rewrite(gostmt.Call.Fun)
	for _, arg := range gostmt.Call.Args {
		// XXX: What if an argument contains chan operations
		t.Rewrite(arg)
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
			t.Rewrite(expr)
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
	t.Rewrite(sendstmt.Chan)
	t.Rewrite(sendstmt.Value)
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
		if err := t.Rewrite(commclause); err != nil {
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
