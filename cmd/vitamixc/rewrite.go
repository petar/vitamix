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

// XXX: Make sure label statements produce an error
// XXX: Take care of label statements
// XXX: Check imports for newline printing insights

func fixChan(fset *token.FileSet, file *ast.File) {
	if err := Rewrite(fset, file); err != nil {
		//fmt.Fprintf(os.Stderr, "Rewrite errors parsing '%s':\n%s\n", file.Name.Name, err)
		fmt.Fprintf(os.Stderr, "—— Encountered errors while parsing\n")
	}
}

// Rewrite creates a new rewriting frame
func Rewrite(fset *token.FileSet, node ast.Node) error {
	rwv := &rewriteVisitor{}
	rwv.frame.Init(fset)
	ast.Walk(rwv, node)
	return rwv.Error()
}


// RecurseRewrite creates a new rewriting frame as a callee from the frame caller
func RecurseRewrite(caller Framed, node ast.Node) error {
	rwv := &rewriteVisitor{}
	rwv.frame.InitRecurse(caller)
	ast.Walk(rwv, node)
	return rwv.Error()
}

// rewriteVisitor is an AST frame that traverses down the AST until it hits a block
// statement, within which it rewrites the statement-level channel operations. 
// This visitor itself does not traverse below the statements of the block statement.
// It does however call another visitor type to continue below those statements.
type rewriteVisitor struct {
	frame
}

// Frame implements Framed.Frame
func (t *rewriteVisitor) Frame() *frame {
	return &t.frame
}

// Visit implements ast.Visistor's Visit method
func (t *rewriteVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return t
	}
	bstmt, ok := node.(*ast.BlockStmt)
	// If node is not a block statement, it means we are recursing down the
	// AST and we haven't hit a block statement yet. 
	if !ok {
		// Keep recursing
		return t
	}

	// Rewrite each statement of a block statement and stop the recursion of this visitor
	var list []ast.Stmt
	for _, stmt := range bstmt.List {
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
		// TODO: Handle the case when an argument contains chan operations
		RecurseProhibit(t, arg)
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
					makeSimpleCallStmt("vtime", "Die", gostmt.Call.Pos()),
				},
			},
		},
	}
	return []ast.Stmt{
		makeSimpleCallStmt("vtime", "Go", gostmt.Pos()),
		gostmt,
	}
}

func (t *rewriteVisitor) rewriteRecvStmt(stmt ast.Stmt) []ast.Stmt {
	// Prohibit inner blocks from having channel operation nodes
	switch q := stmt.(type) {
	case *ast.AssignStmt:
		for _, expr := range q.Lhs {
			// TODO: Handle channel operations inside LHS of assignments
			RecurseProhibit(t, expr)
		}
		for _, expr := range q.Rhs {
			if expr == nil {
				continue
			}
			// TODO: Handle channel operations inside RHS of assignments
			if ue := filterRecvExpr(expr); ue != nil {
				RecurseProhibit(t, ue.X)
			} else {
				RecurseProhibit(t, expr)
			}
		}
	case *ast.ExprStmt:
		if q == nil {
			break
		}
		// TODO: Handle channel operations inside RHS of assignments
		if ue := filterRecvExpr(q.X); ue != nil {
			RecurseProhibit(t, ue.X)
		} else {
			RecurseProhibit(t, q)
		}
	default:
		panic("unreach")
	}
	// Rewrite receive statement itself
	return []ast.Stmt{
		makeSimpleCallStmt("vtime", "Block", stmt.Pos()),
		stmt,
		makeSimpleCallStmt("vtime", "Unblock", stmt.Pos()),
	}
}

func (t *rewriteVisitor) rewriteSendStmt(sendstmt *ast.SendStmt) []ast.Stmt {
	// Rewrite lower level nodes
	// TODO: Allow channel operations inside channel and value fields of send expression
	RecurseProhibit(t, sendstmt.Chan)
	RecurseProhibit(t, sendstmt.Value)
	// Rewrite send statement itself
	return []ast.Stmt{
		makeSimpleCallStmt("vtime", "Block", sendstmt.Pos()),
		sendstmt,
		makeSimpleCallStmt("vtime", "Unblock", sendstmt.Pos()),
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
			[]ast.Stmt{ makeSimpleCallStmt("vtime", "Unblock", comm.Pos()) },
			body...,
		)
	}
	// Surround the select by a block statement and prefix it with a call to vtime.Block
	return []ast.Stmt{
		makeSimpleCallStmt("vtime", "Block", selstmt.Pos()),
		selstmt,
	}
}
