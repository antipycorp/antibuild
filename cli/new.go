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

package cli

import (
	"errors"
	"fmt"
	"regexp"

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
			Options: []string{"none", "homepage", "newspage"},
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
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
