#!/bin/sh

if ! getent group scrimplb >/dev/null; then
  groupadd -r scrimplb
fi

if ! getent passwd scrimplb >/dev/null; then
  useradd -M -r -g scrimplb -s /usr/sbin/nologin -c "Scrimplb user" scrimplb
fi
