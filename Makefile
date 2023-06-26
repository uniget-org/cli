M                   = $(shell printf "\033[34;1m▶\033[0m")
SHELL              := /bin/bash
GIT_BRANCH         ?= $(shell git branch --show-current)
GIT_COMMIT_SHA      = $(shell git rev-parse HEAD)
#VERSION            ?= $(patsubst v%,%,$(GIT_BRANCH))
VERSION            ?= main
DOCKER_TAG         ?= $(subst /,-,$(VERSION))
TOOLS_DIR           = tools
ALL_TOOLS           = $(shell find tools -type f -wholename \*/manifest.yaml | cut -d/ -f1-2 | sort)
ALL_TOOLS_RAW       = $(subst tools/,,$(ALL_TOOLS))
TOOLS              ?= $(shell find tools -type f -wholename \*/manifest.yaml | cut -d/ -f1-2 | sort)
TOOLS_RAW          ?= $(subst tools/,,$(TOOLS))
PREFIX             ?= /docker_setup_install
TARGET             ?= /usr/local

# Pre-defined colors: https://github.com/moby/buildkit/blob/master/util/progress/progressui/colors.go
BUILDKIT_COLORS    ?= run=light-blue:warning=yellow:error=red:cancel=255,165,0
NO_COLOR           ?= ""

OWNER              ?= nicholasdille
PROJECT            ?= docker-setup
REGISTRY           ?= ghcr.io
REPOSITORY_PREFIX  ?= $(OWNER)/$(PROJECT)/

HELPER              = helper
BIN                 = $(HELPER)/usr/local/bin
export PATH        := $(BIN):$(PATH)

SUPPORTED_ARCH     := x86_64 aarch64
SUPPORTED_ALT_ARCH := amd64 arm64
ARCH               ?= $(shell uname -m)
ifeq ($(ARCH),x86_64)
ALT_ARCH           := amd64
endif
ifeq ($(ARCH),aarch64)
ALT_ARCH           := arm64
endif
ifndef ALT_ARCH
$(error ERROR: Unable to determine alternative name for architecture ($(ARCH)))
endif

check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined $1$(if $2, ($2))))

.PHONY:
all: $(ALL_TOOLS_RAW)

.PHONY:
info: ; $(info $(M) Runtime info...)
	@echo "BUILDKIT_COLORS:    $(BUILDKIT_COLORS)"
	@echo "NO_COLOR:           $(NO_COLOR)"
	@echo "GIT_BRANCH:         $(GIT_BRANCH)"
	@echo "GIT_COMMIT_SHA:     $(GIT_COMMIT_SHA)"
	@echo "VERSION:            $(VERSION)"
	@echo "DOCKER_TAG:         $(DOCKER_TAG)"
	@echo "OWNER:              $(OWNER)"
	@echo "PROJECT:            $(PROJECT)"
	@echo "REGISTRY:           $(REGISTRY)"
	@echo "REPOSITORY_PREFIX:  $(REPOSITORY_PREFIX)"
	@echo "TOOLS_RAW:          $(TOOLS_RAW)"
	@echo "SUPPORTED_ARCH:     $(SUPPORTED_ARCH)"
	@echo "SUPPORTED_ALT_ARCH: $(SUPPORTED_ALT_ARCH)"
	@echo "ARCH:               $(ARCH)"
	@echo "ALT_ARCH:           $(ALT_ARCH)"

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

-include .env.mk
include make/dev.mk
include make/metadata.mk
include make/tool.mk
include make/checks.mk
include make/sbom.mk
include make/ghcr.mk
include make/helper.mk
include make/site.mk
include make/hcloud.mk
include make/terraform.mk
include make/remote.mk
include make/go.mk
include make/release.mk
