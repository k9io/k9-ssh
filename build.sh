#!/bin/bash

VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
env CGO_ENABLED=0 go build -ldflags "-s -X main.version=${VERSION}"
