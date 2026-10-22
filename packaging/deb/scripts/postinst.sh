#!/bin/sh
set -e

# dpkg calls "configure <old-version>" on upgrade and "configure" on first install
chown root:go-iotdevice /etc/go-iotdevice/config.yaml

if [ -d /run/systemd/system ]; then
    systemctl daemon-reload
    if [ "$1" = "configure" ] && [ -n "$2" ]; then
        systemctl try-restart go-iotdevice.service || true
    else
        systemctl enable go-iotdevice.service
        systemctl start go-iotdevice.service || true
    fi
fi
