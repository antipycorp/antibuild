package internal

import (
	"encoding/gob"
)

var ()

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}
