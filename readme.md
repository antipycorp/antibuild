# Easy static site generator that uses go html templates and json files

## Usage

### How to install
`go get -u -v gitlab.com/antipy/antibuild`

### How to use the template builder
`antibuild --config {configfile} {options}`

- `{configfile}` will be the config.json containing all the sites

#### Options:

- `--development` refreshes the html files after changing a file in the templates directory or the json directory
- `--host` only works in combination with --development. Hosts a webserver at localhost from the output directory on port 8080, unless the POST envirement variable is set to specify a port, 
### config.json

#### Layout

The config.json has three parts ([type config](#examples)). It should contain a folder, modules and pages part.

___This is all outdated. More documentation will come soon. If you have a question please open an issue.___

```
If you do this you can also leave the root JSONFiles, Templates and/or Slug empty so that all your sub sites start with their own JSONFiles, Templates and/or Slug.

All sites inside the sides element of the site struct, will append on to the previously JSONFiles, templateFiles, and the slug will be appended as well. templateroot, jsonroot, and outroot will be overwritten by parrend, as long as that is not empty. If the sites element is not nil it will not be executed, so the [example below](#examples) will only create one site with the Slug: /index.html, Templates: [ "Layout.html" ], JSONFiles: [ "Layout.json", "pages/Home.json" ]. templateroot, jsonroot, and outroot will always be overwritten by the site's parent.

The templates variable should contain a list of template files relative to the templates directory.

The templates variable should contain a list of json files relative to the json directory. These files will be combined in order as of reference, so if you first json file contains the variable "Name" and a later referenced file has a variable named "Name" the latter will overwrite the first one.

The slug will be the ourput file relative to the output directory

All variable names in the JSON files should be capitalized as otherwise the template won't be able to use them

If staticroot is set on an site, antibuild will move everything from that static page to the outroot
```

### Examples

Type of a site in golang:
```golang
type (
	config struct {
		Folders configFolder  `json:"folders"`
		Modules configModules `json:"modules"`
		Pages   site          `json:"pages"`
	}
	
	configFolder struct {
		Templates string `json:"templates"`
		Data      string `json:"data"`
		Static    string `json:"static"`
		Output    string `json:"output"`
	}

	configModules struct {
		Dependencies map[string]string                 `json:"dependencies"`
		Config       map[string]map[string]interface{} `json:"config"`
	}

	site struct {
		Slug            string   `json:"slug"`
		Templates       []string `json:"templates"`
		Data            []string `json:"data"`
		Languages       []string `json:"languages"`
		DefaultLanguage string   `json:"language_default"`

		Sites []site `json:"sites"`
	}
)
```

Example in the config.json file:
```json
{
    "folders": {
        "templates": "templates/",
        "data": "data/",
        "output": "public/"
    },
    "pages": {
        "slug": "/content",
        "templates": [
            "Layout.html"
        ],
        "data": [
            "Layout.json"
        ],
        "sites": [
            {
                "slug": "/index.html",
                "data": [
                    "pages/Home.json"
                ]
            },
            {
                "Slug": "/alternate.html",
                "data": [
                    "pages/Links.json"
                ]
            }
        ]
    
    },
    "modules":{}
}
```