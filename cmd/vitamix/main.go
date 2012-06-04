// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"strings"
	. "github.com/petar/vitamix/vrewrite"
)

// TODO: Make subdir recursion optional
// Integrate with GOPATH to rewrite internal imports as well

// XXX:
//	* Print out is messy when comments are present
// TODO:
//	* fallthough in select statements is not supported. check for it.
//	* We only catch direct calls of the form 'time.Now()', 
//	  we would not catch indirect calls as in 'f := time.Now; f()'

func FilterGoFiles(fi os.FileInfo) bool {
	name := fi.Name()
	return !fi.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

var (
	flagDump   *bool   = flag.Bool("d", false, "Dump the AST of the source file")
	flagGoPath *string = flag.String("gopath", "", "Specify a GOPATH that overrides the environment variable")
)

func usage() {
	fmt.Printf("%s -d [source_file]\n", os.Args[0])
	fmt.Printf("   e.g. %s -d src.go\n", os.Args[0])
	fmt.Printf("%s [source_file] [dest_file]\n", os.Args[0])
	fmt.Printf("   e.g. %s src.go dest.go\n", os.Args[0])
	fmt.Printf("%s [src_pkg_path] [dest_pkg_path]\n", os.Args[0])
	fmt.Printf("   e.g. %s a/p b/x/y\n", os.Args[0])
	fmt.Printf("%s [src_pkg_go_glob] [dest_pkg_go_glob]\n", os.Args[0])
	fmt.Printf("   e.g. %s a/p/... b/x/y/...\n", os.Args[0])
	os.Exit(1)
}

func dump(filename string) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filename, nil, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %s\n", err)
		os.Exit(1)
	}
	ast.Print(fileSet, file)
}

func main() {
	flag.Parse()
	args := flag.Args()

	root, err := getGoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine source root directory (%s)\n", err)
		os.Exit(1)
	}
	fmt.Printf("Using source root directory `%s`\n", root)

	// Are we in dump mode?
	if *flagDump {
		if len(args) != 1 {
			usage()
		}
		srci, err := os.Stat(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Problem accessing source (%s)\n", err)
			os.Exit(1)
		}
		if srci.IsDir() {
			fmt.Fprintf(os.Stderr, "Expecting individual source file for dumping\n")
			os.Exit(1)
		}
		dump(args[0])
		return
	}

	// Need two arguments for source-to-source transformation
	/*
	if len(args) != 2 {
		usage()
	}
	src, dest := args[1], args[2]
	srci, err := os.Stat(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Problem accessing source (%s)\n", err)
		os.Exit(1)
	}

	??
*/
	// Single source file mode
	/*
	if !srcInfo.IsDir() {
		fileSet := token.NewFileSet()
		file, err := parser.ParseFile(fileSet, src, nil, parser.ParseComments)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Problem parsing file '%s' (%s)\n", src, err)
			os.Exit(1)
		}
		RewriteFile(fileSet, file)
		if err = PrintToFile(dest, fileSet, file); err != nil {
			os.Exit(1)
		}
		return
	}
	*/

	// Determine the Go root directory
	/*
	root, err := getGoRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot determine the source root directory (%s)\n", err)
		os.Exit(1)
	}
	fmt.Printf("Using source root directory `%s`", root)

	if err := processDir(src, dest); err != nil {
		fmt.Fprintf(os.Stderr, "Problem processing directory (%s)\n", err)
		os.Exit(1)
	}
	*/
}

func processDir(src, dest string) error {
	fmt.Printf("Rewriting directory %s ——> %s\n", src, dest)

	// Make destination directory if it doesn't exist
	if err := os.MkdirAll(dest, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Target directory cannot be created (%s).\n", err)
		return err
	}

	// Parse source directory
	fileSet := token.NewFileSet()
	pkgs, err := parser.ParseDir(fileSet, src, FilterGoFiles, parser.ParseComments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Problem parsing directory (%s)\n", err)
		return err
	}

	// In every package, rewrite every file
	for _, pkg := range pkgs {
		for _, fileFile := range pkg.Files {
			RewriteFile(fileSet, fileFile)
			position := fileSet.Position(fileFile.Package)
			_, filename := path.Split(position.Filename)
			fmt.Printf("  %s ==> %s\n", filename, path.Join(dest, filename))
			if err = PrintToFile(path.Join(dest, filename), fileSet, fileFile); err != nil {
				fmt.Fprintf(os.Stderr, "Problem determining source filename (%s)", err)
				os.Exit(1)
			}
		}
	}

	// Recurse the subdirectories of src
	srcDir, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcDir.Close()

	for { 
		fi_, err := srcDir.Readdir(10)
		for _, fi := range fi_ {
			if !fi.IsDir() {
				continue
			}
			if err := processDir(path.Join(src, fi.Name()), path.Join(dest, fi.Name())); err != nil {
				return err
			}
		}
		if err != nil {
			break
		}
	}
	return nil
}
