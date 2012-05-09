// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package main

import (
	//"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path"
	"strings"
)

// TODO:
//	* Remove import of "time" package if not used other than for Now and Sleep
//	* Ensure there is no other package imported as "vtime"
//	* fallthough in select statements is not supported. check for it.
//	* We are assuming that pkg 'time' is imported as 'time'
//	* We only catch direct calls of the form 'time.Now()'.
//	* We would not catch indirect calls as in 'f := time.Now; f()'

func usage() {
	fmt.Printf("%s [source_file_or_dir] [dest_file_or_dir]\n", os.Args[0])
	//flag.PrintDefaults()
	os.Exit(1)
}

func FilterGoFiles(fi os.FileInfo) bool {
	name := fi.Name()
	return !fi.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func main() {
	//flag.Parse()
	if len(os.Args) != 3 {
		usage()
	}
	src, tgt := os.Args[1], os.Args[2]
	srcInfo, err := os.Stat(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Problem accessing source (%s)\n", err)
		os.Exit(1)
	}
	if srcInfo.IsDir() {
		// Verify that tgt is an existing directory
		tgtInfo, err := os.Stat(tgt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Problem accessing target (%s)\n", err)
			os.Exit(1)
		}
		if !tgtInfo.IsDir() {
			fmt.Fprintf(os.Stderr, "Target must be an existing directory\n")
			os.Exit(1)
		}

		fileSet := token.NewFileSet()
		pkgs, err := parser.ParseDir(fileSet, src, FilterGoFiles, parser.ParseComments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Problem parsing directory (%s)\n", err)
			os.Exit(1)
		}
		for _, pkg := range pkgs {
			for _, fileFile := range pkg.Files {
				processFile(fileSet, fileFile)
				position := fileSet.Position(fileFile.Package)
				if err = printToFile(path.Join(tgt, position.Filename), fileSet, fileFile); err != nil {
					os.Exit(1)
				}
			}
		}
	} else {
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, src, nil, parser.ParseComments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Problem parsing file '%s' (%s)\n", src, err)
			os.Exit(1)
		}
		processFile(fileSet, file)
		if err = printToFile(tgt, fileSet, file); err != nil {
			os.Exit(1)
		}
	}

}

func printToFile(name string, fileSet *token.FileSet, fileFile *ast.File) error {
	w, err := os.Create(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Problem creating target file '%s' (%s)\n", name, err)
		return err
	}
	if err = printer.Fprint(w, fileSet, fileFile); err != nil {
		fmt.Fprintf(os.Stderr, "Problem writing to target file '%s' (%s)\n", name, err)
		w.Close()
		return err
	}
	if err = w.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Problem closing target file '%s' (%s)\n", name, err)
		return err
	}
	return nil
}

func processFile(fileSet *token.FileSet, file *ast.File) {
	// Add import of "vtime" package
	addImport(file, "github.com/petar/vitamix/vtime")
	// Rewrite time.Now and time.Sleep calls
	fixCall(file)
	// Rewrite chan operations
	fixChan(fileSet, file)
}
