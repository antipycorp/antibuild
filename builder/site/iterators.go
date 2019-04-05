// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site

import (
	"bytes"
	"strings"

	"gitlab.com/antipy/antibuild/cli/ui"

	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

type (
	// IteratorData is info about an iterator from the config
	IteratorData struct {
		Iterator          string
		IteratorArguments string
		List              []string
	}
)

var (
	//ErrNoIteratorFound means an iterator does not have a function
	ErrNoIteratorFound = errors.NewError("could not get iterator information", 11)
)

// MarshalJSON marshalls the iterator data
func (i *IteratorData) MarshalJSON() ([]byte, error) {
	return []byte("\"" + i.String() + "\""), nil
}

func (i *IteratorData) String() string {
	out := ""

	out += "[" + i.Iterator
	if i.IteratorArguments != "" {
		out += ":" + i.IteratorArguments
	}
	out += "]"

	return out
}

// UnmarshalJSON unmarshalls the iterator data
func (i *IteratorData) UnmarshalJSON(data []byte) error {
	//get the data from for the dataLoader
	i1 := bytes.Index(data, []byte("["))
	i2 := bytes.Index(data, []byte("]"))

	iteratorData := data[i1+1 : i2]

	{
		//get all the arguments for the loader
		sep := bytes.Split(iteratorData, []byte(":"))
		if len(sep) == 0 {
			return ErrNoIteratorFound.SetRoot(string(iteratorData))
		}
		var loader = make([]byte, len(sep[0]))
		copy(loader, sep[0])

		if len(bytes.Split(sep[0], []byte("_"))) == 1 {
			loader = append(sep[0], append([]byte("_"), sep[0]...)...)
		}

		i.Iterator = string(loader)
		i.IteratorArguments = ""

		if len(sep) >= 2 { //only if bigger than 2 this is available
			i.IteratorArguments = string(sep[1])
		}
	}

	return nil
}

func includedVars(d string) []string {
	data := []byte(d)
	var vars []string

	for strings.Count(string(data), "{{") > 0 && strings.Count(string(data), "}}") > 0 {
		i1 := bytes.Index(data, []byte("{{"))
		i2 := bytes.Index(data, []byte("}}"))

		chunk := data[i1+2 : i2]
		data = data[i2+2:]

		vars = append(vars, string(chunk))
	}

	return unique(vars)
}

func unique(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func replaceVar(data string, variable string, value string) string {
	return strings.ReplaceAll(data, "{{"+variable+"}}", value)
}

func replaceVarData(d Data, variable string, value string) Data {
	d.Loader = replaceVar(d.Loader, variable, value)
	d.LoaderArguments = replaceVar(d.LoaderArguments, variable, value)
	d.Parser = replaceVar(d.Parser, variable, value)
	d.ParserArguments = replaceVar(d.ParserArguments, variable, value)
	for pp := range d.PostProcessors {
		dpp := d.PostProcessors[pp]
		dpp.PostProcessor = replaceVar(dpp.PostProcessor, variable, value)
		dpp.PostProcessorArguments = replaceVar(dpp.PostProcessorArguments, variable, value)
		d.PostProcessors[pp] = dpp
	}
	return d
}

func replaceVarIterator(i IteratorData, variable string, value string) IteratorData {
	i.Iterator = replaceVar(i.Iterator, variable, value)
	i.IteratorArguments = replaceVar(i.IteratorArguments, variable, value)

	return i
}

func getReplacers(vars []string, cSite *ConfigSite) ([]map[string]string, errors.Error) {
	var replacers = []map[string]string{
		make(map[string]string),
	}

	for _, variable := range vars {
		var varReplacers []map[string]string

		err := gatherIterators(cSite.Iterators)
		if err != nil {
			return nil, ErrFailedGather.SetRoot(err.Error())
		}

		var i IteratorData
		var ok bool
		if i, ok = cSite.Iterators[variable]; !ok {
			return nil, ErrFailedGather.SetRoot("no iterator defined for " + variable)
		}

		for _, val := range i.List {
			for _, replacer := range replacers {
				r := make(map[string]string)

				for p, q := range replacer {
					r[p] = q
				}

				r[variable] = val
				varReplacers = append(varReplacers, r)
			}
		}

		replacers = varReplacers
	}

	return replacers, nil
}

func doIterators(cSite *ConfigSite, sites *map[string]*ConfigSite, log *ui.UI) errors.Error {
	if len(cSite.IteratorValues) > 0 {
		log.Debugf("Applying existing iterator variables: %v", cSite.IteratorValues)

		for variable, value := range cSite.IteratorValues {
			cSite.Slug = replaceVar(cSite.Slug, variable, value)

			for x := range cSite.Templates {
				cSite.Templates[x] = replaceVar(cSite.Templates[x], variable, value)
			}

			for x := range cSite.Data {
				if cSite.Data[x].ShouldRange != variable {
					cSite.Data[x] = replaceVarData(cSite.Data[x], variable, value)
				}
			}

			for x := range cSite.Iterators {
				cSite.Iterators[x] = replaceVarIterator(cSite.Iterators[x], variable, value)
			}
		}
	}

	var newSites []ConfigSite
	var slugSitesChanged = true

	slugVars := includedVars(cSite.Slug)

	if len(slugVars) > 0 {
		log.Debugf("Generating sub-sites with vars %v", slugVars)
		replacers, err := getReplacers(slugVars, cSite)
		if err != nil {
			return err
		}

		if len(replacers) > 0 {
			for _, replacer := range replacers {
				newSlug := cSite.Slug
				newTemplates := append([]string(nil), cSite.Templates...)
				newData := append([]Data(nil), cSite.Data...)
				newIterators := make(map[string]IteratorData, len(cSite.Iterators))
				for p, q := range cSite.Iterators {
					newIterators[p] = q
				}

				for variable, value := range replacer {
					newSlug = replaceVar(newSlug, variable, value)

					for x := range newTemplates {
						newTemplates[x] = replaceVar(newTemplates[x], variable, value)
					}

					for x := range newData {
						d := newData[x]

						if d.ShouldRange != variable {
							d = replaceVarData(d, variable, value)
							newData[x] = d
						}
					}

					for x := range newIterators {
						newIterators[x] = replaceVarIterator(newIterators[x], variable, value)
					}
				}

				newSite := *cSite

				newSite.Slug = newSlug
				newSite.Templates = newTemplates
				newSite.Data = newData

				newSite.Iterators = cSite.Iterators
				newSite.IteratorValues = make(map[string]string, len(cSite.IteratorValues)+len(replacer))
				for variable, value := range cSite.IteratorValues {
					newSite.IteratorValues[variable] = value
				}
				for variable, value := range replacer {
					newSite.IteratorValues[variable] = value
				}

				newSite.Sites = cSite.Sites

				newSites = append(newSites, newSite)
			}
		}
	} else {
		slugSitesChanged = false
		newSites = append(newSites, *cSite)
	}

	for _, currentSite := range newSites {
		var additionalData []Data

		for x, d := range currentSite.Data {
			if d.ShouldRange != "" {
				log.Debugf("Doing data iterator ranging for variable %s", d.ShouldRange)

				variable := d.ShouldRange

				d.ShouldRange = ""

				err := gatherIterators(currentSite.Iterators)
				if err != nil {
					return ErrFailedGather.SetRoot(err.Error())
				}

				var i IteratorData
				var ok bool
				if i, ok = currentSite.Iterators[variable]; !ok {
					return ErrFailedGather.SetRoot("no iterator defined for " + variable)
				}

				for _, v := range i.List {
					newD := replaceVarData(d, variable, v)
					additionalData = append(additionalData, newD)
				}

				currentSite.Data = remove(currentSite.Data, x)
			}
		}

		currentSite.Data = append(currentSite.Data, additionalData...)

		if slugSitesChanged {
			for _, childSite := range currentSite.Sites {
				err := unfold(childSite, &currentSite, sites, log)
				if err != nil {
					return err
				}
			}
		} else {
			cSite.Data = currentSite.Data
		}
	}

	return nil
}

func remove(s []Data, i int) []Data {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
