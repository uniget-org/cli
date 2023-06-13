#syntax=docker/dockerfile:1.5.2

ARG base=ubuntu-22.04
FROM ghcr.io/nicholasdille/docker-setup/docker-setup:main AS docker-setup

FROM ubuntu:22.04 AS ubuntu
COPY --from=docker-setup /usr/local/bin/docker-setup /usr/local/bin/docker-setup

FROM debian:11.7@sha256:1e5f2d70c9441c971607727f56d0776fb9eecf23cd37b595b26db7a974b2301d AS debian
COPY --from=docker-setup /usr/local/bin/docker-setup /usr/local/bin/docker-setup

FROM alpine:3.18@sha256:02bb6f428431fbc2809c5d1b41eab5a68350194fb508869a33cb1af4444c9b11 AS alpine
COPY --from=docker-setup /usr/local/bin/docker-setup /usr/local/bin/docker-setup