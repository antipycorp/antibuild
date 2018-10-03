package builder

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

//start the template execution
func executeTemplate(config *config) (err error) {
	//check if the output folder is there and delete its contents
	if config.Folders.Output == "" {
		err = os.RemoveAll(config.Folders.Output)
	}
	if err != nil {
		panic(errNoOutput)
	}

	fmt.Println("Start parsing...")

	//start the first page
	err = config.Pages.execute(nil, config)
	if err != nil {
		fmt.Println("failed to parse")
	}

	return
}

//execute each site
func (s *site) execute(parent *site, config *config) error {
	if parent != nil {
		if s.Data != nil {
			s.Data = append(parent.Data, s.Data...)
		} else {
			s.Data = make([]string, len(parent.Data))
			copy(s.Data, parent.Data)
		}
		if s.Templates != nil {
			s.Templates = append(parent.Templates, s.Templates...)
		} else {
			s.Templates = make([]string, len(parent.Templates))
			copy(s.Templates, parent.Templates)
		}

		s.Slug = parent.Slug + s.Slug
	}

	if config.Folders.Static != "" && config.Folders.Output != "" {
		info, err := os.Lstat(config.Folders.Static)
		if err != nil {
			return err
		}

		genCopy(config.Folders.Static, config.Folders.Output, info)
	}

	for i, dataFile := range s.Data {
		if strings.Contains(dataFile, "*") {
			return parseStar(s, config, i)
		}
	}

	if len(s.Sites) != 0 {
		for _, site := range s.Sites {
			err := site.execute(s, config)
			if err != nil {
				return err
			}
		}

		return nil
	}

	var dataFile dataFile
	err := s.gatherDataFiles(&dataFile, config)
	if err != nil {
		return err
	}

	template, err := s.gatherTemplates(config)
	if err != nil {
		return err
	}
	err = s.executeTemplate(template, dataFile, config)
	if err != nil {
		return err
	}

	return nil
}

func parseStar(s *site, config *config, index int) error {
	dataPath := filepath.Dir(filepath.Join(config.Folders.Data, s.Data[index]))
	dataFile := strings.Replace(filepath.Base(s.Data[index]), "*", "([^/]*)", -1)
	re := regexp.MustCompile(dataFile)

	var matches [][][]string

	err := filepath.Walk(dataPath, func(path string, file os.FileInfo, err error) error {
		if path == dataPath {
			return nil
		}

		if file.IsDir() {
			return filepath.SkipDir
		}

		if ok, _ := regexp.MatchString(dataFile, file.Name()); ok {
			matches = append(matches, re.FindAllStringSubmatch(file.Name(), -1))
		}
		return nil
	})
	if err != nil {
		return nil
	}

	for _, file := range matches {
		site := s.copy()
		for _, match := range file {
			site.Slug = strings.Replace(site.Slug, "*", match[1], 1)
			site.Data[index] = strings.Replace(site.Data[index], "*", match[1], 1)
		}

		err := site.execute(nil, config)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *site) gatherDataFiles(dataInput *dataFile, config *config) error {
	for _, dataFileString := range s.Data {
		expression, err := regexp.Compile("\\[(.*?)\\]")
		if err != nil {
			return err
		}

		matches := expression.FindAllStringSubmatch(dataFileString, -1)

		loader := strings.SplitN(matches[0][1], ":", 2)
		if len(loader) == 1 {
			loader[1] = ""
		}
		file := fileLoaders[loader[0]](loader[1])

		parser := strings.SplitN(matches[1][1], ":", 1)
		if len(parser) == 1 {
			parser = append(parser, "")
		}
		parsed := fileParsers[parser[0]](file, parser[1])

		dataInput.Data = combine(dataInput.Data, parsed)
	}

	return nil
}

func (s *site) gatherTemplates(config *config) (*template.Template, error) {

	for i := range s.Templates {
		s.Templates[i] = filepath.Join(config.Folders.Templates, s.Templates[i])
	}

	template, err := template.New("").Funcs(templateFunctions).ParseFiles(s.Templates...)
	if err != nil {
		return nil, fmt.Errorf("could not parse the template files: %v", err.Error())
	}
	return template, nil
}

func (s *site) executeTemplate(template *template.Template, jsonImput dataFile, config *config) error {
	OUTPath := filepath.Join(config.Folders.Output, s.Slug)

	err := os.MkdirAll(filepath.Dir(OUTPath), 0766)
	if err != nil {
		return errors.New("Couldn't create directory: " + err.Error())
	}

	OUTFile, err := os.Create(OUTPath)
	if err != nil {
		return errors.New("Couldn't create file: " + err.Error())
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")

	err = template.ExecuteTemplate(OUTFile, "html", jsonImput.Data)
	if err != nil {
		return errors.New("Could not parse: " + err.Error())
	}

	fmt.Println("Executing template for ", s.Slug)

	return nil
}

func (s *site) copy() site {
	newSite := *s
	for i, site := range s.Sites {
		newSite.Sites[i] = site.copy()
	}
	newSite.Data = make([]string, len(s.Data))
	copy(newSite.Data, s.Data)

	newSite.Templates = make([]string, len(s.Templates))
	copy(newSite.Templates, s.Templates)

	return newSite
}

func (ji *dataFile) UnmarshalJSON(data []byte) error {
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

func combine(a map[string]interface{}, b map[string]interface{}) map[string]interface{} {
	for k, v := range b {
		a[k] = v
	}

	return a
}
