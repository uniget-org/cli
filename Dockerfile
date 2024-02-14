#syntax=docker/dockerfile:1.6.0

FROM --platform=${BUILDPLATFORM} golang:1.22.0@sha256:cefea7fa6852b85f0042ce9d4b883c7e0b03b2bcb25972372d59e4f7c4367c04 AS base
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

FROM ghcr.io/uniget-org/tools/goreleaser:latest AS goreleaser
FROM ghcr.io/uniget-org/tools/cosign:latest AS cosign
FROM ghcr.io/uniget-org/tools/syft:latest AS syft
FROM ghcr.io/uniget-org/tools/gh:latest AS gh

FROM base AS publish
WORKDIR /go/src/github.com/uniget-org/cli
COPY . .
ARG GITHUB_TOKEN
ARG ACTIONS_ID_TOKEN_REQUEST_URL
ARG ACTIONS_ID_TOKEN_REQUEST_TOKEN
ARG GITHUB_REF_NAME
RUN --mount=from=goreleaser,src=/bin/goreleaser,target=/usr/local/bin/goreleaser \
    --mount=from=cosign,src=/bin/cosign,target=/usr/local/bin/cosign \
    --mount=from=syft,src=/bin/syft,target=/usr/local/bin/syft \
    --mount=from=gh,src=/bin/gh,target=/usr/local/bin/gh \
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

FROM ghcr.io/uniget-org/tools/gosec:latest AS gosec

FROM base AS sec
RUN --mount=target=. \
    --mount=from=gosec,src=/bin/gosec,target=/usr/local/bin/gosec \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
gosec ./...
EOF

FROM golangci/golangci-lint:v1.56.1@sha256:30faf70f689bfce7be639f79076c689db45df48549d43ad265ce965612a6bebf AS lint-base

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
