SHELL=bash

.PHONY: help
## help - Show this help.
help:
	@sed -ne '/@sed/!s/## //p' $(MAKEFILE_LIST)

.PHONY: all
## all - Invokes all on genaiz, genaiz-it, genaiz-oauth and genaiz-lib
all: genaiz-lib/all genaiz-oauth/all genaiz/all genaiz-it/all

.PHONY: docker
## docker - Invokes docker builds on genaiz, genaiz-it and genaiz-oauth
docker: genaiz/docker genaiz-it/docker genaiz-oauth/docker

.PHONY: install
## install - Invokes install on genaiz, genaiz-it and genaiz-oauth
install: genaiz/install genaiz-it/install genaiz-oauth/install

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

.PHONY: genaiz-it/all
## genaiz-it/all - Invokes all on genaiz-it
genaiz-it/all:
	cd genaiz-it && make all

.PHONY: genaiz-it/docker
## genaiz-it/docker - Invokes docker build on genaiz-it
genaiz-it/docker:
	cd genaiz-it && make docker

.PHONY: genaiz-it/install
## genaiz-it/install - Invokes install on genaiz-it
genaiz-it/install:
	cd genaiz-it && make install

.PHONY: genaiz-lib/all
## genaiz-lib/all - Invokes all on genaiz-lib
genaiz-lib/all:
	cd genaiz-lib && make all

.PHONY: genaiz-oauth/all
## genaiz-oauth/all - Invokes all on genaiz-oauth
genaiz-oauth/all:
	cd genaiz-oauth && make all

.PHONY: genaiz-oauth/docker
## genaiz-oauth/docker - Invokes docker build on genaiz-oauth
genaiz-oauth/docker:
	cd genaiz-oauth && make docker

.PHONY: genaiz-oauth/install
## genaiz-oauth/install - Invokes install on genaiz-oauth
genaiz-oauth/install:
	cd genaiz-oauth && make install