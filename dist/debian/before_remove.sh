#!/bin/sh
if [ $1 = "remove" ]; then
  if getent passwd scrimp >/dev/null ; then
    userdel scrimp
  fi

  if getent group scrimp >/dev/null ; then
    groupdel scrimp
  fi
fi
