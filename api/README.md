# AntiBuild module API

The API used for creating, and interfacing, with modules written for antibuild

## How to use

All antibuild modules have to import the client
(gitlab.com/antipy/antibuild/cli/api/client), in the examples we import this as
abm(antibuild module). The module has to register a module:

```golang
module := abm.Register("module name")
```

This returns a \*Module. using this module you can register functions. For
example for a file parser(json,yaml, etc) this would be:

```golang
module.DataParserRegister("function name", DataParser)
```

In this case DataParser can be any DataParser.

For some function types it is required to provide a test for a funtion. The
template functions is one example of this. An example from the math module:

```golang
module.TemplateFunctionRegister("add", add, &abm.TFTest{
  Request: abm.TFRequest{
    Data: []interface{}{
      1,
      2,
    }
  },
  Response: &abm.TFResponse{
    Data: 3,
  },
})
```

At the end of your main function you should call module.Start() in order to
start listening for incomming commands.

The functions should Receive the corresponding type of request, and a Response
type. The request will have the nessecary information to process the request. If
you receive some kind of interface you should always check the type, there are
no type guarantees about the underlying types of interface{} You can return the
data using r.AddData(data), this will check if its the correct type, and return
false if it is not the correct type. You can add errors to return log, any error
will result in a failure and results could be disregarded. Only the first error
will be returned to the caller. There are 2 types of errors, r.AddError(message)
is used for returning an invalid input message, r.AddFatal(message) is for when
there is a error occured. You can also use R.AddInfo(message) to add a debug
message to the log, this should be avoided in production, but could, for
example, be enabled in a the module config. These might be printed out on the
host, depending on the AntiBuild config.

An example of a complete module would be:(subset of the math module)

```golang
package main

import (
	abm "gitlab.com/antipy/antibuild/cli/api/client"
)

func main() {
	module := abm.Register("math")

	module.TemplateFunctionRegister("add", add, &abm.TFTest{
		Request: abm.TFRequest{Data: []interface{}{
			1,
			2,
		}}, Response: &abm.TFResponse{
			Data: 3,
		},
	})

	module.Start()
}

func add(w abm.TFRequest, r abm.Response) {
	var args = make([]int, len(w.Data))
	var ok bool

	for i, data := range w.Data {
		if args[i], ok = data.(int); !ok {
			r.AddError(abm.InvalidInput)
			return
		}
	}

	result := args[0] + args[1]

	r.AddData(result)
	return
}
```

For examples of how to register a Config function check out the main repo, under
/modules/language
