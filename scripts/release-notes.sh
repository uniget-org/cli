#!/bin/bash
set -o errexit -o pipefail

TAG="$(
    git tag \
    | sort -V \
    | tail -n 1
)"
PREVIOUS_TAGS="$(
    git tag --list v* \
    | sort -V \
    | head -n -1 \
    | sort -Vr
)"

for PREVIOUS_TAG in ${PREVIOUS_TAGS}; do
    echo "Testing previous tag: ${PREVIOUS_TAG}" >&2

    if \
        TIMESTAMP="$(
            gh release view "${PREVIOUS_TAG}" --json=publishedAt --template='{{.publishedAt}}'
        )"; then
        break
    fi
done
if test -z "${TIMESTAMP}"; then
    echo "ERROR: Unable to find previous release of ${TAG}. Candidates: ${PREVIOUS_TAGS}." >&2
    exit 1
fi
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
echo "## Bugfixes"
echo
gh issue list \
    --search="state:closed closed:>${TIMESTAMP} label:bug -label:wontfix" \
    --json=number,title,url \
    --template='{{range .}}- {{.title}} ({{hyperlink .url (printf "#%v" .number)}}){{"\n"}}{{end}}'

echo
echo "## Features"
echo
gh issue list \
    --search="state:closed closed:>${TIMESTAMP} label:enhancement -label:wontfix" \
    --json=number,title,url \
    --template='{{range .}}- {{.title}} ({{hyperlink .url (printf "#%v" .number)}}){{"\n"}}{{end}}'

echo
echo "## Dependency updates"
echo
gh pr list \
    --state=merged \
    --search="merged:>${TIMESTAMP} label:type/renovate" \
    --json=number,title,url \
    --template='{{range .}}- {{.title}} ({{hyperlink .url (printf "#%v" .number)}}){{"\n"}}{{end}}'

echo
cat <<EOF
## Full Changelog

[Compare with previous release](https://github.com/uniget-org/cli/compare/${PREVIOUS_TAG}...${TAG})
EOF
