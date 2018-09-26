binary = antibuild
output = public/
templateargs = build
templaterefresh = develop

$(binary): ../../main.go 
	go build -o $(binary) ../../main.go 

build: $(binary)
	./$(binary) $(templateargs)

refresh: $(binary)
	./$(binary) $(templaterefresh)
	
clean:
	rm -rf $(output)
	rm $(binary)