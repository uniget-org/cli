GO_SOURCES = $(shell find . -type f -name \*.go)
GO_VERSION = $(shell git describe --tags --abbrev=0 | tr -d v)
GO         = go

.PHONY:
go-info:
	@echo "GO_VERSION: $(GO_VERSION)"

coverage.out.tmp: $(GO_SOURCES)
	@$(GO) test -v -buildvcs -coverprofile ./coverage.out.tmp ./...

coverage.out: coverage.out.tmp
	@cat ./coverage.out.tmp | grep -v '.pb.go' | grep -v 'mock_' > ./coverage.out

.PHONY:
test: $(GO_SOURCES) ; $(info $(M) Running unit tests...)
	@$(GO) test ./...

.PHONY:
cover: coverage.out
	@echo ""
	@$(GO) tool cover -func ./coverage.out

bin/docker-setup: bin/docker-setup-linux-$(ALT_ARCH)
	@\
	cp bin/docker-setup-linux-$(ALT_ARCH) bin/docker-setup; \
	cp bin/docker-setup-linux-$(ALT_ARCH) docker-setup

bin/docker-setup-linux-$(ALT_ARCH):bin/docker-setup-linux-%: make/go.mk $(GO_SOURCES) test ; $(info $(M) Building docker-setup version $(GO_VERSION) for $(ALT_ARCH)...)
	@\
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

.PHONY:
go-clean:
	@rm -rf bin
	@rm docker-setup
	@rm coverage.out

,PHONY:
go-tidy:
	@$(GO) fmt ./...
	@$(GO) mod tidy -v

.PHONY:
go-audit:
	@$(GO) mod verify
	@$(GO) vet ./...
	@$(GO) run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	@$(GO) run golang.org/x/vuln/cmd/govulncheck@latest ./...
	@$(GO) test -buildvcs -vet=off ./...
