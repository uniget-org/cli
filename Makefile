OWNER          = nicholasdille
PROJECT        = docker-setup
REPOSITORY     = $(OWNER)/$(PROJECT)
BIN            = $(PWD)/bin
YQ             = $(BIN)/yq
YQ_VERSION     = 4.22.1
DIST           = $(PWD)/dist
GIT_TAG        = $(shell git describe --tags 2>/dev/null)
GIT_BRANCH     = $(shell git branch --show-current)
DOCKER_TAG     = $(subst /,-,$(GIT_BRANCH))
RESET          = "\\e[39m\\e[49m"
GREEN          = "\\e[92m"
YELLOW         = "\\e[93m"
RED            = "\\e[91m"
GREY           = "\\e[90m"
M              = $(shell printf "\033[34;1m▶\033[0m")

DISTROS        = $(shell ls env/*/Dockerfile | sed -E 's|env/([^/]+)/Dockerfile|\1|')

.PHONY: all check env-% mount mount-% dind dind-% test test-% build build-% record-% $(DISTROS) install

all: check $(DISTROS)

check:
	@shellcheck docker-setup.sh

$(DISTROS): docker-setup.sh tools.json
	@distro=$@ docker buildx bake --load

env-%: %
	@docker run \
		--interactive \
		--tty \
		--rm \
		--privileged \
		--env no_wait=true \
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
		--env no_wait=true \
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
		nicholasdille/docker-setup:$(DOCKER_TAG) --login

test: test-amd64

test-%: check renovate.json build-%
	@docker run \
		--interactive \
		--tty \
		--rm \
		--privileged \
		--platform linux/$* \
		--env no_wait=true \
		--entrypoint bash \
		nicholasdille/docker-setup:$(DOCKER_TAG) --login

build: build-amd64

build-%: tools.json ; $(info $(M) Building $(GIT_BRANCH)...)
	@docker buildx build \
		--tag nicholasdille/docker-setup:$(DOCKER_TAG) \
		--build-arg BRANCH=$(GIT_BRANCH) \
		--build-arg DOCKER_SETUP_VERSION=$(GIT_BRANCH) \
		--platform linux/$* \
		--load \
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

/usr/local/bin/docker-setup: docker-setup.sh /var/cache/docker-setup/tools.json ; $(info $(M) Replacing docker-setup and preserving version...)
	@\
	version="$$(grep docker_setup_version= $@ | sed -E 's/docker_setup_version="([^"]+)"/\1/')"; \
	sudo cp docker-setup.sh $@; \
	sudo sed -i "s/docker_setup_version=\"main\"/docker_setup_version=\"$${version}\"/" $@

/usr/local/bin/docker-setup-%: docker-setup.sh /var/cache/docker-setup/tools.json ; $(info $(M) Replacing docker-setup and setting version $*...)
	@\
	sudo cp docker-setup.sh /usr/local/bin/docker-setup; \
	sudo sed -i "s/docker_setup_version=\"main\"/docker_setup_version=\"$*\"/" /usr/local/bin/docker-setup; \
	sudo touch /var/cache/docker-setup/$*

/var/cache/docker-setup/tools.json: tools.json ; $(info $(M) Replacing tools.json...)
	@sudo cp tools.json $@

install: tools.json ; $(info $(M) Installing locally...)
	@\
	sudo cp docker-setup.sh /usr/local/bin/docker-setup; \
	sudo mkdir -p /var/cache/docker-setup/; \
	sudo cp -r lib /var/cache/docker-setup/; \
	sudo cp tools.json /var/cache/docker-setup

install-%: install ; $(info $(M) Installing locally as $*...)
	@\
	sudo sed -i "s/docker_setup_version=\"main\"/docker_setup_version=\"$*\"/" /usr/local/bin/docker-setup; \
	sudo touch /var/cache/docker-setup/$*

renovate.json: scripts/renovate.sh renovate-root.json tools.json ; $(info $(M) Updating $@...)
	@bash scripts/renovate.sh

$(BIN): ; $(info $(M) Preparing tools...)
	@mkdir -p $(BIN)

$(YQ): $(BIN) ; $(info $(M) Installing yq...)
	@test -f $@ && test -x $@ || ( \
		curl -sLfo $@ https://github.com/mikefarah/yq/releases/download/v$(YQ_VERSION)/yq_linux_amd64; \
		chmod +x $@; \
	)
