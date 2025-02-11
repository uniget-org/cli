#syntax=docker/dockerfile:1.13.0

FROM ghcr.io/uniget-org/tools/goreleaser:2.6.1@sha256:d930684d9061b5617bf1024e42b2853f19d9630dbe6627464032c07da387a9fd AS uniget-goreleaser
FROM ghcr.io/uniget-org/tools/cosign:2.4.2@sha256:4e7ef1c3803d2901825254548962b20d9a28fe49844e2fe5a81b3444b1179665 AS uniget-cosign
FROM ghcr.io/uniget-org/tools/syft:1.19.0@sha256:4733fca235840b2704409595f4df98e66ef1f026ab068ec71121c997e538d00c AS uniget-syft
FROM ghcr.io/uniget-org/tools/gh:2.66.1@sha256:8250541f7bed5d8fcc0d50fefb25790ee6918267c3783f052939593fcdf4c432 AS uniget-gh
FROM ghcr.io/uniget-org/tools/gosec:2.22.0@sha256:f59cb6fa8729f4018811e5476718509397b16fff4e006afe0b7ebe19d9838c18 AS uniget-gosec
FROM ghcr.io/uniget-org/tools/golangci-lint:1.63.4@sha256:0fbf6af165633030827e37136259b55f79c98dd2f439169229482d2cedfa8d44 AS lint-base
FROM golang:1.23.6@sha256:927112936d6b496ed95f55f362cc09da6e3e624ef868814c56d55bd7323e0959 AS latest-golang
FROM alpine:3.21.2@sha256:56fa17d2a7e7f168a043a2712e63aed1f8543aeafdcee47c58dcffe38ed51099 AS latest-alpine
FROM ubuntu:24.04@sha256:72297848456d5d37d1262630108ab308d3e9ec7ed1c3286a32fe09856619a782 AS latest-ubuntu2404

FROM --platform=${BUILDPLATFORM} latest-golang AS base
SHELL [ "/bin/sh", "-o", "errexit", "-c" ]
WORKDIR /src
COPY go.* .
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

FROM base AS build
ARG TARGETOS
ARG TARGETARCH
ARG GOOS=${TARGETOS}
ARG GOARCH=${TARGETARCH}
WORKDIR /go/src/github.com/uniget-org/cli
RUN --mount=target=.,readwrite \
    --mount=from=uniget-goreleaser,src=/bin/goreleaser,target=/usr/local/bin/goreleaser \
    --mount=from=uniget-cosign,src=/bin/cosign,target=/usr/local/bin/cosign \
    --mount=from=uniget-syft,src=/bin/syft,target=/usr/local/bin/syft \
    --mount=from=uniget-gh,src=/bin/gh,target=/usr/local/bin/gh \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
goreleaser healthcheck
goreleaser build \
    --single-target \
    --snapshot
mkdir -p /out
find dist -type f -executable -exec cp {} /out/uniget \;
EOF

FROM base AS publish
ARG GITHUB_TOKEN
ARG ACTIONS_ID_TOKEN_REQUEST_URL
ARG ACTIONS_ID_TOKEN_REQUEST_TOKEN
ARG GITHUB_REF_NAME
WORKDIR /go/src/github.com/uniget-org/cli
RUN --mount=target=.,readwrite \
    --mount=from=uniget-goreleaser,src=/bin/goreleaser,target=/usr/local/bin/goreleaser \
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
COPY --from=build /out/uniget /usr/local/bin/
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
bash scripts/test.sh
EOF

FROM base AS lint
RUN --mount=target=. \
    --mount=from=lint-base,src=/bin/golangci-lint,target=/usr/local/bin/golangci-lint \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/root/.cache/golangci-lint <<EOF
golangci-lint run
EOF

FROM scratch AS unit-test-coverage
COPY --from=unit-test /out/cover.out /cover.out

FROM scratch AS bin-unix
COPY --from=build /out/uniget /

FROM bin-unix AS bin-linux
FROM bin-unix AS bin-darwin

FROM scratch AS bin-windows
COPY --from=build /out/uniget /uniget.exe

FROM bin-${TARGETOS} AS bin

FROM latest-alpine AS ca-certificates
RUN <<EOF
apk update
apk add ca-certificates
EOF

FROM ca-certificates AS uniget
COPY --from=bin /uniget /usr/local/bin/uniget
ENTRYPOINT [ "/usr/local/bin/uniget"]

FROM scratch AS scratch-uniget
COPY --from=ca-certificates /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=bin /uniget /uniget
ENTRYPOINT [ "/uniget"]

FROM latest-alpine AS alpine-uniget
COPY --from=bin /uniget /usr/local/bin/uniget
ENTRYPOINT [ "uniget"]

FROM latest-ubuntu2404 AS ubuntu2404-uniget
COPY --from=bin /uniget /usr/local/bin/uniget
ENTRYPOINT [ "uniget"]

FROM ghcr.io/uniget-org/images/systemd:ubuntu24.04@sha256:99f520ebfebaecba22abbbb6d9aced629e25443acfda92fd6cc9f5fc450ee42e AS systemd-uniget
ARG version
ARG TARGETARCH
RUN <<EOF
case "${TARGETARCH}" in
    amd64) ARCH="x86_64" ;;
    arm64) ARCH="aarch64" ;;
    *) ARCH="${TARGETARCH}" ;;
esac
curl --silent --show-error --location --fail \
    "https://github.com/uniget-org/cli/releases/download/v${version}/uniget_Linux_${ARCH}.tar.gz" \
| tar --extract --gzip --directory=/usr/local/bin uniget
EOF
LABEL \
    org.opencontainers.image.source="https://github.com/uniget-org/cli" \
    org.opencontainers.image.title="uniget CLI" \
    org.opencontainers.image.description="The universal installer and updater for (container) tools" \
    org.opencontainers.image.version="${version}"
