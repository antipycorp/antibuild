// Copyright © 2018-2019 Antipy V.O.F. info@antipy.com
//
// Licensed under the MIT License

package modules

import (
	"encoding/json"

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
	var ret []string

	pipe := it.GetPipe(variable)
	pipeline.ExecPipeline(nil, &ret, pipe)

	return ret
}

func (it *iterator) GetPipe(variable string) pipeline.Pipe {
	pipe := func(fileLoc string) errors.Error {
		_, err := it.host.ExcecuteMethod("iterators_"+it.command, []interface{}{fileLoc, variable})
		if err != nil {
			return errors.Import(err)
		}
		return nil
	}
	return pipe
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
	var ret []byte

	pipe := fl.GetPipe(variable)
	pipeline.ExecPipeline(nil, &ret, pipe)

	return ret
}

func (fl *dataLoader) GetPipe(variable string) pipeline.Pipe {
	pipe := func(fileLoc string) errors.Error {
		_, err := fl.host.ExcecuteMethod("dataLoaders_"+fl.command, []interface{}{fileLoc, variable})
		if err != nil {
			return errors.Import(err)
		}
		return nil
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
	pipeline.ExecPipeline(data, &ret, pipe)

	return ret
}

func (fp *dataParser) GetPipe(variable string) pipeline.Pipe {
	pipe := func(fileLoc string) errors.Error {
		_, err := fp.host.ExcecuteMethod("dataParsers_"+fp.command, []interface{}{fileLoc, variable})
		if err != nil {
			return errors.Import(err)
		}
		return nil
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

	pipe := dpp.GetPipe(variable)
	pipeline.ExecPipeline(data, &ret, pipe)

	return ret
}

func (dpp *dataPostProcessor) GetPipe(variable string) pipeline.Pipe {
	pipe := func(fileLoc string) errors.Error {
		_, err := dpp.host.ExcecuteMethod("dataPostProcessors_"+dpp.command, []interface{}{fileLoc, variable})
		if err != nil {
			return errors.Import(err)
		}
		return nil
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
	pipeline.ExecPipeline(send, &recieve, pipe)

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
	pipe := func(fileLoc string) errors.Error {
		_, err := spp.host.ExcecuteMethod("sitePostProcessors_"+spp.command, []interface{}{fileLoc, variable})
		if err != nil {
			return errors.Import(err)
		}
		return nil
	}
	return pipe
}

//UnmarshalJSON unmarshals the json into a module config
func (mc *ModuleConfig) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &mc.Config); err != nil {
		return err
	}
	return nil
}

//MarshalJSON marschals the data into json
func (mc *ModuleConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(mc.Config)
}
