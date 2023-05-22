GO_SOURCES = $(shell find . -type f -name \*.go)
GO_VERSION = $(shell git describe --tags --abbrev=0)

.PHONY:
go-info:
	@echo "GO_VERSION: $(GO_VERSION)"

docker-setup: make/go.mk $(GO_SOURCES) ; $(info $(M) Building docker-setup version $(GO_VERSION)...)
	@CGO_ENABLED=0 \
		go build -buildvcs=false -ldflags "-X main.version=$(GO_VERSION)" -o docker-setup ./cmd/docker-setup

.PHONY:
go-deps:
	@go get -u ./...
	@go mod tidy
