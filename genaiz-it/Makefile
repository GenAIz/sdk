REGISTRY ?=

.PHONY: build
build: clean
	@echo "Building..."
	go build

.PHONY: clean
clean:
	@echo "Cleaning..."
	go clean -testcache
	rm -f cover.out
	rm -f genaiz

.PHONY: docker
docker: build
	@echo "Building docker image..."
	./buildDocker.sh $(REGISTRY)
