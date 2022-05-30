#syntax=docker/dockerfile:1.3.0

FROM ubuntu:21.04@sha256:ba394fabd516b39ccf8597ec656a9ddd7d0a2688ed8cb373ca7ac9b6fe67848f AS base
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