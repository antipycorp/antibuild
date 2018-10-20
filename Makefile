builderDir := builder
cliDir := cli
binary := antibuild
outbinary := $(binary)
debugflags = -gcflags=all="-N -l" 
filetags = -tags 'module file'
firebasetags = -tags 'module firebase'
jsontags = -tags 'module json'

$(binary): $(shell find . -name '*.go' -type f)
	go build -o $(outbinary)

file: $(shell find . -name '*.go' -type f)
	go build -o $(outbinary) $(filetags)

firebase: $(shell find . -name '*.go' -type f)
	go build -o $(outbinary) $(firebasetags)

json: $(shell find . -name '*.go' -type f)
	go build -o $(outbinary) $(jsontags)