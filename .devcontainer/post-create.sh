#!/usr/bin/env bash
# Codespaces / devcontainer setup: submodules, Go tools, Flutter (Linux + Android targets).
set -euo pipefail

git submodule update --init --recursive

go install golang.org/x/mobile/cmd/gomobile@latest
go install golang.org/x/mobile/cmd/gobind@latest

sudo apt-get update
sudo apt-get install -y --no-install-recommends \
  clang cmake ninja-build pkg-config libgtk-3-dev liblzma-dev libstdc++-12-dev

if [ ! -d "$HOME/flutter" ]; then
  git clone --depth 1 -b stable https://github.com/flutter/flutter.git "$HOME/flutter"
fi
"$HOME/flutter/bin/flutter" config --no-analytics --enable-linux-desktop
"$HOME/flutter/bin/flutter" precache --linux
