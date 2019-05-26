// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package internal

import "gitlab.com/antipy/antibuild/cli/internal/download"

// TemplateEntry is a single entry for a template repository file
type TemplateEntry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      struct {
		Type         string `json:"type"`
		URL          string `json:"url"`
		SubDirectory string `json:"subdirectory"`
	} `json:"source"`
}

// GetTemplateRepository from a url
func GetTemplateRepository(url string) (repo map[string]TemplateEntry, err error) {
	err = download.JSON(url, &repo)
	return
}
