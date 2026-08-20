#syntax=docker/dockerfile:1

FROM ghcr.io/uniget-org/tools/uniget:0.28.10@sha256:7b6ee317304d9b1c5b3361a39789be161006182ff5ffab4f958a4f6b36a64d99 AS uniget
FROM ghcr.io/uniget-org/tools/git:2.54.0 AS git
FROM ghcr.io/uniget-org/tools/go:1.27.0@sha256:80ef1b4892ed19361e3f5e50b4f21b7c2488fc49a7d7543713d9ac6592e865af AS go
FROM ghcr.io/uniget-org/tools/make:4.4.1 AS make

FROM ubuntu:26.04@sha256:678c6550cc43645e08669028bc177f50be4e7c5b8cca677067b1914d4afc7a03
COPY --link --from=uniget / /usr/local
COPY --link --from=git / /usr/local
COPY --link --from=go / /usr/local
COPY --link --from=make / /usr/local