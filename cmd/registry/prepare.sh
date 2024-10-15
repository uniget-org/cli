#!/bin/bash
set -o errexit

regctl registry set --tls disabled 127.0.0.1:5000

regctl image export ghcr.io/uniget-org/tools/jq:latest >jq.tar
tar -xf jq.tar