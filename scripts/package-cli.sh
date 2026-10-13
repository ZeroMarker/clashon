#!/usr/bin/env bash
# Usage: package-cli.sh <version>  -> dist/clashon-cli-<version>-<os>-<arch>.{tar.gz,zip}
set -euo pipefail
v=$1
mkdir -p dist
for target in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64; do
  os=${target%/*} arch=${target#*/}
  name=clashon-cli-$v-$os-$arch
  dir=$(mktemp -d)/$name
  mkdir -p "$dir"
  exe=clashon; [ "$os" = windows ] && exe=clashon.exe
  scripts/build-cli.sh "$os" "$arch" "$dir/$exe" "$v"
  cp LICENSE README.md "$dir/"
  if [ "$os" = windows ]; then
    (cd "$(dirname "$dir")" && zip -qr "$OLDPWD/dist/$name.zip" "$name")
  else
    tar -C "$(dirname "$dir")" -czf "dist/$name.tar.gz" "$name"
  fi
done
ls -l dist
