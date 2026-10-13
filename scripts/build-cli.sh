#!/usr/bin/env bash
# Usage: build-cli.sh <goos> <goarch> <output> [version]
# Builds the clashon CLI (with TUI and embedded mihomo) for one platform.
set -euo pipefail
goos=$1 goarch=$2 out=$3 version=${4:-$(git describe --tags --always --dirty)}
mihomo=$(git -C core/mihomo describe --tags)
CGO_ENABLED=0 GOOS=$goos GOARCH=$goarch go build -tags with_gvisor -trimpath \
  -ldflags "-s -w -X github.com/ZeroMarker/clashon/cli/cmd.Version=$version -X github.com/metacubex/mihomo/constant.Version=$mihomo" \
  -o "$out" ./cli/cmd/clashon
