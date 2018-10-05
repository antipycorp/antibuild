package builder

import (
	"fmt"
	"os"

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

	config.Pages = site.Site{}
	config.Pages.Sites = make([]*site.Site, 1)
	config.Pages.Sites[0] = &sites
	config.Pages.OUTFolder = config.Folders.Output
	config.Pages.TemplateFolder = config.Folders.Templates
	config.Pages.DataFolder = config.Folders.Data
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
