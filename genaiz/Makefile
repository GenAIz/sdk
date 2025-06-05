REGISTRY ?=

.PHONY: all
all: install

.PHONY: build
build: clean
	go build

.PHONY: clean
clean:
	go clean -testcache
	rm -f cover.out
	rm -f genaiz
	rm -rf ../out

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

.PHONY: recipes
recipes:
