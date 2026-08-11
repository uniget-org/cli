#syntax=docker/dockerfile:1.26.0
#check=skip=SecretsUsedInArgOrEnv

FROM registry.gitlab.com/uniget-org/images/ubuntu:24.04 AS ubuntu-uniget
ARG version
RUN <<EOF
curl --silent --show-error --location --fail \
    "https://gitlab.com/uniget-org/cli/-/releases/v${version}/downloads/uniget_Linux_$(uname -m).tar.gz" \
| tar --extract --gzip --directory=/usr/local/bin uniget
useradd --shell=/bin/bash --create-home bob
echo "export UNIGET_USER=1" >>/home/bob/.bashrc
echo "export PATH=\${HOME}/.local/bin:${PATH}" >>/home/bob/.bashrc
EOF
LABEL \
    org.opencontainers.image.source="https://gitlab.com/uniget-org/cli" \
    org.opencontainers.image.title="uniget CLI" \
    org.opencontainers.image.description="The universal installer and updater for (container) tools" \
    org.opencontainers.image.version="${version}"