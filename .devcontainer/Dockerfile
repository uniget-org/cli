#syntax=docker/dockerfile:1

FROM ghcr.io/uniget-org/tools/uniget:0.29.3@sha256:5bca1fb240324c6dab58d4235e6eea4a866edb614a81d9e467b3494264c62158 AS uniget
FROM ghcr.io/uniget-org/tools/git:2.56.0@sha256:8aefdab2251e4b126db5a4e8733c7420ec75194cfda002f3571e8dab29353893 AS git
FROM ghcr.io/uniget-org/tools/go:1.27.1@sha256:7165d7ed4e6a0c73031fdd3684755b2aea3f2c91d3b648c2c9bef54436a78032 AS go
FROM ghcr.io/uniget-org/tools/make:4.4.1 AS make

FROM ubuntu:26.04@sha256:da6fc2be547864451aa253836dd926da33623312df4a9a243e35dc877c378a78
COPY --link --from=uniget / /usr/local
COPY --link --from=git / /usr/local
COPY --link --from=go / /usr/local
COPY --link --from=make / /usr/local