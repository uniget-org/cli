M               = $(shell printf "\033[34;1m▶\033[0m")
GIT_BRANCH     ?= $(shell git branch --show-current)
VERSION        ?= $(patsubst v%,%,$(GIT_BRANCH))
TOOLS_DIR       = tools
TOOLS          ?= $(shell find $(TOOLS_DIR) -mindepth 1 -maxdepth 1 -type d | sort)
TOOLS_RAW      ?= $(subst tools/,,$(TOOLS))
MANIFESTS       = $(addsuffix /manifest.json,$(TOOLS))
DOCKERFILES     = $(addsuffix /Dockerfile,$(TOOLS))
PREFIX         ?= /docker_setup_install
TARGET         ?= /usr/local

OWNER          ?= nicholasdille
PROJECT        ?= docker-setup
REGISTRY       ?= ghcr.io

YQ              = bin/yq

.PHONY:
all: $(TOOLS_RAW)

.PHONY:
vars:
	@echo "VERSION=$(VERSION)"

.PHONY:
clean:
	@\
	rm -f tools.json; \
	for TOOL in $(TOOLS_RAW); do \
		rm -f $(TOOLS_DIR)/$${TOOL}/manifest.json $(TOOLS_DIR)/$${TOOL}/Dockerfile; \
	done

renovate.json: scripts/renovate.sh renovate-root.json tools.json ; $(info $(M) Updating $@...)
	@bash scripts/renovate.sh

tools.json: $(MANIFESTS) ; $(info $(M) Creating $@...)
	@jq --slurp '{"tools": map(.tools[])}' $(MANIFESTS) >tools.json

$(MANIFESTS):%.json: %.yaml $(YQ) ; $(info $(M) Creating $*.json...)
	@$(YQ) --output-format json eval '{"tools":[.]}' $*.yaml >$*.json

$(DOCKERFILES):%: %.template Dockerfile.tail ; $(info $(M) Creating $@...)
	@\
	cat $@.template >$@; \
	echo >>$@; \
	cat Dockerfile.tail >>$@

.PHONY:
login: ; $(info $(M) Logging in to $(REGISTRY)...)
	@\
	docker login $(REGISTRY)

.PHONY:
base: login ; $(info $(M) Building base image...)
	@\
	docker buildx build @base \
		--build-arg prefix_override=$(PREFIX) \
		--build-arg target_override=$(TARGET) \
		--cache-from $(REGISTRY)/$(OWNER)/$(PROJECT)/base:$(VERSION) \
		--tag $(REGISTRY)/$(OWNER)/$(PROJECT)/base:$(VERSION) \
		--push \
		--progress plain \
		>@base/build.log 2>&1 || \
	cat @base/build.log

.PHONY:
tools: $(TOOLS_RAW)

.PHONY:
$(TOOLS_RAW):%: base $(TOOLS_DIR)/%/manifest.json $(TOOLS_DIR)/%/Dockerfile ; $(info $(M) Building image for $*...)
	@\
	VERSION="$$(jq --raw-output '.tools[].version' tools/$*/manifest.json)"; \
	docker buildx build $(TOOLS_DIR)/$@ \
		--build-arg branch=$(GIT_BRANCH) \
		--build-arg ref=$(GIT_BRANCH) \
		--build-arg name=$* \
		--build-arg version=$${VERSION} \
		--cache-from $(REGISTRY)/$(OWNER)/$(PROJECT)/$*:$(VERSION) \
		--tag $(REGISTRY)/$(OWNER)/$(PROJECT)/$*:$(VERSION) \
		--push \
		--progress plain \
		>$(TOOLS_DIR)/$@/build.log 2>&1 || \
	cat $@/build.log

.PHONY:
%-debug: $(TOOLS_DIR)/%/manifest.json $(TOOLS_DIR)/%/Dockerfile ; $(info $(M) Debugging image for $*...)
	@\
	VERSION="$$(jq --raw-output '.tools[].version' $(TOOLS_DIR)/$*/manifest.json)"; \
	docker buildx build $(TOOLS_DIR)/$* \
		--build-arg branch=$(GIT_BRANCH) \
		--build-arg ref=$(GIT_BRANCH) \
		--build-arg name=$* \
		--build-arg version=$${VERSION} \
		--cache-from $(REGISTRY)/$(OWNER)/$(PROJECT)/$*:$(VERSION) \
		--tag $(REGISTRY)/$(OWNER)/$(PROJECT)/$*:$(VERSION) \
		--target prepare \
		--load \
		--progress plain \
		--no-cache && \
	docker container run \
		--interactive \
		--tty \
		--privileged \
		--rm \
		$(REGISTRY)/$(OWNER)/$(PROJECT)/$*:$(VERSION) \
			bash

.PHONY:
usage:
	@\
	export VERSION=$(VERSION); \
	bash scripts/usage.sh $(TOOLS_RAW)

.PHONY:
test: tools.json; $(info $(M) Testing image for all tools...)
	@\
	bash docker-setup.sh build $(REGISTRY)/$(OWNER)/$(PROJECT)/test:$(VERSION) $(TOOLS_RAW) && \
	docker container run \
		--interactive \
		--tty \
		--privileged \
		--rm \
		$(REGISTRY)/$(OWNER)/$(PROJECT)/test:$(VERSION) \
			bash

.PHONY:
debug: base
	@\
	docker container run \
		--interactive \
		--tty \
		--privileged \
		--rm \
		$(REGISTRY)/$(OWNER)/$(PROJECT)/base:$(VERSION) \
			bash
