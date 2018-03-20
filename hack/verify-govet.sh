#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

echo $(go version)


find_files() {
  find . -not \( \
      \( \
        -wholename '*/vendor/*' \
      \) -prune \
    \) -name '*.go'
}


FAILURE=false
test_dirs=$(find_files | cut -d '/' -f 1-2 | sort -u)
for test_dir in $test_dirs
do
  if ! go tool vet -shadow=false -composites=false $test_dir
  then
    FAILURE=true
  fi
done

if $FAILURE
then
  echo "FAILURE: go vet failed!"
  exit 1
else
  echo "SUCCESS: go vet succeeded!"
  exit 0
fi
