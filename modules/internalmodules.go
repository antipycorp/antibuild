// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	abm_file "gitlab.com/antipy/antibuild/std/file/handler"
	abm_json "gitlab.com/antipy/antibuild/std/json/handler"
	abm_language "gitlab.com/antipy/antibuild/std/language/handler"
	abm_markdown "gitlab.com/antipy/antibuild/std/markdown/handler"
	abm_math "gitlab.com/antipy/antibuild/std/math/handler"
	abm_noescape "gitlab.com/antipy/antibuild/std/noescape/handler"
	abm_util "gitlab.com/antipy/antibuild/std/util/handler"
	abm_yaml "gitlab.com/antipy/antibuild/std/yaml/handler"
)

// InternalModules that the are integrated into the antibuild binary
var InternalModules = map[string]internalMod{
	"file": internalMod{
		start:      abm_file.Handler,
		version:    abm_file.Version,
		repository: STDRepo,
	},
	"json": internalMod{
		start:      abm_json.Handler,
		version:    abm_json.Version,
		repository: STDRepo,
	},
	"language": internalMod{
		start:      abm_language.Handler,
		version:    abm_language.Version,
		repository: STDRepo,
	},
	"markdown": internalMod{
		start:      abm_markdown.Handler,
		version:    abm_markdown.Version,
		repository: STDRepo,
	},
	"math": internalMod{
		start:      abm_math.Handler,
		version:    abm_math.Version,
		repository: STDRepo,
	},
	"noescape": internalMod{
		start:      abm_noescape.Handler,
		version:    abm_noescape.Version,
		repository: STDRepo,
	},
	"util": internalMod{
		start:      abm_util.Handler,
		version:    abm_util.Version,
		repository: STDRepo,
	},
	"yaml": internalMod{
		start:      abm_yaml.Handler,
		version:    abm_yaml.Version,
		repository: STDRepo,
	},
}
