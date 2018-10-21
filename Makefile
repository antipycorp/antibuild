builderDir := builder
cliDir := cli
binary := antibuild
outbinary := $(binary)
debugflags = -gcflags=all="-N -l" 
filetags = -tags 'module file'
jsontags = -tags 'module json'
languagetags = -tags 'module language'
noescapetags = -tags 'module noescape'

$(binary): $(shell find . -name '*.go' -type f)
	go build -o $(outbinary)

file: $(shell find . -name '*.go' -type f)
	go build -o $(outbinary) $(filetags)

json: $(shell find . -name '*.go' -type f)
	go build -o $(outbinary) $(jsontags)

language: $(shell find . -name '*.go' -type f)
	go build -o $(outbinary) $(languagetags)

noescape: $(shell find . -name '*.go' -type f)
	go build -o $(outbinary) $(noescapetags)