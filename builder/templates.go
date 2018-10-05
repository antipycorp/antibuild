package builder

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"gitlab.com/antipy/antibuild/cli/builder/site"
)

//start the template execution
func executeTemplate(config *Config) (err error) {
	//check if the output folder is there and delete its contents
	if config.Folders.Output == "" {
		err = os.RemoveAll(config.Folders.Output)
	}

	if err != nil {
		fmt.Println("output not specified")
	}
	sites := config.Pages
	sitemap := site.SiteMap{}

	site.TemplateFunctions = &templateFunctions
	site.FileLoaders = &fileLoaders
	site.FileParsers = &fileParsers
	site.FilePostProcessors = &filePostProcessors

	config.Pages = site.Site{}
	config.Pages.Sites = make([]*site.Site, 1)
	config.Pages.Sites[0] = &sites
	config.Pages.OUTFolder = config.Folders.Output
	config.Pages.TemplateFolder = config.Folders.Templates
	config.Pages.JSONFolder = config.Folders.Data
	config.Pages.Static = config.Folders.Static
	config.Pages.SiteMap = &sitemap

	err = config.Pages.Unfold(nil)
	if err != nil {
		fmt.Println("failed to parse:", err)
	}
	err = config.Pages.Execute()
	if err != nil {
		fmt.Println("failed to Execute function:", err)
	}

	return
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
