GO_SOURCES = $(shell find . -type f -name \*.go)
GO_VERSION = $(shell git describe --tags --abbrev=0 | tr -d v)

.PHONY:
info:
	@echo "GO_VERSION: $(GO_VERSION)"

coverage.out.tmp: \
		$(GO_SOURCES)
	@go test -v -buildvcs -coverprofile ./coverage.out.tmp ./...

coverage.out: coverage.out.tmp
	@cat ./coverage.out.tmp | grep -v '.pb.go' | grep -v 'mock_' > ./coverage.out

.PHONY:
test: \
		$(GO_SOURCES) \
		; $(info $(M) Running unit tests...)
	@go test ./...

.PHONY:
cover: \
		coverage.out
	@echo ""
	@go tool cover -func ./coverage.out

snapshot: \
		make/go.mk \
		$(GO_SOURCES) \
		; $(info $(M) Building snapshot of uniget...)
	@docker buildx bake binary

release: ; $(info $(M) Building uniget...)
	@goreleaser release --clean --snapshot --skip-sbom --skip-publish
	@cp dist/uniget_$$(go env GOOS)_$$(go env GOARCH)/uniget uniget

.PHONY:
deps:
	@go get -u ./...
	@go mod tidy

.PHONY:
clean:
	@rm -rf dist
	@rm uniget
	@rm coverage.out

,PHONY:
tidy:
	@go fmt ./...
	@go mod tidy -v

.PHONY:
audit:
	@go mod verify
	@go vet ./...
	@go run honnef.co/go/tools/cmd/staticcheck@latest -checks=all,-ST1000,-U1000 ./...
	@go run golang.org/x/vuln/cmd/govulncheck@latest ./...
	@go test -buildvcs -vet=off ./...

.PHONY:
debug: snapshot
	@\
	docker build . --target uniget --tag uniget-snapshot --load; \
	docker run -it --rm --entrypoint sh uniget-snapshot
