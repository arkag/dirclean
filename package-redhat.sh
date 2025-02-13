#!/bin/bash

if [ ! -f /etc/redhat-release ]; then
    echo "This script must be run on a RedHat-based system"
    exit 1
fi

VERSION=$(git describe --tags)
rpmbuild -bb --define "_version ${VERSION}" dirclean.spec
