#!/bin/sh
# Prints the debian package version derived from git:
#   tag v3.12.0                -> 3.12.0
#   12 commits after v3.12.0   -> 3.12.0+12.gabc1234  (sorts above the release)
set -e

desc=$(git describe --tags --long --match 'v*')
sha=${desc##*-g}
rest=${desc%-g*}
n=${rest##*-}
tag=${rest%-*}
tag=${tag#v}

if [ "$n" = "0" ]; then
    echo "$tag"
else
    echo "${tag}+${n}.g${sha}"
fi
