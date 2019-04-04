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

package new

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"

	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/builder/config"
	cmdInternal "gitlab.com/antipy/antibuild/cli/cmd/internal"
	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	"gopkg.in/AlecAivazis/survey.v1"
)

var (
	nameregex = regexp.MustCompile("[a-z-]{3,}")

	//ErrInvalidInput is when the template failed building
	ErrInvalidInput = errors.NewError("invalid input", 1)
	//ErrInvalidName is for a failure moving the static folder
	ErrInvalidName = errors.NewError("name does not match the requirements", 2)

	moduleRepositoryURL   string
	templateRepositoryURL string
)

// newCMD represents the new command
var newCMD = &cobra.Command{
	Use:   "new",
	Short: "Make a new antibuild project.",
	Long:  `Generate a new antibuild project. To get started run "antibuild new" and follow the prompts.`,
	Run: func(cmd *cobra.Command, args []string) {
		moduleRepository, err := cmdInternal.GetModuleRepository(moduleRepositoryURL)
		if err != nil {
			fmt.Println("Failed to download module repository list.")
			return
		}

		templateRepository, err := cmdInternal.GetTemplateRepository(templateRepositoryURL)
		if err != nil {
			fmt.Println("Failed to download template repository list.")
			return
		}

		var modules []string
		var templates []string

		for module := range moduleRepository {
			modules = append(modules, module)
		}

		for template := range templateRepository {
			templates = append(templates, template)
		}

		var newSurvey = []*survey.Question{
			{
				Name:   "name",
				Prompt: &survey.Input{Message: "What should the name of the project be?"},
				Validate: func(input interface{}) error {
					var in string
					var ok bool
					if in, ok = input.(string); !ok {
						return ErrInvalidInput.SetRoot("input is of type " + reflect.TypeOf(input).String() + "not string")
					}

					match := nameregex.MatchString(in)

					if !match {
						return ErrInvalidName.SetRoot("the name should be at least 3 characters and only include a-z and -")
					}
					return nil
				},
			},
			{
				Name: "template",
				Prompt: &survey.Select{
					Message: "Choose a starting template:",
					Options: templates,
				},
			},
			{
				Name: "default_modules",
				Prompt: &survey.MultiSelect{
					Message: "Select any modules you want to pre install now (can also not choose any):",
					Options: modules,
				},
			},
		}

		answers := struct {
			Name           string   `survey:"name"`
			Template       string   `survey:"template"`
			DefaultModules []string `survey:"default_modules"`
		}{}

		err = survey.Ask(newSurvey, &answers)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if _, err := ioutil.ReadDir(answers.Name); os.IsNotExist(err) {
			downloadTemplate(templateRepository, answers.Template, answers.Name)

			if len(answers.DefaultModules) > 0 {
				installModules(moduleRepository, answers.DefaultModules, answers.Name)
			}

			fmt.Println("Success. Run these commands to get started:")
			fmt.Println("")
			fmt.Println("cd " + answers.Name)
			fmt.Println("antibuild develop")
			fmt.Println("")
			fmt.Println("Need help? Look at our docs: https://build.antipy.com/get-started")
			fmt.Println("")
			fmt.Println("")
			return
		}

		fmt.Println("Failed.")
	},
}

func downloadTemplate(templateRepository map[string]cmdInternal.TemplateRepositoryEntry, template string, outPath string) bool {
	if _, ok := templateRepository[template]; !ok {
		fmt.Println("The selected template is not available in this repository.")
		return false
	}

	t := templateRepository[template]

	dir, err := ioutil.TempDir("", "antibuild")
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(dir, 777)
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir)
	
	var src string

	switch t.Source.Type {
	case "zip":
		downloadFilePath := filepath.Join(dir, "download.zip")

		err = internal.DownloadFile(downloadFilePath, t.Source.URL, false)
		if err != nil {
			log.Fatal(err)
		}

		_, err = internal.Unzip(downloadFilePath, filepath.Join(dir, "unzip"))
		if err != nil {
			log.Fatal(err)
		}
		src = filepath.Join(dir, filepath.Join("unzip", t.Source.SubDirectory))
		break
	case "git":
		err = internal.DownloadGit(dir, t.Source.URL)
		if err != nil {
			log.Fatal(err)
		}

		err = os.RemoveAll(filepath.Join(dir, ".git"))
		if err != nil {
			log.Fatal(err)
		}
		src = filepath.Join(dir, t.Source.SubDirectory)
		break
	}

	info, err := os.Lstat(src)
	if err != nil {
		log.Fatal(err)
	}
	internal.DirCopy(src, outPath,info)
	if err != nil {
		log.Fatal(err)
	}


	fmt.Println("Downloaded template.")
	return true
}

func installModules(moduleRepository map[string]cmdInternal.ModuleRepositoryEntry, modules []string, outPath string) {
	cfg, err := config.GetConfig(filepath.Join(outPath, "config.json"))
	if err != nil {
		fmt.Println("Could not open config file to add modules. Module installation will be skipped.")
		return
	}

	for _, module := range modules {
		cfg.Modules.Dependencies[module] = moduleRepositoryURL
	}

	err = config.SaveConfig(filepath.Join(outPath, "config.json"), cfg)
	if err != nil {
		fmt.Println("Could not save config file after adding modules. Modules installation will be skipped.")
		return
	}

	fmt.Println("Please run 'antibuild modules install' to install your selected modules.")
}

//SetCommands sets the commands for this package to the cmd argument
func SetCommands(cmd *cobra.Command) {
	newCMD.Flags().StringVarP(&moduleRepositoryURL, "modules", "m", "https://build.antipy.com/dl/modules.json", "The module repository list file to use. Default is \"https://build.antipy.com/dl/modules.json\"")
	newCMD.Flags().StringVarP(&templateRepositoryURL, "templates", "t", "https://build.antipy.com/dl/templates.json", "The template repository list file to use. Default is \"https://build.antipy.com/dl/templates.json\"")
	cmd.AddCommand(newCMD)
}
