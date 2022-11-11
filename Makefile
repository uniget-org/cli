M                   = $(shell printf "\033[34;1m▶\033[0m")
SHELL              := /bin/bash
GIT_BRANCH         ?= $(shell git branch --show-current)
GIT_COMMIT_SHA      = $(shell git rev-parse HEAD)
VERSION            ?= $(patsubst v%,%,$(GIT_BRANCH))
DOCKER_TAG         ?= $(subst /,-,$(VERSION))
TOOLS_DIR           = tools
ALL_TOOLS           = $(shell find tools -type f -wholename \*/manifest.yaml | cut -d/ -f1-2 | sort)
ALL_TOOLS_RAW       = $(subst tools/,,$(ALL_TOOLS))
TOOLS              ?= $(shell find tools -type f -wholename \*/manifest.yaml | cut -d/ -f1-2 | sort)
TOOLS_RAW          ?= $(subst tools/,,$(TOOLS))
PREFIX             ?= /docker_setup_install
TARGET             ?= /usr/local

OWNER              ?= nicholasdille
PROJECT            ?= docker-setup
REGISTRY           ?= ghcr.io
REPOSITORY_PREFIX  ?= $(OWNER)/$(PROJECT)/

HELPER              = helper
BIN                 = $(HELPER)/usr/local/bin
export PATH        := $(BIN):$(PATH)

.PHONY:
all: $(ALL_TOOLS_RAW)

.PHONY:
info: ; $(info $(M) Runtime info...)
	@echo "GIT_BRANCH:        $(GIT_BRANCH)"
	@echo "GIT_COMMIT_SHA:    $(GIT_COMMIT_SHA)"
	@echo "VERSION:           $(VERSION)"
	@echo "DOCKER_TAG:        $(DOCKER_TAG)"
	@echo "OWNER:             $(OWNER)"
	@echo "PROJECT:           $(PROJECT)"
	@echo "REGISTRY:          $(REGISTRY)"
	@echo "REPOSITORY_PREFIX: $(REPOSITORY_PREFIX)"
	@echo "TOOLS_RAW:         $(TOOLS_RAW)"

