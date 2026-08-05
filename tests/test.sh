#!/bin/bash
# Integration test: query this repository for the partial user name "heath"
# and assert a single record for "heaths" is returned.
set -euo pipefail

cd "$(dirname "$0")/.."

output="$(GH_REPO="${GH_REPO:-heaths/gh-users}" ./gh-users heath)"
echo "${output}"

count="$(echo "${output}" | grep -c . || true)"
if [[ "${count}" -ne 1 ]]; then
    echo "error: expected 1 record, got ${count}" >&2
    exit 1
fi

login="$(echo "${output}" | cut -f1)"
if [[ "${login}" != "heaths" ]]; then
    echo "error: expected login 'heaths', got '${login}'" >&2
    exit 1
fi

echo "ok"
