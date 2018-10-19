package site

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"time"
)

type (
	//ConfigSite is the way a site is defined in the config file
	ConfigSite struct {
		Slug      string        `json:"slug"`
		Templates []string      `json:"templates"`
		Data      []datafile    `json:"data"`
		Sites     []*ConfigSite `json:"sites"`
	}

	//Site is the way a site is defined after all of its data and templates have been collected
	Site struct {
		Slug     string
		Template string
		Data     map[string]interface{}
	}

	//FileLoader is a module that loads data
	FileLoader interface {
		Load(string) []byte
	}

	//FileParser is a module that parses loaded data
	FileParser interface {
		Parse([]byte, string) map[string]interface{}
	}

	//FilePostProcessor is a function thats able to post-process data
	FilePostProcessor interface {
		Process(map[string]interface{}, string) map[string]interface{}
	}
)

var (
	//TemplateFunctions are all the template functions defined by modules
	TemplateFunctions = template.FuncMap{}

	//FileLoaders are all the module file loaders
	FileLoaders = make(map[string]FileLoader)
	//FileParsers are all the module file parsers
	FileParsers = make(map[string]FileParser)
	//FilePostProcessors are all the module data post processors
	FilePostProcessors = make(map[string]FilePostProcessor)

	//TemplateFolder is the folder all templates are stored
	TemplateFolder string
	//StaticFolder is the folder all static files are stored
	StaticFolder string
	//OutputFolder is the folder that should be exported to
	OutputFolder string

	globalTemplates = make(map[string]*template.Template)
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
func Unfold(configSite *ConfigSite) ([]*Site, error) {
	return unfold(configSite, nil)
}

func unfold(cSite *ConfigSite, parent *ConfigSite) (sites []*Site, err error) {
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
			return nil, err
		}

		err = gatherTemplates(site, cSite.Templates)
		if err != nil {
			return nil, err
		}
		//append site to the list of sites that will be executed
		sites = append(sites, site)
		return sites, err
	}

	for _, childSite := range cSite.Sites {
		var appendSites []*Site
		appendSites, err = unfold(childSite, cSite)
		sites = append(sites, appendSites...) //!TODO have one slice of sites which will be added on by the childs themselves, avoid large realocs
	}

	return
}

//mergeConfigSite merges the src into the dst
func mergeConfigSite(dst *ConfigSite, src *ConfigSite) {
	if dst.Data != nil {
		dst.Data = append(src.Data, dst.Data...)
	} else {
		dst.Data = make([]datafile, len(src.Data))
		for i, s := range src.Data {
			dst.Data[i] = s
		}
	}
	if dst.Templates != nil {
		dst.Templates = append(src.Templates, dst.Templates...)
	} else {
		dst.Templates = make([]string, len(src.Templates))
		for i, s := range src.Templates {
			dst.Templates[i] = s
		}
	}

	dst.Slug = src.Slug + dst.Slug
}

//collect data objects from modules
//!!! need to implement post processors !!!
func gatherData(site *Site, files []datafile) error {
	for _, datafile := range files {
		fmt.Println(datafile)

		//init data if it is empty
		if site.Data == nil {
			site.Data = make(map[string]interface{})
		}

		//load and parse data
		file := FileLoaders[datafile.loader].Load(datafile.file)
		parsed := FileParsers[datafile.parser].Parse(file, datafile.parseArg)

		//add the parsed data to the site data
		for k, v := range parsed {
			site.Data[k] = v
		}
	}

	return nil
}

func gatherTemplates(site *Site, templates []string) error {
	var newTemplates = make([]string, len(templates))
	for i, template := range templates {
		//prefix the templates with the TemplateFolder
		newTemplates[i] = filepath.Join(TemplateFolder, template)
	}

	var err error
	//get the template with the TemplateFunctions initalized
	id := randString(32)
	globalTemplates[id], err = template.New("").Funcs(TemplateFunctions).ParseFiles(newTemplates...)
	site.Template = id
	if err != nil {
		return fmt.Errorf("could not parse the template files: %v", err.Error())
	}

	return nil
}

/*
	EXECUTE

	Iterate over the []Site and use the data to execute the template and export the result to the output file.
*/

//Execute the templates of a []Site into the final files
func Execute(sites []*Site) error {
	return execute(sites)
}

func execute(sites []*Site) error {
	//move static folder
	if StaticFolder != "" && OutputFolder != "" {
		info, err := os.Lstat(StaticFolder)
		if err != nil {
			return err
		}

		genCopy(StaticFolder, OutputFolder, info)
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

func (site *Site) executeTemplate() error {
	//prefix the slug with the output folder
	fileLocation := filepath.Join(OutputFolder, site.Slug)

	//check all folders in the path of the output file
	err := os.MkdirAll(filepath.Dir(fileLocation), 0766)
	if err != nil {
		return errors.New("couldn't create directory: " + err.Error())
	}

	//create the file
	file, err := os.Create(fileLocation)
	if err != nil {
		return errors.New("couldn't create file: " + err.Error())
	}

	//fill the file by executing the template
	err = globalTemplates[site.Template].ExecuteTemplate(file, "html", site.Data)
	if err != nil {
		return errors.New("could not execute template: " + err.Error())
	}

	return nil
}

/*
	HELPERS

	This should go into a diferent file, but no suitable place has been found
*/

func genCopy(src, dest string, info os.FileInfo) error {
	if info.IsDir() {
		return dirCopy(src, dest, info)
	}
	return fileCopy(src, dest, info)
}

func fileCopy(src, dest string, info os.FileInfo) error {

	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_, err = io.Copy(f, s)
	return err
}

func dirCopy(srcdir, destdir string, info os.FileInfo) error {

	if err := os.MkdirAll(destdir, info.Mode()); err != nil {
		return err
	}

	contents, err := ioutil.ReadDir(srcdir)
	if err != nil {
		return err
	}

	for _, content := range contents {
		cs, cd := filepath.Join(srcdir, content.Name()), filepath.Join(destdir, content.Name())
		if err := genCopy(cs, cd, content); err != nil {
			return err
		}
	}
	return nil
}

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
