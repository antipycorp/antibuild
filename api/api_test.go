package api_test

import (
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/jaicewizard/tt"

	abm "gitlab.com/antipy/antibuild/cli/api/client"
	"gitlab.com/antipy/antibuild/cli/api/file"
	"gitlab.com/antipy/antibuild/cli/api/host"
	"gitlab.com/antipy/antibuild/cli/builder/site"
)

func TestKill(t *testing.T) {
	hin, cout := io.Pipe()
	cin, hout := io.Pipe()

	complete := make(chan bool)
	go func() {
		module := abm.Register("TEST")
		ret := make(chan int)
		go startwrap(module, cin, cout, ret)
		select {
		case <-ret:
			complete <- true
		case <-time.After(1 * time.Second):
			complete <- false
		}
	}()

	module, err := host.Start(hin, hout, nil)
	if err != nil {
		t.Fail()
	}
	module.Kill()

	if <-complete == false {
		t.Fail()
	}
}

func TestError(t *testing.T) {
	complete := make(chan bool)

	hin, cout := io.Pipe()
	cin, hout := io.Pipe()

	testTMP := func(w abm.TFRequest, r abm.Response) {
		r.AddFatal("test error")
	}

	go func() {
		module := abm.Register("TEST")
		module.TemplateFunctionRegister("TESTFUNC", testTMP, &abm.TFTest{
			Request: abm.TFRequest{Data: []interface{}{
				1,
				2,
			}}, Response: &abm.TFResponse{
				Data: 3,
			},
		})

		module.CustomStart(cin, cout)
	}()

	go func() {
		module, err := host.Start(hin, hout, nil)
		if err != nil {
			t.Fail()
		}
		ret := make(chan bool)

		go func() {
			_, err = module.ExcecuteMethod("templateFunctions_TESTFUNC", nil)
			ret <- err != nil //this should return an error.

		}()
		select {
		case res := <-ret:
			complete <- res
		case <-time.After(1 * time.Second):
			complete <- false
		}
	}()

	if <-complete == false {
		t.Fail()
	}
}

func TestFileLoad(t *testing.T) {
	complete := make(chan bool)

	hin, cout := io.Pipe()
	cin, hout := io.Pipe()

	loadFile := func(w abm.FLRequest, r abm.Response) {
		if w.Variable == "" {
			fmt.Println("invalid input")
			r.AddError(abm.InvalidInput)
			return
		}
		r.AddData([]byte("{\"testdata\":\"hey\"}"))
	}

	parseFile := func(w abm.FPRequest, r abm.Response) {
		if w.Data == nil {
			r.AddError(abm.InvalidInput)
			return
		}

		var jsonData map[string]interface{}
		err := json.Unmarshal(w.Data, &jsonData)
		if err != nil {
			r.AddFatal(err.Error())
			return
		}
		var retData = make(map[interface{}]interface{})
		for k, v := range jsonData {
			retData[k] = v
		}
		r.AddData(retData)
	}

	processFile := func(w abm.FPPRequest, r abm.Response) {
		if w.Data == nil {
			r.AddError(abm.InvalidInput)
			return
		}

		r.AddData(w.Data)
	}

	resultCheck := map[interface{}]interface{}{
		"testdata": "hey",
	}

	go func() {
		module := abm.Register("TEST")

		module.DataLoaderRegister("file", loadFile)
		module.DataParserRegister("load", parseFile)
		module.DataPostProcessorRegister("process", processFile)

		module.CustomStart(cin, cout)
	}()

	go func() {
		module, err := host.Start(hin, hout, nil)
		if err != nil {
			t.Fail()
		}

		file, err := file.NewFile([]byte(""))
		if err != nil {
			complete <- false
		}

		_, err = module.ExcecuteMethod("dataLoaders_file", []interface{}{file.GetRef(), "TEST"})
		if err != nil {
			complete <- false
		}

		_, err = module.ExcecuteMethod("dataParsers_load", []interface{}{file.GetRef(), "n0thing"})
		if err != nil {
			complete <- false
		}

		_, err = module.ExcecuteMethod("dataPostProcessors_process", []interface{}{file.GetRef(), "n0thing"})

		var vret tt.Data

		err = file.Retreive(&vret)

		for k, v := range vret {
			if v != resultCheck[k] {
				complete <- false
			}
		}

		complete <- err == nil
	}()

	select {
	case succes := <-complete:
		if !succes {
			t.Fail()
		}
		// case <-time.After(1 * time.Second):
		// 	t.Fail()
	}
}

func TestSiteParse(t *testing.T) {
	complete := make(chan bool)

	hin, cout := io.Pipe()
	cin, hout := io.Pipe()

	resultCheck := []*site.Site{
		&site.Site{
			Data: map[string]interface{}{
				"data": "hey",
				"new":  "yes",
			},
		},
	}
	dataSend := []*site.Site{
		&site.Site{
			Data: map[string]interface{}{
				"data": "hey",
			},
		},
	}

	languageProcess := func(w abm.SPPRequest, r abm.Response) {
		dat := w.Data
		dat[0].Data["new"] = "yes"
		r.AddData(dat)
	}

	go func() {
		module := abm.Register("TEST")

		module.SitePostProcessorRegister("test", languageProcess)

		module.CustomStart(cin, cout)
	}()

	go func() {
		module, err := host.Start(hin, hout, nil)
		if err != nil {
			t.Fail()
		}

		file, err := file.NewFile(dataSend)
		if err != nil {
			complete <- false
		}

		_, err = module.ExcecuteMethod("sitePostProcessors_test", []interface{}{file.GetRef(), "n0thing"})
		if err != nil {
			complete <- false
		}
		var ret []*site.Site

		err = file.Retreive(&ret)

		if err != nil {
			complete <- false
		}

		for k, v := range resultCheck {
			if v.Slug != ret[k].Slug || v.Template != ret[k].Template {
				complete <- false
			}
			for kd, d := range v.Data {
				if d != ret[k].Data[kd] {
					complete <- false
				}
			}
		}
		complete <- true
	}()

	select {
	case succes := <-complete:
		if !succes {
			t.Fail()
		}
	case <-time.After(1 * time.Second):
		t.Fail()
	}
}

func startwrap(m *abm.Module, in io.Reader, out io.Writer, ret chan int) {
	m.CustomStart(in, out)
	ret <- 1
}
