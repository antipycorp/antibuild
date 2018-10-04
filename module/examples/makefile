#!make
GOBUILD=go build -o

build:
	mkdir -p output/windows/amd64
	mkdir -p output/darwin/amd64
	mkdir -p output/linux/amd64
	mkdir -p output/linux/arm
	mkdir -p output/linux/arm64
	@echo "\033[92mBuilding windows/amd64...\033[0m"
	cd $(PACKAGE) && \
		GOOS=windows GOARCH=amd64 $(GOBUILD) ../output/windows/amd64/$(PACKAGE)
	@echo "\033[92mBuilding darwin/amd64...\033[0m"
	cd $(PACKAGE) && \
		GOOS=darwin GOARCH=amd64 $(GOBUILD) ../output/darwin/amd64/$(PACKAGE)
	@echo "\033[92mBuilding linux/amd64...\033[0m"
	cd $(PACKAGE) && \
		GOOS=linux GOARCH=amd64 $(GOBUILD) ../output/linux/amd64/$(PACKAGE)
	@echo "\033[92mBuilding linux/arm...\033[0m"
	cd $(PACKAGE) && \
		GOOS=linux GOARCH=arm $(GOBUILD) ../output/linux/arm/$(PACKAGE)
	@echo "\033[92mBuilding linux/arm64...\033[0m"
	cd $(PACKAGE) && \
		GOOS=linux GOARCH=arm64 $(GOBUILD) ../output/linux/arm64/$(PACKAGE)
	@echo "\033[92mFinshed building.\033[0m"