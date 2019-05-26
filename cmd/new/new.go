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
	"gitlab.com/antipy/antibuild/cli/cli/templates"
	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/internal/download"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	"gitlab.com/antipy/antibuild/cli/internal/zip"
	"gopkg.in/AlecAivazis/survey.v1"
)

const defaultTemplateRepositoryURL = "https://build.antipy.com/dl/templates.json"
const defaultTemplateBranch = "master"

var (
	nameregex = regexp.MustCompile("[a-z-]{3,}")

	//ErrInvalidInput is when the template failed building
	ErrInvalidInput = errors.NewError("invalid input", 1)
	//ErrInvalidName is for a failure moving the static folder
	ErrInvalidName = errors.NewError("name does not match the requirements", 2)

	newProjecQuestion = survey.Question{
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
	}
)

// newCMD represents the new command
var newCMD = &cobra.Command{
	Use:   "new",
	Short: "Make a new antibuild project.",
	Long:  `Generate a new antibuild project. To get started run "antibuild new" and follow the prompts.`,
	Run: func(cmd *cobra.Command, args []string) {
		templateRepositoryURL := *cmd.Flags().StringP("templates", "t", defaultTemplateRepositoryURL,
			"The template repository list file to use. Default is \"https://build.antipy.com/dl/templates.json\"")
		templateBranch := *cmd.Flags().StringP("branch", "b", defaultTemplateBranch,
			"The branch to pull the template from if using git.")

		templateRepository, err := templates.GetTemplateRepository(templateRepositoryURL)
		if err != nil {
			println("Failed to download template repository list.")
			return
		}

		templates := make([]string, 0, len(templateRepository))

		for template := range templateRepository {
			templates = append(templates, template)
		}

		newProjectSurvey := []*survey.Question{
			&newProjecQuestion,
			&survey.Question{
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

		err = survey.Ask(newProjectSurvey, &answers)
		if err != nil {
			println(err.Error())
			return
		}
		path, _ := filepath.Abs(answers.Name)

		if _, err := os.Stat(path); !os.IsNotExist(err) {
			println("Failed.")
			return
		}

		downloadTemplate(templateRepository, answers.Template, answers.Name, templateBranch)

		println("Success. Run these commands to get started:\n")
		println("cd " + answers.Name)
		println("antibuild develop\n")
		println("Need help? Look at our docs: https://build.antipy.com/get-started\n\n")
	},
}

func downloadTemplate(templateRepository map[string]templates.TemplateEntry, template, outPath, templateBranch string) bool {
	if _, ok := templateRepository[template]; !ok {
		println("The selected template is not available in this repository.")
		return false
	}

	t := templateRepository[template]

	dir, err := ioutil.TempDir("", "antibuild")
	if err != nil {
		log.Fatal(err)
	}

	err = os.MkdirAll(dir, 0755)
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

//SetCommands sets the commands for this package to the cmd argument
func SetCommands(cmd *cobra.Command) {
	cmd.AddCommand(newCMD)
}
