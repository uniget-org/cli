M                   = $(shell printf "\033[34;1m▶\033[0m")
SHELL              := /bin/bash
GIT_BRANCH         ?= $(shell git branch --show-current)
GIT_COMMIT_SHA      = $(shell git rev-parse $(GIT_BRANCH))
VERSION            ?= $(patsubst v%,%,$(GIT_BRANCH))
DOCKER_TAG         ?= $(subst /,-,$(VERSION))
TOOLS_DIR           = tools
TOOLS              ?= $(shell find $(TOOLS_DIR) -mindepth 1 -maxdepth 1 -type d | sort)
TOOLS_RAW          ?= $(subst tools/,,$(TOOLS))
MANIFESTS           = $(addsuffix /manifest.json,$(TOOLS))
DOCKERFILES         = $(addsuffix /Dockerfile,$(TOOLS))
SBOMS               = $(addsuffix /sbom.json,$(TOOLS))
PREFIX             ?= /docker_setup_install
TARGET             ?= /usr/local

OWNER              ?= nicholasdille
PROJECT            ?= docker-setup
REGISTRY           ?= ghcr.io
REPOSITORY_PREFIX  ?= $(OWNER)/$(PROJECT)/

BIN                 = bin
YQ                  = $(BIN)/yq
YQ_VERSION         ?= 4.27.3
REGCTL              = $(BIN)/regctl
REGCTL_VERSION     ?= 0.4.4
SHELLCHECK          = $(BIN)/shellcheck
SHELLCHECK_VERSION ?= 0.8.0

.PHONY:
all: $(TOOLS_RAW)

.PHONY:
info: ; $(info $(M) Runtime info...)
	@echo "GIT_BRANCH:        $(GIT_BRANCH)"
	@echo "VERSION:           $(VERSION)"
	@echo "DOCKER_TAG:        $(DOCKER_TAG)"
	@echo "OWNER:             $(OWNER)"
	@echo "PROJECT:           $(PROJECT)"
	@echo "REGISTRY:          $(REGISTRY)"
	@echo "REPOSITORY_PREFIX: $(REPOSITORY_PREFIX)"

.PHONY:
help:
	@echo
	@echo "General targets:"
	@echo "    all (default)                Build all tools"
	@echo "    clean                        Remove all temporary files"
	@echo "    metadata.json                Generate inventory from tools/*/manifest.json"
	@echo "    metadata.json--build         Build metadata image from @metadata/ and metadata.json"
	@echo "    metadata.json--push          Push metadata image"
	@echo
	@echo "Dependency management:"
	@echo "    renovate.json                Generate from tools/*/manifest.json"
	@echo "    tools/<tool>/manifest.json   Generate from tools/*/manifest.yaml"
	@echo
	@echo "Reflection:"
	@echo "    info                         Display configuration data"
	@echo "    list                         List available tools"
	@echo "    size                         Display storage usage"
	@echo
	@echo "Building:"
	@echo "    tools/<tool>/Dockerfile      Generate from tools/*/Dockerfile.template"
	@echo "    base                         Build base container image for all tool installations"
	@echo "    <tool>                       Build container image for specific tool"
	@echo "    <tool>--debug                Build container image specific tool and enter shell"
	@echo "    debug                        Enter shell in base image"
	@echo "    push                         Push all container images"
	@echo "    <tool>--push                 Push container image for specific tool"
	@echo "    <tool>--inspect              Inspect pushed container image for specific tool"
	@echo
	@echo "Security:"
	@echo "    cosign.key                   Create cosign key pair"
	@echo "    sign                         Sign all container images"
	@echo "    <tool>--sign                 Sign container image for specific tool"
	@echo "    sbom                         Create SBoM for all tools"
	@echo "    tools/<tool>/sbom.json       Create SBoM for specific tool"
	@echo "    attest                       Attest SBoM for all tools"
	@echo "    <tool>--attest               Attest SBoM for specific tool"
	@echo "    install                      Push, sign and attest all container images"
	@echo "    <tool>--install              Push, sign and attest container image for specific tool"
	@echo
	@echo "Reminder: foo-% => \$$@=foo-bar \$$*=bar"
	@echo
	@echo "Only some tools: TOOLS_RAW=\$$(jq -r '.tools[].name' metadata.json | grep ^k | xargs echo) make info"
	@echo

