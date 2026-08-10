#!/bin/bash
set -o errexit -o pipefail

unset UNIGET_USER

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

echo "-----------------------------"
uniget --version

echo "-----------------------------"
TEMP_DIR=$(mktemp -d)
echo "Using temp dir: ${TEMP_DIR}"
trap "find $HOME $TEMP_DIR -type f; rm -rf $TEMP_DIR" EXIT

echo "-----------------------------"
echo "Testing download of metadata"
uniget --prefix=${TEMP_DIR} update
check_file "${TEMP_DIR}/var/cache/uniget/metadata.json" "Metadata" || exit 1

echo "-----------------------------"
echo "Testing install and uninstall of gojq"
uniget --prefix=${TEMP_DIR} install gojq
check_dir "${TEMP_DIR}/var/cache/uniget/gojq" "Marker file" || exit 1
check_file "${TEMP_DIR}/var/lib/uniget/manifests/gojq.json" "Manifest" || exit 1
check_file "${TEMP_DIR}/var/lib/uniget/manifests/gojq.txt" "File list" || exit 1
"${TEMP_DIR}/usr/local/bin/gojq" --version || exit 1
uniget --prefix=${TEMP_DIR} uninstall gojq
check_dir "${TEMP_DIR}/var/cache/uniget/gojq" "Marker file" && exit 1
check_file "${TEMP_DIR}/var/lib/uniget/manifests/gojq.json" "Manifest" && exit 1
check_file "${TEMP_DIR}/var/lib/uniget/manifests/gojq.txt" "File list" && exit 1

echo "-----------------------------"
echo "Testing install of yq"
uniget --prefix=${TEMP_DIR} --target=usr install yq
check_dir "${TEMP_DIR}/var/cache/uniget/yq" "Marker file" || exit 1
check_file "${TEMP_DIR}/var/lib/uniget/manifests/yq.json" "Manifest" || exit 1
check_file "${TEMP_DIR}/var/lib/uniget/manifests/yq.txt" "File list" || exit 1
uniget --prefix=${TEMP_DIR} --target=usr version yq || exit 1

echo "-----------------------------"
echo "Testing inspection of jq"
uniget inspect --raw jq | grep "^-rwxr-xr-x bin/jq$" || exit 1

echo "-----------------------------"
echo "Testing install in user context"
uniget --user debug
uniget --user update --quiet
check_file "${HOME}/.cache/uniget/metadata.json" "Metadata" || exit 1
uniget --user install gojq
check_dir "${HOME}/.cache/uniget/gojq" "Marker file" || exit 1
check_file "${HOME}/.local/state/uniget/manifests/gojq.json" "Manifest" || exit 1
check_file "${HOME}/.local/state/uniget/manifests/gojq.txt" "File list" || exit 1
uniget --user uninstall gojq

echo "-----------------------------"
echo "Testing hooks"
mkdir -p \
    ${TEMP_DIR}/etc/uniget/hooks/pre-install.d \
    ${TEMP_DIR}/etc/uniget/hooks/post-install.d \
    ${TEMP_DIR}/etc/uniget/hooks/pre-uninstall.d \
    ${TEMP_DIR}/etc/uniget/hooks/post-uninstall.d
cat >${TEMP_DIR}/etc/uniget/hooks/pre-install.d/test.sh <<EOF
#!/bin/bash

touch /var/log/uniget-hook-pre-install
EOF
cat >${TEMP_DIR}/etc/uniget/hooks/post-install.d/test.sh <<EOF
#!/bin/bash

touch /var/log/uniget-hook-post-install
EOF
cat >${TEMP_DIR}/etc/uniget/hooks/pre-uninstall.d/test.sh <<EOF
#!/bin/bash

touch /var/log/uniget-hook-pre-uninstall
EOF
cat >${TEMP_DIR}/etc/uniget/hooks/post-uninstall.d/test.sh <<EOF
#!/bin/bash

touch /var/log/uniget-hook-post-uninstall
EOF
chmod +x \
    ${TEMP_DIR}/etc/uniget/hooks/pre-install.d/test.sh \
    ${TEMP_DIR}/etc/uniget/hooks/post-install.d/test.sh \
    ${TEMP_DIR}/etc/uniget/hooks/pre-uninstall.d/test.sh \
    ${TEMP_DIR}/etc/uniget/hooks/post-uninstall.d/test.sh
uniget --prefix=${TEMP_DIR} debug
uniget --prefix=${TEMP_DIR} install jq
check_file "/var/log/uniget-hook-pre-install" || exit 1
check_file "/var/log/uniget-hook-post-install" || exit 1
check_file "/var/log/uniget-hook-pre-uninstall" && exit 1
check_file "/var/log/uniget-hook-post-uninstall" && exit 1
rm -f \
    "/var/log/uniget-hook-pre-install" \
    "/var/log/uniget-hook-post-install" \
    "/var/log/uniget-hook-pre-uninstall" \
    "/var/log/uniget-hook-post-uninstall"
uniget --prefix=${TEMP_DIR} uninstall jq
check_file "/var/log/uniget-hook-pre-install" && exit 1
check_file "/var/log/uniget-hook-post-install" && exit 1
check_file "/var/log/uniget-hook-pre-uninstall" || exit 1
check_file "/var/log/uniget-hook-post-uninstall" || exit 1
echo "-----------------------------"
echo "All tests passed successfully"
exit 0
