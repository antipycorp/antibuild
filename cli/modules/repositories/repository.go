// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package repositories

import (
	"gitlab.com/antipy/antibuild/cli/engine/modules"
	"gitlab.com/antipy/antibuild/cli/internal/download"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

type (
	// Entry is a single entry for a module repository file
	Entry struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Source      struct {
			Type         string `json:"type"`
			URL          string `json:"url"`
			SubDirectory string `json:"subdirectory"`
		} `json:"source"`
		CompiledStatic  map[string]map[string]map[string]string `json:"compiled_static"`
		CompiledDynamic struct {
			URL          string              `json:"url"`
			Vesions      []string            `json:"versions"`
			OSArchCombos map[string][]string `json:"os_arch_combos"`
		} `json:"compiled_dynamic"`
		LatestVersion string `json:"latest"`
	}

	// Repository is a repository for modules
	Repository map[string]Entry
)

const (
	// StandardRepository for modules
	StandardRepository = modules.StandardRepository

	// NoRepositorySpecified is when you dont pass the -m flag
	NoRepositorySpecified = ""
)

var (
	//ErrFailedModuleRepositoryDownload means the module repository list download failed
	ErrFailedModuleRepositoryDownload = errors.NewError("failed downloading the module repository list", 22)
)

//Download downloads a json file into a module repository
func (m *Repository) Download(url string) errors.Error {
	err := download.JSON(url, m)
	if err != nil {
		return ErrFailedModuleRepositoryDownload.SetRoot(err.Error())
	}

	return nil
}
