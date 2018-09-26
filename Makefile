builderDir := builder
cliDir := cli
binary := antibuild
outbinary := $(binary)

$(binary): $(shell find . -name '*.go' -type f)
	go build -o $(outbinary) main.go 
