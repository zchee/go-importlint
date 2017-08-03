// Copyright 2017 The go-importlint Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package importlint

import (
	"os"
	"path/filepath"
)

func isNotExist(path string) bool {
	if _, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		return true
	}
	return false
}

func srcDir(path string) string {
	return filepath.Join(path, "src")
}
