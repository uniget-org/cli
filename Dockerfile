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
 && update-alternatives --set iptables /usr/sbin/iptables-legacy

FROM base AS docker-setup
COPY docker/entrypoint.sh /
COPY docker-setup.sh /usr/local/bin/docker-setup
RUN chmod +x /usr/local/bin/docker-setup \
 && mkdir -p /var/cache/docker-setup
COPY tools.json /var/cache/docker-setup/
COPY contrib /var/cache/docker-setup/contrib
ENTRYPOINT [ "bash", "/entrypoint.sh" ]