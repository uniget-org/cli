#syntax=docker/dockerfile:1

FROM ghcr.io/uniget-org/tools/uniget:latest AS uniget
FROM ghcr.io/uniget-org/tools/git:latest AS git
FROM ghcr.io/uniget-org/tools/go:latest AS go
FROM ghcr.io/uniget-org/tools/make:latest AS make

FROM ubuntu:24.04
COPY --link --from=ghcr.io/uniget-org/tools/uniget:latest / /usr/local
COPY --link --from=ghcr.io/uniget-org/tools/git:latest / /usr/local
COPY --link --from=ghcr.io/uniget-org/tools/go:latest / /usr/local
COPY --link --from=ghcr.io/uniget-org/tools/make:latest / /usr/local