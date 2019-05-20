// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package new

import (
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
)

// newCMD represents the new command
var newCMD = &cobra.Command{
	Use:   "new",
	Short: "Make a new antibuild project.",
	Long:  `Generate a new antibuild project. To get started run "antibuild new" and follow the prompts.`,
	Run: func(cmd *cobra.Command, args []string) {
		templateRepositoryURL := *cmd.Flags().StringP("templates", "t", "https://build.antipy.com/dl/templates.json", "The template repository list file to use. Default is \"https://build.antipy.com/dl/templates.json\"")
		templateBranch := *cmd.Flags().StringP("branch", "b", "master", "The branch to pull the template from if using git.")

		templateRepository, err := cmdInternal.GetTemplateRepository(templateRepositoryURL)
		if err != nil {
			println("Failed to download template repository list.")
			return
		}

		var templates []string

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
		}

		answers := struct {
			Name     string `survey:"name"`
			Template string `survey:"template"`
		}{}

		err = survey.Ask(newSurvey, &answers)
		if err != nil {
			println(err.Error())
			return
		}

		if _, err := ioutil.ReadDir(answers.Name); os.IsNotExist(err) {
			downloadTemplate(templateRepository, answers.Template, answers.Name, templateBranch)

			println("Success. Run these commands to get started:\n")
			println("cd " + answers.Name)
			println("antibuild develop\n")
			println("Need help? Look at our docs: https://build.antipy.com/get-started\n\n")
			return
		}

		println("Failed.")
	},
}

func downloadTemplate(templateRepository map[string]cmdInternal.TemplateRepositoryEntry, template string, outPath string, templateBranch string) bool {
	if _, ok := templateRepository[template]; !ok {
		println("The selected template is not available in this repository.")
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
		err = internal.DownloadGit(dir, t.Source.URL, templateBranch)
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

	internal.DirCopy(src, outPath, info)
	if err != nil {
		log.Fatal(err)
	}

	println("Downloaded template.")
	return true
}

func installModules(ms [][3]string, outPath string) {
	cfg, err := config.GetConfig(filepath.Join(outPath, "config.json"))
	if err != nil {
		println("Could not open config file to add modules. Module installation will be skipped.")
		return
	}

	for _, module := range ms {
		cfg.Modules.Dependencies[module[0]] = &config.Module{
			Version:    module[1],
			Repository: module[2],
		}
	}

	err = config.SaveConfig(filepath.Join(outPath, "config.json"), cfg)
	if err != nil {
		println("Could not save config file after adding modules. Modules installation will be skipped.")
		return
	}

	println("Please run 'antibuild modules install' to install your selected modules.")
}

//SetCommands sets the commands for this package to the cmd argument
func SetCommands(cmd *cobra.Command) {
	cmd.AddCommand(newCMD)
}
