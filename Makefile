M                   = $(shell printf "\033[34;1m▶\033[0m")
SHELL              := /bin/bash
GIT_BRANCH         ?= $(shell git branch --show-current)
GIT_COMMIT_SHA      = $(shell git rev-parse HEAD)

OWNER              ?= uniget-org
PROJECT            ?= uniget
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

include make/helper.mk
include make/go.mk
include make/release.mk
