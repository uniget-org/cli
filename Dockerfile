#syntax=docker/dockerfile:1.26.0
#check=skip=SecretsUsedInArgOrEnv

FROM ghcr.io/uniget-org/tools/goreleaser:2.17.1@sha256:f4eb47949ae546ca23bc91bc1d8519f167731764aa8c0118a13927919ba7297d AS uniget-goreleaser
FROM ghcr.io/uniget-org/tools/cosign:3.1.3@sha256:efd829fa1cd8d2dfac7344548a70b80d42a226a7bc9ffc5716911a12043e6a1b AS uniget-cosign
FROM ghcr.io/uniget-org/tools/syft:1.51.0@sha256:e530d8fac4f7857a2b70b3a999f2bb54172a6cbe5946a29301d0409c5c9e752c AS uniget-syft
FROM ghcr.io/uniget-org/tools/gh:2.97.0@sha256:09bef6f0926ead3fa28961acd6c5476d9dd1765a8d1218ad4f1b91ab2d806283 AS uniget-gh
FROM ghcr.io/uniget-org/tools/glab:1.114.0@sha256:ccbf7fdedba5afeeb66d64bdc1eb3172cfac87eae3a291d43062d1c70b04d296 AS uniget-glab
FROM ghcr.io/uniget-org/tools/jq:1.8.2@sha256:346380fefb2967af66774cb15a1df991b7df546ee4d58fd0d73c3d8e985c6b5f AS uniget-jq
FROM ghcr.io/uniget-org/tools/gosec:2.28.0@sha256:c738c824abcc224d54a7b0490b89252ce299e08ed0bc0f9ea1c991556dcc45b4 AS uniget-gosec
FROM ghcr.io/uniget-org/tools/golangci-lint:2.13.0@sha256:cd45ccaaab1184d93befc8c1e72aacc75734f827b03e17b6e83a2f41ce8546f8 AS lint-base
FROM golang:1.27.0@sha256:65b6f280bf050ec5af12716857e8ea8439d694dbba8f31ceeb7630670071f2bb AS latest-golang
FROM alpine:3.24.1@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b AS latest-alpine
FROM ubuntu:26.04@sha256:678c6550cc43645e08669028bc177f50be4e7c5b8cca677067b1914d4afc7a03 AS latest-ubuntu2604

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
WORKDIR /go/src/gitlab.com/uniget-org/cli
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
    --snapshot \
    --clean
mkdir -p /out
find dist -type f -executable -exec cp {} /out/uniget \;
EOF

FROM base AS publish-gitlab
ARG CI_SERVER_HOST
ARG CI_JOB_TOKEN
ARG GITLAB_TOKEN
ARG SIGSTORE_ID_TOKEN
WORKDIR /go/src/gitlab.com/uniget-org/cli
RUN --mount=target=.,readwrite \
    --mount=from=uniget-goreleaser,src=/bin/goreleaser,target=/usr/local/bin/goreleaser \
    --mount=from=uniget-cosign,src=/bin/cosign,target=/usr/local/bin/cosign \
    --mount=from=uniget-syft,src=/bin/syft,target=/usr/local/bin/syft \
    --mount=from=uniget-glab,src=/bin/glab,target=/usr/local/bin/glab \
    --mount=from=uniget-jq,src=/bin/jq,target=/usr/local/bin/jq \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
goreleaser healthcheck
goreleaser release --release-notes=release-notes.md
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
RUN --mount=target=. <<EOF
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

FROM latest-ubuntu2604 AS ubuntu2604-uniget
COPY --from=bin /uniget /usr/local/bin/uniget
ENTRYPOINT [ "uniget"]

FROM ubuntu:rolling AS uniget-release
ARG version
ARG TARGETARCH
RUN <<EOF
apt-get update
apt-get install --yes --no-install-recommends \
    ca-certificates \
    curl \
    tar \
    gzip
case "${TARGETARCH}" in
    amd64) ARCH="x86_64" ;;
    arm64) ARCH="aarch64" ;;
    *) ARCH="${TARGETARCH}" ;;
esac
curl --silent --show-error --location --fail \
    "https://gitlab.com/uniget-org/cli/-/releases/v${version}/downloads/uniget_Linux_${ARCH}.tar.gz" \
| tar --extract --gzip --directory=/usr/local/bin uniget
EOF