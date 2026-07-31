SHELL=bash

.PHONY: help
## help - Show this help.
help:
	@sed -ne '/@sed/!s/## //p' $(MAKEFILE_LIST)

.PHONY: all
## all - Invokes all on genaiz-lib and genaiz
all: genaiz-lib/all genaiz/all

.PHONY: docker
## docker - Invokes docker build on genaiz
docker: genaiz/docker

.PHONY: install
## install - Invokes install on genaiz
install: genaiz/install

.PHONY:
## genaiz/dev - Installs development packaged binaries with supplementary trace tooling
genaiz/dev:
	cd genaiz && make dev

.PHONY: genaiz/all
## genaiz/all - Invokes all on genaiz
genaiz/all:
	cd genaiz && make all

.PHONY: genaiz/docker
## genaiz/docker - Invokes docker build on genaiz
genaiz/docker:
	cd genaiz && make docker

.PHONY: genaiz/install
## genaiz/install - Invokes install on genaiz
genaiz/install:
	cd genaiz && make install

.PHONY: genaiz-lib/all
## genaiz-lib/all - Invokes all on genaiz-lib
genaiz-lib/all:
	cd genaiz-lib && make all
