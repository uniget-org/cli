#syntax=docker/dockerfile:1.8.1

FROM --platform=${BUILDPLATFORM} golang:1.22.4@sha256:3589439790974ec05491b66b656bf1048d0f50dd010a903463e3156ba1fc26de AS base
WORKDIR /src
ENV CGO_ENABLED=0
COPY go.* .
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /go/src/github.com/uniget-org/cli
ARG version=main
ENV CGO_ENABLED=0
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
mkdir -p /out
GOOS=${TARGETOS} \
GOARCH=${TARGETARCH} \
    go build \
        -buildvcs=false \
        -ldflags "-w -s -X main.version=${version}" \
        -o /out/uniget \
        ./cmd/uniget
EOF

FROM ghcr.io/uniget-org/tools/goreleaser:latest AS uniget-goreleaser
FROM ghcr.io/uniget-org/tools/cosign:latest AS uniget-cosign
FROM ghcr.io/uniget-org/tools/syft:latest AS uniget-syft
FROM ghcr.io/uniget-org/tools/gh:latest AS uniget-gh

FROM base AS publish
WORKDIR /go/src/github.com/uniget-org/cli
COPY . .
ARG GITHUB_TOKEN
ARG ACTIONS_ID_TOKEN_REQUEST_URL
ARG ACTIONS_ID_TOKEN_REQUEST_TOKEN
ARG GITHUB_REF_NAME
RUN --mount=from=uniget-goreleaser,src=/bin/goreleaser,target=/usr/local/bin/goreleaser \
    --mount=from=uniget-cosign,src=/bin/cosign,target=/usr/local/bin/cosign \
    --mount=from=uniget-syft,src=/bin/syft,target=/usr/local/bin/syft \
    --mount=from=uniget-gh,src=/bin/gh,target=/usr/local/bin/gh \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
goreleaser healthcheck
goreleaser release
bash scripts/release-notes.sh >release-notes.md
echo "Updating release ${GITHUB_REF_NAME} with release notes"
gh release edit "${GITHUB_REF_NAME}" --notes-file release-notes.md
EOF

FROM base AS unit-test
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
mkdir -p /out
go test \
    -v \
    -coverprofile=/out/cover.out \
    ./...
EOF

FROM base AS cli-test
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
bash scripts/test.sh
EOF

FROM base AS vet
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
mkdir -p /out
go vet \
    ./...
EOF

FROM ghcr.io/uniget-org/tools/gosec:latest AS uniget-gosec

FROM base AS gosec
RUN --mount=target=. \
    --mount=from=uniget-gosec,src=/bin/gosec,target=/usr/local/bin/gosec \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
gosec ./...
EOF

FROM ghcr.io/uniget-org/tools/staticcheck:latest AS uniget-staticcheck

FROM base AS staticcheck
RUN --mount=target=. \
    --mount=from=uniget-staticcheck,src=/bin/staticcheck,target=/usr/local/bin/staticcheck \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
staticcheck ./...
EOF

FROM golangci/golangci-lint:v1.59.1@sha256:b5f8712114561f1e2fbe74d04ed07ddfd992768705033a6251f3c7b848eac38e AS lint-base

FROM base AS lint
RUN --mount=target=. \
    --mount=from=lint-base,src=/usr/bin/golangci-lint,target=/usr/bin/golangci-lint \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/golangci-lint <<EOF
golangci-lint run --timeout 10m0s ./...
EOF

FROM scratch AS unit-test-coverage
COPY --from=unit-test /out/cover.out /cover.out

FROM scratch AS bin-unix
COPY --from=build /out/uniget /

FROM bin-unix AS bin-linux
FROM bin-unix AS bin-darwin

FROM scratch AS bin-windows
COPY --from=build /out/uniget /uniget.exe

FROM bin-${TARGETOS} as bin

FROM alpine:3.20.1@sha256:b89d9c93e9ed3597455c90a0b88a8bbb5cb7188438f70953fede212a0c4394e0 AS ca-certificates
RUN <<EOF
apk update
apk add ca-certificates
EOF

FROM ca-certificates AS uniget
COPY --from=bin /uniget /uniget
ENTRYPOINT [ "/uniget"]

FROM scratch AS scratch-uniget
COPY --from=ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=bin /uniget /uniget
ENTRYPOINT [ "/uniget"]