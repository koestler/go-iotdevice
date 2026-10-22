#!/bin/sh
set -e

# dpkg calls "configure <old-version>" on upgrade and "configure" on first install;
# ignore abort-upgrade, abort-remove, abort-deconfigure, ...
[ "$1" = "configure" ] || exit 0

# serial / gpio access; groups that do not exist on this system are skipped
for g in dialout gpio; do
    if getent group "$g" >/dev/null; then
        id -nG go-iotdevice | tr ' ' '\n' | grep -qx "$g" || usermod -aG "$g" go-iotdevice
    fi
done

# the conffile may have been deleted by the administrator
if [ -f /etc/go-iotdevice/config.yaml ]; then
    chown root:go-iotdevice /etc/go-iotdevice/config.yaml
fi

if command -v systemctl >/dev/null 2>&1; then
    # works without a running systemd as well (chroot, image builds): only creates the symlink
    if [ -z "$2" ]; then
        systemctl enable go-iotdevice.service || true
    fi

    if [ -d /run/systemd/system ]; then
        systemctl daemon-reload
        if [ -n "$2" ]; then
            systemctl try-restart go-iotdevice.service || true
        else
            systemctl start go-iotdevice.service || true
        fi
    fi
fi
