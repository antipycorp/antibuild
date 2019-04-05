// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package internal

import "gitlab.com/antipy/antibuild/cli/internal"

// ModuleRepositoryEntry is a single entry for a module repository file
type ModuleRepositoryEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      struct {
		Type         string `json:"type"`
		URL          string `json:"url"`
		SubDirectory string `json:"subdirectory"`
	} `json:"source"`
	Compiled map[string]map[string]string
}

// TemplateRepositoryEntry is a single entry for a template repository file
type TemplateRepositoryEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      struct {
		Type         string `json:"type"`
		URL          string `json:"url"`
		SubDirectory string `json:"subdirectory"`
	} `json:"source"`
}

// GetModuleRepository from a url
func GetModuleRepository(url string) (repo map[string]ModuleRepositoryEntry, err error) {
	err = internal.DownloadJSON(url, &repo)
	return
}

// GetTemplateRepository from a url
func GetTemplateRepository(url string) (repo map[string]TemplateRepositoryEntry, err error) {
	err = internal.DownloadJSON(url, &repo)
	return
}
