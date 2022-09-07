M                  = $(shell printf "\033[34;1m▶\033[0m")
GIT_BRANCH        ?= $(shell git branch --show-current)
VERSION           ?= $(patsubst v%,%,$(GIT_BRANCH))
TOOLS_DIR          = tools
TOOLS             ?= $(shell find $(TOOLS_DIR) -mindepth 1 -maxdepth 1 -type d | sort)
TOOLS_RAW         ?= $(subst tools/,,$(TOOLS))
MANIFESTS          = $(addsuffix /manifest.json,$(TOOLS))
DOCKERFILES        = $(addsuffix /Dockerfile,$(TOOLS))
SBOMS              = $(patsubst tools/%,sbom/%.json,$(TOOLS))
PREFIX            ?= /docker_setup_install
TARGET            ?= /usr/local

OWNER             ?= nicholasdille
PROJECT           ?= docker-setup
REGISTRY          ?= ghcr.io
REPOSITORY_PREFIX ?= $(OWNER)/$(PROJECT)/

BIN                = bin
YQ                 = $(BIN)/yq
YQ_VERSION        ?= 4.27.3
REGCTL             = $(BIN)/regctl
REGCTL_VERSION    ?= 0.4.4

.PHONY:
all: $(TOOLS_RAW)

.PHONY:
info: ; $(info $(M) Runtime info...)
	@echo "git describe:      $$(git describe)"
	@echo "GIT_BRANCH:        $(GIT_BRANCH)"
	@echo "VERSION:           $(VERSION)"
	@echo "OWNER:             $(OWNER)"
	@echo "PROJECT:           $(PROJECT)"
	@echo "REGISTRY:          $(REGISTRY)"
	@echo "REPOSITORY_PREFIX: $(REPOSITORY_PREFIX)"

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

$(DOCKERFILES):%/Dockerfile: %/Dockerfile.template Dockerfile.tail ; $(info $(M) Creating $@...)
	@\
	cat $@.template >$@; \
	echo >>$@; \
	echo >>$@; \
	if test -f $*/post_install.sh; then echo 'COPY post_install.sh $${prefix}$${docker_setup_post_install}/${name}.json' >>$@; fi; \
	cat Dockerfile.tail >>$@

.PHONY:
login: ; $(info $(M) Logging in to $(REGISTRY)...)
	@\
	docker login $(REGISTRY)

.PHONY:
base: info ; $(info $(M) Building base image $(REGISTRY)/$(REPOSITORY_PREFIX)base:$(VERSION)...)
	@\
	docker build @base \
		--build-arg prefix_override=$(PREFIX) \
		--build-arg target_override=$(TARGET) \
		--cache-from $(REGISTRY)/$(REPOSITORY_PREFIX)base:$(VERSION) \
		--tag $(REGISTRY)/$(REPOSITORY_PREFIX)base:$(VERSION) \
		--progress plain \
		>@base/build.log 2>&1 || \
	cat @base/build.log

.PHONY:
tools: $(TOOLS_RAW)

.PHONY:
$(TOOLS_RAW):%: base $(TOOLS_DIR)/%/manifest.json $(TOOLS_DIR)/%/Dockerfile ; $(info $(M) Building image $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(VERSION)...)
	@\
	TOOL_VERSION="$$(jq --raw-output '.tools[].version' tools/$*/manifest.json)"; \
	DEPS="$$(jq --raw-output '.tools[] | select(.dependencies != null) |.dependencies[]' tools/$*/manifest.json | paste -sd,)"; \
	TAGS="$$(jq --raw-output '.tools[] | select(.tags != null) |.tags[]' tools/$*/manifest.json | paste -sd,)"; \
	echo "Name:         $*"; \
	echo "Version:      $${TOOL_VERSION}"; \
	echo "Dependencies: $${DEPS}"; \
	docker build $(TOOLS_DIR)/$@ \
		--build-arg branch=$(GIT_BRANCH) \
		--build-arg ref=$(GIT_BRANCH) \
		--build-arg name=$* \
		--build-arg version=$${TOOL_VERSION} \
		--build-arg deps=$${DEPS} \
		--build-arg tags=$${TAGS} \
		--cache-from $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(VERSION) \
		--tag $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(VERSION) \
		--progress plain \
		>$(TOOLS_DIR)/$@/build.log 2>&1 || \
	cat $(TOOLS_DIR)/$@/build.log

