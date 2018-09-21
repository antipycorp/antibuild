```json
{
	"templateroot": "templates/", //the root of the template files
	"jsonroot": "json/", //the root of the json files
	"outroot": "public/", //the root of the output

	"Slug": "*.html", //{name of the json file}.html
	"Templates": [
		"Layout.html"
	],
	"JSONfiles": [
		"*.json" // matches any json file ending with .json
	],
	"languages": [ // all the languages available, are inherited same way as templates and jsonfiles, if the languages are not found in the json files it will use the whole JSON file.
		"en",
		"nl"
	],
	"defaultlanguage": "en" // the default language will not have a prefix in the url. eg: /layout.html /nl/layout.html
}
```