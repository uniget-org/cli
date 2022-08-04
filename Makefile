OWNER          = nicholasdille
PROJECT        = docker-setup
REPOSITORY     = $(OWNER)/$(PROJECT)
BIN            = $(PWD)/bin
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

mount: mount-amd64

mount-%: check ubuntu-22.04
	@docker run \
		--interactive \
		--tty \
		--rm \
		--volume /var/run/docker.sock:/var/run/docker.sock \
		--volume "$${PWD}:/src" \
		--workdir /src \
		--platform linux/$* \
		--entrypoint bash \
		nicholasdille/docker-setup:ubuntu-22.04 --login

dind: dind-amd64

dind-%: check build-%
	@docker run \
		--interactive \
		--tty \
		--rm \
		--volume /var/run/docker.sock:/var/run/docker.sock \
		--platform linux/$* \
		--env no_wait=true \
		--entrypoint bash \
		nicholasdille/docker-setup:main --login

test: test-amd64

test-%: check build-%
	@docker run \
		--interactive \
		--tty \
		--rm \
		--privileged \
		--platform linux/$* \
		--env no_wait=true \
		--entrypoint bash \
		nicholasdille/docker-setup:main --login

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

%.json: %.yaml $(YQ) ; $(info $(M) Creating $*.json...)
	@$(YQ) --output-format json eval . $*.yaml >$*.json

$(BIN): ; $(info $(M) Preparing tools...)
	@mkdir -p $(BIN)

$(YQ): $(BIN) ; $(info $(M) Installing yq...)
	@test -f $@ && test -x $@ || ( \
		curl -sLfo $@ https://github.com/mikefarah/yq/releases/download/v$(YQ_VERSION)/yq_linux_amd64; \
		chmod +x $@; \
	)