.PHONY:
clean:
	@\
	rm -f metadata.json; \
	for TOOL in $(TOOLS_RAW); do \
		rm -f \
			$(TOOLS_DIR)/$${TOOL}/manifest.json \
			$(TOOLS_DIR)/$${TOOL}/Dockerfile \
			$(TOOLS_DIR)/$${TOOL}/build.log \
			$(TOOLS_DIR)/$${TOOL}/sbom.json; \
	done

.PHONY:
list:
	@echo "$(TOOLS_RAW)"

.PHONY:
%--show:
	@ls -l $(TOOLS_DIR)/$*

renovate.json: scripts/renovate.sh renovate-root.json metadata.json ; $(info $(M) Updating $@...)
	@bash scripts/renovate.sh

metadata.json: $(MANIFESTS) ; $(info $(M) Creating $@...)
	@jq --slurp '{"tools": map(.tools[])}' $(MANIFESTS) >metadata.json

.PHONY:
metadata.json--build: metadata.json @metadata/Dockerfile ; $(info $(M) Building metadata image for $(GIT_COMMIT_SHA)...)
	@\
	docker build . \
		--file @metadata/Dockerfile \
		--build-arg commit=$(GIT_COMMIT_SHA) \
		--tag $(REGISTRY)/$(REPOSITORY_PREFIX)metadata:$(DOCKER_TAG) \
		--progress plain \
		>@metadata/build.log 2>&1 || \
	cat @metadata/build.log

.PHONY:
metadata.json--push: metadata.json--build ; $(info $(M) Pushing metadata image...)
	@\
	docker push $(REGISTRY)/$(REPOSITORY_PREFIX)metadata:$(DOCKER_TAG)

.PHONY:
metadata.json--sign: cosign.key ; $(info $(M) Signing metadata image...)
	@\
	source .env; \
	cosign sign --key cosign.key $(REGISTRY)/$(REPOSITORY_PREFIX)metadata:$(DOCKER_TAG)

$(MANIFESTS):%.json: %.yaml $(YQ) ; $(info $(M) Creating $*.json...)
	@$(YQ) --output-format json eval '{"tools":[.]}' $*.yaml >$*.json

$(DOCKERFILES):%/Dockerfile: %/Dockerfile.template $(TOOLS_DIR)/Dockerfile.tail ; $(info $(M) Creating $@...)
	@\
	cat $@.template >$@; \
	echo >>$@; \
	echo >>$@; \
	if test -f $*/post_install.sh; then echo 'COPY post_install.sh $${prefix}$${docker_setup_post_install}/$${name}.sh' >>$@; fi; \
	cat $(TOOLS_DIR)/Dockerfile.tail >>$@

.PHONY:
base: info ; $(info $(M) Building base image $(REGISTRY)/$(REPOSITORY_PREFIX)base:$(DOCKER_TAG)...)
	@\
	docker build @base \
		--build-arg prefix_override=$(PREFIX) \
		--build-arg target_override=$(TARGET) \
		--cache-from $(REGISTRY)/$(REPOSITORY_PREFIX)base:$(DOCKER_TAG) \
		--tag $(REGISTRY)/$(REPOSITORY_PREFIX)base:$(DOCKER_TAG) \
		--progress plain \
		>@base/build.log 2>&1 || \
	cat @base/build.log

