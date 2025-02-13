#!/bin/bash

if [ ! -f /etc/debian_version ]; then
    echo "This script must be run on a Debian-based system"
    exit 1
fi

VERSION=$(git describe --tags)
dh_make --createorig -y -p dirclean_${VERSION}
dpkg-buildpackage -us -uc
