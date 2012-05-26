// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package vrewrite

import (
	"go/ast"
	"go/token"
)

func removeImport(file *ast.File, ipath string) {
	var j int
	for j < len(file.Decls) {
		gen, ok := file.Decls[j].(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			j++
			continue
		}
		var i int
		for i < len(gen.Specs) {
			impspec := gen.Specs[i].(*ast.ImportSpec)
			if importPath(impspec) != ipath {
				i++
				continue
			}
			// If we found a match, pop the spec from gen.Specs
			gen.Specs[i] = gen.Specs[len(gen.Specs)-1]
			gen.Specs = gen.Specs[:len(gen.Specs)-1]
		}
		if len(gen.Specs) > 0 {
			j++
			continue
		}
		// Remove entire import decl if no imports left in it
		file.Decls[j] = file.Decls[len(file.Decls)-1]
		file.Decls = file.Decls[:len(file.Decls)-1]
	}
}
