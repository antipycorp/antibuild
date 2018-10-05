package internal

import (
	"encoding/gob"
	"errors"
	"html/template"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
	gob.Register(errors.New("gob"))
	gob.Register(template.HTML("<h1>GOB</h1>"))
	gob.Register(template.HTMLAttr("gob"))
	gob.Register(template.JS("console.log(\"GOB\""))
	gob.Register(template.JSStr("gob"))
	gob.Register(template.URL("https://build.antipy.com"))
}
