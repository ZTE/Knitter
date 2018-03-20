#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

echo $(go version)

# if golint not exist, just install it.
if ! which golint &>/dev/null; then
  echo "Unable to detect 'golint' package"
  echo "To install it, run: 'go get github.com/golang/lint/golint'"
  exit 1
fi

find_files() {
  find . -not \( \
      \( \
        -path '*/test/*' \
       -or -path '*/vendor/*' \
       -or -path '*/mock/*' \
      \) -prune \
    \) -name '*.go'
}


bad_files=$(find_files | xargs -n1 golint)

if [[ -n "${bad_files}" ]]; then
  echo "!!! golint detected the following problems:"
  echo "${bad_files}"
  exit 1
fi

