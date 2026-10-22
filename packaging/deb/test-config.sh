#!/bin/bash
# Checks the packaged default config. Usage: packaging/deb/test-config.sh <go-iotdevice-binary>
set -euo pipefail
BIN="$(cd "$(dirname "${1:?usage: $0 <go-iotdevice-binary>}")" && pwd)/$(basename "$1")"
cd "$(dirname "$0")"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

grep -q '/usr/share/doc/go-iotdevice/full-config.yaml' config.yaml
grep -q '^# *JwtSecret:' config.yaml
grep -q '^# *HtaccessFile: /etc/go-iotdevice/auth.passwd' config.yaml
if grep -q '^Authentication:' config.yaml; then echo 'Authentication must be commented out' >&2; exit 1; fi

# FrontendPath defaults to a relative path; point it to a temp dir
mkdir "$tmp/frontend"
awk -v p="$tmp/frontend" '{print} /^  Port: 8000/{print "  FrontendPath: " p}' config.yaml > "$tmp/default.yaml"
"$BIN" --dry-run -c "$tmp/default.yaml"

# the commented Authentication block must work when uncommented
# shellcheck disable=SC2016
echo 'user:$apr1$x$x' > "$tmp/auth.passwd"
sed -E '/^# Authentication:/,/^$/ s/^# ?//' "$tmp/default.yaml" \
    | sed "s#/etc/go-iotdevice/auth.passwd#$tmp/auth.passwd#" > "$tmp/auth.yaml"
grep -q '^Authentication:' "$tmp/auth.yaml"
"$BIN" --dry-run -c "$tmp/auth.yaml"
echo OK
