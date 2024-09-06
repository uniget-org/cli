#!/bin/bash

base="$( dirname "$( readlink -f "$0" )" )"

docker compose \
    --project-directory "${base}" \
    down