.PHONY:
help:
	@echo
	@echo "General targets:"
	@echo "    all (default)                Build all tools"
	@echo "    help                         Display help for targets"
	@echo "    clean                        Remove all temporary files"
	@echo "    metadata.json                Generate inventory from tools/*/manifest.json"
	@echo "    metadata.json--build         Build metadata image from @metadata/ and metadata.json"
	@echo "    metadata.json--push          Push metadata image"
	@echo "    metadata.json--show          Push metadata image"
	@echo
	@echo "Dependency management:"
	@echo "    renovate.json                Generate from tools/*/manifest.json"
	@echo "    tools/<tool>/manifest.json   Generate from tools/*/manifest.yaml"
	@echo "    tools/<tool>/history.json    Generate history from git"
	@echo
	@echo "Reflection:"
	@echo "    info                         Display configuration data"
	@echo "    list                         List available tools"
	@echo "    size                         Display storage usage"
	@echo "    <tool>--show                 Display directory contents"
	@echo
	@echo "Building:"
	@echo "    check                        Run shellcheck on docker-setup"
	@echo "    tools/<tool>/Dockerfile      Generate from tools/*/Dockerfile.template"
	@echo "    base                         Build base container image for all tool installations"
	@echo "    <tool>                       Build container image for specific tool"
	@echo "    <tool>--debug                Build container image specific tool and enter shell"
	@echo "    <tool>--test                 Test a tool in a container image"
	@echo "    <tool>--deep                 Build container image including all dependencies"
	@echo "    debug                        Enter shell in base image"
	@echo "    push                         Push all container images"
	@echo "    <tool>--push                 Push container image for specific tool"
	@echo "    <tool>--inspect              Inspect pushed container image for specific tool"
	@echo "    check-tools                  Run all checks check-tools-*"
	@echo "    check-tools-homepage         Display tools without a homepage"
	@echo "    check-tools-description      Display tools without a description"
	@echo "    check-tools-deps             Display tools with missing dependencies"
	@echo "    check-tools-tags             Display tools without tags or with a single tag"
	@echo "    check-tools-renovate         Display tools without renovate information"
	@echo "    tag-usage                    Show how many times the tag is used"
	@echo "    assert-no-hardcoded-version  Display tools with hardcoded versions"
	@echo
	@echo "Security:"
	@echo "    cosign.key                   Create cosign key pair"
	@echo "    metadata.json--sign          Sign metadata container image"
	@echo "    sign                         Sign all container images"
	@echo "    <tool>--sign                 Sign container image for specific tool"
	@echo "    sbom                         Create SBoM for all tools"
	@echo "    <tool>--sbom                 Create SBoM for a specific tool"
	@echo "    tools/<tool>/sbom.json       Create SBoM for specific tool"
	@echo "    <tool>--scan                 Scan SBoM for vulnerabilities"
	@echo "    attest                       Attest SBoM for all tools"
	@echo "    <tool>--attest               Attest SBoM for specific tool"
	@echo "    install                      Push, sign and attest all container images"
	@echo "    <tool>--install              Push, sign and attest container image for specific tool"
	@echo
	@echo "Git operations:"
	@echo "    recent                       Show tools changed in the last 3 days"
	@echo "    recent-days--<N>             Show tools changed in the last <N> days"
	@echo
	@echo "Helper tools:"
	@echo "    $(HELPER)/var/lib/docker-setup/manifests/<tool>.json"
	@echo "                                 Install specified tool to helper/"
	@echo
	@echo "GHCR:"
	@echo "    clean-registry-untagged      Remove all untagged container images"
	@echo "    clean-ghcr-unused--<tool>    Remove a tag on all container images"
	@echo "    ghcr-orphaned                List container image without a tools/<tool>/manifest.yaml"
	@echo "    ghcr-exists--<tool>          Check is a container image exists"
	@echo "    ghcr-exists                  Check if all container images exist"
	@echo "    ghcr-inspect                 List tags for all container images"
	@echo "    <tool>--ghcr-tags            Display tags for a container image"
	@echo "    <tool>--ghcr-inspect         Display API object for a container image"
	@echo "    delete-ghcr--<tool>          Delete container image"
	@echo "    ghcr-private                 List all private container images"
	@echo
	@echo "Reminder: foo-% => \$$@=foo-bar \$$*=bar"
	@echo
	@echo "Only some tools: TOOLS_RAW=\$$(jq -r '.tools[].name' metadata.json | grep ^k | xargs echo) make info"
	@echo

.PHONY:
clean:
	@set -o errexit; \
	rm -f metadata.json; \
	rm -rf helper; \
	for TOOL in $(ALL_TOOLS_RAW); do \
		rm -f \
			$(TOOLS_DIR)/$${TOOL}/history.json \
			$(TOOLS_DIR)/$${TOOL}/manifest.json \
			$(TOOLS_DIR)/$${TOOL}/Dockerfile \
			$(TOOLS_DIR)/$${TOOL}/build.log \
			$(TOOLS_DIR)/$${TOOL}/sbom.json; \
	done

.PHONY:
list:
	@echo "$(ALL_TOOLS_RAW)"

.PHONY:
$(addsuffix --show,$(ALL_TOOLS_RAW)):%--show: $(TOOLS_DIR)/$*
	@ls -l $(TOOLS_DIR)/$*

renovate.json: scripts/renovate.sh renovate-root.json metadata.json ; $(info $(M) Updating $@...)
	@bash scripts/renovate.sh

metadata.json: $(HELPER)/var/lib/docker-setup/manifests/jq.json $(addsuffix /manifest.json,$(ALL_TOOLS)) ; $(info $(M) Creating $@...)
	@jq --slurp --arg revision "$(GIT_COMMIT_SHA)" '{"revision": $$revision, "tools": map(.tools[])}' $(addsuffix /manifest.json,$(ALL_TOOLS)) >metadata.json

.PHONY:
metadata.json--show:%--show:
	@less $*

