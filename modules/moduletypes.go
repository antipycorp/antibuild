// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"gitlab.com/antipy/antibuild/cli/builder/site"

	"github.com/jaicewizard/tt"
	"gitlab.com/antipy/antibuild/api/host"
	apiSite "gitlab.com/antipy/antibuild/api/site"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	"gitlab.com/antipy/antibuild/cli/modules/pipeline"
)

type (
	dataLoader struct {
		host    *host.ModuleHost
		command string
	}

	dataParser struct {
		host    *host.ModuleHost
		command string
	}

	dataPostProcessor struct {
		host    *host.ModuleHost
		command string
	}

	sitePostProcessor struct {
		host    *host.ModuleHost
		command string
	}

	templateFunction struct {
		host    *host.ModuleHost
		command string
	}

	iterator struct {
		host    *host.ModuleHost
		command string
	}
)

/*
	template function
*/
func getTemplateFunction(command string, host *host.ModuleHost) *templateFunction {
	return &templateFunction{
		host:    host,
		command: command,
	}
}

func (tf *templateFunction) Run(data ...interface{}) interface{} {
	output, err := tf.host.ExcecuteMethod("templateFunctions_"+tf.command, data)
	if err != nil {
		panic("execute methods: " + err.Error())
	}

	return output
}

/*
	iterators
*/
func getIterator(command string, host *host.ModuleHost) *iterator {
	return &iterator{
		host:    host,
		command: command,
	}
}

func (it *iterator) GetIterations(variable string) []string {
	output, err := it.host.ExcecuteMethod("iterators_"+it.command, []interface{}{variable})
	if err != nil {
		panic("get iterators: " + err.Error())
	}
	return output.([]string)
}

/*
	data loaders and post processors
	all of these are pipe-only.
*/
func getDataLoader(command string, host *host.ModuleHost) *dataLoader {
	return &dataLoader{
		host:    host,
		command: command,
	}
}

func (fl *dataLoader) Load(variable string) []byte {

	pipe := fl.GetPipe(variable)
	ret, _ := pipeline.ExecPipeline(nil, pipe)

	return ret
}

func (fl *dataLoader) GetPipe(variable string) pipeline.Pipe {
	pipe := func(binary []byte) ([]byte, errors.Error) {
		data, err := fl.host.ExcecuteMethod("dataLoaders_"+fl.command, []interface{}{variable})
		if err != nil {
			return nil, errors.Import(err)
		}
		if data == nil { //you cant convert nil to []byte, thus we need this check since it is 100% valid to have empty files
			return nil, nil
		}
		return data.([]byte), nil
	}
	return pipe
}

func getDataParser(command string, host *host.ModuleHost) *dataParser {
	return &dataParser{
		host:    host,
		command: command,
	}
}

func (fp *dataParser) Parse(data []byte, variable string) tt.Data {
	var ret tt.Data

	pipe := fp.GetPipe(variable)
	bytes, _ := pipeline.ExecPipeline(data, pipe)
	ret.GobDecode(bytes)

	return ret
}

func (fp *dataParser) GetPipe(variable string) pipeline.Pipe {
	pipe := func(binary []byte) ([]byte, errors.Error) {
		if binary == nil { // no daata should not be parsed and just return nil
			return nil, nil
		}
		data, err := fp.host.ExcecuteMethod("dataParsers_"+fp.command, []interface{}{variable}, binary...)
		if err != nil {
			return nil, errors.Import(err)
		}
		return data.([]byte), nil
	}
	return pipe
}

func getDataPostProcessor(command string, host *host.ModuleHost) *dataPostProcessor {
	return &dataPostProcessor{
		host:    host,
		command: command,
	}
}

func (dpp *dataPostProcessor) Process(data tt.Data, variable string) tt.Data {
	var ret tt.Data
	bin, _ := data.GobEncode()
	pipe := dpp.GetPipe(variable)
	bytes, _ := pipeline.ExecPipeline(bin, pipe)
	ret.GobDecode(bytes)

	return ret
}

func (dpp *dataPostProcessor) GetPipe(variable string) pipeline.Pipe {
	pipe := func(binary []byte) ([]byte, errors.Error) {
		data, err := dpp.host.ExcecuteMethod("dataPostProcessors_"+dpp.command, []interface{}{variable}, binary...)
		if err != nil {
			return nil, errors.Import(err)
		}
		return data.([]byte), nil
	}
	return pipe
}

func getSitePostProcessor(command string, host *host.ModuleHost) *sitePostProcessor {
	return &sitePostProcessor{
		host:    host,
		command: command,
	}
}

func (spp *sitePostProcessor) Process(data []*site.Site, variable string) []*site.Site {
	send := make([]apiSite.Site, 0, len(data))
	var recieve []apiSite.Site

	for _, d := range data {
		send = append(send, apiSite.Site{
			Slug:     d.Slug,
			Template: d.Template,
			Data:     d.Data,
		})
	}

	pipe := spp.GetPipe(variable)

	bytes, _ := pipeline.ExecPipeline(apiSite.Encode(send), pipe)
	recieve = apiSite.Decode(bytes)

	ret := make([]*site.Site, 0, len(recieve))

	for _, d := range recieve {
		ret = append(ret, &site.Site{
			Slug:     d.Slug,
			Template: d.Template,
			Data:     d.Data,
		})
	}

	return ret
}

func (spp *sitePostProcessor) GetPipe(variable string) pipeline.Pipe {
	pipe := func(binary []byte) ([]byte, errors.Error) {
		data, err := spp.host.ExcecuteMethod("sitePostProcessors_"+spp.command, []interface{}{variable}, binary...)
		if err != nil {
			return nil, errors.Import(err)
		}
		return data.([]byte), nil
	}
	return pipe
}
