DISTROS = $(shell ls test/Dockerfile.* | sed 's|test/Dockerfile.||')

.PHONY: all check

all: check $(DISTROS)

check:
	shellcheck docker-setup.sh

$(DISTROS):
	distro=$@ docker buildx bake