.PHONY:
metadata.json--build: metadata.json @metadata/Dockerfile ; $(info $(M) Building metadata image for $(GIT_COMMIT_SHA)...)
	@set -o errexit; \
	if ! docker build . \
			--file @metadata/Dockerfile \
			--build-arg commit=$(GIT_COMMIT_SHA) \
			--tag $(REGISTRY)/$(REPOSITORY_PREFIX)metadata:$(DOCKER_TAG) \
			--progress plain \
			>@metadata/build.log 2>&1; then \
		cat @metadata/build.log; \
		exit 1; \
	fi

.PHONY:
metadata.json--push: metadata.json--build ; $(info $(M) Pushing metadata image...)
	@docker push $(REGISTRY)/$(REPOSITORY_PREFIX)metadata:$(DOCKER_TAG)

.PHONY:
metadata.json--sign: $(HELPER)/var/lib/docker-setup/manifests/cosign.json cosign.key ; $(info $(M) Signing metadata image...)
	@set -o errexit; \
	source .env; \
	cosign sign --key cosign.key $(REGISTRY)/$(REPOSITORY_PREFIX)metadata:$(DOCKER_TAG)

.SECONDARY: $(addsuffix /history.json,$(ALL_TOOLS))
$(addsuffix /history.json,$(ALL_TOOLS)):$(TOOLS_DIR)/%/history.json: $(HELPER)/var/lib/docker-setup/manifests/jq.json
	@set -o errexit; \
	git log --author=renovate* --pretty="format:%cs %s" -- $(TOOLS_DIR)/$*/manifest.yaml \
	| jq --raw-input --slurp 'split("\n")' >$@

$(addsuffix /manifest.json,$(ALL_TOOLS)):$(TOOLS_DIR)/%/manifest.json: $(HELPER)/var/lib/docker-setup/manifests/jq.json $(HELPER)/var/lib/docker-setup/manifests/yq.json $(TOOLS_DIR)/%/manifest.yaml $(TOOLS_DIR)/%/history.json ; $(info $(M) Creating manifest for $*...)
	@set -o errexit; \
	yq --output-format json eval '{"tools":[.]}' $(TOOLS_DIR)/$*/manifest.yaml \
	| jq --slurp '.[0].tools[0].history = .[1] | .[0]' - $(TOOLS_DIR)/$*/history.json >$(TOOLS_DIR)/$*/manifest.json

$(addsuffix /Dockerfile,$(ALL_TOOLS)):$(TOOLS_DIR)/%/Dockerfile: $(TOOLS_DIR)/%/Dockerfile.template $(TOOLS_DIR)/Dockerfile.tail ; $(info $(M) Creating $@...)
	@set -o errexit; \
	cat $@.template >$@; \
	echo >>$@; \
	echo >>$@; \
	if test -f $*/post_install.sh; then echo 'COPY post_install.sh $${prefix}$${docker_setup_post_install}/$${name}.sh' >>$@; fi; \
	cat $(TOOLS_DIR)/Dockerfile.tail >>$@

.PHONY:
check: $(HELPER)/var/lib/docker-setup/manifests/shellcheck.json
	@shellcheck docker-setup

.PHONY:
check-tools: check-tools-homepage check-tools-description check-tools-deps check-tools-tags check-tools-renovate

.PHONY:
check-tools-homepage: $(HELPER)/var/lib/docker-setup/manifests/jq.json metadata.json
	@\
	TOOLS="$$(jq --raw-output '.tools[] | select(.homepage == null) | .name' metadata.json)"; \
	if test -n "$${TOOLS}"; then \
		echo "$(RED)Tools missing homepage:$(RESET)"; \
		echo "$${TOOLS}" \
		| while read TOOL; do \
			echo "- $${TOOL}"; \
		done; \
		exit 1; \
	fi

.PHONY:
check-tools-description: $(HELPER)/var/lib/docker-setup/manifests/jq.json metadata.json
	@\
	TOOLS="$$(jq --raw-output '.tools[] | select(.description == null) | .name' metadata.json)"; \
	if test -n "$${TOOLS}"; then \
		echo "$(RED)Tools missing description:$(RESET)"; \
		echo "$${TOOLS}" \
		| while read TOOL; do \
			echo "- $${TOOL}"; \
		done; \
	fi

