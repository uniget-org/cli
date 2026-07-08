#syntax=docker/dockerfile:1

FROM ghcr.io/uniget-org/tools/uniget:0.27.10@sha256:336f29dfddfcd59efc259dee59573c6e436388036f1965b2f63d381eb00e9db2 AS uniget
FROM ghcr.io/uniget-org/tools/git:2.54.0 AS git
FROM ghcr.io/uniget-org/tools/go:1.26.4 AS go
FROM ghcr.io/uniget-org/tools/make:4.4.1 AS make

FROM ubuntu:26.04@sha256:b7f48194d4d8b763a478a621cdc81c27be222ba2206ca3ca6bc42b49685f3d9e
COPY --link --from=uniget / /usr/local
COPY --link --from=git / /usr/local
COPY --link --from=go / /usr/local
COPY --link --from=make / /usr/local