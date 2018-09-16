binary = templateBuilder
output = public/
templateargs = --templates templates/ --json json/ --out $(output)
templaterefresh = $(templateargs) --development

$(binary): ../main.go
	go build -o $(binary) ../main.go 

build: $(binary)
	./$(binary) $(templateargs)

refresh: $(binary)
	./$(binary) $(templaterefresh)
clean:
	rm -rf $(output)
	rm $(binary)