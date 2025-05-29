REGISTRY ?=

.PHONY: all
all: install

.PHONY: build
build:
	go clean -testcache
	go build

.PHONY: docker
docker: build
	./buildDocker.sh $(REGISTRY)

.PHONY: test
test: build
	go test ./...

.PHONY: coverage
coverage:
	go test --coverprofile=cover.out ./...
	go tool cover -html=cover.out

.PHONY: install
install: test
	go install