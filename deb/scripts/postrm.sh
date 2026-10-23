#!/bin/sh
set -e

if [ -d /run/systemd/system ]; then
    systemctl daemon-reload || true
fi

if [ "$1" = "purge" ]; then
    rm -rf /etc/go-iotdevice
    userdel go-iotdevice 2>/dev/null || true
    groupdel go-iotdevice 2>/dev/null || true
fi
