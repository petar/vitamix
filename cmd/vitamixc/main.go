// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// TODO:
//	* Remove import of "time" package if not used other than for Now and Sleep
//	* Ensure there is no other package imported as "vtime"
//	* fallthough in select statements is not supported. check for it.
//	* We are assuming that pkg 'time' is imported as 'time'
//	* We only catch direct calls of the form 'time.Now()'.
//	* We would not catch indirect calls as in 'f := time.Now; f()'

var (
	flagSrc  *string = flag.String("src", ".", "Path to source directory")
	flagDest *string = flag.String("dest", "", "Path to destination directory")
)

func usage() {
	fmt.Printf("%s\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(1)
}

func FilterGoFiles(fi os.FileInfo) bool {
	name := fi.Name()
	return !fi.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

func main() {
	flag.Parse()

	if *flagDest == "" {
		usage()
	}
	fileSet := token.NewFileSet()
	pkgs, err := parser.ParseDir(fileSet, *flagSrc, FilterGoFiles, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %s\n", err)
		os.Exit(1)
	}

	for _, pkg := range pkgs {
		transformPackage(fileSet, pkg, *flagDest)
	}
}
