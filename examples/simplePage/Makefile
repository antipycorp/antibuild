binary = $(shell pwd)/antibuild
output = public/
templateargs = build
templaterefresh = develop


$(binary): FORCE
	cd ../../ && \
		make $(binary) binary=$(binary)

build: $(binary)
	cd ../../module/examples/abm_arithmetic && \
		go build -o abm_arithmetic
	mv ../../module/examples/abm_arithmetic/abm_arithmetic .modules 
	$(binary) $(templateargs)

refresh: $(binary)
	$(binary) $(templaterefresh)
	
clean:
	rm -rf $(output)
	rm $(binary)

FORCE: