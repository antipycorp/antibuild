// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package templates

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/internal/download"
	"gitlab.com/antipy/antibuild/cli/internal/zip"
)

// Entry is a single entry for a template repository file
type Entry struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      struct {
		Type         string `json:"type"`
		URL          string `json:"url"`
		SubDirectory string `json:"subdirectory"`
	} `json:"source"`
}

// GetRepository from a url
func GetRepository(url string) (repo map[string]Entry, err error) {
	err = download.JSON(url, &repo)
	return
}

// Download a template
func Download(templateRepository map[string]Entry, template, outPath, templateBranch string) bool {
	if _, ok := templateRepository[template]; !ok {
		println("The selected template is not available in this repository.")
		return false
	}

	t := templateRepository[template]

	dir, err := ioutil.TempDir("", "antibuild")
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(dir, 0744)
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir)
	var src string

	switch t.Source.Type {
	case "zip":
		downloadFilePath := filepath.Join(dir, "download.zip")

		err = download.File(downloadFilePath, t.Source.URL, false)
		if err != nil {
			log.Fatal(err)
		}

		err = zip.Unzip(downloadFilePath, dir)
		if err != nil {
			log.Fatal(err)
		}
		os.Remove(downloadFilePath) //we dont want the zip hanging around in the template
		src = filepath.Join(dir, t.Source.SubDirectory)

	case "git":
		err = download.Git(dir, t.Source.URL, templateBranch)
		if err != nil {
			log.Fatal(err)
		}

		err = os.RemoveAll(filepath.Join(dir, ".git"))
		if err != nil {
			log.Fatal(err)
		}

		src = filepath.Join(dir, t.Source.SubDirectory)
	}

	info, err := os.Lstat(src)
	if err != nil {
		log.Fatal(err)
	}

	internal.DirCopy(src, outPath, info)
	if err != nil {
		log.Fatal(err)
	}

	println("Downloaded template.")
	return true
}