.PHONY:
$(TOOLS_RAW):%: base $(TOOLS_DIR)/%/manifest.json $(TOOLS_DIR)/%/Dockerfile ; $(info $(M) Building image $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG)...)
	@\
	TOOL_VERSION="$$(jq --raw-output '.tools[].version' tools/$*/manifest.json)"; \
	DEPS="$$(jq --raw-output '.tools[] | select(.dependencies != null) |.dependencies[]' tools/$*/manifest.json | paste -sd,)"; \
	TAGS="$$(jq --raw-output '.tools[] | select(.tags != null) |.tags[]' tools/$*/manifest.json | paste -sd,)"; \
	echo "Name:         $*"; \
	echo "Version:      $${TOOL_VERSION}"; \
	echo "Dependencies: $${DEPS}"; \
	docker build $(TOOLS_DIR)/$@ \
		--build-arg branch=$(DOCKER_TAG) \
		--build-arg ref=$(DOCKER_TAG) \
		--build-arg name=$* \
		--build-arg version=$${TOOL_VERSION} \
		--build-arg deps=$${DEPS} \
		--build-arg tags=$${TAGS} \
		--cache-from $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG) \
		--tag $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG) \
		--progress plain \
		>$(TOOLS_DIR)/$@/build.log 2>&1 || \
	cat $(TOOLS_DIR)/$@/build.log

$(addsuffix --deep,$(TOOLS_RAW)):%--deep: metadata.json
	@\
	DEPS="$$(./docker-setup --tools="$*" dependencies)"; \
	echo "Making deps: $${DEPS}."; \
	make $${DEPS}

.PHONY:
push: $(addsuffix --push,$(TOOLS_RAW)) metadata.json--push

.PHONY:
$(addsuffix --push,$(TOOLS_RAW)):%--push: % ; $(info $(M) Pushing image for $*...)
	@\
	docker push $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG)

.PHONY:
$(addsuffix --inspect,$(TOOLS_RAW)):%--inspect: $(REGCTL) ; $(info $(M) Inspecting image for $*...)
	@\
	regctl manifest get $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG)

.PHONY:
install: push sign attest

.PHONY:
recent: recent-days--3

.PHONY:
recent-days--%:
	@\
	CHANGED_TOOLS="$$( \
		git log --pretty=format: --name-only --since="$* days ago" \
		| sort \
		| grep -E "^tools/[^/]+/" \
		| cut -d/ -f2 \
		| uniq \
		| xargs \
	)"; \
	echo "Tools changed in the last $* day(s): $${CHANGED_TOOLS}."; \
	make $${CHANGED_TOOLS}

.PHONY:
$(addsuffix --install,$(TOOLS_RAW)):%--install: %--push %--sign %--attest

.PHONY:
$(addsuffix --debug,$(TOOLS_RAW)):%--debug: $(TOOLS_DIR)/%/manifest.json $(TOOLS_DIR)/%/Dockerfile ; $(info $(M) Debugging image for $*...)
	@\
	TOOL_VERSION="$$(jq --raw-output '.tools[].version' $(TOOLS_DIR)/$*/manifest.json)"; \
	DEPS="$$(jq --raw-output '.tools[] | select(.dependencies != null) |.dependencies[]' tools/$*/manifest.json | paste -sd,)"; \
	TAGS="$$(jq --raw-output '.tools[] | select(.tags != null) |.tags[]' tools/$*/manifest.json | paste -sd,)"; \
	echo "Name:         $*"; \
	echo "Version:      $${TOOL_VERSION}"; \
	echo "Dependencies: $${DEPS}"; \
	docker buildx build $(TOOLS_DIR)/$* \
		--build-arg branch=$(DOCKER_TAG) \
		--build-arg ref=$(DOCKER_TAG) \
		--build-arg name=$* \
		--build-arg version=$${TOOL_VERSION} \
		--build-arg deps=$${DEPS} \
		--build-arg tags=$${TAGS} \
		--cache-from $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG) \
		--tag $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG) \
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
		$(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG) \
			bash

cosign.key: ; $(info $(M) Creating key pair for cosign...)
	@\
	source .env; \
	cosign generate-key-pair

.PHONY:
sign: $(addsuffix --sign,$(TOOLS_RAW))

