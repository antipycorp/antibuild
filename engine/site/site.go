// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site

import (
	"html/template"
	"text/template/parse"

	apiSite "gitlab.com/antipy/antibuild/api/site"
	"gitlab.com/antipy/antibuild/cli/engine/modules"

	"github.com/jaicewizard/tt"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	ui "gitlab.com/antipy/antibuild/cli/internal/log"
)

type (
	//Site is the way a site is defined after all of its data and templates have been collected
	Site struct {
		Slug     string
		Template string
		Data     tt.Data
	}

	//ConfigSite is the way a site is defined in the config file
	ConfigSite struct {
		Iterators      map[string]IteratorData `json:"iterators,omitempty"`
		Slug           string                  `json:"slug,omitempty"`
		Templates      []string                `json:"templates,omitempty"`
		Data           []Data                  `json:"data,omitempty"`
		Sites          []ConfigSite            `json:"sites,omitempty"`
		IteratorValues map[string]string       `json:"-"`
		Dependencies   []string                `json:"-"`
	}
)

var (
	//TemplateFunctions are all the template functions defined by modules
	TemplateFunctions = modules.TemplateFunctions

	//DataLoaders are all the module data loaders
	DataLoaders = modules.DataLoaders
	//DataParsers are all the module file parsers
	DataParsers = modules.DataParsers
	//DataPostProcessors are all the module data post processors
	DataPostProcessors = modules.DataPostProcessors
	//SPPs are all the module site post processors
	SPPs = modules.SPPs

	//Iterators are all the module iterators
	Iterators = modules.Iterators

	//TemplateFolder is the folder all templates are stored
	TemplateFolder string
	//StaticFolder is the folder all static files are stored
	StaticFolder string
	//OutputFolder is the folder that should be exported to
	OutputFolder string

	globalTemplates = make(map[string]*template.Template)
	subTemplates    = make(map[string]string)

	//ErrFailedTemplate is when the template failed building
	ErrFailedTemplate = errors.NewError("failed building template", 1)
	//ErrFailedStatic is for a failure moving the static folder
	ErrFailedStatic = errors.NewError("failed to move static folder", 2)
	//ErrFailedGather is for a failure in gathering files.
	ErrFailedGather = errors.NewError("failed to gather files", 3)
	//ErrFailedCreateFS is for a failure in gathering files.
	ErrFailedCreateFS = errors.NewError("couldn't create directory/file", 4)
	//ErrUsingUnknownModule is when a user uses a module that is not registered.
	ErrUsingUnknownModule = errors.NewError("module is used but not registered", 5)
)

/*
	UNFOLD

	Parse the tree into the individual sites by iterating over the children of parents
	and combining their data until only sites with no more the children remain. Add
	these sites to an array, so there is no more nesting.
	Parse the sites tree from the config file, any final site (no child sites) will
	be parsed into the final list of sites.
*/

//Unfold the ConfigSite into a []ConfigSite
func Unfold(configSite *ConfigSite, log *ui.UI) ([]ConfigSite, errors.Error) {
	var sites []ConfigSite
	globalTemplates = make(map[string]*template.Template, len(sites))

	err := unfold(configSite, nil, &sites, log)
	if err != nil {
		return sites, err
	}

	return sites, nil
}

func unfold(cSite *ConfigSite, parent *ConfigSite, sites *[]ConfigSite, log *ui.UI) (err errors.Error) {
	if parent != nil {
		mergeConfigSite(cSite, parent)
	}
	log.Debugf("Unfolding " + cSite.Slug)

	numIncludedVars := numIncludedVars(*cSite)

	//If this is the last in the chain, add it to the list of return values
	if len(cSite.Sites) == 0 && numIncludedVars == 0 {
		for _, v := range cSite.Data {
			dep := v.Loader +
				v.LoaderArguments +
				v.Parser +
				v.ParserArguments
			for _, pp := range v.PostProcessors {
				dep += pp.PostProcessor + pp.PostProcessorArguments
			}
			cSite.Dependencies = append(cSite.Dependencies, dep)
		}

		//append site to the list of sites that will be executed
		(*sites) = append(*sites, *cSite)
		log.Debug("Unfolded to final site")

		return nil
	}

	if numIncludedVars > 0 {
		itSited, err := doIterators(*cSite, log)
		if err != nil {
			log.Fatalf("failled to do iterators: %v", err)
			return err
		}
		for i := range itSited {
			err := unfold(&itSited[i], nil, sites, log)
			if err != nil {
				return err
			}
		}
		return nil
	}
	// we might not have iterators values that we need right now
	// but if we can parse them now then we dont need duplicate work
	gatherIterators(cSite.Iterators)

	for _, childSite := range cSite.Sites {
		err = unfold(&childSite, cSite, sites, log)
		if err != nil {
			return err
		}
	}

	return
}

//mergeConfigSite merges the src into the dst
func mergeConfigSite(dst *ConfigSite, src *ConfigSite) {
	if dst.Data != nil {
		dst.Data = append(dst.Data, src.Data...) // just append
	} else {
		dst.Data = make([]Data, len(src.Data)) // or make a new one and fill it
		copy(dst.Data, src.Data)
	}

	if dst.Templates != nil {
		dst.Templates = append(dst.Templates, src.Templates...) // just append
	} else {
		dst.Templates = make([]string, len(src.Templates)) // or make a new one and fill it
		copy(dst.Templates, src.Templates)
	}

	if dst.Dependencies != nil {
		dst.Dependencies = append(dst.Dependencies, src.Dependencies...) // just append
	} else {
		dst.Dependencies = make([]string, len(src.Dependencies)) // or make a new one and fill it
		copy(dst.Dependencies, src.Dependencies)
	}

	if dst.IteratorValues == nil {
		dst.IteratorValues = make(map[string]string, len(src.IteratorValues)) // if none exist, make a new one to fill
	}

	for i, s := range src.IteratorValues {
		dst.IteratorValues[i] = s
	}

	if dst.Iterators == nil {
		dst.Iterators = make(map[string]IteratorData, len(src.Iterators)) // if none exist, make a new one to fill
	}

	for i, s := range src.Iterators {
		dst.Iterators[i] = s
	}

	dst.Slug = src.Slug + dst.Slug
}

// RemoveTemplate cleans a template from the template cache
func RemoveTemplate(path string) {
	delete(subTemplates, path)
}

// PostProcess all sites
func PostProcess(sites *[]*Site, spps []string, log *ui.UI) errors.Error {
	send := make([]apiSite.Site, 0, len(*sites))
	var recieve []apiSite.Site

	for _, d := range *sites {
		send = append(send, apiSite.Site{
			Slug:     d.Slug,
			Template: d.Template,
			Data:     d.Data,
		})
	}

	line := make([]modules.Pipe, 0, len(spps))
	for _, spp := range spps {
		if k, ok := SPPs[spp]; ok {
			line = append(line, k.GetPipe(""))
		}
	}

	bytes, err := modules.ExecPipeline(apiSite.Encode(send), line...)
	if err != nil {
		return err
	}

	recieve = apiSite.Decode(bytes)
	*sites = make([]*Site, 0, len(recieve))

	for _, d := range recieve {
		*sites = append(*sites, &Site{
			Slug:     d.Slug,
			Template: d.Template,
			Data:     d.Data,
		})
	}

	return nil
}

//GetTemplateTree gets the template according to
func GetTemplateTree(template string) *parse.Tree {
	if v, ok := globalTemplates[template]; ok {
		return v.Tree
	}
	return nil
}
