# Makefile to build the owserver protocol binding
DIST_FOLDER=./dist
.DEFAULT_GOAL := help

.PHONY: 

all: owserver-pb ## Build package with binary distribution and config


install:  ## Install the plugin into ~/bin/wost/bin and config
	mkdir -p ~/bin/wost/bin
	mkdir -p ~/bin/wost/arm
	mkdir -p ~/bin/wost/config
	cp $(DIST_FOLDER)/bin/* ~/bin/wost/bin/
	cp $(DIST_FOLDER)/arm/* ~/bin/wost/arm/
	cp -n $(DIST_FOLDER)/config/* ~/bin/wost/config/

test: .PHONY ## Run tests (todo fix this)
		go test -v ./...

clean: ## Clean distribution files
	go clean -cache
	go mod tidy
	rm -f $(DIST_FOLDER)/certs/*
	rm -f $(DIST_FOLDER)/logs/*
	rm -f $(DIST_FOLDER)/bin/*
	rm -f $(DIST_FOLDER)/arm/*
	rm -f test/certs/*
	rm -f test/logs/*


owserver-pb: ## Build owserver-pb plugin for amd64 and arm64
	GOOS=linux GOARCH=amd64 go build -o $(DIST_FOLDER)/bin/$@ ./main.go
	GOOS=linux GOARCH=arm go build -o $(DIST_FOLDER)/arm/$@ ./main.go
	@echo "> SUCCESS. Plugin '$@' can be found at $(DIST_FOLDER)/bin/$@ and $(DIST_FOLDER)/arm/$@"

help: ## Show this help
		@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

