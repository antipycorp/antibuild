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
)

//Unfold the ConfigSite into a []ConfigSite
func (s *Site) Unfold(parent *Site) error {
	return unfold(s, parent)
}

//Execute the templates of a []Site into the final files
func Convert(configSites []*ConfigSite) ([]*Site, error) {
	return convert(configSites)
}

//Execute the templates of a []Site into the final files
func (s *Site) Execute() error {
	return execute(s)
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

func (s *Site) copy() *Site {
	newSite := *s
	for i, site := range s.Sites {
		newSite.Sites[i] = site.copy()
	}
	newSite.Data = make([]string, len(s.Data))
	copy(newSite.Data, s.Data)

	newSite.Templates = make([]string, len(s.Templates))
	copy(newSite.Templates, s.Templates)

	return &newSite
}

func convert(configSites []*ConfigSite) ([]*Site, error) {

}

func execute(site *Site) error {
	fmt.Println(*site)
	var (
		err       error
		template  *template.Template
		dataInput dataInput
	)

	if StaticFolder != "" && OutputFolder != "" {
		fmt.Println("copying static files")
		info, err := os.Lstat(StaticFolder)
		if err != nil {
			return err
		}
		genCopy(StaticFolder, OutputFolder, info)
	}

	if TemplateFolder == "" || OutputFolder == "" || len(site.Sites) != 0 {
		goto skip
	}

	err = gatherData(site, &dataInput)
	if err != nil {
		return err
	}

	template, err = gatherTemplates(site)
	if err != nil {
		return err
	}

	err = executeTemplate(site, template, dataInput)
	if err != nil {
		return err
	}
	return nil
skip:
	for _, s := range site.Sites {
		fmt.Println(s.Slug)
		err := execute(s)
		if err != nil {
			return err
		}
	}

	return nil
}

func gatherData(s *Site, dataInput *dataInput) error {
	for _, dataFileString := range s.Data {

		if dataInput.Data == nil {
			dataInput.Data = make(map[string]interface{})
		}

		expression, err := regexp.Compile("\\[(.*?)\\]")
		if err != nil {
			return err
		}

		matches := expression.FindAllStringSubmatch(dataFileString, -1)

		loader := strings.SplitN(matches[0][1], ":", 2)
		if len(loader) == 1 {
			loader[1] = ""
		}

		file := FileLoaders[loader[0]].Load(loader[1])
		parser := strings.SplitN(matches[1][1], ":", 2)
		if len(parser) == 1 {
			parser = append(parser, "")
		}

		parsed := FileParsers[parser[0]].Parse(file, parser[1])
		for k, v := range parsed {
			dataInput.Data[k] = v
		}
	}

	return nil
}

func gatherTemplates(s *Site) (*template.Template, error) {

	for i := range s.Templates {
		s.Templates[i] = filepath.Join(TemplateFolder, s.Templates[i])
	}

	template, err := template.New("").Funcs(TemplateFunctions).ParseFiles(s.Templates...)
	if err != nil {
		return nil, fmt.Errorf("could not parse the template files: %v", err.Error())
	}
	return template, nil
}

func executeTemplate(s *Site, template *template.Template, dataInput dataInput) error {
	OUTPath := filepath.Join(OutputFolder, s.Slug)
	err := os.MkdirAll(filepath.Dir(OUTPath), 0766)
	if err != nil {
		return errors.New("Couldn't create directory: " + err.Error())
	}

	OUTFile, err := os.Create(OUTPath)
	if err != nil {
		return errors.New("Couldn't create file: " + err.Error())
	}

	err = template.ExecuteTemplate(OUTFile, "html", dataInput.Data)
	if err != nil {
		return errors.New("Could not parse: " + err.Error())
	}
	return nil
}

//!this should go into a diferent file, but no suitable place has been found
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
			// If any error, exit immediately
			return err
		}
	}
	return nil
}
