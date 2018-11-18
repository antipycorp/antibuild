// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

type unfoldPair struct {
	in    ConfigSite
	out   []*Site
	res   []*Site
	files map[string]string
}

type loader struct {
	data map[string][]byte
}

type fParser struct{}

func init() {
	files := make(map[string][]byte)
	files["1"] = []byte(`{"data":"nothing"}`)
	FileLoaders["l"] = loader{data: files}
	FileParsers["p"] = fParser{}
	OutputFolder = "/tmp/templates/out/"
	TemplateFolder = "/tmp/templates/"
	err := os.MkdirAll("/tmp/templates/", 0777)
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile("/tmp/templates/t1", []byte("{{define \"html\"}}hello darkness my old friend{{end}}"), 0777)
	if err != nil {
		panic(err)
	}
}

var unfoldTests = []unfoldPair{
	unfoldPair{
		in: ConfigSite{
			Slug: "/index.html",
			Data: []datafile{
				datafile{
					loader:          "l",
					loaderArguments: "1",
					parser:          "p",
				},
			},
			Templates: []string{
				"t1",
			},
		},
		out: []*Site{
			&Site{
				Slug: "/index.html",
				Data: map[string]interface{}{
					"data": "nothing",
				},
			},
		},
		files: map[string]string{
			"/index.html": "hello darkness my old friend",
		},
	},
}

func (l loader) Load(f string) []byte {
	return l.data[f]
}
func (p fParser) Parse(data []byte, useless string) (ret map[string]interface{}) {
	json.Unmarshal(data, &ret)
	return
}

//Testunfold doesnt test template parsing, if anything failled it will be done during execute
func TestUnfold(t *testing.T) {
	for _, test := range unfoldTests {
		fmt.Println(test.in)
		dat, _ := (Unfold(&test.in, nil))
		test.res = dat
		if dat[0].Slug != test.out[0].Slug {
			t.FailNow()
		}

		for k := range test.out[0].Data {
			if test.out[0].Data[k] != dat[0].Data[k] {
				t.FailNow()
			}
		}
	}
}

func TestExecute(t *testing.T) {
	for _, test := range unfoldTests {
		Execute(test.res)
		for file, data := range test.files {
			dat, err := ioutil.ReadFile(OutputFolder + file)
			if err != nil {
				t.FailNow()
			}
			if string(dat) != data {
				t.FailNow()
			}
		}
	}
}
