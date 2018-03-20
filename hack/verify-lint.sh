#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

for d in $(find . -type d -a \( -iwholename './pkg/*' -o -iwholename './knitter-agent/*' -o -iwholename './knitter-monitor/*' -o -iwholename './knitter-manager/*' -o -iwholename './knitter-plugin/*' \)); do
	echo for directory ${d} ...
	gometalinter \
		 --exclude='error return value not checked.*(Close|Log|Print).*\(errcheck\)$' \
		 --exclude='.*_test\.go:.*error return value not checked.*\(errcheck\)$' \
		 --exclude='duplicate of.*_test.go.*\(dupl\)$' \
		 --exclude='.*/mock_.*\.go:.*\(golint\)$' \
		 --exclude='declaration of "err" shadows declaration.*\(vetshadow\)$' \
		 --disable=aligncheck \
		 --disable=gotype \
		 --disable=gas \
		 --cyclo-over=60 \
		 --dupl-threshold=100 \
		 --tests \
		 --deadline=300s "${d}"
done
