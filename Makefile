NAME=metcap
IMG_DEV=mc-dev
IMG_PROD=blufor/$(NAME)
LIB_PATH=github.com/blufor/metcap
VERSION=$(shell cat VERSION)
PATH=$(shell pwd -P)
BUILD=$(shell git rev-parse --short HEAD)
DOCKER=$(shell which docker)
DOCKER_COMPOSE=$(shell which docker-compose)
ECHO=$(shell which echo)
RM=$(shell which rm)
TOUCH=$(shell which touch)
LDFLAGS=--ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"
D_RUN=run --rm -h $(IMG_DEV) --name $(IMG_DEV) --net host -v "$(PATH)/$(NAME).go:/go/$(NAME).go" -v "$(PATH)/bin:/usr/local/bin" -v "$(PATH)/src:/go/src/$(LIB_PATH)" -v "$(PATH)/pkg:/go/pkg" -v "$(PATH)/etc:/etc/$(NAME)"
# D_RUN=run --rm -h $(IMG_DEV) --name $(IMG_DEV) -v "$(PATH)/$(NAME).go:/go/$(NAME).go" -v "$(PATH)/bin:/go/bin" -v "$(PATH)/src:/go/src" -v "$(PATH)/pkg:/go/pkg" -v "$(PATH)/etc:/etc/$(NAME)"


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

.image.dev: Dockerfile.dev
	@$(ECHO) BUILDING DOCKER DEV IMAGE
	$(DOCKER) build -t $(IMG_DEV) - < Dockerfile.dev
	@$(TOUCH) $@

.image: bin/$(NAME) bin/$(NAME)-docker Dockerfile
	@$(ECHO) BUILDING DOCKER PROD IMAGE
	$(DOCKER) build -t $(IMG_PROD):$(VERSION) .
	$(DOCKER) tag $(IMG_PROD):$(VERSION) $(IMG_PROD):latest
	@$(TOUCH) $@

bin/$(NAME): .image.dev pkg/linux_amd64/$(LIB_PATH).a $(NAME).go VERSION
	@$(ECHO) -e "BUILDING BINARY"
	@$(ECHO) -e "Version:\t$(VERSION)"
	@$(ECHO) -e "Build:\t\t$(BUILD)"
	$(DOCKER) $(D_RUN) $(IMG_DEV) time go build -v $(LDFLAGS) -o /usr/local/$@ /go/$(NAME).go

pkg:
	@$(ECHO) GETTING GO IMPORTS
	$(DOCKER) $(D_RUN) $(IMG_DEV) go get -v /go/$(LIB_PATH)

pkg/linux_amd64/$(LIB_PATH).a: pkg $(shell find src -name '*.go')
	@$(ECHO) BUILDING LIBRARY
	$(DOCKER) $(D_RUN) $(IMG_DEV) time go fmt /go/src/$(LIB_PATH)
	$(DOCKER) $(D_RUN) $(IMG_DEV) time go vet /go/src/$(LIB_PATH)
	$(DOCKER) $(D_RUN) $(IMG_DEV) time go install -v -a /go/src/$(LIB_PATH)

.PHONY: test
test: bin/$(NAME)
	-$(DOCKER) $(D_RUN) -it $(IMG_DEV) $(NAME)

.PHONY: svc_start svc_stop svc_rm
svc_start:
	$(DOCKER_COMPOSE) create es redis rabbitmq
	$(DOCKER_COMPOSE) start es redis rabbitmq

svc_stop:
	$(DOCKER_COMPOSE) stop es redis rabbitmq

svc_rm: svc_stop
	$(DOCKER_COMPOSE) rm -f es redis rabbitmq


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
