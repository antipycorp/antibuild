// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site

import (
	"html/template"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/antipy/antibuild/cli/modules/pipeline"

	"github.com/jaicewizard/tt"
	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
)

type (
	//ConfigSite is the way a site is defined in the config file
	ConfigSite struct {
		Iterators map[string]iterator `json:"iterators,omitempty"`
		Slug      string              `json:"slug,omitempty"`
		Templates []string            `json:"templates,omitempty"`
		Data      []data              `json:"data,omitempty"`
		Sites     []*ConfigSite       `json:"sites,omitempty"`
	}

	//Site is the way a site is defined after all of its data and templates have been collected
	Site struct {
		Slug     string
		Template string
		Data     tt.Data
	}

	//DataLoader is a module that loads data
	DataLoader interface {
		Load(string) []byte
		GetPipe(string) pipeline.Pipe
	}

	//DataParser is a module that parses loaded data
	DataParser interface {
		Parse([]byte, string) tt.Data
		GetPipe(string) pipeline.Pipe
	}

	//DPP is a function thats able to post-process data
	DPP interface {
		Process(tt.Data, string) tt.Data
		GetPipe(string) pipeline.Pipe
	}

	//SPP is a function thats able to post-process data
	SPP interface {
		Process([]*Site, string) []*Site
		GetPipe(string) pipeline.Pipe
	}
)

var (
	//TemplateFunctions are all the template functions defined by modules
	TemplateFunctions = template.FuncMap{}

	//DataLoaders are all the module data loaders
	DataLoaders = make(map[string]DataLoader)
	//DataParsers are all the module file parsers
	DataParsers = make(map[string]DataParser)
	//DataPostProcessors are all the module data post processors
	DataPostProcessors = make(map[string]DPP)
	//SPPs are all the module data post processors
	SPPs = make(map[string]SPP)

	//TemplateFolder is the folder all templates are stored
	TemplateFolder string
	//StaticFolder is the folder all static files are stored
	StaticFolder string
	//OutputFolder is the folder that should be exported to
	OutputFolder string

	globalTemplates = make(map[string]*template.Template)

	//ErrFailedTemplate is when the template failed building
	ErrFailedTemplate = errors.NewError("failed building template", 1)
	//ErrFailedStatic is for a failure moving the static folder
	ErrFailedStatic = errors.NewError("failed to move static folder", 2)
	//ErrFailedGather is for a failure in gathering files.
	ErrFailedGather = errors.NewError("failed to gather files", 3)
	//ErrFailedCreateFS is for a failure in gathering files.
	ErrFailedCreateFS = errors.NewError("couldn't create directory/file", 4)
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
func Unfold(configSite *ConfigSite, spps []string) ([]*Site, errors.Error) {
	sites := make([]*Site, 0, len(configSite.Sites)*2)
	globalTemplates = make(map[string]*template.Template, len(sites))

	err := unfold(configSite, nil, &sites)
	if err != nil {
		return sites, err
	}

	for _, spp := range spps {
		if k, ok := SPPs[spp]; ok {
			sites = k.Process(sites, "")
		}
	}

	return sites, nil
}

func unfold(cSite *ConfigSite, parent *ConfigSite, sites *[]*Site) (err errors.Error) {
	if parent != nil {
		mergeConfigSite(cSite, parent)
	}
	//If this is the last in the chain, add it to the list of return values
	if len(cSite.Sites) == 0 {
		site := &Site{
			Slug: cSite.Slug,
		}

		err := gatherData(site, cSite.Data)
		if err != nil {
			return ErrFailedGather.SetRoot(err.Error())
		}

		err = gatherTemplates(site, cSite.Templates)
		if err != nil {
			return ErrFailedGather.SetRoot(err.Error())
		}

		//append site to the list of sites that will be executed
		*sites = append(*sites, site)
		return nil
	}

	for _, childSite := range cSite.Sites {
		err = unfold(childSite, cSite, sites)
		if err != nil {
			return err
		}
	}

	return
}

//mergeConfigSite merges the src into the dst
func mergeConfigSite(dst *ConfigSite, src *ConfigSite) {
	if dst.Data != nil {
		dst.Data = append(src.Data, dst.Data...) // just append
	} else {
		dst.Data = make([]data, len(src.Data)) // or make a new one and fill it
		for i, s := range src.Data {
			dst.Data[i] = s
		}
	}
	if dst.Templates != nil {
		dst.Templates = append(src.Templates, dst.Templates...) // just append
	} else {
		dst.Templates = make([]string, len(src.Templates)) // or make a new one and fill it
		for i, s := range src.Templates {
			dst.Templates[i] = s
		}
	}

	dst.Slug = src.Slug + dst.Slug
}

//collect data objects from modules
func gatherData(site *Site, files []data) errors.Error {
	for _, d := range files {

		//init data if it is empty
		if site.Data == nil {
			site.Data = make(tt.Data)
		}

		var data tt.Data

		fPipe := DataLoaders[d.loader].GetPipe(d.loaderArguments)
		pPipe := DataParsers[d.parser].GetPipe(d.parserArguments)
		var ppPipes []pipeline.Pipe
		var validPPPipes = 0
		for _, dpp := range d.postProcessors {
			ppPipes = append(ppPipes, DataPostProcessors[dpp.postProcessor].GetPipe(dpp.postProcessorArguments))
			if ppPipes != nil {
				validPPPipes++
			}
		}

		if fPipe != nil && pPipe != nil && len(ppPipes) == validPPPipes {
			var pipes = []pipeline.Pipe{
				fPipe,
				pPipe,
			}

			for _, dpp := range ppPipes {
				pipes = append(pipes, dpp)
			}

			pipeline.ExecPipeline(nil, &data, pipes...)
		} else {
			fileData := DataLoaders[d.loader].Load(d.loaderArguments)
			data = DataParsers[d.parser].Parse(fileData, d.parserArguments)
			for _, dpp := range d.postProcessors {
				data = DataPostProcessors[dpp.postProcessor].Process(data, dpp.postProcessorArguments)
			}
		}

		//add the parsed data to the site data
		for k, v := range data {
			site.Data[k] = v
		}
	}

	return nil
}

//TODO optimize the SHIT out od this.
func gatherTemplates(site *Site, templates []string) errors.Error {
	var newTemplates = make([]string, len(templates))
	for i, template := range templates {
		//prefix the templates with the TemplateFolder
		newTemplates[i] = filepath.Join(TemplateFolder, template)
	}

	var err error

	//get the template with the TemplateFunctions initalized
	id := randString(32)
	template, err := template.New("").Funcs(TemplateFunctions).ParseFiles(newTemplates...)
	if err != nil {
		return errors.Import(err)
	}

	globalTemplates[id] = template
	site.Template = id

	return nil
}

/*
	EXECUTE

	Iterate over the []Site and use the data to execute the template and export the result to the output file.
*/

//Execute the templates of a []Site into the final files
func Execute(sites []*Site) errors.Error {
	return execute(sites)
}
func execute(sites []*Site) errors.Error {
	//move static folder
	if StaticFolder != "" && OutputFolder != "" {
		info, err := os.Lstat(StaticFolder)
		if err != nil {
			return ErrFailedStatic.SetRoot(err.Error())
		}

		internal.GenCopy(StaticFolder, OutputFolder, info)
		if err != nil {
			return ErrFailedStatic.SetRoot(err.Error())
		}
	}

	//export every template
	for _, site := range sites {
		err := site.executeTemplate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (site *Site) executeTemplate() errors.Error {
	//prefix the slug with the output folder
	fileLocation := filepath.Join(OutputFolder, site.Slug)

	//check all folders in the path of the output file
	err := os.MkdirAll(filepath.Dir(fileLocation), 0766)
	if err != nil {
		return ErrFailedCreateFS.SetRoot(err.Error())
	}

	//create the file
	file, err := os.Create(fileLocation)
	if err != nil {
		return ErrFailedCreateFS.SetRoot(err.Error())
	}

	//fill the file by executing the template
	err = globalTemplates[site.Template].ExecuteTemplate(file, "html", site.Data)
	if err != nil {
		return ErrFailedTemplate.SetRoot(err.Error())
	}

	return nil
}

/*
	HELPERS

	This should go into a diferent file, but no suitable place has been found
*/

var srcRand = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func randString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, srcRand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = srcRand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
