#!/usr/bin/env bash

set -e

vi VERSION

BRANCH=$(git rev-parse --abbrev-ref HEAD)
REVISION=$(git rev-parse HEAD)
VERSION=$(cat VERSION)
OWNER=$(git log -1 --format="%cN" $REVISION | tr " " ".")
DATE=$(git log -1 --format="%ci" $REVISION | tr " " ".")

sed -i -r "s|^([ ]{4}\-X github\.com/prometheus/common/version\.Version\=).*$|\1$VERSION|" .promu.yml
sed -i -r "s|^([ ]{4}\-X github\.com/prometheus/common/version\.Revision\=).*$|\1$REVISION|" .promu.yml
sed -i -r "s|^([ ]{4}\-X github\.com/prometheus/common/version\.Branch\=).*$|\1$BRANCH|" .promu.yml
sed -i -r "s|^([ ]{4}\-X github\.com/prometheus/common/version\.BuildUser\=).*$|\1$OWNER|" .promu.yml
sed -i -r "s|^([ ]{4}\-X github\.com/prometheus/common/version\.BuildDate\=).*$|\1$DATE|" .promu.yml

git status

