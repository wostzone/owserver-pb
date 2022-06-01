# Makefile to build the owserver protocol binding
DIST_FOLDER=./dist
INSTALL_HOME=~/bin/wosthub
.DEFAULT_GOAL := help

all: owserver ## Build package with binary distribution and config

owserver: ## Build owserver plugin
	go build -o $(DIST_FOLDER)/bin/$@ ./cmd/owserver/main.go
	@echo "> SUCCESS. Plugin '$@' can be found at $(DIST_FOLDER)/bin/$@"

clean: ## Clean distribution files
	go clean -cache -testcache
	go mod tidy
	rm -f $(DIST_FOLDER)/bin/*

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install:  all ## Install the plugin into $(INSTALL_HOME)
	mkdir -p $(INSTALL_HOME)/bin
	mkdir -p $(INSTALL_HOME)/config
	cp $(DIST_FOLDER)/bin/* $(INSTALL_HOME)/bin/
	cp -n $(DIST_FOLDER)/config/* $(INSTALL_HOME)/config/

test: ## Run tests
	go test -race -failfast -p 1 -cover ./...

upgrade: ## Upgrade packages (use with care)
	go get -u all
	go mod tidy
