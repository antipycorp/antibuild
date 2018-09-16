# Easy template builder for generating static pages using go html templates and json files

## Usage

### How to install
`go get -u -v gitlab.com/antipy/templatebuilder`

### How to use the template builder
`templatebuilder --templates {templates directory} --json {json directory} --out {output directory} {options}`

- `{templates dierectory}` will be the root directory containing all the templates.
- `{json directory}` will be the root directory containing all json file, this must containing a config.json.
- `{output directory}` will be the root directory containing all output files

#### Options:

- `--development` refreshes the html files after changing a file in the templates directory or the json directory
- `--host` only works in combination with --development. Hosts a webserver at localhost:8080 from the output directory

### config.json

#### Layout

The config.json should contain a variable of type [type site](#examples), only the sites variable has to be set.

All sites inside the sides element of the site struct, will append on to the previously JSONFiles, templateFiles, and the slug will be appended as well. if the sites element is not nil it will not be executed, so the [example below](#examples) will only create one site with th Slug: /index.html, Templates: [ "Layout.html" ], JSONFiles: [ "Layout.json", "pages/Home.json" ]

The templates variable should contain a list of template files relative to the templates directory.

The templates variable should contain a list of json files relative to the json directory. These files will be combined in order as of reference, so if you first json file contains the variable "Name" and a later referenced file has a variable named "Name" the latter will overwrite the first one.

The slug will be the ourput file relative to the output directory

All variable names in the JSON files should be capitalized as otherwise the template won't be able to use them

### Examples

Type of a site in golang:
```golang
type site struct {
    Slug      string   `json:"Slug"`
    Templates []string `json:"Templates"`
    JSONFiles []string `json:"JSONfiles"`
	Sites     []site   `json:"sites"`
}
```

Example in the config.json file:
```json
{
    "sites": [
        {
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
    ]
}
```