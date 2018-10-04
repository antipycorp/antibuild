package site

import (
	"encoding/json"
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
	SiteMap struct {
		sitemap []string
	}

	//Site contains info about a specific site
	Site struct {
		Slug            string   `json:"slug"`
		Templates       []string `json:"templates"`
		Data            []string `json:"data"`
		Languages       []string `json:"languages"`
		Sites           []*Site  `json:"sites"`
		TemplateFolder  string   `json:"templateroot"`
		DefaultLanguage string   `json:"defaultlanguage"`
		JSONFolder      string
		OUTFolder       string
		Static          string
		language        string
		SiteMap         *SiteMap
	}
	DataInput struct {
		Data map[string]interface{} `json:"Data"`
	}
)

var (
	TemplateFunctions  *template.FuncMap
	FileLoaders        *map[string]func(variable string) []byte
	FileParsers        *map[string]func(data []byte, variable string) map[string]interface{}
	FilePostProcessors *map[string]func(data map[string]interface{}, variable string) map[string]interface{}
)

func (s *Site) Unfold(parent *Site) error {
	return unfold(s, parent)
}

func (s *Site) Execute() error {
	return execute(s)
}

func unfoldStar(site, parent *Site, jIndex int, completeUnfoldChild bool) error {
	jsonPath := filepath.Dir(filepath.Join(site.JSONFolder, site.Data[jIndex]))
	jsonfile := strings.Replace(filepath.Base(site.Data[jIndex]), "*", "([^/]*)", -1)
	re := regexp.MustCompile(jsonfile)
	var matches [][][]string
	err := filepath.Walk(jsonPath, func(path string, file os.FileInfo, err error) error {
		if path == jsonPath {
			return nil
		}
		if file.IsDir() {
			return filepath.SkipDir
		}
		if ok, _ := regexp.MatchString(jsonfile, file.Name()); ok {
			matches = append(matches, re.FindAllStringSubmatch(file.Name(), -1))
		}
		return nil
	})
	if err != nil {
		return err
	}
	if parent == nil {
		return nil
	}
	for i, s := range parent.Sites {
		if s == site {
			parent.Sites = append(parent.Sites[:i], parent.Sites[i+1:]...)
		}
	}
	for _, file := range matches {
		s := site.copy()
		for _, match := range file {
			s.Slug = strings.Replace(s.Slug, "*", match[1], 1)
			s.Data[jIndex] = strings.Replace(s.Data[jIndex], "*", match[1], 1)
		}
		parent.Sites = append(parent.Sites, s)
		partialUnfold(s, parent, completeUnfoldChild)
	}
	return nil

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
		if site.Languages != nil {
			site.Languages = append(parent.Languages, site.Languages...)
		} else {
			site.Languages = make([]string, len(parent.Languages))
			copy(site.Languages, parent.Languages)
		}
		site.Slug = parent.Slug + site.Slug
		if parent.OUTFolder != "" {
			site.OUTFolder = parent.OUTFolder
		}
		if parent.TemplateFolder != "" {
			site.TemplateFolder = parent.TemplateFolder
		}
		if parent.JSONFolder != "" {
			site.JSONFolder = parent.JSONFolder
		}
		if parent.DefaultLanguage != "" {
			site.DefaultLanguage = parent.DefaultLanguage
		}
		site.SiteMap = parent.SiteMap
	}
	return partialUnfold(site, parent, true)
}
func partialUnfold(site, parent *Site, completeUnfoldChild bool) error {
	for jIndex, jsonfile := range site.Data {
		if strings.Contains(jsonfile, "*") {
			fmt.Println("a star!")
			return unfoldStar(site, parent, jIndex, completeUnfoldChild)
		}
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
	if len(site.Languages) != 0 && len(site.Sites) == 0 && site.language == "" {
		for i, s := range parent.Sites {
			if s == site {
				parent.Sites = append(parent.Sites[:i], parent.Sites[i+1:]...)
			}
		}
		for _, lang := range site.Languages {
			s := site.copy()
			s.language = lang
			parent.Sites = append(parent.Sites, s)
			err := partialUnfold(s, parent, false)
			if err != nil {
				return fmt.Errorf("could not execute %s the for lang %s:%s", site.Slug, lang, err)
			}
		}
		return nil
	}
	site.SiteMap.sitemap = append(site.SiteMap.sitemap, site.Slug)
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

func execute(site *Site) error {
	fmt.Println(*site)
	var (
		err       error
		template  *template.Template
		jsonImput DataInput
	)

	if site.Static != "" && site.OUTFolder != "" {
		fmt.Println("copying static files")
		info, err := os.Lstat(site.Static)
		if err != nil {
			return err
		}
		genCopy(site.Static, site.OUTFolder, info)
	}

	if site.TemplateFolder == "" || site.JSONFolder == "" || site.OUTFolder == "" || len(site.Sites) != 0 {
		goto skip
	}
	err = gatherJSON(site, &jsonImput)
	if err != nil {
		return err
	}
	template, err = gatherTemplates(site)
	if err != nil {
		return err
	}
	jsonImput.Data["sitemap"] = site.SiteMap.sitemap
	err = executeTemplate(site, template, jsonImput)
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

func gatherJSON(s *Site, dataInput *DataInput) error {
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
		file := (*FileLoaders)[loader[0]](loader[1])
		parser := strings.SplitN(matches[1][1], ":", 1)
		if len(parser) == 1 {
			parser = append(parser, "")
		}
		parsed := (*FileParsers)[parser[0]](file, parser[1])
		for k, v := range parsed {
			dataInput.Data[k] = v
		}
	}

	return nil
}

func gatherTemplates(s *Site) (*template.Template, error) {

	for i := range s.Templates {
		s.Templates[i] = filepath.Join(s.TemplateFolder, s.Templates[i])
	}

	template, err := template.New("").Funcs(*TemplateFunctions).ParseFiles(s.Templates...)
	if err != nil {
		return nil, fmt.Errorf("could not parse the template files: %v", err.Error())
	}
	return template, nil
}

func executeTemplate(s *Site, template *template.Template, jsonImput DataInput) error {
	OUTPath := filepath.Join(s.OUTFolder, s.Slug)
	if s.language != "" && s.DefaultLanguage != s.language {
		OUTPath = filepath.Join(s.OUTFolder, s.language, s.Slug)
	}

	err := os.MkdirAll(filepath.Dir(OUTPath), 0766)
	if err != nil {
		return errors.New("Couldn't create directory: " + err.Error())
	}

	OUTFile, err := os.Create(OUTPath)
	if err != nil {
		return errors.New("Couldn't create file: " + err.Error())
	}

	err = template.ExecuteTemplate(OUTFile, "html", jsonImput.Data)
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

func (ji *DataInput) UnmarshalJSON(data []byte) error {
	var input map[string]interface{}
	err := json.Unmarshal(data, &input)
	if err != nil {
		return err
	}

	if ji.Data == nil {
		ji.Data = make(map[string]interface{})
	}

	for name, in := range input {
		ji.Data[name] = in
	}
	return nil
}
