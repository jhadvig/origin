#!/bin/bash

# This script sets up a go workspace locally and builds all go components.

set -o errexit
set -o nounset
set -o pipefail

OS_ROOT=$(dirname "${BASH_SOURCE}")/..
source "${OS_ROOT}/hack/common.sh"

"${OS_ROOT}/hack/build-go.sh" cmd/genbashcomp

# Find binary
genbashcomp=$( (ls -t _output/local/go/bin/genbashcomp) 2>/dev/null || true | head -1 )

if [[ ! "$genbashcomp" ]]; then
  {
    echo "It looks as if you don't have a compiled genbashcomp binary"
    echo
    echo "If you are running from a clone of the git repo, please run"
    echo "'./hack/build-go.sh cmd/genbashcomp'."
  } >&2
  exit 1
fi

os::build::gen-doc "${genbashcomp}" "${OS_ROOT}/rel-eng/completions/bash/"
