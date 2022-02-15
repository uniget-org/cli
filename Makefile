DISTROS = $(shell ls env/*/Dockerfile | sed -E 's|env/([^/]+)/Dockerfile|\1|')

.PHONY: all check

all: check $(DISTROS)

check:
	shellcheck docker-setup.sh

$(DISTROS):
	distro=$@ docker buildx bake

test-%: %
	@docker run -it --rm --privileged --env NO_WAIT=true --env SKIP_DOCS=true nicholasdille/docker-setup:$*