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
    git log --tags --simplify-by-decoration --pretty="format:%cI %D" | grep "${PREVIOUS_TAG}$" | cut -d' ' -
)";
echo "Found timestamp: ${TIMESTAMP}" >&2

cat <<EOF
## Installation

\`\`\`bash
curl -sSLf https://github.com/uniget-org/cli/releases/download/${TAG}/uniget_linux_\$(uname -m).tar.gz \\
| sudo tar -xzC /usr/local/bin uniget
\`\`\`

## Signature verification

\`\`\`bash
curl -sSLfO https://github.com/uniget-org/cli/releases/download/${TAG}/uniget_linux_\$(uname -m).tar.gz
curl -sSLfO https://github.com/uniget-org/cli/releases/download/${TAG}/uniget_linux_\$(uname -m).tar.gz.pem
curl -sSLfO https://github.com/uniget-org/cli/releases/download/${TAG}/uniget_linux_\$(uname -m).tar.gz.sig
cosign verify-blob uniget_linux_\$(uname -m).tar.gz \\
    --certificate uniget_linux_\$(uname -m).tar.gz.pem \\
    --signature uniget_linux_\$(uname -m).tar.gz.sig \\
    --certificate-identity 'https://github.com/uniget-org/cli/.github/workflows/release.yml@refs/tags/${TAG}' \\
    --certificate-oidc-issuer https://token.actions.githubusercontent.com
\`\`\`
EOF

echo
echo "## Bugfixes (since ${PREVIOUS_TAG})"
echo
gh issue list \
    --search="state:closed closed:>${TIMESTAMP} label:bug -label:wontfix" \
    --json=number,title,url \
    --template='{{range .}}- {{.title}} ([#{{.number}}]({{.url}})){{"\n"}}{{end}}'
git log --after=${TIMESTAMP} --pretty=format:'- %s [%h](https://github.com/uniget-org/cli/commit/%H)' \
| grep "^- fix" \
| grep -v "^- fix(deps)" \
|| true

echo
echo "## Features (since ${PREVIOUS_TAG})"
echo
gh issue list \
    --search="state:closed closed:>${TIMESTAMP} label:enhancement -label:wontfix" \
    --json=number,title,url \
    --template='{{range .}}- {{.title}} ([#{{.number}}]({{.url}})){{"\n"}}{{end}}'
git log --after=${TIMESTAMP} --pretty=format:'- %s [%h](https://github.com/uniget-org/cli/commit/%H)' \
| grep "^- feat" \
|| true

echo
echo "## Dependency updates (since ${PREVIOUS_TAG})"
echo
gh pr list \
    --state=merged \
    --search="merged:>${TIMESTAMP} label:type/renovate" \
    --json=number,title,url \
    --template='{{range .}}- {{.title}} ([#{{.number}}]({{.url}})){{"\n"}}{{end}}'
git log --after=${TIMESTAMP} --pretty=format:'- %s [%h](https://github.com/uniget-org/cli/commit/%H) (%an)' \
| grep ^chore \
| grep -v 'renovate' \
|| true

echo
cat <<EOF
## Full Changelog (since ${PREVIOUS_TAG})

[Compare with previous release](https://github.com/uniget-org/cli/compare/${PREVIOUS_TAG}...${TAG})
EOF
