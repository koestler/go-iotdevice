#!/bin/bash
# Builds a .deb package.
# Usage: deb/build-deb.sh <amd64|arm64|armhf> [version]
# Requires: go, node (npx), curl, nfpm; env: FRONTEND_VERSION, SWAG_VERSION
set -euo pipefail

# Requires FRONTEND_VERSION (release of github.com/koestler/js-iotdevice that is bundled)
# and SWAG_VERSION in the environment
: "${FRONTEND_VERSION:?must be set}" "${SWAG_VERSION:?must be set}"

ARCH="${1:?usage: $0 <amd64|arm64|armhf> [version]}"
cd "$(dirname "$0")/.."
VERSION="${2:-$(deb/version.sh)}"

case "$ARCH" in
amd64) GOARCH=amd64; GOARM="" ;;
arm64) GOARCH=arm64; GOARM="" ;;
armhf) GOARCH=arm; GOARM=7 ;;
*) echo "unsupported arch: $ARCH" >&2; exit 1 ;;
esac

echo "building go-iotdevice $VERSION for $ARCH (frontend $FRONTEND_VERSION)"

# swagger documentation
go install "github.com/swaggo/swag/cmd/swag@$SWAG_VERSION"
PATH="$(go env GOPATH)/bin:$PATH" go generate docs.go

# binary
mkdir -p "dist/$ARCH"
CGO_ENABLED=0 GOOS=linux GOARCH="$GOARCH" GOARM="$GOARM" go build \
    -ldflags="-s -w -X main.buildVersion=$VERSION -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -tags=jsoniter -o "dist/$ARCH/go-iotdevice"

# frontend
rm -rf dist/frontend-build dist/frontend-*.stamp
mkdir -p dist/frontend-build dist/frontend-download
base="https://github.com/koestler/js-iotdevice/releases/download/$FRONTEND_VERSION"
tarball="js-iotdevice-$FRONTEND_VERSION.tar.gz"
curl -fsSL -o "dist/frontend-download/$tarball" "$base/$tarball"
curl -fsSL -o dist/frontend-download/SHA256SUMS "$base/SHA256SUMS"
(cd dist/frontend-download && sha256sum -c --ignore-missing SHA256SUMS)
tar -xzf "dist/frontend-download/$tarball" -C dist/frontend-build
test -f dist/frontend-build/index.html

# independent of the umask / tar modes: readable by the unprivileged service user
chmod -R u=rwX,go=rX dist/frontend-build

# Debian expects a compressed changelog (-n: reproducible, no timestamp)
gzip -9n -c CHANGELOG.md > dist/changelog.gz

# package
export ARCH VERSION
nfpm package -f deb/nfpm.yaml -p deb \
    -t "dist/go-iotdevice_${VERSION}_${ARCH}.deb"
ls -l "dist/go-iotdevice_${VERSION}_${ARCH}.deb"