.PHONY:
check-tools-deps: $(HELPER)/var/lib/docker-setup/manifests/jq.json
	@\
	TOOLS="$$(jq --raw-output '.tools[] | select(.dependencies != null) | .name' metadata.json)"; \
	if test -n "$${TOOLS}"; then \
		for TOOL in $${TOOLS}; do \
			DEPS="$$(jq --raw-output --arg tool $${TOOL} '.tools[] | select(.name == $$tool) | .dependencies[]' metadata.json)"; \
			for DEP in $${DEPS}; do \
				if ! test -f "$(TOOLS_DIR)/$${DEP}/manifest.yaml"; then \
					echo "$(RED)Dependency <$${DEP}> for tool <$${TOOL}> does not exist.$(RESET)"; \
				fi; \
			done; \
		done; \
	fi

.PHONY:
check-tools-tags: $(HELPER)/var/lib/docker-setup/manifests/jq.json metadata.json
	@\
	TOOLS="$$(jq --raw-output '.tools[] | select(.tags == null) | .name' metadata.json)"; \
	if test -n "$${TOOLS}"; then \
		echo "$(YELLOW)Tools missing tags:$(RESET)"; \
		echo "$${TOOLS}" \
		| while read TOOL; do \
			echo "- $${TOOL}"; \
		done; \
	fi; \
	TOOLS="$$(jq --raw-output '.tools[] | select(.tags | length < 2) | .name' metadata.json)"; \
	if test -n "$${TOOLS}"; then \
		echo "$(YELLOW)Tools with only one tag:$(RESET)"; \
		echo "$${TOOLS}" \
		| while read TOOL; do \
			echo "- $${TOOL}"; \
		done; \
	fi

.PHONY:
tag-usage: $(HELPER)/var/lib/docker-setup/manifests/jq.json
	@\
	jq --raw-output '.tools[] | .tags[]' metadata.json \
	| sort \
	| uniq \
	| while read -r TAG; do \
		jq --raw-output --arg tag $${TAG} '"\($$tag): \([.tools[] | select(.tags[] | contains($$tag)) | .name] | length)"' metadata.json; \
	done

.PHONY:
check-tools-renovate: $(HELPER)/var/lib/docker-setup/manifests/jq.json metadata.json
	@\
	TOOLS="$$(jq --raw-output '.tools[] | select(.renovate == null) | .name' metadata.json)"; \
	if test -n "$${TOOLS}"; then \
		echo "$(YELLOW)Tools missing renovate:$(RESET)"; \
		echo "$${TOOLS}" \
		| while read TOOL; do \
			echo "- $${TOOL}"; \
		done; \
	fi

.PHONY:
assert-no-hardcoded-version:
	@\
	find tools -type f -name Dockerfile.template -exec grep -P '\d+\.\d+(\.\d+)?' {} \; \
	| grep -v "^#syntax=" \
	| grep -v "^FROM " \
	| grep -v "^ARG " \
	| grep -v "127.0.0.1"

.PHONY:
base: info ; $(info $(M) Building base image $(REGISTRY)/$(REPOSITORY_PREFIX)base:$(DOCKER_TAG)...)
	@set -o errexit; \
	if ! docker build @base \
			--build-arg prefix_override=$(PREFIX) \
			--build-arg target_override=$(TARGET) \
			--cache-from $(REGISTRY)/$(REPOSITORY_PREFIX)base:$(DOCKER_TAG) \
			--tag $(REGISTRY)/$(REPOSITORY_PREFIX)base:$(DOCKER_TAG) \
			--progress plain \
			>@base/build.log 2>&1; then \
		cat @base/build.log; \
		exit 1; \
	fi

