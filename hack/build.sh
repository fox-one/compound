#!/usr/bin/env bash

set -e

TAG=${1:-dev}
VERSION=$(git describe --tags --abbrev=0)
COMMIT=$(git rev-parse --short HEAD)

CONFIG=config."${TAG}".yaml
if [ -f "${CONFIG}" ]; then
  trap 'rm -f config_gen.go' EXIT
  if ! type "config-gen" > /dev/null 2>/dev/null; then
    env GO111MODULE=off go get -u github.com/fox-one/pkg/config/config-gen
  fi
  echo "use config ${CONFIG}"
  config-gen --config "${CONFIG}" --tag "${TAG}"
fi

echo "build rings with version ${VERSION} & commit ${COMMIT} & tag ${TAG}"
go build -a \
         --tags "${TAG}" \
         --ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
         -o builds/rings

if [ -f "config_gen.go" ]; then
  trap 'rm -f config_gen.go' EXIT
fi
