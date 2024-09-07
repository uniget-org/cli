#syntax=docker/dockerfile:1.9.0

FROM ghcr.io/uniget-org/tools/goreleaser:2.2.0@sha256:744c3af45765f16fd44eedbf8b80e0df756ed1306698c9cb1fdd9c5f003a927c AS uniget-goreleaser
FROM ghcr.io/uniget-org/tools/cosign:2.4.0@sha256:f98cc3d9f9a8c8ddddd3d77ee0bb80a4950b7874ffe1cd490162372a0217592a AS uniget-cosign
FROM ghcr.io/uniget-org/tools/syft:1.11.1@sha256:0e2bfb951a24695e9c2ca5dcc9240b52877356c1b6cefa5b325407384b3118ec AS uniget-syft
FROM ghcr.io/uniget-org/tools/gh:2.55.0@sha256:49a411599c389e008c4b0124d5f516d11abf261982db3ea1cdb389af2005ce11 AS uniget-gh
FROM ghcr.io/uniget-org/tools/gosec:2.21.1@sha256:e97faed0bcc4a451fbae7432e1fb7f0f496f7b22659981fe5f931939d6e3da80 AS uniget-gosec
FROM ghcr.io/uniget-org/tools/staticcheck:2024.1.1@sha256:319956a914f39999a22e2c676f70f06e14351007b7d2d689f78b181246109e7d AS uniget-staticcheck
FROM golangci/golangci-lint:v1.60.3@sha256:e64f6ba5950132542e3b6745c18879282444699f89db69ac9b46f267c1aeb3fd AS lint-base
FROM golang:1.23.1@sha256:4a3c2bcd243d3dbb7b15237eecb0792db3614900037998c2cd6a579c46888c1e AS latest-golang
FROM alpine:3.20.3@sha256:beefdbd8a1da6d2915566fde36db9db0b524eb737fc57cd1367effd16dc0d06d AS latest-alpine
FROM ubuntu:24.04@sha256:8a37d68f4f73ebf3d4efafbcf66379bf3728902a8038616808f04e34a9ab63ee AS latest-ubuntu

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

FROM base AS vet
RUN --mount=target=. \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
mkdir -p /out
go vet \
    ./...
EOF


FROM base AS gosec
RUN --mount=target=. \
    --mount=from=uniget-gosec,src=/bin/gosec,target=/usr/local/bin/gosec \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
gosec ./...
EOF


FROM base AS staticcheck
RUN --mount=target=. \
    --mount=from=uniget-staticcheck,src=/bin/staticcheck,target=/usr/local/bin/staticcheck \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build <<EOF
staticcheck ./...
EOF


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

FROM bin-${TARGETOS} AS bin

FROM latest-alpine AS ca-certificates
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

# docker run -d --name systemd --security-opt seccomp=unconfined --tmpfs /run --tmpfs /run/lock -v /sys/fs/cgroup:/sys/fs/cgroup:ro -t systemd
# docker run -dt --privileged -v /sys/fs/cgroup:/sys/fs/cgroup systemd
FROM latest-ubuntu AS systemd
ENV container=docker \
    LC_ALL=C \
    DEBIAN_FRONTEND=noninteractive
RUN <<EOF
apt-get update
apt-get -y install --no-install-recommends \
    ca-certificates \
    systemd \
    systemd-sysv \
    systemd-cron \
    dbus \
    sudo
cd /lib/systemd/system/sysinit.target.wants/
ls | grep -v systemd-tmpfiles-setup | xargs rm -f $1
rm -f /lib/systemd/system/multi-user.target.wants/*
rm -f /etc/systemd/system/*.wants/*
rm -f /lib/systemd/system/local-fs.target.wants/*
rm -f /lib/systemd/system/sockets.target.wants/*udev*
rm -f /lib/systemd/system/sockets.target.wants/*initctl*
rm -f /lib/systemd/system/basic.target.wants/*
rm -f /lib/systemd/system/anaconda.target.wants/*
rm -f /lib/systemd/system/plymouth*
rm -f /lib/systemd/system/systemd-update-utmp*
systemctl set-default multi-user.target
EOF
STOPSIGNAL SIGRTMIN+3
VOLUME [ "/sys/fs/cgroup" ]
CMD ["/bin/bash", "-c", "exec /sbin/init --log-target=journal 3>&1"]

FROM systemd AS systemd-uniget
COPY --from=bin /uniget /uniget