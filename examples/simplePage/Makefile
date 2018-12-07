binary = $(shell pwd)/antibuild
output = public/
templateargs = build
templaterefresh = develop

modulebuild = go build
examplessource = gitlab.com/antipy/antibuild/api/examples
modulesfolder = .modules

yaml = /abm_yaml
yamlargs = -o $(modulesfolder)$(yaml) $(examplessource)$(yaml)

firebase = /abm_firebase
firebaseargs = -o $(modulesfolder)$(firebase) $(examplessource)$(firebase)

math = /abm_math
mathargs = -o $(modulesfolder)$(math) $(examplessource)$(math)

$(binary): FORCE
	cd ../../ && \
		make $(binary) binary=$(binary)

setup: 
	mkdir -p .modules
#	$(modulebuild) $(yamlargs)
#	$(modulebuild) $(firebaseargs)
#	$(modulebuild) $(mathargs)

build: $(binary) setup
	$(binary) $(templateargs)

refresh: $(binary) setup
	$(binary) $(templaterefresh)

stress: export STRESS = 1
stress: $(binary) setup 	
	$(binary) $(templaterefresh)

debug: $(binary) setup
	dlv debug gitlab.com/antipy/antibuild/cli -- build

clean:
	rm -rf $(output)
	rm $(binary)

FORCE:

