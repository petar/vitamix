// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package main

import (
	//"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path"
	"strings"
	. "github.com/petar/vitamix/vrewrite"
)

// XXX:
//	* Print out is messy when comments are present
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
				RewriteFile(fileSet, fileFile)
				position := fileSet.Position(fileFile.Package)
				if err = PrintToFile(path.Join(tgt, position.Filename), fileSet, fileFile); err != nil {
					fmt.Fprintf(os.Stderr, "Problem determining source filename (%s)", err)
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
		RewriteFile(fileSet, file)
		if err = PrintToFile(tgt, fileSet, file); err != nil {
			os.Exit(1)
		}
	}

}
