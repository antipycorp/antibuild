```json
{
	"Slug": "/content",
	"Templates": [
		"Layout.html"
	],
	"JSONfiles": [
		"Layout.json"
	],
	"sites": [
		{
			"Slug": "/index.html",
			"JSONfiles": [
				"pages/Home.json"
			]
		}, 	//it will inherit /content from the parent slug, same for Layout.json ans Layout.html
		{
			"Slug": "/alternate.html",
			"JSONfiles": [
				"pages/links.json"
			]
		}, 	//it will inherit /content from the parent slug, same for Layout.json ans Layout.html,
			//since link.json has the same variable inside as Layout, links.json will overwrite Layout.json's links Headers
	]
}
```