binary = templateBuilder
output = public/
templateargs = --config config.json
templaterefresh = $(templateargs) --development --host

$(binary):  ../../main.go ../../site/site.go
	go build -o $(binary) ../../main.go

build: $(binary)
	./$(binary) $(templateargs)

refresh: $(binary)
	./$(binary) $(templaterefresh)
	
clean:
	rm -rf $(output)
	rm $(binary)