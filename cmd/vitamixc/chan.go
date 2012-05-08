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
	/*
	if err := v.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "Rewrite channel operations · errors parsing '%s':\n%s\n", file.Name.Name, err)
	}
	*/
}

// fixChanVisitor is an AST visitor that rewrites channel operations
type fixChanVisitor struct {
	fileSet   *token.FileSet
	recursion int
	errs      ErrorQueue
}

func (t *fixChanVisitor) Error() error {
	if t.errs.Len() > 0 {
		return &t.errs
	}
	return nil
}

func (t *fixChanVisitor) saveError(pos token.Pos, msg string) {
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

func (t *fixChanVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return t
	}
	bstmt, ok := node.(*ast.BlockStmt)
	// If node is not a block statement, 
	if !ok {
		// Check that we haven't reached a channel operation (statement or expression) out of context
		if filterChanStmtOrExpr(node) != nil {
			t.saveError(node.Pos(), "Channel operation out of context")
			return nil
		}
		// Continue the walk recursively
		return t
	}
	for i, stmt := range bstmt.List {
		switch q := stmt.(type) {
		case *ast.SelectStmt:
			t.fixSelectStmt(bstmt, i, q)
			// Continue walking recursively from the comm clauses down
			for _, commclause := range q.Body.List {
				t.walk(commclause)
			}
		case *ast.SendStmt:
			t.fixSendStmt(bstmt, i, q)
		default:
			if filterRecvStmt(stmt) != nil {
				t.fixRecvStmt(bstmt, i, stmt)
			} else {
				// Continue the walk recursively below this stmt
				t.walk(stmt)
			}
		}
	}
	// Do not continue the parent walk recursively
	return nil
}

func (t *fixChanVisitor) walk(x ast.Node) {
	t.recursion++
	ast.Walk(t, x)
	t.recursion--
}

func (t *fixChanVisitor) printf(fmt_ string, args_ ...interface{}) {
	for i := 0; i < t.recursion; i++ {
		os.Stderr.WriteString("··")
	}
	fmt.Fprintf(os.Stderr, fmt_, args_...)
}

func (t *fixChanVisitor) fixRecvStmt(bstmt *ast.BlockStmt, i int, recvstmt ast.Stmt) {
	// We cannot surround the recv statement in a block statement, since it may
	// be an assignment statement to new variables, so we include the Block/Unblock
	// calls in the original enclosing block statement
	list := append(bstmt.List[:i], 
		makeSimpleCallStmt("vtime", "Block"),
		recvstmt,
		makeSimpleCallStmt("vtime", "Unblock"),
	)
	list = append(list, bstmt.List[i+1:]...)
	bstmt.List = list
}

func (t *fixChanVisitor) fixSendStmt(bstmt *ast.BlockStmt, i int, sendstmt *ast.SendStmt) {
	bstmt.List[i] = &ast.BlockStmt{
		List: []ast.Stmt{
			makeSimpleCallStmt("vtime", "Block"),
			sendstmt,
			makeSimpleCallStmt("vtime", "Unblock"),
		},
	}
}

func (t *fixChanVisitor) fixSelectStmt(bstmt *ast.BlockStmt, i int, selstmt *ast.SelectStmt) {
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
