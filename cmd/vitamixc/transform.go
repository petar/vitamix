// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
)

// TODO:
//	* Remove import of "time" package if not used other than for Now and Sleep
//	* Ensure there is no other package imported as "vtime"
//	* fallthough in select statements is not supported. check for it.

func transformPackage(fileSet *token.FileSet, pkg *ast.Package, destDir string) {
	for fileName, fileFile := range pkg.Files {
		fmt.Printf("——— virtualizing '%s' ———\n", fileName)
		transformFile(fileSet, fileFile, destDir)
	}
}

func transformFile(fileSet *token.FileSet, file *ast.File, destDir string) {
	// Add import of "vtime" package
	addImport(file, "github.com/petar/vitamix/vtime")
	// Rewrite go statements
	fixGo(file)
	// Rewrite time.Now and time.Sleep calls
	fixCall(file)
	// Rewrite chan operations
	fixChan(fileSet, file)

	printer.Fprint(os.Stdout, fileSet, file)
}
