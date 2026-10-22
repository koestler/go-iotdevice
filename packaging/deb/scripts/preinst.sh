#!/bin/sh
set -e

getent group gpio >/dev/null || groupadd --system gpio
getent group go-iotdevice >/dev/null || groupadd --system go-iotdevice
getent passwd go-iotdevice >/dev/null || useradd --system --no-create-home \
    --home-dir /nonexistent --shell /usr/sbin/nologin \
    --gid go-iotdevice -G dialout,gpio go-iotdevice
