#!/usr/bin/env bash
set -euo pipefail

repo_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
temp_dir=$(mktemp -d)
trap 'rm -rf "$temp_dir"' EXIT

cat >"$temp_dir/gh" <<'EOF'
#!/usr/bin/env bash
case "$*" in
  *'search=hea in:login'*)
    printf '%s\n' heaths heathcliff
    ;;
  *'search=oct in:login'*)
    printf '%s\n' octocat heaths
    ;;
esac
EOF
chmod +x "$temp_dir/gh"

output=$(PATH="$temp_dir:$PATH" "$repo_root/scripts/query-users.sh" hea oct)
expected='query{user0:user(login:"heaths"){...User} user1:user(login:"heathcliff"){...User} user2:user(login:"octocat"){...User}}fragment User on User{login name email status{message}}'

[[ "$output" == "$expected" ]]

if PATH="$temp_dir:$PATH" "$repo_root/scripts/query-users.sh" >/dev/null 2>&1; then
  exit 1
fi

cat >"$temp_dir/gh" <<'EOF'
#!/usr/bin/env bash
exit 1
EOF

if PATH="$temp_dir:$PATH" "$repo_root/scripts/query-users.sh" failure >/dev/null 2>&1; then
  exit 1
fi
