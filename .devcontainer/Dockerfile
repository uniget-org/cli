#syntax=docker/dockerfile:1

FROM ghcr.io/uniget-org/tools/uniget:0.28.10@sha256:7b6ee317304d9b1c5b3361a39789be161006182ff5ffab4f958a4f6b36a64d99 AS uniget
FROM ghcr.io/uniget-org/tools/git:2.54.0 AS git
FROM ghcr.io/uniget-org/tools/go:1.26.6@sha256:6b71128f934990b726ec836bbd6598b439bc9f3621d1a1625c788822cede1671 AS go
FROM ghcr.io/uniget-org/tools/make:4.4.1 AS make

FROM ubuntu:26.04@sha256:2260313b31c8c011cd2eebe728008efac1b3982be73eb71348ea2648d2c0e09b
COPY --link --from=uniget / /usr/local
COPY --link --from=git / /usr/local
COPY --link --from=go / /usr/local
COPY --link --from=make / /usr/local