#syntax=docker/dockerfile:1

FROM ghcr.io/uniget-org/tools/uniget:0.29.2@sha256:70b551f090a7d0aef6e39404b52c414a50fdc4c7c300685814eca2533bbe99a5 AS uniget
FROM ghcr.io/uniget-org/tools/git:2.55.0@sha256:693cfe5a38d9731ea967ae55d874f5b5e2cb37edd03052845721e60fe70d2cea AS git
FROM ghcr.io/uniget-org/tools/go:1.27.1@sha256:7165d7ed4e6a0c73031fdd3684755b2aea3f2c91d3b648c2c9bef54436a78032 AS go
FROM ghcr.io/uniget-org/tools/make:4.4.1 AS make

FROM ubuntu:26.04@sha256:da6fc2be547864451aa253836dd926da33623312df4a9a243e35dc877c378a78
COPY --link --from=uniget / /usr/local
COPY --link --from=git / /usr/local
COPY --link --from=go / /usr/local
COPY --link --from=make / /usr/local