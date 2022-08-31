M              = $(shell printf "\033[34;1m▶\033[0m")
GIT_BRANCH     = $(shell git branch --show-current)
TOOLS_DIR      = tools
TOOLS          = $(shell find $(TOOLS_DIR) -mindepth 1 -maxdepth 1 -type d | sort)
MANIFESTS      = $(addsuffix /manifest.json,$(TOOLS))
DOCKERFILES    = $(addsuffix /Dockerfile,$(TOOLS))
PREFIX         = /docker_setup_install
TARGET         = /usr/local

OWNER          = nicholasdille
PROJECT        = docker-setup
REGISTRY       = ghcr.io

YQ             = bin/yq

.PHONY:
all: $(TOOLS)

.PHONY:
debug:
	@echo "TOOLS=$(TOOLS)"
	@echo "MANIFESTS=$(MANIFESTS)"
	@echo "DOCKERFILES=$(DOCKERFILES)"

.PHONY:
clean:
	@\
	rm -f tools.json; \
	for TOOL in $(TOOLS); do \
		rm -f $${TOOL}/manifest.json $${TOOL}/Dockerfile; \
	done

renovate.json: $(MANIFESTS) ; $(info $(M) Updating $@...)
	@echo "NOT IMPLEMENTED YET"

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
		--cache-from $(REGISTRY)/$(OWNER)/$(PROJECT)/base:$(GIT_BRANCH) \
		--tag $(REGISTRY)/$(OWNER)/$(PROJECT)/base:$(GIT_BRANCH) \
		--push \
		--progress plain \
		>@base/build.log 2>&1 || \
	cat @base/build.log

.PHONY:
tools: $(TOOLS)

$(TOOLS):%: base $(TOOLS_DIR)/%/manifest.json $(TOOLS_DIR)/%/Dockerfile ; $(info $(M) Building image for $@...)
	@\
	docker buildx build $@ \
		--build-arg branch=$(GIT_BRANCH) \
		--build-arg ref=$(GIT_BRANCH) \
		--cache-from $(REGISTRY)/$(OWNER)/$(PROJECT)/$@:$(GIT_BRANCH) \
		--tag $(REGISTRY)/$(OWNER)/$(PROJECT)/$@:$(GIT_BRANCH) \
		--push \
		--progress plain \
		>$@/build.log 2>&1 || \
	cat $@/build.log

.PHONY:
debug-%: base $(TOOLS_DIR)/%/manifest.json $(TOOLS_DIR)/%/Dockerfile ; $(info $(M) Debugging image for $*...)
	@\
	docker buildx build $* \
		--build-arg branch=$(GIT_BRANCH) \
		--build-arg ref=$(GIT_BRANCH) \
		--cache-from $(REGISTRY)/$(OWNER)/$(PROJECT)/$*:$(GIT_BRANCH) \
		--tag $(REGISTRY)/$(OWNER)/$(PROJECT)/$*:$(GIT_BRANCH) \
		--target prepare \
		--load \
		--progress plain \
		--no-cache && \
	docker container run \
		--interactive \
		--tty \
		--privileged \
		--rm \
		$(REGISTRY)/$(OWNER)/$(PROJECT)/$*:$(GIT_BRANCH) \
			bash

.PHONY:
usage:
	@\
	for TOOL in $(TOOLS); do \
		regctl manifest get $(REGISTRY)/$(OWNER)/$(PROJECT)/$${TOOL}:$(GIT_BRANCH) --format raw-body \
		| jq -r '.layers[].size' \
		| paste -sd+ \
		| bc \
		| numfmt --to=iec-i --suffix=B --padding=20 \
		| xargs -I{} echo -n "{}"; \
		echo " $${TOOL}"; \
	done \
	| column --table --table-right=1

.PHONY:
test: $(TOOLS) tools.json; $(info $(M) Testing image for all tools...)
	@\
	bash cli.sh build-image $(REGISTRY)/$(OWNER)/$(PROJECT)/test:$(GIT_BRANCH) $(TOOLS) && \
	docker container run \
		--interactive \
		--tty \
		--privileged \
		--rm \
		$(REGISTRY)/$(OWNER)/$(PROJECT)/test:$(GIT_BRANCH) \
			bash