#!/usr/bin/env bash
# Usage: package-windows.sh <version>  (Git Bash, after `flutter build windows --release`)
# -> dist/ClashOn-<version>-windows-x64-setup.exe and a portable zip
set -euo pipefail
v=$1
release=app/build/windows/x64/runner/Release
scripts/build-cli.sh windows amd64 "$release/clashon.exe" "$v"
mkdir -p dist
name=ClashOn-$v-windows-x64
(cd "$release" && 7z a -tzip -bso0 "$OLDPWD/dist/$name-portable.zip" .)
iscc //Qp //DAppVersion="$v" //DSourceDir="$(cygpath -w "$PWD/$release")" \
  //O"$(cygpath -w "$PWD/dist")" //F"$name-setup" packaging/windows/clashon.iss
ls -l dist
