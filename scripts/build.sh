#!/usr/bin/env bash

set -euo pipefail

oses=${GOOS:-"linux windows darwin"}
archs=${GOARCH:-amd64}

for os in $oses; do
  for arch in $archs; do 
    echo "building $os-$arch";
    GOOS=$os GOARCH=$arch CGO_ENABLED=0 go build -ldflags "$LDFLAGS" -o ./bin/duffle-$os-$arch ./cmd/duffle; \
  done; \
  if [ $os = 'windows' ]; then
    mv ./bin/duffle-$os-$arch ./bin/duffle-$os-$arch.exe; \
  fi; \
done
