#!/bin/bash
set -o errexit -o pipefail

TAG="$(
    git tag \
    | tail -n 2 \
    | head -n 1
)"

TIMESTAMP="$(
    gh release view "${TAG}" --json=publishedAt --template='{{.publishedAt}}'
)"

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