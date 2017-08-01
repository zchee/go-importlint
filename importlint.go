// Copyright 2017 The go-importlint Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package importlint

import (
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

type BuildContext struct {
	ctxt *build.Context

	gbroot  string
	gbpaths []string
}

func NewBuildContext(dir string) BuildContext {
	bc := context(&build.Default)
	if root, yes := isGb(dir); yes {
		bc.gbroot = root
		bc.gbpaths = []string{root, filepath.Join(root, "vendor")}
		bc.ctxt.GOPATH = root + string(filepath.ListSeparator) + filepath.Join(root, "vendor")
		bc.ctxt.SplitPathList = bc.splitPathList
		bc.ctxt.JoinPath = bc.joinPath
	}

	return bc
}

func context(bctx *build.Context) BuildContext {
	bc := BuildContext{
		ctxt: new(build.Context),
	}

	bc.ctxt.GOARCH = bctx.GOARCH
	bc.ctxt.GOOS = bctx.GOOS
	bc.ctxt.GOROOT = bctx.GOROOT
	bc.ctxt.GOPATH = bctx.GOPATH
	bc.ctxt.CgoEnabled = bctx.CgoEnabled
	bc.ctxt.UseAllFiles = bctx.UseAllFiles
	bc.ctxt.Compiler = bctx.Compiler
	bc.ctxt.BuildTags = bctx.BuildTags
	bc.ctxt.ReleaseTags = bctx.ReleaseTags
	bc.ctxt.InstallSuffix = bctx.InstallSuffix

	return bc
}

func match(s, prefix string) (string, bool) {
	rest := strings.TrimPrefix(s, prefix)
	return rest, len(rest) < len(s)
}

type FindMode int

const (
	ExcludeVendor FindMode = 1 << iota
)

// FindAllPackage returns a list of all packages in all of the GOPATH trees
// in the given build context. If prefix is non-empty, only packages
// whose import paths begin with prefix are returned.
func FindAllPackage(bc BuildContext, ignores []string, mode FindMode) ([]*build.Package, error) {
	var (
		pkgs []*build.Package
		done = make(map[string]bool)
	)

	filepath.Walk(bc.gbroot, func(path string, fi os.FileInfo, err error) error {
		if err != nil || !fi.IsDir() {
			return nil
		}

		// avoid .foo, _foo, and testdata directory trees.
		_, elem := filepath.Split(path)
		if elem == "pkg" || strings.HasPrefix(elem, ".") || strings.HasPrefix(elem, "_") || elem == "testdata" || (mode&ExcludeVendor != 0 && elem == "vendor") || matchIgnore(elem, ignores) {
			return filepath.SkipDir
		}

		name := filepath.ToSlash(path[len(bc.gbroot):])
		if done[name] {
			return nil
		}
		done[name] = true

		pkg, err := bc.ctxt.ImportDir(path, build.IgnoreVendor)
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
