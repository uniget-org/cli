#syntax=docker/dockerfile:1.5.2

ARG base=ubuntu-22.04
FROM ghcr.io/nicholasdille/docker-setup/docker-setup:main AS docker-setup

FROM ubuntu:22.04 AS ubuntu
COPY --from=docker-setup /usr/local/bin/docker-setup /usr/local/bin/docker-setup

FROM debian:11.7@sha256:1e5f2d70c9441c971607727f56d0776fb9eecf23cd37b595b26db7a974b2301d AS debian
COPY --from=docker-setup /usr/local/bin/docker-setup /usr/local/bin/docker-setup

FROM alpine:3.18@sha256:ac03b2a7eecaa3b1871d4c3971bf93dbd37ab9d69a4031b25eae3c8a9783f58a AS alpine
COPY --from=docker-setup /usr/local/bin/docker-setup /usr/local/bin/docker-setup