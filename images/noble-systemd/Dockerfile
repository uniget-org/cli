#syntax=docker/dockerfile:1.27.0
#check=skip=SecretsUsedInArgOrEnv

FROM registry.gitlab.com/uniget-org/images/systemd:26.04@sha256:aff2c472db6ced9caae4999a332cd67b9db04f50c439b557c3201ab9026515b2
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
