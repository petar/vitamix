package main

import (
	"errors"
	"path"
	"os"
	"strings"
)

// getGoRoot determines which GOPATH shuold be used, based on the
// the source directory and the current directory
func getGoRoot() (root string, err error) {
	wdir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	gopath := *flagGoPath
	if gopath == "" {
		gopath = os.ExpandEnv("$GOPATH")
	}
	for _, root = range strings.Split(gopath, ":") {
		root = path.Clean(root)
		if strings.HasPrefix(wdir, root) {
			return root, nil
		}
	}
	return "", errors.New("source dir does not belong to a GOPATH")
}