.PHONY:
$(ALL_TOOLS_RAW):%: $(HELPER)/var/lib/docker-setup/manifests/jq.json base $(TOOLS_DIR)/%/manifest.json $(TOOLS_DIR)/%/Dockerfile ; $(info $(M) Building image $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG)...)
	@set -o errexit; \
	TOOL_VERSION="$$(jq --raw-output '.tools[].version' tools/$*/manifest.json)"; \
	DEPS="$$(jq --raw-output '.tools[] | select(.dependencies != null) |.dependencies[]' tools/$*/manifest.json | paste -sd,)"; \
	TAGS="$$(jq --raw-output '.tools[] | select(.tags != null) |.tags[]' tools/$*/manifest.json | paste -sd,)"; \
	echo "Name:         $*"; \
	echo "Version:      $${TOOL_VERSION}"; \
	echo "Dependencies: $${DEPS}"; \
	if ! docker build $(TOOLS_DIR)/$@ \
			--build-arg branch=$(DOCKER_TAG) \
			--build-arg ref=$(DOCKER_TAG) \
			--build-arg name=$* \
			--build-arg version=$${TOOL_VERSION} \
			--build-arg deps=$${DEPS} \
			--build-arg tags=$${TAGS} \
			--cache-from $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG) \
			--tag $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG) \
			--progress plain \
			>$(TOOLS_DIR)/$@/build.log 2>&1; then \
		cat $(TOOLS_DIR)/$@/build.log; \
		exit 1; \
	fi

$(addsuffix --deep,$(ALL_TOOLS_RAW)):%--deep: info metadata.json
	@set -o errexit; \
	DEPS="$$(./docker-setup --tools="$*" dependencies)"; \
	echo "Making deps: $${DEPS}."; \
	make $${DEPS}

.PHONY:
push: $(addsuffix --push,$(TOOLS_RAW)) metadata.json--push

.PHONY:
$(addsuffix --push,$(ALL_TOOLS_RAW)):%--push: % ; $(info $(M) Pushing image for $*...)
	@docker push $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG)

.PHONY:
$(addsuffix --inspect,$(ALL_TOOLS_RAW)):%--inspect: $(HELPER)/var/lib/docker-setup/manifests/regclient.json ; $(info $(M) Inspecting image for $*...)
	@regctl manifest get $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG)

.PHONY:
recent: recent-days--3

.PHONY:
recent-days--%:
	@set -o errexit; \
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
push-new: $(HELPER)/var/lib/docker-setup/manifests/regclient.json
	@ \
	CONFIG_DIGEST="$$( \
		regctl manifest get $(REGISTRY)/$(REPOSITORY_PREFIX)metadata:$(DOCKER_TAG) --format raw-body \
		| jq --raw-output '.config.digest' \
	)"; \
	OLD_COMMIT_SHA="$$( \
		regctl blob get $(REGISTRY)/$(REPOSITORY_PREFIX)metadata:$(DOCKER_TAG) $${CONFIG_DIGEST} \
		| jq --raw-output '.config.Labels."org.opencontainers.image.revision"' \
	)"; \
	CHANGED_TOOLS="$$( \
		git log --pretty=format: --name-only $${OLD_COMMIT_SHA}..$${GITHUB_SHA} \
		| sort \
		| grep -E "^tools/[^/]+/" \
		| cut -d/ -f2 \
		| uniq \
		| xargs \
	)"; \
	TOOLS_RAW="$${CHANGED_TOOLS}" make push metadata.json--push

.PHONY:
install: push sign attest

.PHONY:
$(addsuffix --install,$(ALL_TOOLS_RAW)):%--install: %--push %--sign %--attest

.PHONY:
$(addsuffix --debug,$(ALL_TOOLS_RAW)):%--debug: $(HELPER)/var/lib/docker-setup/manifests/jq.json $(TOOLS_DIR)/%/manifest.json $(TOOLS_DIR)/%/Dockerfile ; $(info $(M) Debugging image for $*...)
	@set -o errexit; \
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

cosign.key: $(HELPER)/var/lib/docker-setup/manifests/cosign.json ; $(info $(M) Creating key pair for cosign...)
	@set -o errexit; \
	source .env; \
	cosign generate-key-pair

.PHONY:
sign: $(addsuffix --sign,$(TOOLS_RAW))

.PHONY:
$(addsuffix --sign,$(ALL_TOOLS_RAW)):%--sign: $(HELPER)/var/lib/docker-setup/manifests/cosign.json cosign.key ; $(info $(M) Signing image for $*...)
	@set -o errexit; \
	source .env; \
	cosign sign --key cosign.key $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG)

.PHONY:
sbom: $(addsuffix /sbom.json,$(TOOLS))

