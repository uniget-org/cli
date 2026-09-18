#syntax=docker/dockerfile:1.27.0
#check=skip=SecretsUsedInArgOrEnv

FROM registry.gitlab.com/uniget-org/images/ubuntu:26.04@sha256:7b6eb462bb70782bd4f510b044fb4d8ab45a5239e2b1cb7b2bdacd8cfc9fd2d2
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