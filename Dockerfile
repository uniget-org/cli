#syntax=docker/dockerfile:1.5.2

ARG base=ubuntu-22.04
FROM ghcr.io/nicholasdille/docker-setup/docker-setup:main AS docker-setup

FROM golang:1.20.5 AS binary
ARG version=dev
WORKDIR /go/src/github.com/nicholasdille/docker-setup
COPY go.* .
RUN go mod download
COPY . .
RUN make bin/docker-setup GO_VERSION=${version}

FROM ubuntu:22.04 AS ubuntu-test
COPY --link --from=binary /go/src/github.com/nicholasdille/docker-setup/bin/docker-setup /usr/local/bin/docker-setup

FROM ubuntu:22.04 AS ubuntu
COPY --from=docker-setup /usr/local/bin/docker-setup /usr/local/bin/docker-setup

FROM debian:11.7@sha256:1e5f2d70c9441c971607727f56d0776fb9eecf23cd37b595b26db7a974b2301d AS debian
COPY --from=docker-setup /usr/local/bin/docker-setup /usr/local/bin/docker-setup

FROM alpine:3.18@sha256:82d1e9d7ed48a7523bdebc18cf6290bdb97b82302a8a9c27d4fe885949ea94d1 AS alpine
COPY --from=docker-setup /usr/local/bin/docker-setup /usr/local/bin/docker-setup