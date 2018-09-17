```json
{
	"templateroot": "templates/", //the root of the template files
	"jsonroot": "json/", //the root of the json files
	"outroot": "public/", //the root of the output
	"staticroot": "static/", //the root of any static files

	"Slug": "/index.html", // the output file relative to the specified output folder
	"Templates": [
		"Layout.html" // all the templates needed to build this page
	],
	"JSONfiles": [
		"Layout.json",
		"pages/Home.json"  // all the JSON files needed to build this page
	]
}
```