binary = antibuild
abargs = build


$(binary): $(shell find ../../ -name "*.go")
	go build -o $(binary) ../../main.go

build: $(binary)
	./$(binary) $(abargs)

clean:
	rm $(binary) $(binary).log