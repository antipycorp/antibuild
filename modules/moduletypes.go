package modules

import (
	"encoding/json"

	"github.com/jaicewizard/tt"
	"gitlab.com/antipy/antibuild/api/host"
	"gitlab.com/antipy/antibuild/cli/builder/site"
	"gitlab.com/antipy/antibuild/cli/internal/errors"
	"gitlab.com/antipy/antibuild/cli/modules/pipeline"
)

type (
	fileLoader struct {
		host    *host.ModuleHost
		command string
	}

	fileParser struct {
		host    *host.ModuleHost
		command string
	}

	filePostProcessor struct {
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
)

/*
	templaate function
*/
func getTemplateFunction(command string, host *host.ModuleHost) *templateFunction {
	return &templateFunction{
		host:    host,
		command: command,
	}
}

func (tf *templateFunction) Load(data ...interface{}) []byte {
	output, err := tf.host.ExcecuteMethod("templateFunctions_"+tf.command, data)
	if err != nil {
		panic("execute methods: " + err.Error())
	}

	//check if return type is correct
	var outputFinal []byte
	var ok bool
	if outputFinal, ok = output.([]byte); ok != true {
		panic("fileLoader_" + tf.command + " did not return a []byte")
	}

	return outputFinal
}

/*
	file loaders and post processors
	all of these are pipe-only.
*/
func getFileLoader(command string, host *host.ModuleHost) *fileLoader {
	return &fileLoader{
		host:    host,
		command: command,
	}
}

func (fl *fileLoader) Load(variable string) []byte {
	var ret []byte

	pipe := fl.GetPipe(variable)
	pipeline.ExecPipeline(nil, &ret, pipe)

	return ret
}

func (fl *fileLoader) GetPipe(variable string) pipeline.Pipe {
	pipe := func(fileLoc string) errors.Error {
		_, err := fl.host.ExcecuteMethod("fileLoaders_"+fl.command, []interface{}{fileLoc, variable})
		if err != nil {
			return errors.Import(err)
		}
		return nil
	}
	return pipe
}

func getFileParser(command string, host *host.ModuleHost) *fileParser {
	return &fileParser{
		host:    host,
		command: command,
	}
}

func (fp *fileParser) Parse(data []byte, variable string) tt.Data {
	var ret tt.Data

	pipe := fp.GetPipe(variable)
	pipeline.ExecPipeline(data, &ret, pipe)

	return ret
}

func (fp *fileParser) GetPipe(variable string) pipeline.Pipe {
	pipe := func(fileLoc string) errors.Error {
		_, err := fp.host.ExcecuteMethod("fileParsers_"+fp.command, []interface{}{fileLoc, variable})
		if err != nil {
			return errors.Import(err)
		}
		return nil
	}
	return pipe
}

func getFilePostProcessor(command string, host *host.ModuleHost) *filePostProcessor {
	return &filePostProcessor{
		host:    host,
		command: command,
	}
}

func (fpp *filePostProcessor) Process(data tt.Data, variable string) tt.Data {
	var ret tt.Data

	pipe := fpp.GetPipe(variable)
	pipeline.ExecPipeline(data, &ret, pipe)

	return ret
}

func (fpp *filePostProcessor) GetPipe(variable string) pipeline.Pipe {
	pipe := func(fileLoc string) errors.Error {
		_, err := fpp.host.ExcecuteMethod("filePostProcessors_"+fpp.command, []interface{}{fileLoc, variable})
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
	var ret []*site.Site

	pipe := spp.GetPipe(variable)
	pipeline.ExecPipeline(data, &ret, pipe)

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

//UnmarshalJSON unmarschals the json into a moduleconfig
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
