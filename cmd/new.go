// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"os"
	"path/filepath"
	"reflect"
	"regexp"

	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/cmd/templates"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	"gopkg.in/AlecAivazis/survey.v1"
)

type newAnswers struct {
	Name     string `survey:"name"`
	Template string `survey:"template"`
}

const defaultTemplateRepositoryURL = "https://build.antipy.com/dl/templates.json"
const defaultTemplateBranch = "master"

var (
	nameregex = regexp.MustCompile("[a-z-]{3,}")

	//ErrInvalidInput is when the template failed building
	ErrInvalidInput = errors.NewError("invalid input", 1)
	//ErrInvalidName is for a failure moving the static folder
	ErrInvalidName = errors.NewError("name does not match the requirements", 2)
)

func newCommandRun(command *cobra.Command, args []string) {
	templateRepositoryURL := *command.Flags().StringP("templates", "t", defaultTemplateRepositoryURL,
		"The template repository list file to use. Default is \"https://build.antipy.com/dl/templates.json\"")
	templateBranch := *command.Flags().StringP("branch", "b", defaultTemplateBranch,
		"The branch to pull the template from if using git.")

	templateRepository, err := templates.GetRepository(templateRepositoryURL)
	if err != nil {
		println("Failed to download template repository list.")
		return
	}

	templateOptions := make([]string, 0, len(templateRepository))

	for template := range templateRepository {
		templateOptions = append(templateOptions, template)
	}

	newProjectSurvey := []*survey.Question{
		&survey.Question{
			Name:     "name",
			Prompt:   &survey.Input{Message: "What should the name of the project be?"},
			Validate: nameValidator,
		},
		&survey.Question{
			Name: "template",
			Prompt: &survey.Select{
				Message: "Choose a starting template:",
				Options: templateOptions,
			},
		},
	}

	answers := newAnswers{}

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

	templates.Download(templateRepository, answers.Template, answers.Name, templateBranch)

	println("Success. Run these commands to get started:\n")
	println("cd " + answers.Name)
	println("antibuild develop\n")
	println("Need help? Look at our docs: https://build.antipy.com/get-started\n\n")
}

func nameValidator(input interface{}) error {
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
}
