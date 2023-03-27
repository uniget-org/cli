#syntax=docker/dockerfile:1.5.2

ARG base=ubuntu-22.04

FROM ubuntu:22.04 AS ubuntu-22.04
RUN <<EOF
apt-get update
apt-get -y install --no-install-recommends \
    curl \
    ca-certificates \
    bsdextrautils
EOF

FROM debian:11.6@sha256:7b991788987ad860810df60927e1adbaf8e156520177bd4db82409f81dd3b721 AS debian-11.5
RUN <<EOF
apt-get update
apt-get -y install --no-install-recommends \
    curl \
    ca-certificates \
    bsdextrautils
EOF

FROM alpine:3.17@sha256:ff6bdca1701f3a8a67e328815ff2346b0e4067d32ec36b7992c1fdc001dc8517 AS alpine-3.16
RUN <<EOF
apk update
apk add \
    bash \
    curl \
    ca-certificates \
    util-linux-misc
EOF

FROM ubuntu-22.04 AS dev
RUN <<EOF
apt-get update
apt-get -y install --no-install-recommends \
    make
EOF
WORKDIR /src
COPY . .

FROM ${base} AS local
COPY docker-setup /usr/local/bin/
COPY tools/Dockerfile.template /var/cache/docker-setup/

FROM ${base} AS release
RUN <<EOF
curl --silent --location --fail --output "/usr/local/bin/docker-setup" \
    "https://github.com/nicholasdille/docker-setup/raw/main/docker-setup"
chmod +x "/usr/local/bin/docker-setup"
mkdir -p /var/cache/docker-setup
curl --silent --location --fail --output "/var/cache/docker-setup/Dockerfile.template" \
    "https://github.com/nicholasdille/docker-setup/raw/main/tools/Dockerfile.template"
EOF

FROM ghcr.io/nicholasdille/docker-setup/regclient:main AS regclient
FROM ghcr.io/nicholasdille/docker-setup/jq:main AS jq
FROM ghcr.io/nicholasdille/docker-setup/yq:main AS yq
FROM ghcr.io/nicholasdille/docker-setup/metadata:main AS metadata

FROM local AS local-dogfood
COPY --link --from=regclient / /
COPY --link --from=jq / /
COPY --link --from=yq / /
COPY --link --from=metadata / /var/cache/docker-setup/

FROM release AS release-dogfood
COPY --link --from=regclient / /
COPY --link --from=jq / /
COPY --link --from=yq / /
COPY --link --from=metadata / /var/cache/docker-setup/