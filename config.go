// Copyright 2017 The go-importlint Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package importlint

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
	"github.com/pkg/errors"
)

type Config struct {
	Project string              `yaml:"project"`
	Layer   map[string][]string `yaml:"layer"`
}

func ParseConfig(path string) (*Config, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read %s file", path)
	}

	conf := new(Config)
	if err := yaml.Unmarshal(buf, &conf); err != nil {
		return nil, err
	}

	return conf, nil
}
