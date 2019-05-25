// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package site_test

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"gitlab.com/antipy/antibuild/cli/engine/modules"
	ui "gitlab.com/antipy/antibuild/cli/internal/log"

	"github.com/jaicewizard/tt"

	"gitlab.com/antipy/antibuild/cli/engine"
)

type unfoldPair struct {
	in    engine.ConfigSite
	out   []*engine.Site
	res   []*engine.Site
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
	tmpDir, err := ioutil.TempDir("", "templates")
	if err != nil {
		panic(err)
	}

	files := make(map[string][]byte)
	files["1"] = []byte(`{"data":"nothing"}`)
	engine.DataLoaders["l"] = loader{data: files}
	engine.DataParsers["p"] = parser{}
	engine.Iterators["ls"] = iterator{}
	engine.OutputFolder = tmpDir + "/out/"
	engine.TemplateFolder = tmpDir
	err = ioutil.WriteFile(tmpDir+"/t1", []byte("{{define \"html\"}}hello darkness my old friend{{end}}"), 0777)
	if err != nil {
		panic(err)
	}
}

var unfoldTests = []unfoldPair{
	unfoldPair{
		in: engine.ConfigSite{
			Slug: "/index.html",
			Data: []engine.Data{
				engine.Data{
					Loader:          "l",
					LoaderArguments: "1",
					Parser:          "p",
				},
			},
			Templates: []string{
				"t1",
			},
		},
		out: []*engine.Site{
			&engine.Site{
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
		in: engine.ConfigSite{
			Iterators: map[string]engine.IteratorData{
				"article": engine.IteratorData{
					Iterator: "ls",
				},
			},
			Slug: "/{{article}}/index.html",
			Data: []engine.Data{
				engine.Data{
					Loader:          "l",
					LoaderArguments: "1",
					Parser:          "p",
				},
			},
			Templates: []string{
				"t1",
			},
		},
		out: []*engine.Site{
			&engine.Site{
				Slug: "/hello/index.html",
				Data: tt.Data{
					"data": "nothing",
				},
			},
			&engine.Site{
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

func (l loader) GetPipe(variable string) modules.Pipe {
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

func (p parser) GetPipe(variable string) modules.Pipe {
	return nil
}

func (i iterator) GetIterations(location string) []string {
	return []string{
		"hello",
		"world",
	}
}

func (i iterator) GetPipe(variable string) modules.Pipe {
	return nil
}

//Testunfold doesn't test template parsing, if anything failed it will be done during execute
func TestUnfold(t *testing.T) {
	for _, test := range unfoldTests {
		in := engine.DeepCopy(test.in)
		dat, err := engine.Unfold(&in, testUI)
		if err != nil {
			t.Fatal(err.Error())
		}

		for _, d := range dat {
			s, err := engine.Gather(d, testUI)
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
		in := engine.DeepCopy(test.in)
		dat, err := engine.Unfold(&in, testUI)
		if err != nil {
			t.Fatal(err.Error())
		}

		for _, d := range dat {
			s, err := engine.Gather(d, testUI)
			if err != nil {
				t.Fatal(err.Error())
			}

			test.res = append(test.res, s)
		}
		engine.Execute(test.res, testUI)

		for file, data := range test.files {
			dat, err := ioutil.ReadFile(engine.OutputFolder + file)
			if err != nil {
				t.Fatal(err.Error())
			}

			if string(dat) != data {
				t.FailNow()
			}
		}
	}
}

var benchMarks = [...]engine.ConfigSite{
	engine.ConfigSite{
		Slug: "/index.html",
		Data: []engine.Data{
			engine.Data{
				Loader:          "l",
				LoaderArguments: "1",
				Parser:          "p",
			},
		},
		Templates: []string{
			"t1",
		},
	},
	engine.ConfigSite{
		Iterators: map[string]engine.IteratorData{
			"article": engine.IteratorData{
				Iterator:          "ls",
				IteratorArguments: "/tmp/templates/iterators",
			},
		},
		Slug: "/{{article}}/index.html",
		Data: []engine.Data{
			engine.Data{
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
			s := engine.DeepCopy(benchMarks[benchID])
			engine.Unfold(&s, testUI)
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
			s := engine.DeepCopy(benchMarks[benchID])

			engine.Unfold(&s, testUI)
		}

		var sites = make([][]engine.ConfigSite, b.N)
		for n := 0; n < b.N; n++ {
			s := engine.DeepCopy(benchMarks[benchID])
			sites[n], _ = engine.Unfold(&s, testUI)
		}

		b.ResetTimer()
		for n := 0; n < b.N; n++ {
			for _, d := range sites[n] {
				_, err := engine.Gather(d, testUI)
				if err != nil {
					b.Fatal(err.Error())
				}
			}
		}
	}
}
