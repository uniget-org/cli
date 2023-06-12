GO_SOURCES = $(shell find . -type f -name \*.go)
GO_VERSION = $(shell git describe --tags --abbrev=0)
GO         = go

.PHONY:
go-info:
	@echo "GO_VERSION: $(GO_VERSION)"

bin/docker-setup: bin/docker-setup-linux-$(ALT_ARCH)
	@\
	cp bin/docker-setup-linux-$(ALT_ARCH) bin/docker-setup; \
	cp bin/docker-setup-linux-$(ALT_ARCH) docker-setup

bin/docker-setup-linux-$(ALT_ARCH):bin/docker-setup-linux-%: make/go.mk $(GO_SOURCES) ; $(info $(M) Building docker-setup version $(GO_VERSION) for $(ALT_ARCH)...)
	@\
	go test github.com/nicholasdille/docker-setup/pkg/... github.com/nicholasdille/docker-setup/cmd/docker-setup; \
	export GOOS=linux; \
	export GOARCH=$*; \
	CGO_ENABLED=0 \
		$(GO) build \
			-buildvcs=false \
			-ldflags "-w -s -X main.version=$(GO_VERSION)" \
			-o bin/docker-setup-$${GOOS}-$${GOARCH} \
			./cmd/docker-setup

.PHONY:
go-deps:
	@$(GO) get -u ./...
	@$(GO) mod tidy
