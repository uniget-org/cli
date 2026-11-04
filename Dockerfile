#syntax=docker/dockerfile:1.27.1
#check=skip=SecretsUsedInArgOrEnv

FROM ghcr.io/uniget-org/tools/goreleaser:2.19.0@sha256:74404af9f131fc0f43fbab410142e7557339c4113c02f4836705042381af64f7 AS uniget-goreleaser
FROM ghcr.io/uniget-org/tools/cosign:3.1.4@sha256:41ccfe66b2bf4407b1e187cf8f4992e0227c78ade884a9b755a71d1f8d5d303c AS uniget-cosign
FROM ghcr.io/uniget-org/tools/syft:1.56.0@sha256:d5284e0f13a6a80af481dfed3ec9d71d68cc0910b3a32e6aeaa734f7fc2b2ca1 AS uniget-syft
FROM ghcr.io/uniget-org/tools/gh:2.103.3@sha256:ee468049fd5d2624c72eeab7e17b9d62d219cd9ae6cc6a5586c7a7529be38620 AS uniget-gh
FROM ghcr.io/uniget-org/tools/glab:1.128.0@sha256:e45f834a8801d5aae77950cfd1e3733dbd90049f3f62056338c8e4cfdabfde5a AS uniget-glab
FROM ghcr.io/uniget-org/tools/jq:1.8.2@sha256:346380fefb2967af66774cb15a1df991b7df546ee4d58fd0d73c3d8e985c6b5f AS uniget-jq
FROM ghcr.io/uniget-org/tools/gosec:2.30.0@sha256:01e3d6135280936caa2a8231971d35b9fa1b7b6da63cd09ad00e15af629af682 AS uniget-gosec
FROM ghcr.io/uniget-org/tools/golangci-lint:2.15.0@sha256:f18da029d37e65a96d3276f3f837c306b6790466b85366bb9ebfe1765f9431a5 AS lint-base
FROM golang:1.27.1@sha256:e0174e51e81218523251d85d248a90d24c3d5e81543b4f07a5d66229397db190 AS latest-golang
FROM alpine:3.24.2@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 AS latest-alpine
FROM ubuntu:26.04@sha256:f144425ff09be612d6d9ad965196e9cdc23dae1f42110a8a11a3e9a8198759f7 AS latest-ubuntu2604

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

FROM base AS publish
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