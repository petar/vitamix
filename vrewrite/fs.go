package vrewrite

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
)

// PrintToFile writes the file AST node to the named file
func PrintToFile(name string, fileSet *token.FileSet, fileFile *ast.File) error {
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
