binary = $(shell pwd)/antibuild
output = public/
templateargs = build
templaterefresh = develop


$(binary): FORCE
	cd ../../ && \
		make $(binary) binary=$(binary)

setup: 
	mkdir -p .modules

build: $(binary) setup
	$(binary) $(templateargs)

refresh: $(binary) setup
	$(binary) $(templaterefresh)
	
clean:
	rm -rf $(output)
	rm $(binary)

FORCE: