#syntax=docker/dockerfile:1

FROM ghcr.io/uniget-org/tools/uniget:0.27.10@sha256:336f29dfddfcd59efc259dee59573c6e436388036f1965b2f63d381eb00e9db2 AS uniget
FROM ghcr.io/uniget-org/tools/git:2.54.0 AS git
FROM ghcr.io/uniget-org/tools/go:1.26.4 AS go
FROM ghcr.io/uniget-org/tools/make:4.4.1 AS make

FROM ubuntu:26.04@sha256:53958ec7b67c2c9355df922dd08dbf0360611f8c3cdb656875e81873db9ffdba
COPY --link --from=uniget / /usr/local
COPY --link --from=git / /usr/local
COPY --link --from=go / /usr/local
COPY --link --from=make / /usr/local