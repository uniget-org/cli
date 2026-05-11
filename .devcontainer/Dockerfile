#syntax=docker/dockerfile:1

FROM ghcr.io/uniget-org/tools/uniget:latest AS uniget
FROM ghcr.io/uniget-org/tools/git:latest AS git
FROM ghcr.io/uniget-org/tools/go:latest AS go
FROM ghcr.io/uniget-org/tools/make:latest AS make

FROM ubuntu:26.04@sha256:f3d28607ddd78734bb7f71f117f3c6706c666b8b76cbff7c9ff6e5718d46ff64
COPY --link --from=ghcr.io/uniget-org/tools/uniget:latest / /usr/local
COPY --link --from=ghcr.io/uniget-org/tools/git:latest / /usr/local
COPY --link --from=ghcr.io/uniget-org/tools/go:latest / /usr/local
COPY --link --from=ghcr.io/uniget-org/tools/make:latest / /usr/local