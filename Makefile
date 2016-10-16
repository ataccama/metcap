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
SUDO=$(shell which sudo)
LDFLAGS=--ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"
D_RUN=run --rm -h $(IMG_DEV) --name $(IMG_DEV) --net host -v "$(PATH)/$(NAME).go:/go/$(NAME).go" -v "$(PATH)/bin:/usr/local/bin" -v "$(PATH)/src:/go/src/$(LIB_PATH)" -v "$(PATH)/pkg:/go/pkg" -v "$(PATH)/etc:/etc/$(NAME)"
# D_RUN=run --rm -h $(IMG_DEV) --name $(IMG_DEV) -v "$(PATH)/$(NAME).go:/go/$(NAME).go" -v "$(PATH)/bin:/go/bin" -v "$(PATH)/src:/go/src" -v "$(PATH)/pkg:/go/pkg" -v "$(PATH)/etc:/etc/$(NAME)"


.DEFAULT_GOAL := default
.PHONY: default
default: binary .image

.PHONY: prepare
prepare: .image.dev pkg

.PHONY: lib
lib: pkg/linux_amd64/$(LIB_PATH).a

.PHONY: binary
binary:	bin/$(NAME)

.PHONY: build
build: lib binary

.image.dev: Dockerfile.dev
	@$(ECHO) == BUILDING DOCKER DEV IMAGE
	$(DOCKER) build -t $(IMG_DEV) - < Dockerfile.dev
	@$(TOUCH) $@
	@$(ECHO)

.image: bin/$(NAME) bin/$(NAME)-docker Dockerfile
	@$(ECHO) == BUILDING DOCKER PROD IMAGE
	$(DOCKER) build -t $(IMG_PROD):$(VERSION) .
	$(DOCKER) tag $(IMG_PROD):$(VERSION) $(IMG_PROD):latest
	@$(TOUCH) $@
	@$(ECHO)

bin/$(NAME): .image.dev pkg/linux_amd64/$(LIB_PATH).a $(NAME).go VERSION
	@$(ECHO) -e "BUILDING BINARY"
	@$(ECHO) -e "Version:\t$(VERSION)"
	@$(ECHO) -e "Build:\t\t$(BUILD)"
	$(DOCKER) $(D_RUN) $(IMG_DEV) time go build -v $(LDFLAGS) -o /usr/local/$@ /go/$(NAME).go
	@$(ECHO)

pkg/linux_amd64/$(LIB_PATH).a: $(shell find src -name '*.go')
	@$(ECHO) == FORMATTING
	$(DOCKER) $(D_RUN) $(IMG_DEV) time go fmt $(LIB_PATH)
	@$(ECHO)
	@$(ECHO) == VETTING
	$(DOCKER) $(D_RUN) $(IMG_DEV) time go vet $(LIB_PATH)
	@$(ECHO)
	@$(ECHO) == BUILDING LIBRARY
	$(DOCKER) $(D_RUN) $(IMG_DEV) time go install -v -a $(LIB_PATH)
	@$(ECHO)

.PHONY: test
test: bin/$(NAME)
	-$(DOCKER) $(D_RUN) -it $(IMG_DEV) $(NAME)
	@$(ECHO)

.PHONY: svc_start svc_stop svc_rm
svc_start:
	$(DOCKER_COMPOSE) create es redis rabbitmq
	@$(ECHO)
	$(DOCKER_COMPOSE) start es redis rabbitmq
	@$(ECHO)

svc_stop:
	$(DOCKER_COMPOSE) stop es redis rabbitmq
	@$(ECHO)

svc_rm: svc_stop
	$(DOCKER_COMPOSE) rm -f es redis rabbitmq
	@$(ECHO)

.PHONY: push
push:
	@$(ECHO) == PUSHING VERSION
	$(DOCKER) push $(IMG_PROD):$(VERSION)
	@$(ECHO)
	@$(ECHO) == LATEST LATEST
	$(DOCKER) push $(IMG_PROD):latest
	@$(ECHO)

.PHONY: enter
enter: .image.dev
	@$(ECHO) == ENTERING CONTAINER
	$(DOCKER) $(D_RUN) -it $(IMG_DEV)

.PHONY: rmi
rmi:
	@$(ECHO) == REMOVING IMAGE
	-$(DOCKER) rmi -f $(IMG_DEV) $(IMG_PROD)
	-$(RM) -f .image.dev .image
	@$(ECHO)

.PHONY: clean
clean: rmi
	@$(ECHO) == CLEANING
	$(SUDO) $(RM) -rf pkg bin/$(NAME)
	@$(ECHO)
