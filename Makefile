NAME=metcap
IMG_DEV=mc-dev
IMG_PROD=blufor/$(NAME)
LIB_PATH=github.com/blufor/metcap
DOCKERFILE=Dockerfile
DOCKERFILE_DEV=Dockerfile.dev

VERSION=$(shell cat VERSION)
PATH=$(shell pwd -P)
BUILD=$(shell git rev-parse --short HEAD)
DOCKER=$(shell which docker)
DOCKER_COMPOSE=$(shell which docker-compose)
ECHO=$(shell which echo)
GIT=$(shell which git)
RM=$(shell which rm)
FIND=$(shell which find)
XARGS=$(shell which xargs)
WC=$(shell which wc)
SORT=$(shell which sort)
TOUCH=$(shell which touch)
SUDO=$(shell which sudo)

LDFLAGS=--ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"
D_RUN=run --rm -h $(IMG_DEV) \
--name $(IMG_DEV) \
--net host \
-v "$(PATH)/bin:/usr/local/bin" \
-v "$(PATH)/src:/go/src/$(LIB_PATH)" \
-v "$(PATH)/etc:/etc/$(NAME)" \
-v "$(PATH)/tmp:/tmp"

.DEFAULT_GOAL := default
.PHONY: default
default: build image

.PHONY: build
build:	bin/$(NAME)

.PHONY: prepare
prepare: .image.dev
.image.dev: $(DOCKERFILE_DEV)
	### BUILDING DOCKER DEV IMAGE
	$(DOCKER) build -t $(IMG_DEV) - < $(DOCKERFILE_DEV)
	@$(TOUCH) $@
	@$(ECHO)

.PHONY: image
image: .image
.image: bin/$(NAME) bin/$(NAME)-docker $(DOCKERFILE)
	### BUILDING DOCKER PROD IMAGE
	$(DOCKER) build -t $(IMG_PROD):$(VERSION) .
	$(DOCKER) tag $(IMG_PROD):$(VERSION) $(IMG_PROD):latest
	@$(TOUCH) $@
	@$(ECHO)

bin/$(NAME): .image.dev $(shell find src -name '*.go') VERSION
	### FORMATTING
	$(DOCKER) $(D_RUN) $(IMG_DEV) go fmt $(LIB_PATH) $(LIB_PATH)/cmd/metcap
	@$(ECHO)
	### VETTING
	$(DOCKER) $(D_RUN) $(IMG_DEV) go vet $(LIB_PATH) $(LIB_PATH)/cmd/metcap
	@$(ECHO)
	### BUILDING BINARY
	### Version: $(VERSION)
	### Build:   $(BUILD)
	$(DOCKER) $(D_RUN) $(IMG_DEV) time go build $(LDFLAGS) -o /usr/local/$@ $(LIB_PATH)/cmd/metcap
	@$(ECHO)

.PHONY: run
run: bin/$(NAME)
	-$(DOCKER) $(D_RUN) -it $(IMG_DEV) $(NAME)
	@$(ECHO)

.PHONY: check
check: bin/$(NAME) $(shell find src -name '*.go')
	### CHECKING GIT STATUS
	@$(GIT) diff --quiet || ( $(GIT) status && false )
	@$(ECHO)
	### LINE REPORT
	@$(FIND) $(PATH) -name '*go' | $(XARGS) $(WC) -l | $(SORT) -n

.PHONY: svc_start svc_stop svc_rm
svc_start:
	cd docker
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
push: check .image
	### PUSHING SOURCE VERSION $(VERSION) \($(BUILD)\)
	$(GIT) push
	@$(ECHO)
	### PUSHING IMAGE VERSION $(VERSION) \($(BUILD)\)
	$(DOCKER) push $(IMG_PROD):$(VERSION)
	@$(ECHO)
	### PUSHING IMAGE LATEST
	$(DOCKER) push $(IMG_PROD):latest
	@$(ECHO)

.PHONY: enter
enter: .image.dev
	### ENTERING CONTAINER
	$(DOCKER) $(D_RUN) -it $(IMG_DEV)

.PHONY: rmi
rmi:
	### REMOVING BUILD IMAGE
	-$(DOCKER) rmi -f $(IMG_PROD)
	-$(RM) -f .image
	@$(ECHO)

.PHONY: clean
clean: rmi
	### REMOVING BUILD BINARY
	$(RM) -f bin/$(NAME)
	@$(ECHO)

.PHONY: mrproper
mrproper: clean rmi
	### REMOVING WORK IMAGE
	-$(DOCKER) rmi -f $(IMG_PROD)
	$(RM) -f  .image.dev
