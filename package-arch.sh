#!/bin/bash

VERSION=$(git describe --tags)
makepkg -si --noconfirm