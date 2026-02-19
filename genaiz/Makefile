REGISTRY ?=

.PHONY: all
all: clean report install

.PHONY: help
## help - Show this help.
help:
	@sed -ne '/@sed/!s/## //p' $(MAKEFILE_LIST)

.PHONY: build
## build - Produces the main binary genaiz in the source folder
build: clean
	@echo "Building..."
	go build -tags prod

.PHONY: clean
## clean - Removes the built binaries and coverage reports, and resets the test cache of the go compiler
clean:
	@echo "Cleaning up..."
	go clean -testcache
	rm -f cover.out
	rm -f genaiz

.PHONY: docker
## docker - Produces a docker image following the buildDocker.sh script
docker: build
	@echo "Building docker image..."
	./buildDocker.sh $(REGISTRY)

.PHONY: test
## test - Invokes test on all packages with inlining disabled
test:
	@echo "Testing..."
	# gcflags=-1 disables inlining to allow Monkey Patching
	go test -gcflags=-l -tags prod ./...

.PHONY: coverage
## coverage - Invokes the go coverage tool to produce an html report of the cover.out file
coverage: report
	@echo "Opening coverage..."
	go tool cover -html=cover.out

.PHONY: dev
## dev - Installs development based binaries with supplementary trace tooling
dev:
	@echo "Installing development binaries..."
	go install -tags dev

.PHONY: install
## install - Installs the binary under the users bin directory ($HOME/go/bin)
install: build
	@echo "Installing..."
	go install -tags prod

.PHONY: recipes
## recipes - Should install embedded recipes and recipe packs under the users share folder ($HOME/.local/share/genaiz)
recipes:

.PHONY: report
## report - produces a coverage report invoking the go test tool, produced as cover.out in source folder
report:
	@echo "Running coverage, report only..."
	# gcflags=-1 disables inlining to allow Monkey Patching
	go test -tags prod -gcflags=-l --coverprofile=cover.out ./...
	@echo "Removing test executable"
	rm -f genaiz
