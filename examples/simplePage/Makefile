binary = $(shell pwd)/antibuild
output = public/
templateargs = build
templaterefresh = develop


$(binary): FORCE
	cd ../../ && \
		make $(binary) binary=$(binary)

build: $(binary)
<<<<<<< HEAD
	cd ../../module/examples/abm_arithmetic && \
		go build -o abm_arithmetic
	mv ../../module/examples/abm_arithmetic/abm_arithmetic .modules 
=======
	cd ../../module/examples/abm_math && \
		go build -o abm_math
	mv ../../module/examples/abm_math/abm_math .modules 
	cd ../../module/examples/abm_file && \
		go build -o abm_file
	mv ../../module/examples/abm_file/abm_file .modules 
	cd ../../module/examples/abm_json && \
		go build -o abm_json
	mv ../../module/examples/abm_json/abm_json .modules 
>>>>>>> 3772f17e3634776f6360e41104e8fb7214e47ba4
	$(binary) $(templateargs)

refresh: $(binary)
	$(binary) $(templaterefresh)
	
clean:
	rm -rf $(output)
	rm $(binary)

FORCE: