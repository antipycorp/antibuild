// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

var newSurvey = []*survey.Question{
	{
		Name:   "name",
		Prompt: &survey.Input{Message: "What should the name of the project be?"},
		Validate: func(input interface{}) error {
			if _, err := input.(string); err == false {
				return errors.New("not a string")
			}

			match, err := regexp.MatchString("\\A[a-z-]{3,}\\z", input.(string))

			if err != nil {
				return err
			} else if match == false {
				return errors.New("Should be at least 3 characters and should only include a-z and - (dash)")
			}

			return nil
		},
	},
	{
		Name: "template",
		Prompt: &survey.Select{
			Message: "Choose a starting template:",
			Options: []string{"basic", "homepage", "newspage"},
			Default: "homepage",
		},
	},
	{
		Name: "default_modules",
		Prompt: &survey.MultiSelect{
			Message: "Select any modules you want to pre install now (can also not choose any):",
			Options: []string{"arithmetic", "languages", "escaper", "markdown", "firebase"},
		},
	},
}

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Make a new antibuild project.",
	Long:  `Generate a new antibuild project. To get started run "antibuild new" and follow the prompts.`,
	Run: func(cmd *cobra.Command, args []string) {
		answers := struct {
			Name           string   `survey:"name"`
			Template       string   `survey:"template"`
			DefaultModules []string `survey:"default_modules"`
		}{}

		err := survey.Ask(newSurvey, &answers)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if _, err := ioutil.ReadDir(answers.Name); os.IsNotExist(err) {
			dir, err := ioutil.TempDir("", "antibuild")
			if err != nil {
				log.Fatal(err)
			}

			defer os.RemoveAll(dir) // clean up

			downloadFilePath := filepath.Join(dir, "download.zip")

			err = downloadFile(downloadFilePath, "https://build.antipy.com/cli/examples/basic.zip")
			if err != nil {
				log.Fatal(err)
			}

			_, err = unzip(downloadFilePath, answers.Name)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("Success. Run these commands to get started:")
			fmt.Println("")
			fmt.Println("cd " + answers.Name)
			fmt.Println("antibuild develop")
			fmt.Println("")
			fmt.Println("Need help? Look at our docs: https://build.antipy.com/documentation")
			fmt.Println("")
			fmt.Println("")
			return
		}

		fmt.Println("Failed.")
	},
}

func unzip(src string, dest string) ([]string, error) {
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}
		defer rc.Close()

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {

			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)

		} else {

			// Make File
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return filenames, err
			}

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return filenames, err
			}

			_, err = io.Copy(outFile, rc)

			// Close the file without defer to close before next iteration of loop
			outFile.Close()

			if err != nil {
				return filenames, err
			}

		}
	}
	return filenames, nil
}

func downloadFile(filepath string, url string) error {
	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