.PHONY:
$(addsuffix --sbom,$(ALL_TOOLS_RAW)):%--sbom: $(TOOLS_DIR)/%/sbom.json

$(addsuffix /sbom.json,$(ALL_TOOLS)):$(TOOLS_DIR)/%/sbom.json: $(HELPER)/var/lib/docker-setup/manifests/syft.json $(TOOLS_DIR)/%/manifest.json $(TOOLS_DIR)/%/Dockerfile ; $(info $(M) Creating sbom for $*...)
	@set -o errexit; \
	syft packages --output cyclonedx-json --file $@ $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG); \
	test -s $(TOOLS_DIR)/$*/sbom.json || rm $(TOOLS_DIR)/$*/sbom.json

.PHONY:
attest: $(addsuffix --attest,$(TOOLS_RAW))

.PHONY:
$(addsuffix --scan,$(ALL_TOOLS_RAW)):%--scan: $(HELPER)/var/lib/docker-setup/manifests/grype.json $(TOOLS_DIR)/%/sbom.json
	@set -o errexit; \
	grype sbom:$(TOOLS_DIR)/$*/sbom.json --add-cpes-if-none --fail-on high --output table

.PHONY:
$(addsuffix --attest,$(ALL_TOOLS_RAW)):%--attest: $(HELPER)/var/lib/docker-setup/manifests/cosign.json sbom/%.json cosign.key ; $(info $(M) Attesting sbom for $*...)
	@set -o errexit; \
	source .env; \
	cosign attest --predicate sbom/$*.json --type cyclonedx --key cosign.key $(REGISTRY)/$(REPOSITORY_PREFIX)$*:$(DOCKER_TAG)

.PHONY:
size:
	@set -o errexit; \
	export VERSION=$(VERSION); \
	bash scripts/usage.sh $(TOOLS_RAW)

.PHONY:
debug: base
	@docker container run \
		--interactive \
		--tty \
		--privileged \
		--rm \
		$(REGISTRY)/$(REPOSITORY_PREFIX)base:$(DOCKER_TAG) \
			bash

.PHONY:
$(addsuffix --test,$(ALL_TOOLS_RAW)):%--test: % ; $(info $(M) Testing $*...)
	@set -o errexit; \
	if ! test -f "$(TOOLS_DIR)/$*/test.sh"; then \
		echo "Nothing to test."; \
		exit; \
	fi; \
	./docker-setup --tools=$* build test-$*; \
	bash $(TOOLS_DIR)/$*/test.sh test-$*

.PHONY:
clean-registry-untagged: $(HELPER)/var/lib/docker-setup/manifests/yq.json $(HELPER)/var/lib/docker-setup/manifests/gh.json $(HELPER)/var/lib/docker-setup/manifests/jq.json $(HELPER)/var/lib/docker-setup/manifests/curl.json
	@set -o errexit; \
	TOKEN="$$(yq '."github.com".oauth_token' "$${HOME}/.config/gh/hosts.yml")"; \
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
clean-ghcr-unused--%: $(HELPER)/var/lib/docker-setup/manifests/yq.json $(HELPER)/var/lib/docker-setup/manifests/gh.json $(HELPER)/var/lib/docker-setup/manifests/jq.json $(HELPER)/var/lib/docker-setup/manifests/curl.json
	@set -o errexit; \
	echo "Removing tag $*"; \
	TOKEN="$$(yq '."github.com".oauth_token' "$${HOME}/.config/gh/hosts.yml")"; \
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
ghcr-orphaned: $(HELPER)/var/lib/docker-setup/manifests/gh.json $(HELPER)/var/lib/docker-setup/manifests/jq.json
	@set -o errexit; \
	gh api --paginate /user/packages?package_type=container | jq --raw-output '.[].name' \
	| cut -d/ -f2 \
	| while read NAME; do \
		test "$${NAME}" == "base" && continue; \
		test "$${NAME}" == "metadata" && continue; \
		if ! test -f "$(TOOLS_DIR)/$${NAME}/manifest.yaml"; then \
			echo "Missing tool for $${NAME}"; \
			exit 1; \
		fi; \
	done

