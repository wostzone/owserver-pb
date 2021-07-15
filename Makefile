# Makefile to build the owserver protocol binding
DIST_FOLDER=./dist
.DEFAULT_GOAL := help

.PHONY: 

all: owserver-pb ## Build package with binary distribution and config


install:  ## Install the plugin into ~/bin/wost/bin and config
	mkdir -p ~/bin/wost/bin
	mkdir -p ~/bin/wost/config
	cp $(DIST_FOLDER)/bin/* ~/bin/wost/bin/
	cp -n $(DIST_FOLDER)/config/* ~/bin/wost/config/

test: .PHONY ## Run tests (todo fix this)
		go test -v ./...

clean: ## Clean distribution files
	go clean -cache
	go mod tidy
	rm -f $(DIST_FOLDER)/certs/*
	rm -f $(DIST_FOLDER)/logs/*
	rm -f $(DIST_FOLDER)/bin/*
	rm -f test/certs/*
	rm -f test/logs/*


owserver-pb: ## Build owserver-pb plugin 
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/owserver-pb/main.go
	@echo "> SUCCESS. Plugin '$@' can be found at $(DIST_FOLDER)/bin/$@"

help: ## Show this help
		@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

