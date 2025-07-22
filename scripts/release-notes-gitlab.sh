#!/bin/bash
set -o errexit -o pipefail

export NO_COLOR=true

TAG="$(
    git tag \
    | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+' \
    | sort -V \
    | tail -n 1
)"
if test -z "${PREVIOUS_TAG}"; then
    PREVIOUS_TAG="$(
        git tag --list v* \
        | grep -v -- - \
        | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+' \
        | sort -V \
        | head -n -1 \
        | sort -Vr \
        | head -n 1
    )"
fi
echo "Creating release notes for ${PREVIOUS_TAG} -> ${TAG}" >&2

TIMESTAMP="$(
    glab api "projects/:id/releases/${PREVIOUS_TAG}" | jq -r .released_at
)";
echo "Found timestamp: ${TIMESTAMP}" >&2

cat <<EOF
## Installation

\`\`\`bash
curl -sSLf https://gitlab.com/uniget-org/cli-build-test/-/releases/${TAG}/downloads/uniget_Linux_\$(uname -m).tar.gz \\
| sudo tar -xzC /usr/local/bin uniget
\`\`\`

## Signature verification

\`\`\`bash
curl -sSLfO https://gitlab.com/uniget-org/cli-build-test/-/releases/${TAG}/downloads/uniget_Linux_\$(uname -m).tar.gz
curl -sSLfO https://gitlab.com/uniget-org/cli-build-test/-/releases/${TAG}/downloads/uniget_Linux_\$(uname -m).tar.gz.pem
curl -sSLfO https://gitlab.com/uniget-org/cli-build-test/-/releases/${TAG}/downloads/uniget_Linux_\$(uname -m).tar.gz.sig
cosign verify-blob uniget_linux_\$(uname -m).tar.gz \\
    --certificate uniget_linux_\$(uname -m).tar.gz.pem \\
    --signature uniget_linux_\$(uname -m).tar.gz.sig \\
    --certificate-identity 'https://gitlab.com/uniget-org/cli-build-test//.gitlab-ci.yml@refs/tags/${TAG}' \\
    --certificate-oidc-issuer https://gitlab.com
\`\`\`
EOF

echo
echo "## Bugfixes (since ${PREVIOUS_TAG})"
echo
glab api "projects/:id/issues?state=closed&updated_after=${TIMESTAMP}&labels=bug&not[labels]=wontfix" \
| jq -r '.[] | "\(.title) ([\(.id)](\(.url)))"'
git log --after=${TIMESTAMP} --pretty=format:'- %s [%h](https://github.com/uniget-org/cli/commit/%H)' \
| grep "^- fix" \
| grep -v "^- fix(deps)" \
|| true

echo
echo "## Features (since ${PREVIOUS_TAG})"
echo
glab api "projects/:id/issues?state=closed&updated_after=${TIMESTAMP}&labels=enhancement&not[labels]=wontfix" \
| jq -r '.[] | "\(.title) ([\(.id)](\(.url)))"'
git log --after=${TIMESTAMP} --pretty=format:'- %s [%h](https://github.com/uniget-org/cli/commit/%H)' \
| grep "^- feat" \
|| true

echo
echo "## Dependency updates (since ${PREVIOUS_TAG})"
echo
glab api "projects/:id/merge_requests?state=merged&updated_after=2025-07-22T00:00:00Z&labels=type/renovate" \
| jq -r '.[] | "\(.title) ([\(.id)](\(.url)))"'
git log --after=${TIMESTAMP} --pretty=format:'- %s [%h](https://github.com/uniget-org/cli/commit/%H) (%an)' \
| grep ^chore \
| grep -v 'renovate' \
|| true

echo
cat <<EOF
## Full Changelog (since ${PREVIOUS_TAG})

[Compare with previous release](https://gitlab.com/uniget-org/cli-build-test/-/compare/${PREVIOUS_TAG}...${TAG})
EOF
