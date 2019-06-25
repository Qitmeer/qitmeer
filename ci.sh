#!/usr/bin/env bash
set -ex

export GO111MODULE=on
#go mod init qitmeer
go mod tidy

if [ ! -x "$(type -p golangci-lint)" ]; then
  exit 1
fi

golangci-lint --version
golangci-lint run -v --deadline=2m --disable-all --enable=govet --tests=false --enable=gosimple ./... && \
linter_targets=$(go list ./...) && \
go test $linter_targets && \
echo -e "\n Success!"


