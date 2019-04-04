builderDir := builder
cliDir := cli
binary := antibuild
outbinary := $(binary)
debugflags = -gcflags=all="-N -l" 

$(binary): $(shell find . -name '*.go' -type f)
	go build -o $(outbinary)

clean:
	rm $(binary)

build: build_darwin build_linux build_windows

build_darwin:
	export GOOS=darwin; \
		make build_amd64;

build_linux:
	export GOOS=linux; \
		make build_amd64;  \
		make build_386;  \
		make build_arm64;  \
		make build_arm;

build_windows:
	export GOOS=windows; \
		make build_amd64; \
		make build_386;

build_amd64:
	export GOARCH=amd64; \
		make build_internal;

build_386:
	export GOARCH=386; \
		make build_internal;

build_arm64:
	export GOARCH=amd64; \
		make build_internal;

build_arm:
	export GOARCH=amd64; \
		make build_internal;

build_internal:
	echo "Building antibuild for ${GOOS}/${GOARCH}";
	go build -o ./dist/${GOOS}/${GOARCH}/antibuild main.go

bin:
	go build -o antibuild main.go
	mv antibuild ~/bin

test:
	go test ./...	

bench:
	go test ./... -run=xxx -bench=. > newBench.txt
	#wget https://gitlab.com/antipy/antibuild/cli/-/jobs/artifacts/iterators/raw/bench.txt?job=bench
	#benchcmp bench.txt newBench.txt > benchcmp
	mv newBench.txt bench.txt