.PHONY: login base $(TOOLS_RAW) ; $(info $(M) Pushing images...)
push:
	@\
	echo "NOT IMPLEMENTED YET"

.PHONY:
%--debug: $(TOOLS_DIR)/%/manifest.json $(TOOLS_DIR)/%/Dockerfile ; $(info $(M) Debugging image for $*...)
	@\
	TOOL_VERSION="$$(jq --raw-output '.tools[].version' $(TOOLS_DIR)/$*/manifest.json)"; \
	DEPS="$$(jq --raw-output '.tools[] | select(.dependencies != null) |.dependencies[]' tools/$*/manifest.json | paste -sd,)"; \
	TAGS="$$(jq --raw-output '.tools[] | select(.tags != null) |.tags[]' tools/$*/manifest.json | paste -sd,)"; \
	echo "Name:         $*"; \
	echo "Version:      $${TOOL_VERSION}"; \
	echo "Dependencies: $${DEPS}"; \
	docker buildx build $(TOOLS_DIR)/$* \
		--build-arg branch=$(GIT_BRANCH) \
		--build-arg ref=$(GIT_BRANCH) \
		--build-arg name=$* \
		--build-arg version=$${TOOL_VERSION} \
		--build-arg deps=$${DEPS} \
		--build-arg tags=$${TAGS} \
		--cache-from $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(VERSION) \
		--tag $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(VERSION) \
		--target prepare \
		--load \
		--progress plain \
		--no-cache && \
	docker container run \
		--interactive \
		--tty \
		--privileged \
		--env name=$* \
		--env version=$${TOOL_VERSION} \
		--rm \
		$(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(VERSION) \
			bash

cosign.key:
	@\
	cosign generate-key-pair

.PHONY:
%--sign: cosign.key ; $(info $(M) Signing image for $*...)
	@\
	cosign sign --key cosign.key $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(VERSION)

$(SBOMS):sbom/%.json: $(TOOLS_DIR)/%/manifest.json $(TOOLS_DIR)/%/Dockerfile ; $(info $(M) Creating sbom for $*...)
	@\
	mkdir -p sbom; \
	syft packages --output cyclonedx-json --file sbom/$*.json $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(VERSION)

.PHONY:
%--attest: sbom/%.json cosign.key ; $(info $(M) Attesting sbom for $*...)
	@\
	cosign attest --predicate sbom/$*.json --type cyclonedx --key cosign.key $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(VERSION)

.PHONY:
usage:
	@\
	export VERSION=$(VERSION); \
	bash scripts/usage.sh $(TOOLS_RAW)

.PHONY:
test: tools.json; $(info $(M) Testing image for all tools...)
	@\
	bash docker-setup.sh build $(REGISTRY)/$(REPOSITORY_PREFIX)test:$(VERSION) $(TOOLS_RAW) && \
	docker container run \
		--interactive \
		--tty \
		--privileged \
		--rm \
		$(REGISTRY)/$(REPOSITORY_PREFIX)test:$(VERSION) \
			bash

.PHONY:
debug: base
	@\
	docker container run \
		--interactive \
		--tty \
		--privileged \
		--rm \
		$(REGISTRY)/$(REPOSITORY_PREFIX)base:$(VERSION) \
			bash

$(YQ): ; $(info $(M) Installing yq...)
	@\
	mkdir -p $(BIN); \
	test -f $@ && test -x $@ || ( \
		curl -sLfo $@ https://github.com/mikefarah/yq/releases/download/v$(YQ_VERSION)/yq_linux_amd64; \
		chmod +x $@; \
	)

$(REGCTL):
	@\
	mkdir -p $(BIN); \
	test -f $@ && test -x $@ || ( \
		curl --silent --location --output $@ "https://github.com/regclient/regclient/releases/download/v${REGCTL_VERSION}/regctl-linux-amd64"; \
		chmod +x $@; \
	)