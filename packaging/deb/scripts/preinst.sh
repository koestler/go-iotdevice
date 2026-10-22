#!/bin/sh
set -e

# serial / gpio group membership is added in postinst for groups that exist
getent group go-iotdevice >/dev/null || groupadd --system go-iotdevice
getent passwd go-iotdevice >/dev/null || useradd --system --no-create-home \
    --home-dir /nonexistent --shell /usr/sbin/nologin \
    --gid go-iotdevice go-iotdevice
