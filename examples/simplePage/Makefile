binary = $(shell pwd)/antibuild
output = public/
templateargs = build
templaterefresh = develop


$(binary): FORCE
	cd ../../ && \
		make $(binary) binary=$(binary)

build: $(binary)
	cd ../../module/examples/abm_math && \
		go build -o abm_math
	mv ../../module/examples/abm_math/abm_math .modules 
	$(binary) $(templateargs)

refresh: $(binary)
	$(binary) $(templaterefresh)
	
clean:
	rm -rf $(output)
	rm $(binary)

FORCE: