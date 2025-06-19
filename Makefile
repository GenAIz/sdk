SHELL=bash

BROKER_VERSION ?= "latest"
DOCKER_UID ?= $(shell id -u)
DOCKER_GID ?= $(shell stat -c "%g" /var/run/docker.sock)
REGISTRY ?= "localhost:5000"

.PHONY: compose
compose:
	ls -l /var/run/docker.sock >/dev/null 2>&1 || (>&2 echo "docker is not running" && exit 1)
	cd genaiz && $(MAKE) docker
	cd genaiz-it && $(MAKE) docker
	docker container prune --force
	docker image prune --force

.PHONY: mock_login_test
mock_login_test:
	@echo "Running mock account login profile test..."
	source .env && \
		rm -rf "$TESTS_MOUNT/mock-account-login" && \
		mkdir -p "$$TESTS_MOUNT/mock-account-login"
	DOCKER_UID=$(DOCKER_UID) DOCKER_GID=$(DOCKER_GID) docker compose --profile mock-account-login --profile mock-broker --profile report up

.PHONY: mock_build_simple_test
mock_build_simple_test:
	@echo "Running mock build simple profile test..."
	@echo "Using docker user and group: $(DOCKER_UID):$(DOCKER_GID)"
	source .env && \
		rm -rf "$$TESTS_MOUNT/mock-build-simple" && \
		mkdir -p "$$TESTS_MOUNT/mock-build-simple"
	DOCKER_UID=$(DOCKER_UID) DOCKER_GID=$(DOCKER_GID) docker compose --profile mock-build-simple --profile report up

.PHONY: mock_create_simple_test
mock_create_simple_test:
	@echo "Running mock create simple profile test..."
	source .env && \
		rm -rf "$$TESTS_MOUNT/mock-create-simple" && \
		mkdir -p "$$TESTS_MOUNT/mock-create-simple"
	docker compose --profile mock-create-simple --profile report up

.PHONY: mock_publish_simple_test
mock_publish_simple_test:
	@echo "Running mock publish simple profile test..."
	@echo "Using docker user and group: $(DOCKER_UID):$(DOCKER_GID)"
	source .env && \
		rm -rf "$$REGISTRY_MOUNT" && \
		mkdir -p "$$REGISTRY_MOUNT/auth" && \
		rm -rf "$$TESTS_MOUNT/mock-publish-simple" && \
		rm -rf "$$TESTS_MOUNT/.cache/genaiz" && \
		mkdir -p "$$TESTS_MOUNT/mock-publish-simple/" && \
		mkdir -p "$$TESTS_MOUNT/.cache/genaiz"
	DOCKER_UID=$(DOCKER_UID) DOCKER_GID=$(DOCKER_GID) docker compose --profile mock-publish-simple --profile mock-services --profile report up

.PHONY: mock_tests
mock_tests:
	@echo "Running isolated tests..."
	@echo "Using mocked services"
	source .env && mkdir -p "$$TESTS_MOUNT"

.PHONY: broker_tests
broker_tests:
	@echo "Running broker tests..."
	@echo "Using broker version XXX"
	@echo "Using mocked Docker registry"

.PHONY: registry_tests
registry_tests:
	@echo "Running registry tests..."
	@echo "Using broker version XXX"
	@echo "Using registry url XXX"