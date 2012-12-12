// Copyright 2012 Petar Maymounkov. All rights reserved.
// Use of this source code is governed by a 
// license that can be found in the LICENSE file.

package main

import (
	"errors"
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

// XXX:
//	* No subdir recursion
//	* Print out is messy when comments are present

// TODO:
//	* fallthrough in select statements is not supported. detect it and complain.
//	* We only catch direct calls of the form 'time.Now()', 
//	  we would not catch indirect calls as in 'f := time.Now; f()'

func FilterGoFiles(fi os.FileInfo) bool {
	name := fi.Name()
	return !fi.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
}

const help = `vitamix InputSourceDir OutputSourceDir PkgPattern`

func usage() {
	println(help)
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
	if flag.NArg() != 3 {
		usage()
	}

	inSrcDir, outSrcDir, pkgPttrn := flag.Arg(0), flag.Arg(1), flag.Arg(2)

	pkgPaths, err := matchPattern(inSrcDir, pkgPttrn)
	if err != nil {
		println("Problem finding packages:", err.Error())
		os.Exit(1)
	}

	for _, pkgPath := range pkgPaths {
		println(pkgPath)
		if err = rewriteDir(path.Join(inSrcDir, pkgPath), path.Join(outSrcDir, pkgPath)); err != nil {
			println("Problem processing", pkgPath, ":", err.Error())
			os.Exit(1)
		}
	}
}

func matchPattern(srcDir, pkgPttrn string) ([]string, error) {
	var ellipses bool
	if strings.HasSuffix(pkgPttrn, "...") {
		ellipses = true
		pkgPttrn = path.Clean(pkgPttrn[:len(pkgPttrn)-len("...")])
	}
	var r []string
	q := []string{pkgPttrn}
	for len(q) > 0 {
		var p string
		p, q = q[0], q[1:]
		fi, err := os.Stat(path.Join(srcDir, p))
		if err != nil {
			return nil, err
		}
		if !fi.IsDir() {
			return nil, errors.New("not a directory")
		}
		r = append(r, p)
		if ellipses {
			if err = descendPkg(srcDir, p, &q); err != nil {
				return nil, err
			}
		}
	}
	return r, nil
}

func descendPkg(srcDir, p string, q *[]string) error {
	d, err := os.Open(path.Join(srcDir, p))
	if err != nil {
		return err
	}
	defer d.Close()

	fifi, err := d.Readdir(0)
	if err != nil {
		return err
	}

	for _, fi := range fifi {
		if fi.IsDir() {
			*q = append(*q, path.Join(p, fi.Name()))
		}
	}
	return nil
}

func rewriteDir(src, dest string) error {
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
	return nil
}
