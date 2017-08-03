// Copyright 2017 The go-importlint Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package importlint

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/buildutil"
)

type FindMode int

const (
	ExcludeVendor FindMode = 1 << iota
)

// FindAllPackage returns a list of all packages in all of the GOPATH trees
// in the given build context. If prefix is non-empty, only packages
// whose import paths begin with prefix are returned.
func (bc *BuildContext) FindAllPackage(ignores []string, mode FindMode) ([]*build.Package, error) {
	pkgs := []*build.Package{}
	done := make(map[string]bool)

	filepath.Walk(bc.root, func(path string, fi os.FileInfo, err error) error {
		if err != nil || !fi.IsDir() {
			return nil
		}

		// avoid .foo, _foo, and testdata directory trees.
		_, elem := filepath.Split(path)
		if elem == "pkg" || strings.HasPrefix(elem, ".") || strings.HasPrefix(elem, "_") || elem == "testdata" || (mode&ExcludeVendor != 0 && elem == "vendor") || matchIgnore(elem, ignores) {
			return filepath.SkipDir
		}

		name := filepath.ToSlash(path[len(bc.root):])
		if done[name] {
			return nil
		}
		done[name] = true

		if path != bc.root {
			// TODO(zchee): O(n)
			for _, gopath := range bc.gopaths {
				// check contains path in "src" directory
				if strings.Contains(path, srcDir(gopath)) {
					break
				}
				return filepath.SkipDir
			}
		}

		pkg, err := bc.ctxt.ImportDir(path, build.ImportMode(0))
		if err != nil && strings.Contains(err.Error(), "no buildable Go source files") {
			return nil
		}
		pkgs = append(pkgs, pkg)

		return nil
	})

	return pkgs, nil
}

func matchIgnore(elem string, ignores []string) bool {
	for _, e := range ignores {
		if elem == e {
			return true
		}
	}
	return false
}

// ParseDir wrapper of buildutil.ParseFile with BuildContext.
func ParseDir(fset *token.FileSet, bctx *BuildContext, path string, filter func(os.FileInfo) bool, mode parser.Mode) (map[string]*ast.Package, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	list, err := fd.Readdir(-1)
	if err != nil {
		return nil, err
	}

	pkgs := make(map[string]*ast.Package)
	var firstErr error
	for _, d := range list {
		if strings.HasSuffix(d.Name(), ".go") && (filter == nil || filter(d)) {
			filename := filepath.Join(path, d.Name())
			if src, err := buildutil.ParseFile(fset, bctx.ctxt, nil, filepath.Dir(filename), filename, mode); err == nil {
				name := src.Name.Name
				pkg, found := pkgs[name]
				if !found {
					pkg = &ast.Package{
						Name:  name,
						Files: make(map[string]*ast.File),
					}
					pkgs[name] = pkg
				}
				pkg.Files[filename] = src
			} else if firstErr == nil {
				firstErr = err
			}
		}
	}

	return pkgs, firstErr
}

func CheckDependency(pkgs map[string]*ast.Package, conf *Config) {
	importmap := make(map[string][]string)

	for name, obj := range pkgs {
		fmt.Printf("name: %+v\n", name)
		fmt.Printf("obj.Name: %+v\n", obj.Name)
		if conf.Layer[name] != nil {
			for id, file := range obj.Files {
				for _, imppkg := range file.Imports {
					importmap[id] = append(importmap[id], imppkg.Path.Value)
				}
			}
		}
	}
}
