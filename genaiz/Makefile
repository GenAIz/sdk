REGISTRY ?=

.PHONY: all
all: install

.PHONY: build
build: clean
	@echo "Building..."
	go build

.PHONY: clean
clean:
	@echo "Cleaning up..."
	go clean -testcache
	rm -f cover.out
	rm -f genaiz

.PHONY: docker
docker: build
	@echo "Building docker image..."
	./buildDocker.sh $(REGISTRY)

.PHONY: test
test:
	@echo "Testing..."
	# gcflags=-1 disables inlining to allow Monkey Patching
	go test -gcflags=-l ./...

.PHONY: coverage
coverage: clean
	@echo "Running coverage..."
	# gcflags=-1 disables inlining to allow Monkey Patching
	go test -gcflags=-l --coverprofile=cover.out ./...
	go tool cover -html=cover.out

.PHONY: install
install: clean
	@echo "Installing..."
	go install

.PHONY: recipes
recipes:
