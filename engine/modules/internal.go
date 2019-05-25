// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"bufio"
	"io"

	"github.com/blang/semver"

	"gitlab.com/antipy/antibuild/api/host"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	abm_file "gitlab.com/antipy/antibuild/std/file/handler"
	abm_json "gitlab.com/antipy/antibuild/std/json/handler"
	abm_language "gitlab.com/antipy/antibuild/std/language/handler"
	abm_markdown "gitlab.com/antipy/antibuild/std/markdown/handler"
	abm_math "gitlab.com/antipy/antibuild/std/math/handler"
	abm_noescape "gitlab.com/antipy/antibuild/std/noescape/handler"
	abm_util "gitlab.com/antipy/antibuild/std/util/handler"
	abm_yaml "gitlab.com/antipy/antibuild/std/yaml/handler"
)

type (
	internalModule struct {
		start      func(io.Reader, io.Writer)
		version    string
		repository string
	}
)

const (
	// internalRepo is the default module repository
	internalRepo = STDRepo
)

const (
	//HaveSameVersion is for when you have the same version as internal module
	HaveSameVersion = iota
	//HaveHigherVersion is for when you have a higher version as internal module
	HaveHigherVersion
	//HaveLowerVersion is for when you have a lower version as internal module
	HaveLowerVersion
	//HaveNoVersion is for when you dont have the same module as internal module
	HaveNoVersion
)

var (
	//ErrNotExistInternal means the module does not exist as an internal module
	ErrNotExistInternal = errors.NewError("module does not exist as an internal module", 1)
)

// InternalModules that the are integrated into the antibuild binary
var InternalModules = map[string]internalModule{
	"file": internalModule{
		start:      abm_file.Handler,
		version:    abm_file.Version,
		repository: internalRepo,
	},
	"json": internalModule{
		start:      abm_json.Handler,
		version:    abm_json.Version,
		repository: internalRepo,
	},
	"language": internalModule{
		start:      abm_language.Handler,
		version:    abm_language.Version,
		repository: internalRepo,
	},
	"markdown": internalModule{
		start:      abm_markdown.Handler,
		version:    abm_markdown.Version,
		repository: internalRepo,
	},
	"math": internalModule{
		start:      abm_math.Handler,
		version:    abm_math.Version,
		repository: internalRepo,
	},
	"noescape": internalModule{
		start:      abm_noescape.Handler,
		version:    abm_noescape.Version,
		repository: internalRepo,
	},
	"util": internalModule{
		start:      abm_util.Handler,
		version:    abm_util.Version,
		repository: internalRepo,
	},
	"yaml": internalModule{
		start:      abm_yaml.Handler,
		version:    abm_yaml.Version,
		repository: internalRepo,
	},
}

func (mod internalModule) load() (io.Reader, io.Writer) {
	in, stdin := io.Pipe()
	stdout, out := io.Pipe()

	in2 := bufio.NewReader(in)
	stdout2 := bufio.NewReader(stdout)

	go mod.start(in2, out)

	return stdout2, stdin
}

//LoadInternalModule loads an internal module, fails if no module is available
func LoadInternalModule(meta *Module, name string, log host.Logger) (io.Reader, io.Writer, string, errors.Error) {
	if v, ok := InternalModules[name]; ok {

		if meta.Version == v.version {
			stdout2, stdin := v.load()

			return stdout2, stdin, meta.Version, nil
		}
	}
	return nil, nil, "", ErrNotExistInternal
}

// MatchesInternalModule checks if a module is available as internal module
// The version MUST be semver compatible otherwise it will fail
func MatchesInternalModule(name, version, repo string) int {

	if internal, ok := InternalModules[name]; ok && internal.repository == repo {
		if internal.version == version {
			return HaveSameVersion
		}
		// The version should be semver,
		iv := semver.MustParse(version)
		v := semver.MustParse(version)
		if iv.GT(v) {
			return HaveHigherVersion
		}
		if iv.LT(v) {
			return HaveLowerVersion
		}
	}

	return HaveNoVersion
}
