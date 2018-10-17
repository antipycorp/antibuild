## Example project using basic built-in modules
This is the config file for this example, it uses most built-in modules, it shows you how you can load files, configure slugs and nest them.

```json
{
    "folders": {
        "templates": "templates/",
        "output": "public/",
        "modules": ".modules/"
    },
    "modules": {
        "dependencies": {
            "file": "0.0.1",
            "json": "0.0.1",
            "math": "0.0.1",
            "yaml": "0.0.1",
            "language": "0.0.1"
        },
        "config": {
            "language": {
                "languages": [
                    "en",
                    "nl"
                ],
                "default": "en"
            }
        },
        "site_post_processors": [
            "language_language"
        ]
    },
    "pages": {
        "slug": "",
        "templates": [
            "Layout.html"
        ],
        "data": [
            "[file_file:data/Layout.json][json_json]"
        ],
        "sites": [
            {
                "slug": "/index.html",
                "templates": [],
                "data": [
                    "[file_file:data/pages/Home.yml][yaml_yaml]"
                ]
            }
        ]
    }
}
```

These are the data files for this project, the lists are rangable and everything under a language will only be used for that language.
```json
{
    "Name": "examples",
    "Headers": [
        {
            "Name": "Home"
        },
        {
            "Name": "example"
        }
    ]
}
```
This data file is written in yaml, as you can see in the config file you can load these using the yaml module instead of the json one.
```yaml
nl: 
  Page: 
    Name: Thuis

en:
  Page: 
    Name: Home
```

this is the html template for the this exameple

```html
{{define "html"}}
    <!DOCTYPE html>
    <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <meta http-equiv="X-UA-Compatible" content="ie=edge">
            <!-- you can acces all variables in both JSON files,
            .Page.Name is from the pages/Home.yaml and .Name is from Layout.json
            .Page.Name is loaded per language, so the /nl/ version would use Thuis  
            -->

            <title>{{.Page.Name}} | {{.Name}}</title>
        </head>
        <body>
            <div>
                <!--you can use all variables as normal, you can itterateover them etc-->
                {{range .Headers }}
                    <h1>{{.Name}}</h1>
                {{end}}

                <!-- these are all the functions provided by the math_ -->
                {{math_add 1 2}} 
                {{math_subtract 3 1}}
                {{math_multiply 2 3}}
                {{math_divide 6 2}}
                {{math_modulo 5 2}}
                {{math_power 3 3}}
            </div>
        </body>
    </html>
{{end}}
```

