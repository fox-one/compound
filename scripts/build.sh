#!/usr/bin/env bash

set -e

TAG=${1:-dev}
VERSION=$(git describe --tags)
COMMIT=$(git rev-parse --short HEAD)
BINARY=uniswap."${TAG}"

trap 'rm -f config_gen.go' EXIT
# go get -u github.com/fox-one/pkg/config/config-gen
config-gen --config config."${TAG}".yaml --tag "${TAG}"

export GOOS=linux
export GOARCH=amd64

echo "build ${BINARY} with version ${VERSION} & commit ${COMMIT}"
go build --tags "${TAG}" \
         --ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
         -o "${BINARY}"

# brew install upx
upx -q "${BINARY}"
