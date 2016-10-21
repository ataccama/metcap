NAME=metcap
IMG_DEV=mc-dev
IMG_PROD=blufor/$(NAME)
LIB_PATH=github.com/blufor/metcap
DOCKERFILE=Dockerfile
DOCKERFILE_DEV=Dockerfile.dev
VERSION=$(shell cat VERSION)
PWD=$(shell pwd -P)
BUILD=$(shell git rev-parse --short HEAD)
DOCKER=$(shell which docker)
DOCKER_COMPOSE=$(shell which docker-compose)
ECHO=$(shell which echo)
GIT=$(shell which git)
RM=$(shell which rm)
CP=$(shell which cp)
MKDIR=$(shell which mkdir)
CHOWN=$(shell which chown)
FIND=$(shell which find)
XARGS=$(shell which xargs)
WC=$(shell which wc)
SORT=$(shell which sort)
TOUCH=$(shell which touch)
SUDO=$(shell which sudo)
FPM=$(shell ( which fpm | grep rvm | sed s/bin/wrappers/ ) || which fpm)
LDFLAGS=--ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"
D_RUN=run --rm -h $(IMG_DEV) \
--name $(IMG_DEV) \
--net host \
-v "$(PWD):/go/src/$(LIB_PATH)" \
-v "$(PWD)/bin:/usr/local/bin" \
-v "$(PWD)/etc:/etc/$(NAME)" \
-v "$(PWD)/tmp:/tmp"

ifndef ARCH
ARCH := amd64
endif

ifeq ($(ARCH),386)
DEB_ARCH := i386
RPM_ARCH := i386
else
DEB_ARCH := amd64
RPM_ARCH := x86_64
endif

FPM_FLAGS := --log error -s dir -C pkg/$(ARCH) -n $(NAME) -v $(VERSION) -a $(ARCH) -m "radek@blufor.cz" --license GPL3

.DEFAULT_GOAL := default
.PHONY: default
default: binary

.PHONY: release
release: lint binary deb tar rpm

.PHONY: rmi
rmi:
	### REMOVING BUILT IMAGE
	-$(DOCKER) rmi -f $(IMG_PROD)
	$(RM) -f .image
	@$(ECHO)

.PHONY: clean
clean: rmi
	### REMOVING BUILT BINARY
	$(RM) -f bin/$(NAME)-*
	$(RM) -fr pkg
	@$(ECHO)

.PHONY: mrproper
mrproper: clean rmi
	### REMOVING WORK IMAGE
	-$(DOCKER) rmi -f $(IMG_PROD)
	$(RM) -f  .image.dev

.PHONY: prepare
prepare: .image.dev
.image.dev: $(DOCKERFILE_DEV)
	### BUILDING DOCKER DEV IMAGE
	$(DOCKER) build -t $(IMG_DEV) - < $(DOCKERFILE_DEV)
	@$(TOUCH) $@
	@$(ECHO)

.PHONY: lint
lint: $(shell find $(PWD) -name '*.go')
	### FORMATTING GO CODE
	$(DOCKER) $(D_RUN) $(IMG_DEV) go fmt $(LIB_PATH) $(LIB_PATH)/cmd/metcap
	$(DOCKER) $(D_RUN) $(IMG_DEV) go vet $(LIB_PATH) $(LIB_PATH)/cmd/metcap
	@$(ECHO)

.PHONY: binary
binary: bin/$(NAME)-$(ARCH)
bin/$(NAME)-$(ARCH): VERSION .image.dev $(shell find $(PWD) -name '*.go')
	### BUILDING BINARY
	### Version: $(VERSION)
	### Build:   $(BUILD)
	### Arch:    $(ARCH)
	$(DOCKER) $(D_RUN) -e 'GOARCH=$(ARCH)' $(IMG_DEV) go build $(LDFLAGS) -o /usr/local/$@ $(LIB_PATH)/cmd/metcap
	@$(ECHO)

