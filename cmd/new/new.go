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
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
	"gitlab.com/antipy/antibuild/cli/internal"
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
			Options: []string{"file", "json", "yaml"},
		},
	},
}

// newCMD represents the new command
var newCMD = &cobra.Command{
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

			err = internal.DownloadFile(downloadFilePath, "https://build.antipy.com/cli/examples/basic.zip", false)
			if err != nil {
				log.Fatal(err)
			}

			_, err = internal.Unzip(downloadFilePath, answers.Name)
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

//SetCommands sets the commands for this package to the cmd argument
func SetCommands(cmd *cobra.Command){
	cmd.AddCommand(newCMD)
}