.PHONY:
%--sign: cosign.key ; $(info $(M) Signing image for $*...)
	@\
	source .env; \
	cosign sign --key cosign.key $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG)

.PHONY:
sbom: $(SBOMS)

.PHONY:
$(addsuffix --sbom,$(TOOLS_RAW)):%--sbom: $(TOOLS_DIR)/%/sbom.json

$(SBOMS):$(TOOLS_DIR)/%/sbom.json: $(TOOLS_DIR)/%/manifest.json $(TOOLS_DIR)/%/Dockerfile ; $(info $(M) Creating sbom for $*...)
	@\
	mkdir -p sbom; \
	syft packages --output cyclonedx-json --file $@ $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG); \
	test -s $(TOOLS_DIR)/$*/sbom.json || rm $(TOOLS_DIR)/$*/sbom.json

.PHONY:
attest: $(addsuffix --attest,$(TOOLS_RAW))

.PHONY:
%--attest: sbom/%.json cosign.key ; $(info $(M) Attesting sbom for $*...)
	@\
	source .env; \
	cosign attest --predicate sbom/$*.json --type cyclonedx --key cosign.key $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG)

.PHONY:
size:
	@\
	export VERSION=$(VERSION); \
	bash scripts/usage.sh $(TOOLS_RAW)

.PHONY:
debug: base
	@\
	docker container run \
		--interactive \
		--tty \
		--privileged \
		--rm \
		$(REGISTRY)/$(REPOSITORY_PREFIX)base:$(DOCKER_TAG) \
			bash

.PHONY:
$(addsuffix --test,$(TOOLS_RAW)):%--test: % ; $(info $(M) Testing $*...)
	@\
	if ! test -f "$(TOOLS_DIR)/$*/test.sh"; then \
		echo "Nothing to test."; \
		exit; \
	fi; \
	./docker-setup --tools=$* build test-$*; \
	bash $(TOOLS_DIR)/$*/test.sh test-$*

.PHONY:
clean-registry-untagged: $(YQ)
	@set -o errexit; \
	TOKEN="$$($(YQ) '."github.com".oauth_token' "$${HOME}/.config/gh/hosts.yml")"; \
	test -n "$${TOKEN}"; \
	test "$${TOKEN}" != "null"; \
	gh api --paginate /user/packages?package_type=container | jq --raw-output '.[].name' \
	| while read NAME; do \
		echo "### Package $${NAME}"; \
		gh api --paginate "user/packages/container/$${NAME////%2F}/versions" \
		| jq --raw-output '.[] | select(.metadata.container.tags | length == 0) | .id' \
		| xargs -I{} \
			curl "https://api.github.com/users/nicholasdille/packages/container/$${NAME////%2F}/versions/{}" \
				--silent \
				--header "Authorization: Bearer $${TOKEN}" \
				--request DELETE \
				--header "Accept: application/vnd.github+json"; \
	done

.PHONY:
clean-ghcr-unused--%: $(YQ)
	@set -o errexit; \
	echo "Removing tag $*"; \
	TOKEN="$$($(YQ) '."github.com".oauth_token' "$${HOME}/.config/gh/hosts.yml")"; \
	test -n "$${TOKEN}"; \
	test "$${TOKEN}" != "null"; \
	gh api --paginate /user/packages?package_type=container | jq --raw-output '.[].name' \
	| while read NAME; do \
		echo "### Package $${NAME}"; \
		gh api --paginate "user/packages/container/$${NAME////%2F}/versions" \
		| jq --raw-output --arg tag "$*" '.[] | select(.metadata.container.tags[] | contains($$tag)) | .id' \
		| xargs -I{} \
			curl "https://api.github.com/users/nicholasdille/packages/container/$${NAME////%2F}/versions/{}" \
				--silent \
				--header "Authorization: Bearer $${TOKEN}" \
				--request DELETE \
				--header "Accept: application/vnd.github+json"; \
	done

