#syntax=docker/dockerfile:1

FROM ghcr.io/uniget-org/tools/uniget:0.29.0@sha256:749756a882537fc64cfa343db7846c2db43fb7fa2de209092a9b6433756912e5 AS uniget
FROM ghcr.io/uniget-org/tools/git:2.55.0@sha256:693cfe5a38d9731ea967ae55d874f5b5e2cb37edd03052845721e60fe70d2cea AS git
FROM ghcr.io/uniget-org/tools/go:1.27.1@sha256:7165d7ed4e6a0c73031fdd3684755b2aea3f2c91d3b648c2c9bef54436a78032 AS go
FROM ghcr.io/uniget-org/tools/make:4.4.1 AS make

FROM ubuntu:26.04@sha256:9559ceb7c21e528e233e8dff26a0fb2682f4094cce06176eeb075d87a22b31de
COPY --link --from=uniget / /usr/local
COPY --link --from=git / /usr/local
COPY --link --from=go / /usr/local
COPY --link --from=make / /usr/local