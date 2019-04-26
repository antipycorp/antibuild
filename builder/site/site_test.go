// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"gitlab.com/antipy/antibuild/cli/ui"

	"github.com/jaicewizard/tt"

	"gitlab.com/antipy/antibuild/cli/builder/site"
	"gitlab.com/antipy/antibuild/cli/modules/pipeline"
)

type unfoldPair struct {
	in    site.ConfigSite
	out   []*site.Site
	res   []*site.Site
	files map[string]string
}

type loader struct {
	data map[string][]byte
}

type parser struct{}

type iterator struct{}

var (
	testUI = &ui.UI{
		LogFile:   nil,
		PrettyLog: false,
	}
)

func init() {
	files := make(map[string][]byte)
	files["1"] = []byte(`{"data":"nothing"}`)
	site.DataLoaders["l"] = loader{data: files}
	site.DataParsers["p"] = parser{}
	site.Iterators["ls"] = iterator{}
	site.OutputFolder = "/tmp/templates/out/"
	site.TemplateFolder = "/tmp/templates/"
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
		in: site.ConfigSite{
			Slug: "/index.html",
			Data: []site.Data{
				site.Data{
					Loader:          "l",
					LoaderArguments: "1",
					Parser:          "p",
				},
			},
			Templates: []string{
				"t1",
			},
		},
		out: []*site.Site{
			&site.Site{
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
	unfoldPair{
		in: site.ConfigSite{
			Iterators: map[string]site.IteratorData{
				"article": site.IteratorData{
					Iterator: "ls",
				},
			},
			Slug: "/{{article}}/index.html",
			Data: []site.Data{
				site.Data{
					Loader:          "l",
					LoaderArguments: "1",
					Parser:          "p",
				},
			},
			Templates: []string{
				"t1",
			},
		},
		out: []*site.Site{
			&site.Site{
				Slug: "/hello/index.html",
				Data: tt.Data{
					"data": "nothing",
				},
			},
			&site.Site{
				Slug: "/world/index.html",
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

func (p parser) Parse(data []byte, useless string) tt.Data {
	var jsonData map[string]interface{}

	err := json.Unmarshal(data, &jsonData)
	if err != nil {
		panic(err)
	}

	var retData = make(tt.Data, len(jsonData))
	for k, v := range jsonData {
		retData[k] = v
	}

	return retData
}

func (p parser) GetPipe(variable string) pipeline.Pipe {
	return nil
}

func (i iterator) GetIterations(location string) []string {
	return []string{
		"hello",
		"world",
	}
}

func (i iterator) GetPipe(variable string) pipeline.Pipe {
	return nil
}

//Testunfold doesn't test template parsing, if anything failed it will be done during execute
func TestUnfold(t *testing.T) {
	for _, test := range unfoldTests {
		in := site.DeepCopy(test.in)
		dat, err := site.Unfold(&in, testUI)
		if err != nil {
			t.Fatal(err.Error())
		}

		for _, d := range dat {
			s, err := site.Gather(d, testUI)
			if err != nil {
				t.Fatal(err.Error())
			}
			test.res = append(test.res, s)
		}
		if len(test.out) != len(test.res) {
			for _, v := range test.res {
				print("\n" + v.Slug)
			}
			t.FailNow()
		}
		for i := 0; i < len(test.res); i++ {
			if test.out[i].Slug != test.res[i].Slug {
				print("should be: "+test.out[i].Slug+" but is: "+test.res[i].Slug+" for ", i, "\n")
				for _, v := range test.res {
					print(v.Slug + "\n")
				}
				t.FailNow()
			}
		}

		for k := range test.out[0].Data {
			if test.out[0].Data[k] != test.res[0].Data[k] {
				t.FailNow()
			}
		}
	}
}

func TestExecute(t *testing.T) {
	for _, test := range unfoldTests {
		in := site.DeepCopy(test.in)
		dat, err := site.Unfold(&in, testUI)
		if err != nil {
			t.Fatal(err.Error())
		}

		for _, d := range dat {
			s, err := site.Gather(d, testUI)
			if err != nil {
				t.Fatal(err.Error())
			}

			test.res = append(test.res, s)
		}
		site.Execute(test.res, testUI)

		for file, data := range test.files {
			dat, err := ioutil.ReadFile(site.OutputFolder + file)
			if err != nil {
				t.Fatal(err.Error())
			}

			if string(dat) != data {
				t.FailNow()
			}
		}
	}
}

var benchMarks = [...]site.ConfigSite{
	site.ConfigSite{
		Slug: "/index.html",
		Data: []site.Data{
			site.Data{
				Loader:          "l",
				LoaderArguments: "1",
				Parser:          "p",
			},
		},
		Templates: []string{
			"t1",
		},
	},
	site.ConfigSite{
		Iterators: map[string]site.IteratorData{
			"article": site.IteratorData{
				Iterator:          "ls",
				IteratorArguments: "/tmp/templates/iterators",
			},
		},
		Slug: "/{{article}}/index.html",
		Data: []site.Data{
			site.Data{
				Loader:          "l",
				LoaderArguments: "1",
				Parser:          "p",
			},
		},
		Templates: []string{
			"t1",
		},
	},
}

func BenchmarkUnfold(b *testing.B) {
	b.Run("simple-basic", genUnfold(0))
	b.Run("simple-iterator", genUnfold(1))
}

func genUnfold(benchID int) func(*testing.B) {
	return func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			s := site.DeepCopy(benchMarks[benchID])
			site.Unfold(&s, testUI)
		}
	}
}

func BenchmarkGather(b *testing.B) {
	b.Run("simple-basic", genGather(0))
	b.Run("simple-iterator", genGather(1))
}

func genGather(benchID int) func(*testing.B) {
	return func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			s := site.DeepCopy(benchMarks[benchID])

			site.Unfold(&s, testUI)
		}
		
		var sites = make([][]*site.ConfigSite, b.N)
		for n := 0; n < b.N; n++ {
			s := site.DeepCopy(benchMarks[benchID])
			sites[n], _ = site.Unfold(&s, testUI)
		}

		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			for _, d := range sites[n] {
				_, err := site.Gather(d, testUI)
				if err != nil {
					b.Fatal(err.Error())
				}
			}
		}
	}
}
