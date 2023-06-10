GO_SOURCES = $(shell find . -type f -name \*.go)
GO_VERSION = $(shell git describe --tags --abbrev=0)

.PHONY:
go-info:
	@echo "GO_VERSION: $(GO_VERSION)"

bin/docker-setup: bin/docker-setup-linux-$(ALT_ARCH)

bin/docker-setup-linux-$(ALT_ARCH):bin/docker-setup-linux-%: make/go.mk $(GO_SOURCES) ; $(info $(M) Building docker-setup version $(GO_VERSION) for $(ALT_ARCH)...)
	@\
	CGO_ENABLED=0 \
	GOOS=linux \
	GOARCH=$* \
		go build -buildvcs=false -ldflags "-w -s -X main.version=$(GO_VERSION)" -o bin/docker-setup-$${GOOS}-$${GOARCH} ./cmd/docker-setup

.PHONY:
go-deps:
	@go get -u ./...
	@go mod tidy
