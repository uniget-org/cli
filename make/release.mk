LATEST_TAG     := $(shell git describe --abbrev=0)
LATEST_VERSION := $(shell echo $(LATEST_TAG) | tr -d v)
PRERELEASE     := $(shell semver get prerelease $(LATEST_VERSION))

NEXT_PATCH     := $(shell semver bump patch $(LATEST_VERSION))
NEXT_MINOR     := $(shell semver bump minor $(LATEST_VERSION))
NEXT_MAJOR     := $(shell semver bump major $(LATEST_VERSION))
NEXT_PRE_PATCH := $(shell semver bump prerelease rc.. $(NEXT_PATCH))
NEXT_PRE_MINOR := $(shell semver bump prerelease rc.. $(NEXT_MINOR))
NEXT_PRE_MAJOR := $(shell semver bump prerelease rc.. $(NEXT_MAJOR))
NEXT_GA        := $(shell semver bump release $(LATEST_VERSION))

GIT ?= git
ifeq ($(DEBUG_RELEASE), true)
	GIT := echo git
endif

.PHONY:
release-debug:
	@echo "LATEST_TAG:      $(LATEST_TAG)"
	@echo "LATEST_VERSION:  $(LATEST_VERSION)"
	@echo "PRERELEASE:      $(PRERELEASE)"
	@echo
	@echo "NEXT_PATCH:      $(NEXT_PATCH)"
	@echo "NEXT_MINOR:      $(NEXT_MINOR)"
	@echo "NEXT_MAJOR:      $(NEXT_MAJOR)"
	@echo "NEXT_PRE_PATCH:  $(NEXT_PRE_PATCH)"
	@echo "NEXT_PRE_MINOR:  $(NEXT_PRE_MINOR)"
	@echo "NEXT_PRE_MAJOR:  $(NEXT_PRE_MAJOR)"
	@echo "NEXT_GA:         $(NEXT_GA)"
	@echo
	@echo "GIT:             $(GIT)"

.PHONY:
patch: ; $(info $(M) Creating patch release...)
	@make tag--$(NEXT_PATCH)

.PHONY:
minor: ; $(info $(M) Creating minor release...)
	@make tag--$(NEXT_MINOR)

.PHONY:
major: ; $(info $(M) Creating major release...)
	@make tag--$(NEXT_MAJOR)

.PHONY:
patch-pre: ; $(info $(M) Creating patch prerelease...)
	@make tag--$(NEXT_PRE_PATCH)

.PHONY:
minor-pre: ; $(info $(M) Creating patch prerelease...)
	@make tag--$(NEXT_PRE_MINOR)

.PHONY:
major-pre: ; $(info $(M) Creating patch prerelease...)
	@make tag--$(NEXT_PRE_MAJOR)

.PHONY:
ga: ; $(info $(M) Creating patch prerelease...)
	@echo "Release: Remove <$(PRERELEASE)> from $(LATEST_VERSION)"; \
	if test -z "$(PRERELEASE)"; then \
		echo "ERROR: Release is only possible from a prerelease."; \
		exit 1; \
	else \
		make tag--$(NEXT_GA); \
	fi

.PHONY:
tag--%: ; $(info $(M) Creating tag v$*...)
	@if git show-ref --tags refs/tags/v$* >/dev/null 2>&1; then \
		echo "Tag v$* already exists"; \
		exit 1; \
	fi
	@$(GIT) tag -a -m $* v$*

.PHONY:
push--%: ; $(info $(M) Pushing tag v$*...)
	@$(GIT) push origin v$*

.PHONY:
retag--%: ; $(info $(M) Creating tag v$*...)
	@$(GIT) tag -a -m $* -f v$*

.PHONY:
repush--%: ; $(info $(M) Pushing tag v$*...)
	@$(GIT) push origin v$* -f