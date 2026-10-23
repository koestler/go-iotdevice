#!/bin/bash
# Installs a .deb into an architecture-matched debian:trixie container (built locally with systemd) as PID 1 and verifies the service.
# Usage: deb/test-install.sh <path-to.deb>
# shellcheck disable=SC2016 # commands are intentionally expanded inside the container
set -euo pipefail

DEB="$(cd "$(dirname "${1:?usage: $0 <path-to.deb>}")" && pwd)/$(basename "$1")"
IMAGE="go-iotdevice-systemd-test:trixie"
NAME="go-iotdevice-deb-test-$$"

case "$DEB" in
*_amd64.deb) PLATFORM=linux/amd64 ;;
*_arm64.deb) PLATFORM=linux/arm64 ;;
*_armhf.deb) PLATFORM=linux/arm/v7 ;;
*) echo "unsupported Debian package architecture: $DEB" >&2; exit 1 ;;
esac

cleanup() { docker rm -f "$NAME" >/dev/null 2>&1 || true; }
trap cleanup EXIT

docker build --platform="$PLATFORM" -q -t "$IMAGE" - >/dev/null <<'DOCKERFILE'
FROM debian:trixie
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update -qq && apt-get install -y -qq systemd systemd-sysv dbus \
    && rm -rf /var/lib/apt/lists/*
STOPSIGNAL SIGRTMIN+3
CMD ["/sbin/init"]
DOCKERFILE

docker run --platform="$PLATFORM" -d --name "$NAME" --privileged --cgroupns=host \
    -v /sys/fs/cgroup:/sys/fs/cgroup:rw -v "$DEB:/opt/test.deb:ro" "$IMAGE" >/dev/null
run() { docker exec "$NAME" bash -euxc "$1"; }

# wait for systemd
for _ in $(seq 1 30); do
    state=$(docker exec "$NAME" systemctl is-system-running 2>/dev/null || true)
    case "$state" in running|degraded) break ;; esac
    sleep 1
done

run 'apt-get update -qq && apt-get install -y -qq curl ca-certificates >/dev/null'

echo "== install"
run 'apt-get install -y /opt/test.deb'
run 'getent passwd go-iotdevice
id -nG go-iotdevice | tr " " "\n" | grep -qx dialout
[ -f /usr/share/doc/go-iotdevice/changelog.gz ]'
run '[ "$(stat -c "%a %U:%G" /etc/go-iotdevice/config.yaml)" = "640 root:go-iotdevice" ]'
run 'systemctl is-enabled go-iotdevice'
run 'for i in $(seq 1 30); do systemctl is-active --quiet go-iotdevice && exit 0; sleep 1; done; systemctl status go-iotdevice --no-pager || true; journalctl -u go-iotdevice --no-pager || true; exit 1'

echo "== http"
run 'for i in $(seq 1 30); do curl -fs http://127.0.0.1:8000/ >/dev/null && break; sleep 1; done
curl -fs http://127.0.0.1:8000/ | grep -qi "<html"
curl -fs http://127.0.0.1:8000/api/v2/docs/swagger.yaml | grep -q swagger
curl -fs http://127.0.0.1:8000/api/v2/config/frontend | grep -q projectTitle'

echo "== dry-run as service user"
run 'runuser -u go-iotdevice -- /usr/bin/go-iotdevice --dry-run -c /etc/go-iotdevice/config.yaml'

echo "== authentication block can be enabled"
run 'cp /etc/go-iotdevice/config.yaml /tmp/auth.yaml
sed -i -E "/^# Authentication:/,/^$/ s/^# ?//" /tmp/auth.yaml
grep -q "^Authentication:" /tmp/auth.yaml
echo "user:\$apr1\$x\$x" > /tmp/auth.passwd
sed -i "s#/etc/go-iotdevice/auth.passwd#/tmp/auth.passwd#" /tmp/auth.yaml
/usr/bin/go-iotdevice --dry-run -c /tmp/auth.yaml'

echo "== modified conffile is preserved on reinstall"
run 'echo "# local change" >> /etc/go-iotdevice/config.yaml
apt-get install -y --reinstall -o Dpkg::Options::=--force-confold /opt/test.deb
grep -q "# local change" /etc/go-iotdevice/config.yaml
systemctl is-active go-iotdevice'

echo "== remove"
run 'apt-get remove -y go-iotdevice
[ ! -e /usr/bin/go-iotdevice ]
[ -f /etc/go-iotdevice/config.yaml ]
getent passwd go-iotdevice
if systemctl is-active go-iotdevice; then exit 1; fi'

echo "== purge"
run 'apt-get purge -y go-iotdevice
[ ! -e /etc/go-iotdevice ]
if getent passwd go-iotdevice; then exit 1; fi
if getent group go-iotdevice; then exit 1; fi'

echo "OK"
