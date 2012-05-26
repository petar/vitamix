// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package vrewrite

import (
	"go/ast"
	"go/token"
)

func RewriteFile(fileSet *token.FileSet, file *ast.File) error {

	// addImport will automatically rename any existing package references with
	// conflicting name vtime to vtime_
	addImport(file, "github.com/petar/vitamix/vtime")

	// rewriteTimeCalls will rewrite time.Now and time.Sleep to
	// vtime.Now and vtime.Sleep
	needVtime, needTime := rewriteTimeCalls(file)

	if !needVtime {
		removeImport(file, "github.com/petar/vitamix/vtime")
	}

	if !needTime {
		removeImport(file, "time")
	}

	rewriteChanOps(fileSet, file)

	return nil
}

func RewritePackage(fileSet *token.FileSet, pkg *ast.Package) error {
	var err error
	for _, fileFile := range pkg.Files {
		if err0 := RewriteFile(fileSet, fileFile); err0 != nil {
			err = err0
		}
	}
	return err
}
