#!/usr/bin/env bash

GOOS=linux
for GOARCH in amd64 arm64; do
  GOOS=$GOOS GOARCH=$GOARCH go build -ldflags '-s -w' -o bin/nebula-easy-$GOOS-$GOARCH
done
