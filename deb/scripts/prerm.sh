#!/bin/sh
set -e

# on upgrade the service is restarted by postinst (try-restart), so only stop on removal
if [ "$1" = "remove" ] && command -v systemctl >/dev/null 2>&1; then
    if [ -d /run/systemd/system ]; then
        systemctl stop go-iotdevice.service || true
    fi
    # postinst enables the service without a running systemd (chroot, image builds), so undo it the same way
    systemctl disable go-iotdevice.service || true
fi
