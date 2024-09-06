#!/bin/bash

base="$( dirname "$( readlink -f "$0" )" )"

registry=127.0.0.1:5000
repository=uniget-org/tools
tool=test
version=1.0.0

regctl registry set "${registry}" --tls=disabled

docker compose \
    --project-directory "${base}" \
    up -d

docker build "${base}" \
    --file "${base}/Dockerfile" \
    --tag "${registry}/${repository}/${tool}:${version}" \
    --push
