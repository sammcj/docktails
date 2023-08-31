# HELP
# This will output the help for each task
# thanks to https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help

help: ## This help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

dev:
	go run docktails.go

build:
	go build docktails.go
	ls -la docktails

build-linux:
	GOOS=linux GOARCH=amd64 go build -o docktails docktails.go
	ls -la docktails

build-mac:
	GOOS=darwin GOARCH=amd64 go build -o docktails docktails.go
	ls -la docktails

install:
	make build
	cp docktails /usr/local/bin/docktails
	echo "installed to /usr/local/bin/docktails"
