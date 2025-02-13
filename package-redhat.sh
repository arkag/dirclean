#!/bin/bash

VERSION=$(git describe --tags)
rpmbuild -bb --define "_version ${VERSION}" dirclean.spec