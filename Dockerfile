#syntax=docker/dockerfile:1.17.1

FROM ghcr.io/uniget-org/tools/goreleaser:2.11.2@sha256:5d4991014db30e89df65c53862ca7f4eebf7c916a8f5abf789e40fca20537983 AS uniget-goreleaser
FROM ghcr.io/uniget-org/tools/cosign:2.5.3@sha256:0dac9f5ab733ce3daad6a6abcfa064e2347ef7c36a464859d92f0203a52948d6 AS uniget-cosign
FROM ghcr.io/uniget-org/tools/syft:1.32.0@sha256:3a91709310a12fc2faef98bd05e3a34704e90913058e3bf43f543ee6165a5042 AS uniget-syft
FROM ghcr.io/uniget-org/tools/gh:2.78.0@sha256:a32bb98ec5f30de12136d84a148ad4929a5296571b535fb4ba6432d1ccd87cd0 AS uniget-gh
FROM ghcr.io/uniget-org/tools/glab:1.67.0@sha256:efc1c4f8d3a22e584504d7e5fea46e4a9ffd2f5e062dd3e730e98613484b6e56 AS uniget-glab
FROM ghcr.io/uniget-org/tools/jq:1.8.1@sha256:79febf71d7a0b349a4a05653af6ecb76a0472d62b8d6e1e643af9dc060c7aad8 AS uniget-jq
FROM ghcr.io/uniget-org/tools/gosec:2.22.8@sha256:c44a6848c8e60c1efeb6fce973eec2d18035acbf495e13c9ecd0d0446b74a37d AS uniget-gosec
FROM ghcr.io/uniget-org/tools/golangci-lint:2.4.0@sha256:2e99ccf7298cda09af51e9ba994b4002adfebee18ec7536300e95703e906bd58 AS lint-base
FROM golang:1.25.0@sha256:5502b0e56fca23feba76dbc5387ba59c593c02ccc2f0f7355871ea9a0852cebe AS latest-golang
FROM alpine:3.22.1@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1 AS latest-alpine
FROM ubuntu:24.04@sha256:9cbed754112939e914291337b5e554b07ad7c392491dba6daf25eef1332a22e8 AS latest-ubuntu2404

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

FROM base AS publish-github
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
    --mount=from=uniget-jq,src=/bin/jq,target=/usr/local/bin/jq \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
goreleaser healthcheck
goreleaser release
bash scripts/release-notes-github.sh >release-notes.md
echo "Updating release ${GITHUB_REF_NAME} with release notes"
gh release edit "${GITHUB_REF_NAME}" --notes-file release-notes.md
EOF

FROM base AS publish-gitlab
ARG CI_SERVER_HOST
ARG CI_JOB_TOKEN
ARG GITLAB_TOKEN
ARG SIGSTORE_ID_TOKEN
WORKDIR /go/src/github.com/uniget-org/cli
RUN --mount=target=.,readwrite \
    --mount=from=uniget-goreleaser,src=/bin/goreleaser,target=/usr/local/bin/goreleaser \
    --mount=from=uniget-cosign,src=/bin/cosign,target=/usr/local/bin/cosign \
    --mount=from=uniget-syft,src=/bin/syft,target=/usr/local/bin/syft \
    --mount=from=uniget-glab,src=/bin/glab,target=/usr/local/bin/glab \
    --mount=from=uniget-jq,src=/bin/jq,target=/usr/local/bin/jq \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
export GLAB_SEND_TELEMETRY=0
glab auth login --hostname="${CI_SERVER_HOST}" --job-token="${CI_JOB_TOKEN}"
bash scripts/release-notes-gitlab.sh >/tmp/release-notes.md
goreleaser healthcheck --config=.goreleaser-gitlab.yaml
goreleaser release --config=.goreleaser-gitlab.yaml --release-notes=/tmp/release-notes.md
cp /tmp/release-notes.md .
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
