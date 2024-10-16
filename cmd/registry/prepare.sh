#!/bin/bash
set -o errexit

echo "### Fetching jq version from uniget metadata"
METADATA_DIGEST="$(
    regctl manifest get ghcr.io/uniget-org/tools/metadata:main --platform local --format raw-body \
    | jq --raw-output '.layers[0].digest'
)"
JQ_VERSION="$(
    regctl blob get ghcr.io/uniget-org/tools/metadata "${METADATA_DIGEST}" \
    | tar --extract --gzip --to-stdout metadata.json \
    | jq --raw-output '.tools[] | select(.name == "jq") | .version'
)"
echo "    jq version: ${JQ_VERSION}"

echo "### Downloading and extracting jq image"
regctl image export "ghcr.io/uniget-org/tools/jq:${JQ_VERSION}" \
| tar --extract

echo "### Parsing index.json to find image index"
INDEX_DIGEST="$(
    jq --raw-output '.manifests[0].digest' index.json \
    | cut -d: -f2
)"
echo "    image index digest: ${INDEX_DIGEST}"
sed -i "s|index.json|blobs/sha256/${INDEX_DIGEST}|g" main.go

echo "### Parsing image index to find manifest for amd64"
MANIFEST_AMD64_DIGEST="$(
    jq --raw-output '.manifests[] | select(.platform.architecture == "amd64") | .digest' "blobs/sha256/${INDEX_DIGEST}" \
    | cut -d: -f2
)"
echo "    manifest digest for amd64: ${MANIFEST_AMD64_DIGEST}"
sed -i "s|manifest_amd64.json|blobs/sha256/${MANIFEST_AMD64_DIGEST}|g" main.go
echo "    replaced manifest digest for amd64"

echo "### Parsing image index to find manifest for arm64"
MANIFEST_ARM64_DIGEST="$(
    jq --raw-output '.manifests[] | select(.platform.architecture == "arm64") | .digest' "blobs/sha256/${INDEX_DIGEST}" \
    | cut -d: -f2
)"
echo "    manifest digest for arm64: ${MANIFEST_ARM64_DIGEST}"
sed -i "s|manifest_arm64.json|blobs/sha256/${MANIFEST_ARM64_DIGEST}|g" main.go
echo "    replaced manifest digest for arm64"

echo "### Parsing manifest for amd64 to find config"
CONFIG_AMD64_DIGEST="$(
    jq --raw-output '.config.digest' "blobs/sha256/${MANIFEST_AMD64_DIGEST}" \
    | cut -d: -f2
)"
echo "    config digest for amd64: ${CONFIG_AMD64_DIGEST}"
sed -i "s|config_amd64.json|blobs/sha256/${CONFIG_AMD64_DIGEST}|g" main.go
echo "    replaced config digest for amd64"

echo "### Parsing manifest for arm64 to find config"
CONFIG_ARM64_DIGEST="$(
    jq --raw-output '.config.digest' "blobs/sha256/${MANIFEST_ARM64_DIGEST}" \
    | cut -d: -f2
)"
echo "    config digest for arm64: ${CONFIG_ARM64_DIGEST}"
sed -i "s|config_arm64.json|blobs/sha256/${CONFIG_ARM64_DIGEST}|g" main.go
echo "    replaced config digest for arm64"

echo "### Parsing manifest for amd64 to layer digest"
LAYER_AMD64_DIGEST="$(
    jq --raw-output '.layers[0].digest' "blobs/sha256/${MANIFEST_AMD64_DIGEST}" \
    | cut -d: -f2
)"
echo "    layer digest for amd64: ${LAYER_AMD64_DIGEST}"
sed -i "s|layer_amd64.tar.gz|blobs/sha256/${LAYER_AMD64_DIGEST}|g" main.go
echo "    replaced layer digest for amd64"

echo "### Parsing manifest for arm64 to layer digest"
LAYER_ARM64_DIGEST="$(
    jq --raw-output '.layers[0].digest' "blobs/sha256/${MANIFEST_ARM64_DIGEST}" \
    | cut -d: -f2
)"
echo "    layer digest for arm64: ${LAYER_ARM64_DIGEST}"
sed -i "s|layer_arm64.tar.gz|blobs/sha256/${LAYER_ARM64_DIGEST}|g" main.go
echo "    replaced layer digest for arm64"
