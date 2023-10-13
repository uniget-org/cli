#!/bin/bash
set -o errexit -o pipefail

TAG="$(
    git tag \
    | tail -n 1
)"
PREVIOUS_TAG="$(
    git tag \
    | tail -n 2 \
    | head -n 1
)"

TIMESTAMP="$(
    gh release view "${PREVIOUS_TAG}" --json=publishedAt --template='{{.publishedAt}}'
)"

cat <<EOF
## Installation

```bash
curl -sSLf https://github.com/uniget-org/cli/releases/download/${TAG}/uniget_linux_\$(uname -m).tar.gz \
| sudo tar -xzC /usr/local/bin uniget
```

### Signature verification

```bash
curl -sSLfO https://github.com/uniget-org/cli/releases/download/${TAG}/uniget_linux_\$(uname -m).tar.gz
curl -sSLfO https://github.com/uniget-org/cli/releases/download/${TAG}/uniget_linux_\$(uname -m).tar.gz.pem
curl -sSLfO https://github.com/uniget-org/cli/releases/download/${TAG}/uniget_linux_\$(uname -m).tar.gz.sig
cosign verify-blob uniget_linux_\$(uname -m).tar.gz \
    --certificate uniget_linux_\$(uname -m).tar.gz.pem \
    --signature uniget_linux_\$(uname -m).tar.gz.sig \
    --certificate-identity 'https://github.com/uniget-org/cli/.github/workflows/release.yml@refs/tags/${TAG}' \
    --certificate-oidc-issuer https://token.actions.githubusercontent.com
```
EOF

echo "## Bugs fixed"
gh issue list \
    --search="state:closed closed:>${TIMESTAMP} label:bug -label:wontfix" \
    --json=number,title,url \
    --template='{{range .}}- {{.title}} ({{hyperlink .url (printf "#%v" .number)}}){{"\n"}}{{end}}'

echo "## Enhancements added"
gh issue list \
    --search="state:closed closed:>${TIMESTAMP} label:enhancement -label:wontfix" \
    --json=number,title,url \
    --template='{{range .}}- {{.title}} ({{hyperlink .url (printf "#%v" .number)}}){{"\n"}}{{end}}'

echo "## Dependencies updated"
gh pr list \
    --state=merged \
    --search="merged:>${TIMESTAMP} label:type/renovate" \
    --json=number,title,url \
    --template='{{range .}}- {{.title}} ({{hyperlink .url (printf "#%v" .number)}}){{"\n"}}{{end}}'

cat <<EOF
## Full Changelog
[Compare with previous release](https://github.com/uniget-org/cli/compare/${PREVIOUS_TAG}...${TAG})
EOF
