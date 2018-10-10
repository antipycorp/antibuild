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
	cd ../../module/examples/abm_file && \
		go build -o abm_file
	mv ../../module/examples/abm_file/abm_file .modules 
	cd ../../module/examples/abm_json && \
		go build -o abm_json
	mv ../../module/examples/abm_json/abm_json .modules 
	cd ../../module/examples/abm_yaml && \
		go build -o abm_yaml
	mv ../../module/examples/abm_yaml/abm_yaml .modules 
	cd ../../module/examples/abm_firebase && \
		go build -o abm_firebase
	mv ../../module/examples/abm_firebase/abm_firebase .modules 
	cd ../../module/examples/abm_testspp && \
		go build -o abm_testspp
	mv ../../module/examples/abm_testspp/abm_testspp .modules 
	$(binary) $(templateargs)

refresh: $(binary)
	$(binary) $(templaterefresh)
	
clean:
	rm -rf $(output)
	rm $(binary)

FORCE: