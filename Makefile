# Makefile to build the owserver protocol binding
DIST_FOLDER=./dist
INSTALL_HOME=~/bin/wosthub
.DEFAULT_GOAL := help

.PHONY: 

all: owserver-pb ## Build package with binary distribution and config

install:  all ## Install the plugin into ~/bin/wost/bin and config
	mkdir -p $(INSTALL_HOME)/bin
	mkdir -p $(INSTALL_HOME)/config
	cp $(DIST_FOLDER)/bin/* $(INSTALL_HOME)/bin/
	cp -n $(DIST_FOLDER)/config/* $(INSTALL_HOME)/config/

test: .PHONY ## Run tests
	go test -race -failfast -p 1 -cover ./...

clean: ## Clean distribution files
	go clean -cache -testcache
	go mod tidy
	rm -f $(DIST_FOLDER)/certs/*
	rm -f $(DIST_FOLDER)/logs/*
	rm -f $(DIST_FOLDER)/bin/*
	rm -f test/certs/*
	rm -f test/logs/*


owserver-pb: ## Build owserver-pb plugin 
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/owserver-pb/main.go
	@echo "> SUCCESS. Plugin '$@' can be found at $(DIST_FOLDER)/bin/$@"

upgrade: ## Upgrade packages (use with care)
	go get -u all
	go mod tidy

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
