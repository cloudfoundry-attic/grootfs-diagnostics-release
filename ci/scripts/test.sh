#!/bin/bash

set -e

main() {
  export GOPATH=${PWD}:$GOPATH
  ginkgo -r
}

pushd grootfs-diagnostics-develop
main
popd
