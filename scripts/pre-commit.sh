#!/usr/bin/env bash
set -euo pipefail

gofmt -w .
goimports -w .

git diff --cached --name-only --diff-filter=ACM | grep '\.go$' | xargs -r git add || true