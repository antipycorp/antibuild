// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site

import (
	"os"
	"path/filepath"

	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	ui "gitlab.com/antipy/antibuild/cli/internal/log"
)

//Execute the templates of a []Site into the final files
func Execute(sites []*Site, log *ui.UI) errors.Error {
	return execute(sites, log)
}

func execute(sites []*Site, log *ui.UI) errors.Error {
	// copy static folder
	if StaticFolder != "" && OutputFolder != "" {
		log.Debug("Copying static folder")

		info, err := os.Lstat(StaticFolder)
		if err != nil {
			return ErrFailedStatic.SetRoot(err.Error())
		}

		err = internal.GenCopy(StaticFolder, OutputFolder, info)
		if err != nil {
			return ErrFailedStatic.SetRoot(err.Error())
		}

		log.Debug("Finished copying static folder")
	}

	//export every template
	for _, site := range sites {
		log.Debugf("Building page for %s", site.Slug)

		err := executeTemplate(site)
		if err != nil {
			return err
		}
	}
	return nil
}

func executeTemplate(site *Site) errors.Error {
	//prefix the slug with the output folder
	fileLocation := filepath.Join(OutputFolder, site.Slug)

	//check all folders in the path of the output file
	err := os.MkdirAll(filepath.Dir(fileLocation), 0766)
	if err != nil {
		return ErrFailedCreateFS.SetRoot(err.Error())
	}

	//create the file
	file, err := os.Create(fileLocation)
	if err != nil {
		return ErrFailedCreateFS.SetRoot(err.Error())
	}

	//fill the file by executing the template
	err = globalTemplates[site.Template].ExecuteTemplate(file, "html", site.Data)
	if err != nil {
		return ErrFailedTemplate.SetRoot(err.Error())
	}

	file.Close() // we have to do this otherwise antibuild silently crashes on NTFS filesystems

	return nil
}
