// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site

import (
	"bytes"
	"gitlab.com/antipy/antibuild/cli/ui"
	"strings"

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

//0xc00020a808
var (
	//ErrNoIteratorFound means an iterator does not have a function
	ErrNoIteratorFound = errors.NewError("could not get iterator information", 11)
)

// MarshalJSON marshalls the iterator data
func (i IteratorData) MarshalJSON() ([]byte, error) {
	return []byte("\"" + i.String() + "\""), nil
}

func (i IteratorData) String() string {
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
	i1 := bytes.IndexByte(data, '[')
	i2 := bytes.IndexByte(data, ']')

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

func includedVars(data []byte) []string {
	var vars []string

	left, right := bytes.Count(data, []byte("{{")), bytes.Count(data, []byte("}}"))
	if left != right {
		//return error
	}
	vars = make([]string, 0, left)
	for ; left > 0; left-- {
		i1 := bytes.Index(data, []byte("{{"))
		i2 := bytes.Index(data, []byte("}}"))

		chunk := data[i1+2 : i2]
		data = data[i2+2:]

		vars = append(vars, string(chunk))
	}

	return vars
}

//fastNumIncluded vars just simply counts the number of "{{" occurances and doesnt check for wrong formatting
func fastNumIncludedVars(data []byte) int {
	return bytes.Count(data, []byte("{{"))
}

func unique(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := make([]string, 0, len(stringSlice))

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
	d.LoaderArguments = replaceVar(d.LoaderArguments, variable, value)
	d.ParserArguments = replaceVar(d.ParserArguments, variable, value)

	for pp := range d.PostProcessors {
		dpp := d.PostProcessors[pp]
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

func numIncludedVars(cSite ConfigSite) int {
	usedVars := fastNumIncludedVars([]byte(cSite.Slug))
	for _, v := range cSite.Data {
		usedVars += fastNumIncludedVars([]byte(v.LoaderArguments + v.ParserArguments))
	}

	for _, v := range cSite.Iterators {
		usedVars += fastNumIncludedVars([]byte(v.IteratorArguments))
	}

	return usedVars
}

func totalIncludedVars(cSite *ConfigSite) []string {
	usedVars := includedVars([]byte(cSite.Slug))
	for _, v := range cSite.Data {
		usedVars = append(usedVars, includedVars([]byte(v.LoaderArguments+v.ParserArguments))...)
	}

	for _, v := range cSite.Iterators {
		usedVars = append(usedVars, includedVars([]byte(v.IteratorArguments))...)
	}

	usedVars = unique(usedVars)
	return usedVars
}

func doIteratorVariables(cSite *ConfigSite) errors.Error {
	for i, v := range cSite.Iterators {
		vars := includedVars([]byte(v.IteratorArguments))
		for _, ivar := range vars {
			var value string
			var ok bool
			if value, ok = cSite.IteratorValues[ivar]; !ok {
				return ErrNoIteratorFound.SetRoot("no variable defined for " + ivar)
			}

			v.IteratorArguments = replaceVar(v.IteratorArguments, ivar, value)
		}
		cSite.Iterators[i] = v
	}
	return nil
}

//DeepCopy deep copies the site.
func DeepCopy(cSite ConfigSite) ConfigSite {
	nsd := make([]Data, len(cSite.Data))
	for k, v := range cSite.Data {
		npp := make([]DataPostProcessor, len(v.PostProcessors))
		copy(npp, v.PostProcessors)
		v.PostProcessors = npp
		nsd[k] = v
	}
	cSite.Data = nsd

	niv := make(map[string]string, len(cSite.IteratorValues))
	for k, v := range cSite.IteratorValues {
		niv[k] = v
	}
	cSite.IteratorValues = niv

	nid := make(map[string]IteratorData, len(cSite.Iterators))
	for k, v := range cSite.Iterators {
		nid[k] = IteratorData{
			Iterator:          v.Iterator,
			IteratorArguments: v.IteratorArguments,
		}
	}
	cSite.Iterators = nid

	ns := make([]ConfigSite, len(cSite.Sites))
	for i, v := range cSite.Sites {
		ns[i] = DeepCopy(v)
	}
	cSite.Sites = ns

	ndp := make([]string, len(cSite.Dependencies))
	copy(ndp, cSite.Dependencies)
	cSite.Dependencies = ndp

	return cSite
}

func doIterators2(cSite ConfigSite, log *ui.UI) ([]ConfigSite, errors.Error) {
	err := doIteratorVariables(&cSite)
	if err != nil {
		return nil, err
	}
	gatherIterators(cSite.Iterators)

	var usedIterators []string

	var newData []Data
	for _, d := range cSite.Data {
		if d.ShouldRange != "" {
			rangeVar := d.ShouldRange
			var i IteratorData
			var ok bool
			if i, ok = cSite.Iterators[rangeVar]; !ok {
				return nil, ErrNoIteratorFound.SetRoot("no iterator defined for " + rangeVar)
			}

			d.ShouldRange = ""

			for _, v := range i.List {
				nd := replaceVarData(d, rangeVar, v)
				newData = append(newData, nd)
			}
			usedIterators = append(usedIterators, rangeVar)
		} else {
			newData = append(newData, d)
		}
	}

	cSite.Data = newData

	usedVars := includedVars([]byte(cSite.Slug))

	usedIterators = unique(append(usedIterators, usedVars...)) //these are the variables that are used inside the site and should be fine

	for _, d := range cSite.Data {
		usedVars = append(usedVars, includedVars([]byte(d.LoaderArguments+d.ParserArguments))...)
		for _, v := range d.PostProcessors {
			usedVars = append(usedVars, includedVars([]byte(v.PostProcessorArguments))...)
		}
	}

	usedVars = unique(usedVars)

	options := make([][]string, len(usedVars))

	for i, iOpts := range usedVars {
		options[i] = cSite.Iterators[iOpts].List
		if len(options[i]) == 0 {
			var value string
			var ok bool
			if value, ok = cSite.IteratorValues[iOpts]; !ok {
				return nil, ErrNoIteratorFound.SetRoot("no variable defined for " + iOpts)
			}

			options[i] = []string{value}
		}

		if len(options[i]) == 0 {
			return nil, ErrNoIteratorFound.SetRoot("no iterator defined for " + iOpts)
		}
	}

	for _, v := range usedIterators {
		cSite.Dependencies = append(cSite.Dependencies, cSite.Iterators[v].Iterator+cSite.Iterators[v].IteratorArguments)
		delete(cSite.Iterators, v)
	}

	if len(cSite.Iterators) != 0 {
		for k, v := range cSite.Iterators {
			log.Error("the iterator for variable " + k + ": " + v.IteratorArguments + " is never used inside a slug")
		}
	}

	olen := 1
	for _, iOpts := range options {
		olen *= len(iOpts)
	}

	var sites = make([]ConfigSite, olen)
	sites[0] = cSite
	lastUpperBound := 0
	currentLowerBound := 0
	for vi, iOpts := range options {
		shouldDownGrade := false
		lastUpperBound += len(iOpts) - 1
		if lastUpperBound == 0 {
			lastUpperBound = 1
			shouldDownGrade = true
		}
		variable := usedVars[vi]

		for i := range iOpts {
			/* if i == 0 {
				continue
			} */
			base := (lastUpperBound) * i
			for i2 := 0; i2 < lastUpperBound; i2++ {
				sites[base+i2] = DeepCopy(sites[i2])
			}
		}
		for _, value := range iOpts {
			for i := currentLowerBound; i < currentLowerBound+lastUpperBound; i++ {
				sites[i].Slug = replaceVar(sites[i].Slug, variable, value)
				for di, d := range sites[i].Data {
					sites[i].Data[di] = replaceVarData(d, variable, value)
				}
				if sites[i].IteratorValues == nil {
					sites[i].IteratorValues = make(map[string]string)
				}
				sites[i].IteratorValues[variable] = value
			}
			currentLowerBound += (lastUpperBound)
		}
		if shouldDownGrade {
			lastUpperBound = 0
			currentLowerBound = 0
		}

	}
	return sites, nil
}
