// Copyright © 2018 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/jaicewizard/tt"

	"gitlab.com/antipy/antibuild/api/site"
	"gitlab.com/antipy/antibuild/cli/modules/pipeline"
)

type unfoldPair struct {
	in    ConfigSite
	out   []*site.Site
	res   []*site.Site
	files map[string]string
}

type loader struct {
	data map[string][]byte
}

type fParser struct{}

func init() {
	files := make(map[string][]byte)
	files["1"] = []byte(`{"data":"nothing"}`)
	DataLoaders["l"] = loader{data: files}
	DataParsers["p"] = fParser{}
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
		out: []*site.Site{
			&Site{
				Slug: "/index.html",
				Data: tt.Data{
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

func (l loader) GetPipe(variable string) pipeline.Pipe {
	return nil
}

func (p fParser) Parse(data []byte, useless string) tt.Data {
	var jsonData map[string]interface{}

	json.Unmarshal(data, &jsonData)
	var retData = make(tt.Data, len(jsonData))
	for k, v := range jsonData {
		retData[k] = v
	}

	return retData
}
func (p fParser) GetPipe(variable string) pipeline.Pipe {
	return nil
}

//Testunfold doesnt test template parsing, if anything failed it will be done during execute
func TestUnfold(t *testing.T) {
	for _, test := range unfoldTests {
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
