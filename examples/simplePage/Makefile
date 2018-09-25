binary = antibuild
output = public/
templateargs = build
templaterefresh = develop

$(binary): ../../main.go 
	go build -o $(binary) ../../main.go 

build: clean $(binary)
	./$(binary) $(templateargs)

refresh: clean $(binary)
	./$(binary) $(templaterefresh)
	
clean:
	rm -rf $(output)
	rm $(binary)