#!/bin/bash
set -o errexit

###################################
# Check prerequisites
# - Docker must be present
# - Docker must be at least 23.0.0
#
if ! type docker >/dev/null 2>&1; then
    echo "ERROR: Docker is required for this script."
    exit 1
fi
DOCKER_VERSION="$(
    { echo "23.0.0"; docker version --format "{{json .}}" | jq --raw-output '.Client.Version'; } \
    | sort -V \
    | head -n 1
)"
if test "${DOCKER_VERSION}" != "23.0.0"; then
    echo "ERROR: Script requires at least Docker 23.0.0"
    exit 1
fi
if ! type regctl >/dev/null 2>&1; then
    echo "ERROR: regclient is required for this script."
    exit 1
fi

######
# XXX
#
arch="$(uname -m)"
case "${arch}" in
    x86_64)
        alt_arch=amd64
        ;;
    aarch64)
        alt_arch=arm64
        ;;
    *)
        echo "ERROR: Unsupported architecture (${arch})."
        exit 1
        ;;
esac

######
# XXX
#
NAME_BASE=nicholasdille/docker-setup
NAME_SOURCE="${NAME_BASE}/base:latest"
IMAGES=(
    "ghcr.io/nicholasdille/docker-setup/kubectl:main"
    "ghcr.io/nicholasdille/docker-setup/helm:main"
    "ghcr.io/nicholasdille/docker-setup/kubeswitch:main"
)
REGISTRY=127.0.0.1:5000
FULL_NAME_SOURCE="${REGISTRY}/${NAME_SOURCE}"

######
# XXX
#
if test "$(docker container ls --filter name=registry | wc -l)" -eq 1; then
    docker container run --detach --name registry --publish "${REGISTRY}:5000" registry
fi
regctl registry set "${REGISTRY}" --tls disabled

######
# XXX
#
TEMP="$(mktemp -d)"
function cleanup() {
    test -d "${TEMP}" && rm -rf "${TEMP}"
}
trap cleanup EXIT

######
# XXX
#
echo "Building base image"
docker image build \
    --cache-from "${FULL_NAME_SOURCE}" \
    --tag "${FULL_NAME_SOURCE}" \
    --push \
    @base

######
# XXX
#
echo "Fetching manifest (list)"
regctl manifest get "${FULL_NAME_SOURCE}" --platform "linux/${alt_arch}" --format raw-body \
>"${TEMP}/manifest.json"
echo "Fetching image config"
cat "${TEMP}/manifest.json" \
| jq --raw-output '.config.digest' \
| xargs -I{} regctl blob get "${FULL_NAME_SOURCE}" {} \
>"${TEMP}/config.json"

######
# XXX
#
for IMAGE in ${IMAGES[*]}; do
    echo
    echo "Image ${IMAGE}"

    ######
    # XXX
    #
    REPOSITORY_TAG="$(
        echo "${IMAGE}" \
        | cut -d/ -f2-
    )"
    LOCAL_NAME="${REGISTRY}/${REPOSITORY_TAG}"
    echo "+ Name with local registry: ${LOCAL_NAME}"

    ######
    # XXX
    #
    echo "+ Copy image to local registry"
    regctl image copy "${IMAGE}" "${REGISTRY}/${REPOSITORY_TAG}"

    ######
    # XXX
    #
    echo "+ Get manifest and config"
    MANIFEST_FILE="${TEMP}/image_manifest.json"
    regctl manifest get "${LOCAL_NAME}" --platform "linux/${alt_arch}" --format raw-body >"${MANIFEST_FILE}"
    CONFIG_DIGEST="$(cat "${MANIFEST_FILE}" | jq --raw-output '.config.digest')"
    echo "  Config digest: ${CONFIG_DIGEST}"
    CONFIG="$(regctl blob get "${LOCAL_NAME}" "${CONFIG_DIGEST}")"
    
    ######
    # XXX
    #
    echo "+ Get layer info"
    LAYER_DIGEST="$(jq --raw-output '.layers[-1].digest' "${MANIFEST_FILE}")"
    LAYER_TYPE="$(jq --raw-output '.layers[-1].mediaType' "${MANIFEST_FILE}")"
    LAYER_SIZE="$(jq --raw-output '.layers[-1].size' "${MANIFEST_FILE}")"
    echo "  Digest: ${LAYER_DIGEST}"
    echo "  Size: ${LAYER_SIZE}"
    LAYER_COMMAND=$(jq --raw-output '.history[-1]' <<<"${CONFIG}")
    LAYER_DIFF=$(jq --raw-output '.rootfs.diff_ids[-1]' <<<"${CONFIG}")

    ######
    # XXX
    #
    echo "+ Mount blob to target repository"
    "${HOME}/.local/bin/regctl" blob copy "${LOCAL_NAME}" "${FULL_NAME_SOURCE}" "${LAYER_DIGEST}"

    ######
    # XXX
    #
    echo "+ Add layer to target manifest"
    mv "${TEMP}/manifest.json" "${TEMP}/manifest.json.bak"
    cat "${TEMP}/manifest.json.bak" \
    | jq '.layers += [{"mediaType": $type, "size": $size | tonumber, "digest": $digest}]' \
            --arg type "${LAYER_TYPE}" \
            --arg digest "${LAYER_DIGEST}" \
            --arg size "${LAYER_SIZE}" \
    >"${TEMP}/manifest.json"

    ######
    # XXX
    #
    echo "+ Add config to target config"
    mv "${TEMP}/config.json" "${TEMP}/config.json.bak"
    cat "${TEMP}/config.json.bak" \
    | jq --arg command "${LAYER_COMMAND}" '.history += [$command | fromjson]' \
    | jq --arg diff "${LAYER_DIFF}" '.rootfs.diff_ids += [$diff]' \
    >"${TEMP}/config.json"

done
echo

######
# XXX
#
echo "Upload config"
NEW_CONFIG_DIGEST="$(
    cat "${TEMP}/config.json" \
    | regctl blob put "${FULL_NAME_SOURCE}"
)"

######
# XXX
#
echo "Update and upload manifest"
cat "${TEMP}/manifest.json" \
| jq --arg digest "${NEW_CONFIG_DIGEST}" '.config.digest = $digest' \
| regctl manifest put "${FULL_NAME_SOURCE}"
