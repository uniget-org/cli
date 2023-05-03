#!/bin/bash
set -o errexit

docker run --interactive --tty --rm \
    --mount "type=bind,src=${HOME}/go/pkg/mod,dst=/go/pkg/mod" \
    --mount "type=bind,src=${HOME}/.cache/go-build,dst=/.cache/go-build" \
    --mount "type=bind,src=${PWD},dst=/src" \
    --workdir /src \
    golang \
        go "$@"