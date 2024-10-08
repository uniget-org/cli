#!/bin/bash
set -o errexit -o pipefail

export UNIGET_USER=false

function check_dir() {
    local dir=$1
    local name=$2

    : "${name:=Directory}"

    if test -d "${dir}"; then
        echo "${name} ${dir} exists"
    else
        echo "${name} ${dir} does not exist"
        return 1
    fi
}

function check_file() {
    local file=$1
    local name=$2

    : "${name:=File}"

    if test -f "${file}"; then
        echo "${name} ${file} exists"
    else
        echo "${name} ${file} does not exist"
        return 1
    fi
}

TEMP_DIR=$(mktemp -d)
echo "Using temp dir: ${TEMP_DIR}"
trap "find $TEMP_DIR -type f; rm -rf $TEMP_DIR" EXIT

uniget --prefix=${TEMP_DIR} update
check_file "${TEMP_DIR}/var/cache/uniget/metadata.json" "Metadata" || exit 1

uniget --prefix=${TEMP_DIR} install dummy
check_dir "${TEMP_DIR}/var/cache/uniget/dummy" "Marker file" || exit 1
check_file "${TEMP_DIR}/var/lib/uniget/manifests/dummy.json" "Manifest" || exit 1
check_file "${TEMP_DIR}/var/lib/uniget/manifests/dummy.txt" "File list" || exit 1

uniget --prefix=${TEMP_DIR} uninstall dummy
check_dir "${TEMP_DIR}/var/cache/uniget/dummy" "Marker file" && exit 1
check_file "${TEMP_DIR}/var/lib/uniget/manifests/dummy.json" "Manifest" && exit 1
check_file "${TEMP_DIR}/var/lib/uniget/manifests/dummy.txt" "File list" && exit 1

uniget --prefix=${TEMP_DIR} install gojq
"${TEMP_DIR}/usr/local/bin/gojq" --version || exit 1

uniget --prefix=${TEMP_DIR} --target=usr install yq
check_dir "${TEMP_DIR}/var/cache/uniget/yq" "Marker file" || exit 1
check_file "${TEMP_DIR}/var/lib/uniget/manifests/yq.json" "Manifest" || exit 1
check_file "${TEMP_DIR}/var/lib/uniget/manifests/yq.txt" "File list" || exit 1
uniget --prefix=${TEMP_DIR} --target=usr version yq || exit 1

uniget inspect jq | grep "bin/jq$" || exit 1

uniget --user update
test -f "${HOME}/.cache/uniget/metadata.json"
uniget --user install dummy
test -d "${HOME}/.cache/uniget/dummy"
test -f "${HOME}/.local/state/uniget/manifests/dummy.json"
test -f "${HOME}/.local/state/uniget/manifests/dummy.txt"
uniget --user uninstall dummy

echo "-----------------------------"
echo "All tests passed successfully"
exit 0
