#!/bin/sh

if ! getent group scrimp >/dev/null; then
  groupadd -r scrimp
fi

if ! getent passwd scrimp >/dev/null; then
  useradd -M -r -g scrimp -s /sbin/nologin -c "Scrimplb user" scrimp
fi
