#!/bin/bash

if [ ! -f /etc/arch-release ]; then
    echo "This script must be run on an Arch Linux system"
    exit 1
fi

VERSION=$(git describe --tags)
makepkg -si --noconfirm
