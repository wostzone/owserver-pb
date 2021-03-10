DIST_FOLDER=./dist



all: FORCE ## Build package with binary distribution and config
all: owserver

owserver:
	GOOS=linux GOARCH=amd64 go build -o $(DIST_FOLDER)/bin/$@ ./main.go
	GOOS=linux GOARCH=arm go build -o $(DIST_FOLDER)/arm/$@ ./main.go
	@echo "> SUCCESS. Plugin '$@' can be found at $(DIST_FOLDER)/bin/$@ and $(DIST_FOLDER)/arm/$@"

clean: ## Clean distribution files
	$(GOCLEAN)
	rm -f $(DIST_FOLDER)/certs/*
	rm -f $(DIST_FOLDER)/logs/*
	rm -f $(DIST_FOLDER)/bin/*
	rm -f $(DIST_FOLDER)/arm/*

FORCE:
