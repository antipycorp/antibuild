builderDir := builder
cliDir := cli
binary := antibuild
outbinary := $(binary)
debugflags = -gcflags=all="-N -l" 

$(binary): $(shell find . -name '*.go' -type f)
	go build -o $(outbinary)

clean:
	rm $(binary)