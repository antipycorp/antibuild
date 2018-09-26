binary = $(shell pwd)/antibuild
output = public/
templateargs = build
templaterefresh = develop


$(binary): FORCE
	cd ../../ && \
		make $(binary) binary=$(binary)

build: $(binary)
	$(binary) $(templateargs)

refresh: $(binary)
	$(binary) $(templaterefresh)
	
clean:
	rm -rf $(output)
	rm $(binary)

FORCE: