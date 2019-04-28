// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site

import (
	"html/template"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"text/template/parse"
	"time"

	"gitlab.com/antipy/antibuild/cli/modules/pipeline"

	"github.com/jaicewizard/tt"
	"gitlab.com/antipy/antibuild/cli/internal"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	"gitlab.com/antipy/antibuild/cli/ui"
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

	//Iterator is a function thats able to post-process data
	Iterator interface {
		GetIterations(string) []string
	}

	pipeQue struct {
		pipe []pipeline.Pipe
		data tt.Data
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
	//SPPs are all the module site post processors
	SPPs = make(map[string]SPP)

	//Iterators are all the module iterators
	Iterators = make(map[string]Iterator)

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

	pipeLines = make(map[string]pipeQue)
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
	} else {
		// we might not have iterators values that we need right now
		// but if we can parse them now then we dont need duplicate work
		gatherIterators(cSite.Iterators)
	}

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
		for i, s := range src.Data {
			dst.Data[i] = s
		}
	}

	if dst.Templates != nil {
		dst.Templates = append(dst.Templates, src.Templates...) // just append
	} else {
		dst.Templates = make([]string, len(src.Templates)) // or make a new one and fill it
		for i, s := range src.Templates {
			dst.Templates[i] = s
		}
	}

	if dst.Dependencies != nil {
		dst.Dependencies = append(dst.Dependencies, src.Dependencies...) // just append
	} else {
		dst.Dependencies = make([]string, len(src.Dependencies)) // or make a new one and fill it
		for i, s := range src.Dependencies {
			dst.Dependencies[i] = s
		}
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
		var ppPipes []pipeline.Pipe
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

		if fPipe != nil && pPipe != nil && len(ppPipes) == validPPPipes {
			var pipes = []pipeline.Pipe{
				fPipe,
				pPipe,
			}

			for _, dpp := range ppPipes {
				pipes = append(pipes, dpp)
			}

			bytes, _ := pipeline.ExecPipeline(nil, pipes...)
			data.GobDecode(bytes)
		} else {
			fileData := DataLoaders[d.Loader].Load(d.LoaderArguments)
			data = DataParsers[d.Parser].Parse(fileData, d.ParserArguments)
			for _, dpp := range d.PostProcessors {
				data = DataPostProcessors[dpp.PostProcessor].Process(data, dpp.PostProcessorArguments)
			}
		}

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

	//get the template with the TemplateFunctions initalized
	id := randString(32)

	globalTemplates[id] = finalTemplate
	site.Template = id

	return nil
}

// PostProcess all sites
func PostProcess(sites *[]*Site, spps []string, log *ui.UI) errors.Error {
	for _, spp := range spps {
		if k, ok := SPPs[spp]; ok {
			*sites = k.Process(*sites, "")
		}
	}

	return nil
}

/*
	EXECUTE

	Iterate over the []Site and use the data to execute the template and export the result to the output file.
*/

//Execute the templates of a []Site into the final files
func Execute(sites []*Site, log *ui.UI) errors.Error {
	return execute(sites, log)
}

func execute(sites []*Site, log *ui.UI) errors.Error {
	// copy static folder
	if StaticFolder != "" && OutputFolder != "" {
		log.Debug("Copying static folder")

		info, err := os.Lstat(StaticFolder)
		if err != nil {
			return ErrFailedStatic.SetRoot(err.Error())
		}

		internal.GenCopy(StaticFolder, OutputFolder, info)
		if err != nil {
			return ErrFailedStatic.SetRoot(err.Error())
		}

		log.Debug("Finished copying static folder")
	}

	//export every template
	for _, site := range sites {
		log.Debugf("Building page for %s", site.Slug)

		err := executeTemplate(site)
		if err != nil {
			return err
		}
	}
	return nil
}

func executeTemplate(site *Site) errors.Error {
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

//GetTemplateTree gets the template according to
func GetTemplateTree(template string) *parse.Tree {
	if v, ok := globalTemplates[template]; ok {
		return v.Tree
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
