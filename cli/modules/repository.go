// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"gitlab.com/antipy/antibuild/cli/internal/download"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

type (

	//ModuleRepository is a repository for modules
	ModuleRepository map[string]ModuleEntry
)

var (
	//ErrFailedGitRepositoryDownload means that the git repository could not be cloned
	ErrFailedGitRepositoryDownload = errors.NewError("failed to clone the git repository", 11)

	//ErrFailedModuleRepositoryDownload means the module repository list download failed
	ErrFailedModuleRepositoryDownload = errors.NewError("failed downloading the module repository list", 22)
)

//Download downloads a json file into a module repository
func (m *ModuleRepository) Download(url string) errors.Error {
	err := download.JSON(url, m)
	if err != nil {
		return ErrFailedModuleRepositoryDownload.SetRoot(err.Error())
	}

	return nil
}
