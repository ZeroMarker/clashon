#!/usr/bin/env bash
# Usage: package-macos.sh <version>  (run after `flutter build macos --release`)
# -> dist/ClashOn-<version>-macos.dmg (universal, ad-hoc signed)
set -euo pipefail
v=$1
app=app/build/macos/Build/Products/Release/ClashOn.app
tmp=$(mktemp -d)
scripts/build-cli.sh darwin amd64 "$tmp/clashon-amd64" "$v"
scripts/build-cli.sh darwin arm64 "$tmp/clashon-arm64" "$v"
# Resources, not MacOS: the case-insensitive filesystem would clash with ClashOn.
lipo -create -output "$app/Contents/Resources/clashon" "$tmp/clashon-amd64" "$tmp/clashon-arm64"
codesign --force -s - "$app/Contents/Resources/clashon"
codesign --force --deep -s - "$app"
mkdir -p dist "$tmp/dmg"
cp -R "$app" "$tmp/dmg/"
ln -s /Applications "$tmp/dmg/Applications"
hdiutil create -volname ClashOn -srcfolder "$tmp/dmg" -ov -format UDZO "dist/ClashOn-$v-macos.dmg"
ls -l dist
