VERSION=$(shell cat VERSION)
BUILD=$(shell git rev-parse --short HEAD)
LDFLAGS=--ldflags "-X main.Version=$(VERSION) -X main.Build=$(BUILD)"
IMG=mc-dev
DOCKER=$(shell which docker)
RUN=run --rm -v ${PWD}/src:/go/src -v $(PWD)/etc:/etc/metrics-capacitor
NAME=metrics-capacitor
LIB_PATH=github.com/metrics-capacitor/metrics-capacitor

PHONY: prepare build enter clean mrproper

.DEFAULT_GOAL: prepare build

.image:
	@echo BUILDING DOCKER IMG
	docker build -t $(IMG) - < Dockerfile.dev
	@touch $@

bin:
	mkdir -p $@

bin/$(NAME): bin
	@echo BUILDING SOURCE
	@echo "Version:\t$(VERSION)"
	@echo "Build:\t\t$(BUILD)\n"
	$(DOCKER) $(RUN) $(IMG) bash -c "cd /go && go build $(LDFLAGS) -o $@ $(LIB_PATH)/*.go"

pkg:
	@echo GETTING GO IMPORTS
	@$(DOCKER) $(RUN) $(IMG) bash -c "cd /go && go get -v $(LIB_PATH)"

enter: .image
	@echo ENTERING CONTAINER
	$(DOCKER) $(RUN) -it $(IMG)

clean:
	@echo REMOVING ENV
	docker rmi $(IMG)
	rm -f .image

mrproper: clean
	@echo REMOVING LIBS
	rm -rf pkg bin/*

prepare: .image pkg bin
build:	bin/$(NAME)
image: .image
