package site

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type (
	//ConfigSite is the way a site is defined in the config file
	ConfigSite struct {
		Slug      string        `json:"slug"`
		Templates []string      `json:"templates"`
		Data      []string      `json:"data"`
		Sites     []*ConfigSite `json:"sites"`
	}

	//Site is the way a site is defined after all of its data and templates have been collected
	Site struct {
		Slug     string
		Template *template.Template
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

	//the splitter that is used to look what files should be parsed.
	dataFileStringSplitter *regexp.Regexp
)

func init() {
	var err error
	dataFileStringSplitter, err = regexp.Compile("\\[(.*?)\\]")
	if err != nil {
		panic(err)
	}
}

/*
	UNFOLD

	Parse the tree into the individual sites by iterating over the children of parents
	and combining their data until only sites with no more the children remain. Add
	these sites to an array, so there is no more nesting. The “Sites” tag of ConfigSite
	should now be disregarded.
*/

//Unfold the ConfigSite into a []ConfigSite
func (site *Site) Unfold(parent *Site) error {
	return unfold(site, parent)
}

func unfold(site, parent *Site) error {
	if parent != nil {
		if site.Data != nil {
			site.Data = append(parent.Data, site.Data...)
		} else {
			site.Data = make([]string, len(parent.Data))
			copy(site.Data, parent.Data)
		}
		if site.Templates != nil {
			site.Templates = append(parent.Templates, site.Templates...)
		} else {
			site.Templates = make([]string, len(parent.Templates))
			copy(site.Templates, parent.Templates)
		}

		site.Slug = parent.Slug + site.Slug
	}

	return partialUnfold(site, parent, true)
}
func partialUnfold(site, parent *Site, completeUnfoldChild bool) error {
	/*
		for jIndex, datafile := range site.Data {
			if strings.Contains(datafile, "*") {
				fmt.Println("a star!")
				return unfoldStar(site, parent, jIndex, completeUnfoldChild)
			}
		}
	*/
	sites := make([]*Site, len(site.Sites))
	for i, s := range site.Sites {
		sites[i] = s
		//fmt.Println("supposed to be DOING NEXT:", s.Slug)
	}

	fmt.Println(site.Sites)
	for _, s := range site.Sites {
		if completeUnfoldChild {
			err := unfold(s, site)
			if err != nil {
				return err
			}
		} else {
			err := partialUnfold(s, site, false)
			if err != nil {
				return err
			}
		}
	}
	if len(site.Sites) != 0 {
		return nil
	}
	return nil
}

func (site *Site) copy() *Site {
	newSite := *site
	for i, site := range site.Sites {
		newSite.Sites[i] = site.copy()
	}
	newSite.Data = make([]string, len(site.Data))
	copy(newSite.Data, site.Data)

	newSite.Templates = make([]string, len(site.Templates))
	copy(newSite.Templates, site.Templates)

	return &newSite
}

/*
	CONVERT

	Iterate over the []ConfigSite and use modules/file loaders and template
	parsers to collect all of the templates and data points. Put these into
	an array of Site to be executed later.
*/

//Convert the []ConfigSite into a []Site by collecting all data and templates
func Convert(configSites []*ConfigSite) ([]*Site, error) {
	return convert(configSites)
}

func convert(configSites []*ConfigSite) ([]*Site, error) {
	//Make the array that will store the converted sites
	sites := make([]*Site, len(configSites))

	for _, configSite := range configSites {
		//initalize a new Site with the correct slug
		site := &Site{
			Slug: configSite.Slug,
		}

		//collect data
		err := gatherData(site, configSite.Data)
		if err != nil {
			return nil, err
		}

		//collect template
		err = gatherTemplates(site, configSite.Data)
		if err != nil {
			return nil, err
		}

		//add to finished sites
		sites = append(sites, site)
	}

	return sites, nil
}

//collect data objects from modules
//!!! need to implement post processors !!!
func gatherData(site *Site, files []string) error {
	for _, dataFileString := range files {
		//init data if it is empty
		if site.Data == nil {
			site.Data = make(map[string]interface{})
		}

		//Split the dataFileString into its components using regexp
		matches := dataFileStringSplitter.FindAllStringSubmatch(dataFileString, -1)

		//split the loader by : and if there is no variable, define an empty one
		loader := strings.SplitN(matches[0][1], ":", 2)
		if len(loader) == 1 {
			loader[1] = ""
		}

		//get the file using the defined module
		file := FileLoaders[loader[0]].Load(loader[1])

		//split the parser by : and if there is no variable, define an empty one
		parser := strings.SplitN(matches[1][1], ":", 2)
		if len(parser) == 1 {
			parser = append(parser, "")
		}

		//parse the file using the defined module
		parsed := FileParsers[parser[0]].Parse(file, parser[1])

		//add the parsed data to the site data
		for k, v := range parsed {
			site.Data[k] = v
		}
	}

	return nil
}

func gatherTemplates(site *Site, templates []string) error {
	for i, template := range templates {
		//prefix the templates with the TemplateFolder
		templates[i] = filepath.Join(TemplateFolder, template)
	}

	var err error
	//get the template with the TemplateFunctions initalized
	site.Template, err = template.New("").Funcs(TemplateFunctions).ParseFiles(templates...)
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
		err = site.executeTemplate()
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
	err = site.Template.ExecuteTemplate(file, "html", site.Data)
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