.PHONY: deb
deb: pkg/$(NAME)-$(VERSION).$(DEB_ARCH).deb
pkg/$(NAME)-$(VERSION).$(DEB_ARCH).deb: etc/* bin/$(NAME)-$(ARCH)
	### BUILDING DEB PACKAGE: $@
	$(RM) -f $@
	$(MKDIR) -p pkg/$(ARCH)/etc/$(NAME) pkg/$(ARCH)/etc/default pkg/$(ARCH)/etc/init.d pkg/$(ARCH)/usr/bin
	$(CP) etc/* pkg/$(ARCH)/etc/$(NAME)/
	$(CP) scripts/deb-init.sh pkg/$(ARCH)/etc/init.d/$(NAME)
	$(CP) bin/$(NAME)-$(ARCH) pkg/$(ARCH)/usr/bin/$(NAME)
	$(ECHO) 'DAEMON_ARGS=""' > pkg/$(ARCH)/etc/default/$(NAME)
	$(FPM) $(FPM_FLAGS) -t deb --deb-user root --deb-group root --provides $(NAME) --after-install scripts/after-install.sh -p $@
	$(RM) -rf pkg/$(ARCH)
	@$(ECHO)

.PHONY: rpm
rpm: pkg/$(NAME)-$(VERSION).$(RPM_ARCH).rpm
pkg/$(NAME)-$(VERSION).$(RPM_ARCH).rpm: etc/* bin/$(NAME)-$(ARCH)
	### BUILDING RPM PACKAGE: $@
	$(RM) -f $@
	$(MKDIR) -p pkg/$(ARCH)/etc/$(NAME) pkg/$(ARCH)/etc/sysconfig pkg/$(ARCH)/etc/init.d pkg/$(ARCH)/usr/bin
	$(CP) etc/* pkg/$(ARCH)/etc/$(NAME)/
	$(CP) scripts/rpm-init.sh pkg/$(ARCH)/etc/init.d/$(NAME)
	$(CP) bin/$(NAME)-$(ARCH) pkg/$(ARCH)/usr/bin/$(NAME)
	$(ECHO) 'METCAP_ARGS=""' > pkg/$(ARCH)/etc/sysconfig/$(NAME)
	$(FPM) $(FPM_FLAGS) -t rpm --rpm-user root --rpm-group root --provides /usr/bin/$(NAME) --after-install scripts/after-install.sh -p $@
	$(RM) -rf pkg/$(ARCH)
	@$(ECHO)


.PHONY: tar
tar: pkg/$(NAME)-$(VERSION).$(ARCH).tar.gz
pkg/$(NAME)-$(VERSION).$(ARCH).tar.gz: etc/* bin/$(NAME)-$(ARCH)
	### BUILDING TAR ARCHIVE: $@
	$(RM) -f $@
	$(MKDIR) -p pkg/$(ARCH)/etc/$(NAME) pkg/$(ARCH)/usr/bin
	$(CP) etc/* pkg/$(ARCH)/etc/$(NAME)/
	$(CP) bin/$(NAME)-$(ARCH) pkg/$(ARCH)/usr/bin/$(NAME)
	$(FPM) $(FPM_FLAGS) -t tar -p $@
	$(RM) -rf pkg/$(ARCH)
	@$(ECHO)

.PHONY: image
image: .image
.image: scripts/docker-entrypoint.sh $(DOCKERFILE)
	### BUILDING DOCKER PROD IMAGE
	$(DOCKER) build -t $(IMG_PROD):$(VERSION) .
	$(DOCKER) tag $(IMG_PROD):$(VERSION) $(IMG_PROD):latest
	@$(TOUCH) $@
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

.PHONY: check
check: bin/$(NAME) $(shell find $(PWD) -name '*.go')
	### CHECKING GIT STATUS
	@$(GIT) diff --quiet || ( $(GIT) status && false )
	@$(ECHO)
	### LINE REPORT
	@$(FIND) $(PWD) -name '*go' | $(XARGS) $(WC) -l | $(SORT) -n

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
