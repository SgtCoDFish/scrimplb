#!/bin/sh
if [ $1 = "remove" ]; then
  if getent passwd scrimplb >/dev/null ; then
    userdel scrimplb
  fi

  if getent group scrimplb >/dev/null ; then
    groupdel scrimplb
  fi
fi
