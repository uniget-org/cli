DISTROS = $(shell ls env/*/Dockerfile | sed -E 's|env/([^/]+)/Dockerfile|\1|')

.PHONY: all check

all: check $(DISTROS)

check:
	@shellcheck docker-setup.sh

$(DISTROS):
	@distro=$@ docker buildx bake

test-%: %
	@docker run -it --rm --privileged --env NO_WAIT=true --env SKIP_DOCS=true nicholasdille/docker-setup:$*

CHANGELOG.md:
	@docker run \
		--interactive \
		--rm \
		--volume "$${PWD}:/usr/local/src/your-app" \
		--env CHANGELOG_GITHUB_TOKEN=$${GITHUB_TOKEN} \
        githubchangeloggenerator/github-changelog-generator \
        	--user nicholasdille \
            --project docker-setup