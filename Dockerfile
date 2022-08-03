#syntax=docker/dockerfile:1.4.2

FROM ubuntu:22.04@sha256:34fea4f31bf187bc915536831fd0afc9d214755bf700b5cdb1336c82516d154e AS base
ARG DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
 && apt-get -y install --no-install-recommends \
        curl \
        ca-certificates \
        iptables \
        git \
        tzdata \
        unzip \
        ncurses-bin \
        asciinema \
        time \
        jq \
        less \
        bash-completion \
        gettext-base \
        vim-tiny \
        xz-utils \
        bsdextrautils \
 && update-alternatives --set iptables /usr/sbin/iptables-legacy

FROM base AS docker-setup

COPY docker-setup.sh /usr/local/bin/docker-setup
RUN chmod +x /usr/local/bin/docker-setup \
 && mkdir -p /var/cache/docker-setup
COPY lib /var/cache/docker-setup/lib
COPY tools.json /var/cache/docker-setup/
COPY completion/bash/docker-setup.sh /etc/bash_completion.d/

COPY docker/entrypoint.sh /
ENTRYPOINT [ "bash", "/entrypoint.sh" ]

ARG BRANCH
ARG DOCKER_SETUP_VERSION
LABEL org.opencontainers.image.source="https://github.com/nicholasdille/docker-setup" \
      org.opencontainers.image.ref.name="${BRANCH}" \
      org.opencontainers.image.title="docker-setup" \
      org.opencontainers.image.description="The container tools installer and updater" \
      org.opencontainers.image.version="${DOCKER_SETUP_VERSION}"
