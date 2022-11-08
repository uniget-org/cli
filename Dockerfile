#syntax=docker/dockerfile:1.4.3

FROM ubuntu:22.04 AS base
RUN <<EOF
apt-get update
apt-get -y install --no-install-recommends \
    curl \
    ca-certificates \
    bsdextrautils
EOF

FROM base AS dev
RUN <<EOF
apt-get update
apt-get -y install --no-install-recommends \
    make
EOF
WORKDIR /src
COPY . .

FROM base AS local
COPY docker-setup /usr/local/bin/
COPY tools/Dockerfile.template /var/cache/docker-setup/

FROM base AS release
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