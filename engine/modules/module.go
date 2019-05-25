// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"bytes"
	"strings"

	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

type (
	// Modules is the part of the config file that handles modules
	Modules struct {
		Dependencies map[string]*Module                `json:"dependencies"`
		Config       map[string]map[string]interface{} `json:"config,omitempty"`
		SPPs         []string                          `json:"spps,omitempty"`
	}

	// Module with info about the path and version
	//TODO: add the module name to this
	Module struct {
		Repository string
		Version    string
	}
)

var moduleParseError = errors.NewError("failled to parse a module string", 10)

// UnmarshalJSON on a module
func (m *Module) UnmarshalJSON(data []byte) error {
	data = bytes.Trim(data, "\"")
	m.fromBytes(data)
	return nil
}

func (m *Module) fromBytes(data []byte) error {
	split := bytes.Split(data, []byte("@"))
	if len(split) < 1 || len(split) > 2 {
		return moduleParseError
	}
	if len(split[0]) == 0 {
		return moduleParseError
	}

	m.Repository = string(split[0])
	if len(split) == 2 {
		if len(split[1]) == 0 {
			return moduleParseError
		}
		m.Version = string(split[1])
	} else {
		m.Version = "latest"
	}

	return nil
}

func (m *Module) bytes() []byte {
	vlen := 0
	if len(m.Version) != 0 {
		vlen = 1 + len(m.Version)
	}
	res := make([]byte, 0, 1+len(m.Repository)+vlen+1)
	res = append(res, byte('"'))
	res = append(res, []byte(m.Repository)...)

	if len(m.Version) != 0 {
		res = append(res, byte('@'))
		res = append(res, []byte(m.Repository)...)
	}
	res = append(res, byte('"'))
	return res
}

// ParseModuleString for config and cli
func ParseModuleString(moduleString string) (m *Module, err errors.Error) {
	m = new(Module)

	err = errors.Import(m.fromBytes([]byte(strings.Trim(moduleString, "\""))))

	return
}

// MarshalJSON on a module
func (m *Module) MarshalJSON() ([]byte, error) {
	return m.bytes(), nil
}
