#syntax=docker/dockerfile:1.17.0

FROM ghcr.io/uniget-org/tools/goreleaser:2.10.2@sha256:f14152ad7efa60e2faf576e43be18e9be50809a419f28ff26a4c8bab5f2678e5 AS uniget-goreleaser
FROM ghcr.io/uniget-org/tools/cosign:2.5.2@sha256:542b27f86ae008cf72ac99a8f5ff3eeaf66792774d49fc3386e00281d548e7d2 AS uniget-cosign
FROM ghcr.io/uniget-org/tools/syft:1.27.1@sha256:116193c7927832ca4d856a914599da64a49b372b9f33fe4c3b4cb07611c614ac AS uniget-syft
FROM ghcr.io/uniget-org/tools/gh:2.74.2@sha256:2e9679ca385be7428f1c3124f9e2fe27e15610fc260200800a403ff868491d0c AS uniget-gh
FROM ghcr.io/uniget-org/tools/gosec:2.22.5@sha256:256b610fd5cb5d982416eb5cf497dd712d7d5f9eb8a79696b639b34be32de1b8 AS uniget-gosec
FROM ghcr.io/uniget-org/tools/golangci-lint:2.1.6@sha256:1c6d9dc4141e2ecb774122a88c0e08fe957c357b469466f318702cc610adc843 AS lint-base
FROM golang:1.24.4@sha256:10c131810f80a4802c49cab0961bbe18a16f4bb2fb99ef16deaa23e4246fc817 AS latest-golang
FROM alpine:3.22.0@sha256:8a1f59ffb675680d47db6337b49d22281a139e9d709335b492be023728e11715 AS latest-alpine
FROM ubuntu:24.04@sha256:b59d21599a2b151e23eea5f6602f4af4d7d31c4e236d22bf0b62b86d2e386b8f AS latest-ubuntu2404

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
    "https://github.com/uniget-org/cli/releases/download/v${version}/uniget_Linux_${ARCH}.tar.gz" \
| tar --extract --gzip --directory=/usr/local/bin uniget
EOF

FROM registry.gitlab.com/uniget-org/images/ubuntu:24.04 AS noble-uniget
ARG version
COPY --from=uniget-release /usr/local/bin/uniget /usr/local/bin/uniget
LABEL \
    org.opencontainers.image.source="https://github.com/uniget-org/cli" \
    org.opencontainers.image.title="uniget CLI" \
    org.opencontainers.image.description="The universal installer and updater for (container) tools" \
    org.opencontainers.image.version="${version}"

FROM registry.gitlab.com/uniget-org/images/systemd:24.04 AS systemd-uniget
ARG version
COPY --from=uniget-release /usr/local/bin/uniget /usr/local/bin/uniget
LABEL \
    org.opencontainers.image.source="https://github.com/uniget-org/cli" \
    org.opencontainers.image.title="uniget CLI" \
    org.opencontainers.image.description="The universal installer and updater for (container) tools" \
    org.opencontainers.image.version="${version}"
