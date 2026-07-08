#syntax=docker/dockerfile:1

FROM ghcr.io/uniget-org/tools/uniget:0.27.11@sha256:1855c4eb050a4f50547099a77713d55a21e74cda75eb9b0afab573379b3fae7a AS uniget
FROM ghcr.io/uniget-org/tools/git:2.54.0 AS git
FROM ghcr.io/uniget-org/tools/go:1.26.5@sha256:38928173f0f2291a7faa1f971b4bf3d12e9a825886bcd65aa726ca09b1697903 AS go
FROM ghcr.io/uniget-org/tools/make:4.4.1 AS make

FROM ubuntu:26.04@sha256:b7f48194d4d8b763a478a621cdc81c27be222ba2206ca3ca6bc42b49685f3d9e
COPY --link --from=uniget / /usr/local
COPY --link --from=git / /usr/local
COPY --link --from=go / /usr/local
COPY --link --from=make / /usr/local