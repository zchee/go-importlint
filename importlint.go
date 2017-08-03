// Copyright 2017 The go-importlint Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package importlint

import (
	"go/build"
	"path/filepath"
	"strings"
)

type BuildContext struct {
	ctxt    *build.Context
	root    string
	gopaths []string
}

func NewBuildContext(root string) BuildContext {
	bc := BuildContext{
		ctxt:    &build.Default,
		root:    root,
		gopaths: []string{build.Default.GOPATH},
	}

	if gbroot, yes := isGb(root); yes { // gb directory structure
		bc.root = gbroot
		bc.gopaths = []string{root, filepath.Join(root, "vendor")}
		bc.ctxt.GOPATH = root + string(filepath.ListSeparator) + filepath.Join(root, "vendor")
		bc.ctxt.SplitPathList = bc.splitPathList
		bc.ctxt.JoinPath = bc.joinPath
	} else { // general directory structure
		// split GOPATH if users set multiple directory path
		if paths := strings.Split(bc.ctxt.GOPATH, string(filepath.ListSeparator)); len(paths) >= 2 {
			bc.gopaths = paths
		}
	}

	return bc
}

func (b *BuildContext) Context() *build.Context {
	return b.ctxt
}
