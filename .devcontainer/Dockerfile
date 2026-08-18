#syntax=docker/dockerfile:1

FROM ghcr.io/uniget-org/tools/uniget:0.28.10@sha256:7b6ee317304d9b1c5b3361a39789be161006182ff5ffab4f958a4f6b36a64d99 AS uniget
FROM ghcr.io/uniget-org/tools/git:2.54.0 AS git
FROM ghcr.io/uniget-org/tools/go:1.26.5@sha256:38928173f0f2291a7faa1f971b4bf3d12e9a825886bcd65aa726ca09b1697903 AS go
FROM ghcr.io/uniget-org/tools/make:4.4.1 AS make

FROM ubuntu:26.04@sha256:678c6550cc43645e08669028bc177f50be4e7c5b8cca677067b1914d4afc7a03
COPY --link --from=uniget / /usr/local
COPY --link --from=git / /usr/local
COPY --link --from=go / /usr/local
COPY --link --from=make / /usr/local