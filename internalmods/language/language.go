// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package language

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"

	abm "gitlab.com/antipy/antibuild/api/client"
	"gitlab.com/antipy/antibuild/cli/builder/site"
)

var languages []string
var defaultLanguage string

//Start starts the file module
func Start(in io.Reader, out io.Writer) {
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

	module.CustomStart(in, out)
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
			var ok bool
			var langData map[string]interface{}
			var langDataINTF map[interface{}]interface{}

			for i, v := range page.Data {
				if i == language { //if this the language we asked for

					fmt.Fprint(os.Stderr, reflect.TypeOf(v), " is the type\n")
					if langData, ok = v.(map[string]interface{}); !ok {
						/* This is really shitty, but some serializers support non-strings as keys,
						and after the first layer it will use whatever it finds cool.
						hence we also chec if it can be converted to a string*/
						if langDataINTF, ok = v.(map[interface{}]interface{}); !ok {
							fmt.Fprint(os.Stderr, v, "\n")
							fmt.Fprint(os.Stderr, langDataINTF, "\n")
							r.Error = abm.ErrInvalidInput
							return
						}
						// Finaly converting the key to a string, I really hate this
						for i, v := range langDataINTF {
							if k, ok := i.(string); ok {
								newData[k] = v
							} else {
								r.Error = abm.ErrInvalidInput
								return
							}

						}
						continue

					}

					for datk, v := range langData {
						newData[datk] = v
					}
					continue
				}
				newData[i] = v
			}

			r.Data = append(r.Data, &site.Site{
				Slug:     filepath.Join(slugLanguage, page.Slug),
				Template: page.Template,
				Data:     newData,
			})
		}
	}
}