.PHONY:
ghcr-exists--%: $(HELPER)/var/lib/docker-setup/manifests/gh.json
	@gh api --paginate "user/packages/container/docker-setup%2F$*" >/dev/null 2>&1

.PHONY:
ghcr-exists: $(addprefix ghcr-exists--,$(TOOLS_RAW))

.PHONY:
ghcr-inspect: $(HELPER)/var/lib/docker-setup/manifests/gh.json $(HELPER)/var/lib/docker-setup/manifests/jq.json
	@set -o errexit; \
	gh api --paginate /user/packages?package_type=container | jq --raw-output '.[].name' \
	| while read NAME; do \
		echo "### Package $${NAME}"; \
		gh api --paginate "user/packages/container/$${NAME////%2F}/versions" \
		| jq --raw-output '.[].metadata.container.tags[]'; \
	done

.PHONY:
$(addsuffix --ghcr-tags,$(ALL_TOOLS_RAW)):%--ghcr-tags: $(HELPER)/var/lib/docker-setup/manifests/gh.json $(HELPER)/var/lib/docker-setup/manifests/jq.json
	@set -o errexit; \
	gh api --paginate "user/packages/container/docker-setup%2F$*/versions" \
	| jq --raw-output '.[] | "\(.metadata.container.tags[]);\(.name);\(.id)"' \
	| column --separator ";" --table --table-columns Tag,SHA256,ID

.PHONY:
$(addsuffix --ghcr-inspect,$(ALL_TOOLS_RAW)):%--ghcr-inspect: $(HELPER)/var/lib/docker-setup/manifests/gh.json $(HELPER)/var/lib/docker-setup/manifests/yq.json
	@set -o errexit; \
	gh api --paginate "user/packages/container/docker-setup%2F$*" \
	| yq --prettyPrint

.PHONY:
delete-ghcr--%: $(HELPER)/var/lib/docker-setup/manifests/yq.json $(HELPER)/var/lib/docker-setup/manifests/gh.json $(HELPER)/var/lib/docker-setup/manifests/jq.json $(HELPER)/var/lib/docker-setup/manifests/curl.json
	@set -o errexit; \
	TOKEN="$$(yq '."github.com".oauth_token' "$${HOME}/.config/gh/hosts.yml")"; \
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
ghcr-private: $(HELPER)/var/lib/docker-setup/manifests/gh.json $(HELPER)/var/lib/docker-setup/manifests/jq.json
	@set -o errexit; \
	gh api --paginate "user/packages?package_type=container&visibility=private" \
	| jq '.[] | "\(.name);\(.html_url)"' \
	| column --separator ";" --table --table-columns Name,Url

.PHONY:
$(addsuffix --ghcr-private,$(ALL_TOOLS_RAW)): $(HELPER)/var/lib/docker-setup/manifests/gh.json $(HELPER)/var/lib/docker-setup/manifests/jq.json ; $(info $(M) Testing that $* is publicly visible...)
	@gh api "user/packages/container/docker-setup%2F$*" \
	| jq --exit-status 'select(.visibility == "public")' >/dev/null 2>&1

$(HELPER)/var/lib/docker-setup/manifests/%.json:
	@docker_setup_cache="$${PWD}/cache" ./docker-setup --tools=$* --prefix=$(HELPER) install | cat

$(HELPER)/var/lib/docker-setup/manifests/regclient.json $(HELPER)/var/lib/docker-setup/manifests/jq.json:
	@set -o errexit; \
	mkdir -p $(HELPER)/usr/bin $(HELPER)/usr/local/bin $(HELPER)/var/lib/docker-setup/manifests; \
	curl --silent --location --output "$(HELPER)/usr/bin/regctl" "https://github.com/regclient/regclient/releases/latest/download/regctl-linux-amd64"; \
	curl --silent --location --output "$(HELPER)/usr/bin/jq" "https://github.com/stedolan/jq/releases/latest/download/jq-linux64"; \
	chmod +x "$(HELPER)/usr/bin/regctl" "$(HELPER)/usr/bin/jq"; \
	PATH="$(HELPER)/usr/bin:$${PATH}" docker_setup_cache="$${PWD}/cache" ./docker-setup --tools=regclient,jq --prefix=$(HELPER) install | cat