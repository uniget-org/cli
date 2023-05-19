.PHONY:
run--%:
	@\
	go run ./cmd/$*

.PHONY:
run: run--docker-setup

.PHONY:
run-in-docker: run-in-docker--docker-setup

.PHONY:
run-in-docker--%:
	@\
	docker run --interactive --tty --rm \
	    --mount type=bind,src=$${HOME}/go/pkg/mod,dst=/go/pkg/mod \
	    --mount type=bind,src=$${HOME}/.cache/go-build,dst=/.cache/go-build \
		--mount type=bind,src=$${PWD},dst=/src \
		--workdir /src \
		golang \
			go run ./cmd/$*

.PHONY:
build: docker-setup

GO_SOURCES = $(shell find . -type f -name \*.go)
docker-setup: make/go.mk $(GO_SOURCES)
	@\
	docker run --rm \
	    --mount type=bind,src=$${HOME}/go/pkg/mod,dst=/go/pkg/mod \
	    --mount type=bind,src=$${HOME}/.cache/go-build,dst=/.cache/go-build \
		--mount type=bind,src=$${PWD},dst=/src \
		--workdir /src \
		--user $$(id -u):$$(id -g) \
		--env CGO_ENABLED=0 \
		golang \
			go build -buildvcs=false -ldflags "-X main.version=$(VERSION)" -o docker-setup ./cmd/docker-setup
