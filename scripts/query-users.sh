#!/usr/bin/env bash
set -euo pipefail

if (( $# == 0 )); then
  echo "usage: ${0##*/} PARTIAL_USERNAME..." >&2
  exit 2
fi

search_query='
query($search: String!, $endCursor: String) {
  search(query: $search, type: USER, first: 100, after: $endCursor) {
    nodes {
      ... on User {
        login
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}'

logins=
for partial_username in "$@"; do
  matches=$(gh api graphql --paginate \
    -f query="$search_query" \
    -F "search=${partial_username} in:login" \
    --jq '.data.search.nodes[].login')
  if [[ -n $matches ]]; then
    logins+="$matches"$'\n'
  fi
done

printf '%s' "$logins" |
  awk '
    BEGIN {
      printf "query{"
    }
    $0 != "" && !seen[$0]++ {
      printf "%suser%d:user(login:\"%s\"){...User}", separator, count++, $0
      separator = " "
    }
    END {
      print "}fragment User on User{login name email status{message}}"
    }
  '
