#!/bin/sh
set -e

# on upgrade the service is restarted by postinst (try-restart), so only stop on removal
if [ "$1" = "remove" ] && [ -d /run/systemd/system ]; then
    systemctl stop go-iotdevice.service || true
    systemctl disable go-iotdevice.service || true
fi
