// Copyright 2017 The go-importlint Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package importlint

import (
	"errors"
	"fmt"
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

// isGb check the current buffer directory whether gb directory structure.
// Return the gb project root path and boolean.
func isGb(dir string) (string, bool) {
	root, err := findGbProjectRoot(dir)
	if err != nil {
		return "", false
	}

	// Check root directory whether "vendor", and overwrite root path to parent of vendor directory.
	if filepath.Base(root) == "vendor" {
		root = filepath.Dir(root)
	}
	// findGbProjectRoot gets the GOPATH root if go directory structure.
	// Recheck use vendor directory.
	vendorDir := filepath.Join(root, "vendor")
	if isNotExist(vendorDir) {
		return dir, false
	}
	return root, true
}

// findGbProjectRoot works upwards from path seaching for the src/ directory
// which identifies the project root.
// Code taken directly from constabulary/gb.
//  github.com/constabulary/gb/cmd/path.go
func findGbProjectRoot(path string) (string, error) {
	if path == "" {
		return "", errors.New("project root is blank")
	}
	start := path
	for path != filepath.Dir(path) {
		root := filepath.Join(path, "src")
		if isNotExist(root) {
			path = filepath.Dir(path)
			continue
		}
		return path, nil
	}
	return "", fmt.Errorf("could not find project root in %q or its parents", start)
}

func (b *BuildContext) splitPathList(list string) []string {
	if b.gbroot != "" {
		return b.gbpaths
	}
	return filepath.SplitList(list)
}

func (b *BuildContext) joinPath(elem ...string) string {
	res := filepath.Join(elem...)

	if b.gbroot != "" {
		// Want to rewrite "$GBROOT/(vendor/)?pkg/$GOOS_$GOARCH(_)?"
		// into "$GBROOT/pkg/$GOOS-$GOARCH(-)?".
		// Note: gb doesn't use vendor/pkg.
		if gbrel, err := filepath.Rel(b.gbroot, res); err == nil {
			gbrel = filepath.ToSlash(gbrel)
			gbrel, _ = match(gbrel, "vendor/")
			if gbrel, ok := match(gbrel, fmt.Sprintf("pkg/%s_%s", b.ctxt.GOOS, b.ctxt.GOARCH)); ok {
				gbrel, hasSuffix := match(gbrel, "_")

				// Reassemble into result.
				if hasSuffix {
					gbrel = "-" + gbrel
				}
				gbrel = fmt.Sprintf("pkg/%s-%s/", b.ctxt.GOOS, b.ctxt.GOARCH) + gbrel
				gbrel = filepath.FromSlash(gbrel)
				res = filepath.Join(b.gbroot, gbrel)
			}
		}
	}

	return res
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
		if strings.HasPrefix(elem, ".") || strings.HasPrefix(elem, "_") || elem == "testdata" || (mode&ExcludeVendor != 0 && elem == "vendor") || matchIgnore(elem, ignores) {
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
