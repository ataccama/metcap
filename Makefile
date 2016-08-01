NAME=metrics-capacitor
IMG_DEV=mc-dev
IMG_PROD=blufor/$(NAME)
LIB_PATH=github.com/metrics-capacitor/metrics-capacitor
VERSION=$(shell cat VERSION)
BUILD=$(shell git rev-parse --short HEAD)
DOCKER=$(shell which docker)
LDFLAGS=--ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"
D_RUN=run --rm -h $(IMG) -v "${PWD}/src:/go/src" -v "${PWD}/pkg:/go/pkg" -v "$(PWD)/etc:/etc/metrics-capacitor"

all: prepare build
.PHONY: prepare build enter rmi clean push
.DEFAULT_GOAL: prepare build
prepare: .image.dev pkg bin
build:	bin/$(NAME) .image

.image.dev:
	@echo BUILDING DOCKER DEV IMAGE
	$(DOCKER) build -t $(IMG_DEV) - < Dockerfile.dev
	@touch $@

.image:
	@echo BUILDING DOCKER PROD IMAGE
	$(DOCKER) build -t $(IMG_PROD):$(VERSION) - < Dockerfile
	$(DOCKER) tag $(IMG_PROD):$(VERSION) $(IMG_PROD):latest
	@touch $@

push:
	$(DOCKER) push $(IMG_PROD):$(VERSION)

bin:
	mkdir -p $@

bin/$(NAME): bin
	@echo BUILDING SOURCE
	@echo "Version:\t$(VERSION)"
	@echo "Build:\t\t$(BUILD)\n"
	$(DOCKER) $(D_RUN) $(IMG_DEV) bash -c 'cd /go && go build -v $(LDFLAGS) -o $@ $(LIB_PATH)'

pkg:
	@echo GETTING GO IMPORTS
	@$(DOCKER) $(D_RUN) $(IMG_DEV) bash -c 'cd /go && go get -v $(LIB_PATH)'

enter: .image
	@echo ENTERING CONTAINER
	$(DOCKER) $(D_RUN) -it $(IMG_DEV)

rmi:
	@echo REMOVING IMAGE
	$(DOCKER) rmi $(IMG)
	rm -f .image

clean: rmi
	@echo CLEANING
	rm -rf pkg bin/$(NAME)
