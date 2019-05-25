// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site

import (
	"html/template"
	"io/ioutil"
	"path/filepath"

	"github.com/jaicewizard/tt"
	"gitlab.com/antipy/antibuild/cli/engine/modules"
	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	ui "gitlab.com/antipy/antibuild/cli/internal/log"
)

// Gather after unfolding
func Gather(cSite ConfigSite, log *ui.UI) (*Site, errors.Error) {
	log.Debugf("Gathering information for %s", cSite.Slug)
	log.Debugf("Site data: %v", cSite)

	site := &Site{
		Slug: cSite.Slug,
	}

	err := gatherData(site, cSite.Data)
	if err != nil {
		return nil, ErrFailedGather.SetRoot(err.Error())
	}

	err = gatherTemplates(site, cSite.Templates)
	if err != nil {
		return nil, ErrFailedGather.SetRoot(err.Error())
	}

	log.Debugf("Finished gathering")

	return site, nil
}

//collect data objects from modules
func gatherData(site *Site, files []Data) errors.Error {
	for _, d := range files {

		//init data if it is empty
		if site.Data == nil {
			site.Data = make(tt.Data)
		}

		var data tt.Data

		if _, ok := DataLoaders[d.Loader]; !ok {
			return ErrUsingUnknownModule.SetRoot(d.Loader)
		}

		if _, ok := DataParsers[d.Parser]; !ok {
			return ErrUsingUnknownModule.SetRoot(d.Parser)
		}

		fPipe := DataLoaders[d.Loader].GetPipe(d.LoaderArguments)
		pPipe := DataParsers[d.Parser].GetPipe(d.ParserArguments)
		var ppPipes []modules.Pipe
		var validPPPipes = 0
		for _, dpp := range d.PostProcessors {
			if _, ok := DataPostProcessors[dpp.PostProcessor]; !ok {
				return ErrUsingUnknownModule.SetRoot(dpp.PostProcessor)
			}

			ppPipes = append(ppPipes, DataPostProcessors[dpp.PostProcessor].GetPipe(dpp.PostProcessorArguments))
			if ppPipes != nil {
				validPPPipes++
			}
		}

		var pipes = []modules.Pipe{
			fPipe,
			pPipe,
		}

		pipes = append(pipes, ppPipes...)

		bytes, _ := modules.ExecPipeline(nil, pipes...)
		data.GobDecode(bytes)

		//add the parsed data to the site data
		for k, v := range data {
			site.Data[k] = v
		}
	}

	return nil
}

func gatherTemplates(site *Site, templates []string) errors.Error {
	finalTemplate := template.New("").Funcs(TemplateFunctions)
	for _, tPath := range templates {
		//prefix the templates with the TemplateFolder
		path := filepath.Join(TemplateFolder, tPath)
		if _, ok := subTemplates[path]; !ok {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return errors.Import(err)
			}
			subTemplates[path] = string(bytes)
		}
		finalTemplate.Parse(subTemplates[path])
	}
	//TODO: make the ID be the template path
	//get the template with the TemplateFunctions initalized
	id := internal.RandString(32)

	globalTemplates[id] = finalTemplate
	site.Template = id

	return nil
}

//collect iterators from modules skiping over any already set values for obvious reasons
func gatherIterators(iterators map[string]IteratorData) errors.Error {
	for n, i := range iterators {
		if len(i.List) != 0 {
			continue
		}

		if _, ok := Iterators[i.Iterator]; !ok {
			return ErrUsingUnknownModule.SetRoot(i.Iterator)
		}

		i.List = Iterators[i.Iterator].GetIterations(i.IteratorArguments)

		iterators[n] = i
	}

	return nil
}
