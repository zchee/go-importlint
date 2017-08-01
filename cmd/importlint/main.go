// Copyright 2017 The go-importlint Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"flag"
	"fmt"
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

	bc := importlint.NewBuildContext(path)
	pkgs, err := bc.FindAllPackage(nil, importlint.ExcludeVendor)
	if err != nil {
		log.Fatal(errors.Wrapf(err, "could not find packages on %s", path))
	}

	buf := new(bytes.Buffer)
	for _, pkg := range pkgs {
		if !pkg.IsCommand() {
			buf.WriteString(pkg.Dir + "\n")
		}
	}

	fmt.Print(buf.String())

	spew.Dump(bc.Context())
}
