// Copyright 2017 The go-importlint Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"go/parser"
	"go/token"
	"log"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	importlint "github.com/zchee/go-importlint"
)

func main() {
	flag.Parse()

	var path string
	if flag.NArg() > 0 {
		path = flag.Arg(0)
	} else {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(errors.Wrap(err, "could not get cwd directory"))
		}
		path = wd
	}

	fset := token.NewFileSet()
	bctx := importlint.NewBuildContext(path)
	pkgs, err := importlint.ParseDir(fset, &bctx, path, nil, parser.AllErrors)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "could not parse %s directly", path))
	}

	spew.Dump(pkgs)
}
