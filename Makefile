NAME=metrics-capacitor
IMG_DEV=mc-dev
IMG_PROD=blufor/$(NAME)
LIB_PATH=github.com/metrics-capacitor/metrics-capacitor
VERSION=$(shell cat VERSION)
PATH=$(shell pwd -P)
BUILD=$(shell git rev-parse --short HEAD)
DOCKER=$(shell which docker)
LDFLAGS=--ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"
D_RUN=run --rm -h $(IMG_DEV) -v "$(PATH)/metrics-capacitor.go:/go/metrics-capacitor.go" -v "$(PATH)/bin:/go/bin" -v "$(PATH)/src:/go/src" -v "$(PATH)/pkg:/go/pkg" -v "$(PATH)/etc:/etc/metrics-capacitor"

all: prepare build
.PHONY: prepare build enter rmi clean push binary
.DEFAULT_GOAL: prepare build
prepare: .image.dev pkg
lib: pkg/linux_amd64/$(LIB_PATH).a
binary:	bin/$(NAME) .image
build: lib binary


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

bin/$(NAME): .image.dev pkg/linux_amd64/$(LIB_PATH).a
	@echo BUILDING SOURCE
	@echo "Version:\t$(VERSION)"
	@echo "Build:\t\t$(BUILD)\n"
	@$(DOCKER) $(D_RUN) $(IMG_DEV) bash -c 'cd /go && go build -v $(LDFLAGS) -o $@ /go/$(NAME).go'

pkg:
	@echo GETTING GO IMPORTS
	@$(DOCKER) $(D_RUN) $(IMG_DEV) bash -c 'cd /go && go get -v $(LIB_PATH)'

pkg/linux_amd64/$(LIB_PATH).a: pkg src/$(LIB_PATH)/*.go
	@$(DOCKER) $(D_RUN) $(IMG_DEV) bash -c 'cd /go && go install -v -a $(LIB_PATH)'

enter: .image.dev
	@echo ENTERING CONTAINER
	$(DOCKER) $(D_RUN) -it $(IMG_DEV)

rmi:
	@echo REMOVING IMAGE
	$(DOCKER) rmi $(IMG_DEV)
	rm -f .image.dev

clean: rmi
	@echo CLEANING
	rm -rf pkg bin/$(NAME)
