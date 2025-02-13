#!/bin/bash

VERSION=$(git describe --tags)
dh_make --createorig -y -p dirclean_${VERSION}
dpkg-buildpackage -us -uc