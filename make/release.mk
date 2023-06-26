LATEST_TAG := $(shell git describe --abbrev=0)
LATEST_VERSION := $(shell echo $(LATEST_TAG) | tr -d v)

.PHONY:
patch: \
		$(HELPER)/var/lib/docker-setup/manifests/semver.json \
		; $(info $(M) Creating patch release...)
	@make tag-$$(semver bump patch $(LATEST_VERSION))

.PHONY:
minor: ; $(info $(M) Creating minor release...)
	@make tag-$$(semver bump minor $(LATEST_VERSION))

.PHONY:
major: ; $(info $(M) Creating major release...)
	@make tag-$$(semver bump major $(LATEST_VERSION))

tag-%: ; $(info $(M) Creating tag v$*...)
	@if git show-ref --tags refs/tags/v$* >/dev/null 2>&1; then \
		echo "Tag v$* already exists"; \
		exit 1; \
	fi
	@git tag -a -m $* v$*
	@git push origin v$*