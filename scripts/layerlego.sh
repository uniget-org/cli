#!/bin/bash
set -o errexit

###################################
# Check prerequisites
# - Docker must be present
# - Docker must be at least 23.0.0
#
echo "Checking prerequisites"
if ! type docker >/dev/null 2>&1; then
    echo "ERROR: Docker is required for this script."
    exit 1
fi
MINIMUM_DOCKER_VERSION=23.0.0
DOCKER_VERSION="$(
    { echo "${MINIMUM_DOCKER_VERSION}"; docker version --format "{{json .}}" | jq --raw-output '.Client.Version'; } \
    | sort -V \
    | head -n 1
)"
if test "${DOCKER_VERSION}" != "${MINIMUM_DOCKER_VERSION}"; then
    echo "ERROR: Script requires at least Docker v${MINIMUM_DOCKER_VERSION}"
    exit 1
fi
if ! type regctl >/dev/null 2>&1; then
    echo "ERROR: regclient is required for this script."
    exit 1
fi
MINIMUM_REGCTL_VERSION=0.4.8
REGCTL_VERSION="$(
    { echo "${MINIMUM_REGCTL_VERSION}"; regctl version; } \
    | sort -V \
    | head -n 1
)"
if test "${REGCTL_VERSION}" != "${MINIMUM_REGCTL_VERSION}"; then
    echo "ERROR: Script requires at least regctl v${MINIMUM_REGCTL_VERSION}"
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
echo "Populating variables"
NAME_BASE=nicholasdille/docker-setup
NAME_SOURCE="${NAME_BASE}/base:latest"
TOOLS_IMAGE_PREFIX="ghcr.io/nicholasdille/docker-setup/"
TOOLS_IMAGE_SUFFIX=":main"
TOOLS=(
    "dotnet"
    "powershell"
    "scala"
    "nodejs"
    "npm"
    "nvm"
    "grunt"
    "serverless"
    "newman"
    "yarn"
    "nx"
    "az"
    "pipx"
    "aws2"
    "dependency-check"
    "docker"
    "buildx"
    "docker-compose"
    "docker-machine"
    "jenkins-remoting"
    "cosign"
    "rekor"
    "trivy"
    "syft"
    "grype"
    "go"
    "kubectl"
    "helm"
    "sops"
    "terraform"
    "terragrunt"
    "sonar-scanner"
    "jf"
    "oc"
    "jaxb"
    "gradle"
    "maven"
)
IMAGES=()
for TOOL in ${TOOLS[*]}; do
    IMAGES+=( "${TOOLS_IMAGE_PREFIX}${TOOL}${TOOLS_IMAGE_SUFFIX}" )
done
for IMAGE in ${IMAGES[*]}; do
    echo "Checking ${IMAGE}"
    if ! regctl manifest head "${IMAGE}" >/dev/null; then
        echo "ERROR: Image ${IMAGE} does not exist."
        exit 1
    fi
done \
| pv --name "Checking images" --progress --timer --eta --line-mode --size "${#IMAGES[*]}" >/dev/null
REGISTRY=127.0.0.1:5000
FULL_NAME_SOURCE="${REGISTRY}/${NAME_SOURCE}"

######
# XXX
#
echo "Starting registry"
if test "$(docker container ls --all --filter status=exited --filter name=registry | wc -l)" -gt 1; then
    docker container rm registry
fi
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
if ! test -f Dockerfile.base; then
    echo "ERROR: Script expects a Dockerfile.base in the current directory (${PWD})."
    exit 1
fi
echo "Building base image"
docker image build \
    --file Dockerfile.base \
    --cache-from "${FULL_NAME_SOURCE}" \
    --tag "${FULL_NAME_SOURCE}" \
    --push \
    .

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
    echo "Image ${IMAGE}"
    REPOSITORY_TAG="$(
        echo "${IMAGE}" \
        | cut -d/ -f2-
    )"
    regctl image copy "${IMAGE}" "${REGISTRY}/${REPOSITORY_TAG}" --platform linux/amd64
done \
| pv --name "Importing images" --progress --timer --eta --line-mode --size "${#IMAGES[*]}" >/dev/null

######
# XXX
#
SECONDS=0
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
    regctl blob copy "${LOCAL_NAME}" "${FULL_NAME_SOURCE}" "${LAYER_DIGEST}"

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
echo "Upload config for ${FULL_NAME_SOURCE}"
NEW_CONFIG_DIGEST="$(
    cat "${TEMP}/config.json" \
    | regctl blob put "${FULL_NAME_SOURCE}"
)"

######
# XXX
#
echo "Update and upload manifest for ${FULL_NAME_SOURCE}"
cat "${TEMP}/manifest.json" \
| jq --arg digest "${NEW_CONFIG_DIGEST}" '.config.digest = $digest' \
| regctl manifest put "${FULL_NAME_SOURCE}"

######
# XXX
#
echo
echo "Finished after ${SECONDS} second(s)"