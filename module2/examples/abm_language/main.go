// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"gitlab.com/antipy/antibuild/cli/builder/site"
	abm "gitlab.com/antipy/antibuild/cli/module/client"
)

var languages []string
var defaultLanguage string

func main() {
	module := abm.Register("language")

	module.ConfigFunctionRegister(func(input map[string]interface{}) error {
		var ok bool
		var languagesInterface []interface{}

		if languagesInterface, ok = input["languages"].([]interface{}); !ok {
			return abm.ErrInvalidInput
		}

		for _, languageInterface := range languagesInterface {
			if language, ok := languageInterface.(string); ok {
				languages = append(languages, language)
			} else {
				return abm.ErrInvalidInput
			}
		}

		if languages == nil {
			return abm.ErrInvalidInput
		}

		if defaultLanguage, ok = input["default"].(string); !ok {
			return abm.ErrInvalidInput
		}

		if defaultLanguage == "" {
			return nil
		}

		for _, language := range languages {
			if language == defaultLanguage {
				return nil
			}
		}

		return abm.ErrInvalidInput
	})
	module.SitePostProcessor("language", languageProcess)

	module.Start()
}

func languageProcess(w abm.SPPRequest, r *abm.SPPResponse) {
	var siteData = w.Data

	if languages == nil {
		r.Error = abm.ErrNoConfig
		return
	}

	for _, page := range siteData {
		for _, language := range languages {
			slugLanguage := language
			if language == defaultLanguage {
				slugLanguage = ""
			}

			var newData = make(map[string]interface{})
			for i, v := range page.Data {
				newData[i] = v
			}

			for _, lang := range languages {
				if lang != language {
					delete(newData, lang)
				}
			}

			var ok bool
			var correctLanguageData = make(map[interface{}]interface{})
			if newData[language] != nil {
				fmt.Fprint(os.Stderr, reflect.TypeOf(newData[language]), "\n")
				if correctLanguageData, ok = newData[language].(map[interface{}]interface{}); !ok {
					r.Error = abm.ErrInvalidInput
					return
				}
			}

			for i, v := range correctLanguageData {
				if k, ok := i.(string); ok {
					newData[k] = v
				} else {
					r.Error = abm.ErrInvalidInput
					return
				}

			}

			r.Data = append(r.Data, &site.Site{
				Slug:     filepath.Join(slugLanguage, page.Slug),
				Template: page.Template,
				Data:     newData,
			})
		}
	}
}
