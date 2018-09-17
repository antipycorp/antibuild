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

The config.json should contain a variable of type [type site](#examples). the root site should contain a templateroot, jsonroot, and outroot. If you want multiple distinct websites you can leve these out but make sure to have them futher up in the config file!!! If you do this you can also leave the root JSONFiles, Templates and/or Slug empty so that all your sub sites start with their own JSONFiles, Templates and/or Slug.

All sites inside the sides element of the site struct, will append on to the previously JSONFiles, templateFiles, and the slug will be appended as well. templateroot, jsonroot, and outroot will be overwritten by parrend, as long as that is not empty. If the sites element is not nil it will not be executed, so the [example below](#examples) will only create one site with the Slug: /index.html, Templates: [ "Layout.html" ], JSONFiles: [ "Layout.json", "pages/Home.json" ]. templateroot, jsonroot, and outroot will always be overwritten by the site's parent.

The templates variable should contain a list of template files relative to the templates directory.

The templates variable should contain a list of json files relative to the json directory. These files will be combined in order as of reference, so if you first json file contains the variable "Name" and a later referenced file has a variable named "Name" the latter will overwrite the first one.

The slug will be the ourput file relative to the output directory

All variable names in the JSON files should be capitalized as otherwise the template won't be able to use them

If staticroot is set on an site, antibuild will move everything from that static page to the outroot

### Examples

Type of a site in golang:
```golang
type site struct {
	Slug           string   `json:"Slug"`
	Templates      []string `json:"Templates"`
	JSONFiles      []string `json:"JSONfiles"`
	Sites          []site   `json:"sites"`
	TemplateFolder string   `json:"templateroot"`
	JSONFolder     string   `json:"jsonroot"`
	OUTFolder      string   `json:"outroot"`
	Static         string   `json:"staticroot"`
}
```

Example in the config.json file:
```json
{
	"templateroot": "templates/",
	"jsonroot": "json/",
	"outroot": "public/",
	"staticroot": "static/",
	
	"Slug": "/index",
	"Templates": [
		"Layout.html"
	],
	"JSONfiles": [
		"Layout.json"
	],
	"sites": [
		{
			"Slug": ".html",
			"JSONfiles": [
				"pages/Home.json"
			]
		}
	]
}
```