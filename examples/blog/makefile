binary = antibuild
abargs = build


$(binary): $(shell find ../../ -name "*.go")
	go build -o $(binary) ../../main.go

build: $(binary)
	./$(binary) $(abargs)

debug: export DEBUG = 1
debug: export GODEBUG = gctrace=1
debug: $(binary)
<<<<<<< Updated upstream
	./$(binary) build
=======
	./$(binary) develop
>>>>>>> Stashed changes

develop:
	./$(binary) develop

clean:
	rm $(binary) $(binary).log
