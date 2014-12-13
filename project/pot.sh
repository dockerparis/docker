#!/usr/bin/env bash
#
# this is a hack to have ncurses dynamically linked to our exec

DEST=bundles/pot

rm -rf ${DEST}
mkdir -p ${DEST}/libs
cp $(ldd $(find bundles -name 'docker-*' -executable) | grep '=>' | awk '// { if ($4) { print $3; } }' | grep -v not | sort | uniq) ${DEST}/libs
cp $(find bundles -name 'docker-*' -executable) ${DEST}/
cp hack/pot-wrapper ${DEST}/docker
