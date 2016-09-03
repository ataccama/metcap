NAME=metrics-capacitor
IMG_DEV=mc-dev
IMG_PROD=blufor/$(NAME)
LIB_PATH=github.com/metrics-capacitor/metrics-capacitor
VERSION=$(shell cat VERSION)
PATH=$(shell pwd -P)
BUILD=$(shell git rev-parse --short HEAD)
DOCKER=$(shell which docker)
DOCKER_COMPOSE=$(shell which docker-compose)
ECHO=$(shell which echo)
RM=$(shell which rm)
TOUCH=$(shell which touch)
LDFLAGS=--ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"
D_RUN=run --rm -h $(IMG_DEV) --name $(IMG_DEV) --net host -v "$(PATH)/metrics-capacitor.go:/go/metrics-capacitor.go" -v "$(PATH)/bin:/go/bin" -v "$(PATH)/src:/go/src" -v "$(PATH)/pkg:/go/pkg" -v "$(PATH)/etc:/etc/metrics-capacitor"
# D_RUN=run --rm -h $(IMG_DEV) --name $(IMG_DEV) -v "$(PATH)/metrics-capacitor.go:/go/metrics-capacitor.go" -v "$(PATH)/bin:/go/bin" -v "$(PATH)/src:/go/src" -v "$(PATH)/pkg:/go/pkg" -v "$(PATH)/etc:/etc/metrics-capacitor"


.DEFAULT_GOAL := binary
.PHONY: default
default: binary

.PHONY: prepare
prepare: .image.dev pkg

.PHONY: lib
lib: pkg/linux_amd64/$(LIB_PATH).a

.PHONY: binary
binary:	bin/$(NAME)

.PHONY: build
build: lib binary

.PHONY: compose
compose: lib binary
	$(DOCKER_COMPOSE) up --abort-on-container-exit --force-recreate --remove-orphans --build

.image.dev:
	@$(ECHO) BUILDING DOCKER DEV IMAGE
	$(DOCKER) build -t $(IMG_DEV) - < Dockerfile.dev
	@$(TOUCH) $@

.image: bin/$(NAME) bin/$(NAME)-docker
	@$(ECHO) BUILDING DOCKER PROD IMAGE
	$(DOCKER) build -t $(IMG_PROD):$(VERSION) .
	$(DOCKER) tag $(IMG_PROD):$(VERSION) $(IMG_PROD):latest
	@$(TOUCH) $@

bin/$(NAME): .image.dev pkg/linux_amd64/$(LIB_PATH).a $(NAME).go VERSION
	@$(ECHO) -e "\nBUILDING SOURCE"
	@$(ECHO) -e "Version:\t$(VERSION)"
	@$(ECHO) -e "Build:\t\t$(BUILD)\n"
	$(DOCKER) $(D_RUN) $(IMG_DEV) bash -c 'cd /go && time go build -v $(LDFLAGS) -o $@ /go/$(NAME).go'

pkg:
	@$(ECHO) GETTING GO IMPORTS
	$(DOCKER) $(D_RUN) $(IMG_DEV) bash -c 'cd /go && go get -v $(LIB_PATH)'

sources := $(shell find src/$(LIB_PATH) -name '*.go')
pkg/linux_amd64/$(LIB_PATH).a: pkg $(sources)
	$(DOCKER) $(D_RUN) $(IMG_DEV) bash -c 'cd /go && time go fmt $(LIB_PATH)'
	$(DOCKER) $(D_RUN) $(IMG_DEV) bash -c 'cd /go && time go install -v -a $(LIB_PATH)'

.PHONY: test
test: bin/$(NAME)
	-$(DOCKER) $(D_RUN) -it $(IMG_DEV) /go/bin/metrics-capacitor

.PHONY: push
push:
	$(DOCKER) push $(IMG_PROD):$(VERSION)
	$(DOCKER) push $(IMG_PROD):latest

.PHONY: enter
enter: .image.dev
	@$(ECHO) ENTERING CONTAINER
	$(DOCKER) $(D_RUN) -it $(IMG_DEV)

.PHONY: rmi
rmi:
	@$(ECHO) REMOVING IMAGE
	$(DOCKER) rmi $(IMG_DEV)
	$(RM) -f .image.dev

.PHONY: clean
clean: rmi
	@$(ECHO) CLEANING
	$(RM) -rf pkg bin/$(NAME)
