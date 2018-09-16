# easy template builder for generating static pages.

## usage

### how to install
`go get -u -v gitlab.com/antipy/templatebuilder`

### how to use the template builder
`templatebuilder --templates {templates directory} --json {json directory} --out {output directory} {options}`

- `{templates dierectory}` will be the root directory containing all the templates.
- `{json directory}` will be the root directory containing all json file, this must containing a config.json.
- `{output directory}` will be the root directory containing all output files

#### options:

- `--development` refreshes the html files after changing a file in the templates directory or the json directory

### config.json

#### layout

The config.json should contain a variable called "sites" which should be an array of  [type site](#examples)

The templates variable should contain a list of template files relative to the templates directory.

The templates variable should contain a list of json files relative to the json directory. These files will be combined in order as of reference, so if you first json file contains the variable "Name" and a later referenced file has a variable named "Name" the latter will overwrite the first one.

The slug will be the ourput file relative to the output directory

All variable names in the JSON files should be capitalized as otherwise the template won't be able to use them

### examples

type site:

```golang
type site struct {
    Slug      string   `json:"Slug"`
    Templates []string `json:"Templates"`
    JSONFiles []string `json:"JSONfiles"`
}
```

config.json:
```json
{
    "sites": [
        {
            "Slug": "/index.html",
			"Templates": [
				"Layout.html"
			],
			"JSONfiles": [
				"Layout.json",
				"pages/Home.json"
			]
		}
    ]
}
```