#syntax=docker/dockerfile:1

FROM ghcr.io/uniget-org/tools/uniget:0.28.11@sha256:cae627fed5b69d705c4d8fa87dd302f7ba05ae76d33073481f1c6de9d90e678b AS uniget
FROM ghcr.io/uniget-org/tools/git:2.54.0 AS git
FROM ghcr.io/uniget-org/tools/go:1.27.1@sha256:7165d7ed4e6a0c73031fdd3684755b2aea3f2c91d3b648c2c9bef54436a78032 AS go
FROM ghcr.io/uniget-org/tools/make:4.4.1 AS make

FROM ubuntu:26.04@sha256:2260313b31c8c011cd2eebe728008efac1b3982be73eb71348ea2648d2c0e09b
COPY --link --from=uniget / /usr/local
COPY --link --from=git / /usr/local
COPY --link --from=go / /usr/local
COPY --link --from=make / /usr/local