.PHONY:
ghcr-orphaned:
	@set -o errexit; \
	gh api --paginate /user/packages?package_type=container | jq --raw-output '.[].name' \
	| cut -d/ -f2 \
	| while read NAME; do \
		test "$${NAME}" == "base" && continue; \
		test "$${NAME}" == "metadata" && continue; \
		if ! test -f "$(TOOLS_DIR)/$${NAME}/manifest.json"; then \
			echo "Missing tool for $${NAME}"; \
			exit 1; \
		fi; \
	done

.PHONY:
ghcr-exists--%:
	@set -o errexit; \
	gh api --paginate "user/packages/container/docker-setup%2F$*" >/dev/null 2>&1 || exit 1

.PHONY:
ghcr-exists: $(addprefix ghcr-exists--,$(TOOLS_RAW))

.PHONY:
ghcr-inspect:
	@set -o errexit; \
	gh api --paginate /user/packages?package_type=container | jq --raw-output '.[].name' \
	| while read NAME; do \
		echo "### Package $${NAME}"; \
		gh api --paginate "user/packages/container/$${NAME////%2F}/versions" \
		| jq --raw-output '.[].metadata.container.tags[]'; \
	done

.PHONY:
ghcr-tags--%:
	@set -o errexit; \
	gh api --paginate "user/packages/container/docker-setup%2F$*/versions" \
	| jq --raw-output '.[] | "\(.metadata.container.tags[]);\(.name);\(.id)"' \
	| column --separator ";" --table --table-columns Tag,SHA256,ID

.PHONY:
ghcr-inspect--%: $(YQ)
	@set -o errexit; \
	gh api --paginate "user/packages/container/docker-setup%2F$*" \
	| $(YQ) --prettyPrint

.PHONY:
delete-ghcr--%: $(YQ)
	@\
	TOKEN="$$($(YQ) '."github.com".oauth_token' "$${HOME}/.config/gh/hosts.yml")"; \
	test -n "$${TOKEN}"; \
	test "$${TOKEN}" != "null"; \
	PARAM=$*; \
	NAME="$${PARAM%%:*}"; \
	TAG="$${PARAM#*:}"; \
	echo "Removing $${NAME}:$${TAG}"; \
	gh api --paginate "user/packages/container/docker-setup%2F$${NAME}/versions" \
	| jq --raw-output --arg tag "$${TAG}" '.[] | select(.metadata.container.tags[] | contains($$tag)) | .id' \
	| xargs -I{} \
		curl "https://api.github.com/users/nicholasdille/packages/container/docker-setup%2F$${NAME}/versions/{}" \
			--silent \
			--header "Authorization: Bearer $${TOKEN}" \
			--request DELETE \
			--header "Accept: application/vnd.github+json"

.PHONY:
ghcr-tags--%:
	@set -o errexit; \
	gh api --paginate "user/packages/container/docker-setup%2F$*/versions" \
	| jq --raw-output '.[] | "\(.metadata.container.tags[]);\(.name);\(.id)"' \
	| column --separator ";" --table --table-columns Tag,SHA256,ID

.PHONY:
ghcr-private:
	@set -o errexit; \
	gh api --paginate "user/packages?package_type=container&visibility=private" \
	| jq '.[] | "\(.name);\(.html_url)"' \
	| column --separator ";" --table --table-columns Name,Url

.PHONY:
ghcr-private--%: ; $(info $(M) Testing that $* is publicly visible...)
	@\
	gh api "user/packages/container/docker-setup%2F$*" \
	| jq --exit-status 'select(.visibility == "public")' >/dev/null 2>&1

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

$(SHELLCHECK):
	@\
	mkdir -p $(BIN); \
	test -f $@ && test -x $@ || ( \
		curl --silent --location "https://github.com/koalaman/shellcheck/releases/download/v$(SHELLCHECK_VERSION)/shellcheck-v$(SHELLCHECK_VERSION).linux.x86_64.tar.xz" \
		| tar --extract --xz --directory=$(BIN) --strip-components=1 "shellcheck-v$(SHELLCHECK_VERSION)/shellcheck"; \
		chmod +x $@; \
	)