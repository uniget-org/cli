OWNER          = nicholasdille
PROJECT        = docker-setup
REPOSITORY     = $(OWNER)/$(PROJECT)
BIN            = $(PWD)/bin
SEMVER_VERSION = 3.3.0
SEMVER         = $(BIN)/semver
HUB_VERSION    = 2.14.2
HUB            = $(BIN)/hub
GH             = $(BIN)/gh
GH_VERSION     = 2.5.1
YQ             = $(BIN)/yq
YQ_VERSION     = 4.22.1
DIST           = $(PWD)/dist
GIT_TAG        = $(shell git describe --tags 2>/dev/null)
RESET          = "\\e[39m\\e[49m"
GREEN          = "\\e[92m"
YELLOW         = "\\e[93m"
RED            = "\\e[91m"
GREY           = "\\e[90m"
M              = $(shell printf "\033[34;1m▶\033[0m")

DISTROS        = $(shell ls env/*/Dockerfile | sed -E 's|env/([^/]+)/Dockerfile|\1|')

.PHONY: all check env-% test test-% build build-% record-%

all: check $(DISTROS)

check:
	@shellcheck docker-setup.sh

$(DISTROS): docker-setup.sh tools.json
	@distro=$@ docker buildx bake

env-%: %
	@docker run \
		--interactive \
		--tty \
		--rm \
		--privileged \
		--env NO_WAIT=true \
		--env SKIP_DOCS=true \
		--volume "${PWD}/.downloads:/var/cache/docker-setup/downloads" \
		nicholasdille/docker-setup:$*

CHANGELOG.md:
	@docker run \
		--interactive \
		--rm \
		--volume "$${PWD}:/usr/local/src/your-app" \
		--env CHANGELOG_GITHUB_TOKEN=$${GITHUB_TOKEN} \
        githubchangeloggenerator/github-changelog-generator \
        	--user nicholasdille \
            --project docker-setup

build: docker-setup.sh tools.json
	@docker image build \
		--tag nicholasdille/docker-setup:main \
		.

test: test-amd64

test-%: check build-%
	@docker run \
		--interactive \
		--tty \
		--rm \
		--privileged \
		--platform linux/$* \
		--entrypoint bash \
		nicholasdille/docker-setup:main

build-%: tools.json
	@docker image build \
		--tag nicholasdille/docker-setup:main \
		--platform linux/$* \
		.

record-%: build-%
	@docker run \
		--interactive \
		--tty \
		--rm \
		--privileged \
		--volume "${HOME}/.config/asciinema:/root/.config/asciinema" \
		--entrypoint bash \
		nicholasdille/docker-setup:$*

%.json: %.yaml $(YQ)
	@$(YQ) --output-format json eval . $*.yaml >$*.json

$(BIN): ; $(info $(M) Preparing tools...)
	@mkdir -p $(BIN)

$(SEMVER): $(BIN) ; $(info $(M) Installing semver...)
	@test -f $@ && test -x $@ || ( \
		curl -sLf https://github.com/fsaintjacques/semver-tool/raw/$(SEMVER_VERSION)/src/semver > $@; \
		chmod +x $@; \
	)

$(HUB): $(BIN) ; $(info $(M) Installing hub...)
	@test -f $@ && test -x $@ || ( \
		curl -sLf https://github.com/github/hub/releases/download/v$(HUB_VERSION)/hub-linux-amd64-$(HUB_VERSION).tgz \
		| tar -xzC "$(PWD)" hub-linux-amd64-$(HUB_VERSION)/bin/hub --strip-components=1; \
		chmod +x $@; \
	)

$(GH): $(BIN) ; $(info $(M) Installing gh...)
	@test -f $@ && test -x $@ || ( \
		curl -sLf https://github.com/cli/cli/releases/download/v$(GH_VERSION)/gh_$(GH_VERSION)_linux_amd64.tar.gz \
		| tar -xzC "$(PWD)" gh_$(GH_VERSION)_linux_amd64/bin/gh --strip-components=1; \
		chmod +x $@; \
	)

$(YQ): $(BIN) ; $(info $(M) Installing yq...)
	@test -f $@ && test -x $@ || ( \
		curl -sLfo $@ https://github.com/mikefarah/yq/releases/download/v$(YQ_VERSION)/yq_linux_amd64; \
		chmod +x $@; \
	)

.PHONY: clean info dist check-changes bump-% release-% tag-% release check-tag

clean:
	@rm -rf $(DIST)

info:
	@\
	echo "$(GREEN)Repository$(RESET): $(REPOSITORY)"; \
	echo "$(GREEN)Git tag$(RESET)   : $(GIT_TAG)"

dist: $(DIST) dist/docker-setup.sh.sha256 dist/contrib.tar.gz.sha256

$(DIST):
	@mkdir -p $(DIST)

dist/docker-setup.sh: docker-setup.sh ; $(info $(M) Creating $@...)
	@cat docker-setup.sh | sed "s/^DOCKER_SETUP_VERSION=main$$/DOCKER_SETUP_VERSION=$(GIT_TAG)/' >$@

dist/contrib.tar.gz: ; $(info $(M) Creating $@...)
	@tar -czf $@ contrib

%.sha256: % ; $(info $(M) Creating SHA256 for $*...)
	@echo sha256sum $* > $@

check-changes: ; $(info $(M) Checking for uncommitted changes...)
	@if test -n "$$(git status --short | grep -vE "(Makefile|.gitignore)")"; then \
		git status --short; \
		false; \
	fi

check-untagged: $(SEMVER) ; $(info $(M) Checking for untagged commits in $(GIT_TAG)...)
	@if ! $(SEMVER) get prerel $(GIT_TAG) | grep -v --quiet "^[0-9]*-g[0-9a-f]*$$"; then \
		PAGER= git log --oneline -n $$($(SEMVER) get prerel $(GIT_TAG) | cut -d- -f1); \
		false; \
	fi

check-new-tag: $(GH) ; $(info $(M) Checking for new tag $(GIT_TAG)...)
	@if git tag | grep -q v1.1.0-rc.5; then \
		echo "$(RED)ERROR: Tag $(GIT_TAG) already exists.$(RESET)"; \
		false; \
	fi

check-gh-token: ; $(info $(M) Checking for GitHub token...)
	@if test -z "$${GITHUB_TOKEN}"; then \
		echo "$(RED)ERROR: Missing GITHUB_TOKEN.$(RESET)"; \
		false; \
	fi

check-new-release: $(GH) check-gh-token ; $(info $(M) Checking for new release for tag $(GIT_TAG)...)
	@if test "$$($(GH) release view --repo $(REPOSITORY) --json tagName --jq '.tagName' $(GIT_TAG) 2>&1)" = "$(GIT_TAG)"; then \
		echo "$(RED)ERROR: Release for tag $(GIT_TAG) already exists.$(RESET)"; \
		false; \
	fi

bump-%: $(SEMVER) ; $(info $(M) Bumping $* for version $(GIT_TAG)...)
	@$(SEMVER) bump $* $(GIT_TAG)

tag-%: | check-changes ; $(info $(M) Tagging as $*...)
	@git tag | grep -q "$(GIT_TAG)" || git tag --annotate --sign $* --message "Version $*"; \
	git push origin $*

release-%: $(GH) check-changes check-untagged tag-% dist check-gh-token check-new-release ; $(info $(M) Uploading release for $(GIT_TAG)...)
	@PRERELEASE="$$(if test -n "$$($(SEMVER) get prerel $(GIT_TAG))"; then echo "--prerelease"; fi)"; \
	$(GH) release create --repo $(REPOSITORY) --title "Version $*" $${PRERELEASE} $* dist/* >/dev/null 2>&1

release: release-$(GIT_TAG) ; $(info $(M) Releasing version $(GIT_TAG)...)
