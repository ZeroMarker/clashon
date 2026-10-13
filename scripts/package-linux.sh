#!/usr/bin/env bash
# Usage: package-linux.sh <version>  (run after `flutter build linux --release`)
# -> dist/clashon-<version>-linux-amd64.tar.gz and dist/clashon_<version>_amd64.deb
set -euo pipefail
v=$1
bundle=app/build/linux/x64/release/bundle
scripts/build-cli.sh linux amd64 "$bundle/clashon" "$v"
mkdir -p dist

# Portable tarball.
name=clashon-$v-linux-amd64
stage=$(mktemp -d)
cp -r "$bundle" "$stage/$name"
cp app/assets/icon.png "$stage/$name/clashon.png"
tar -C "$stage" -czf "dist/$name.tar.gz" "$name"

# Debian package: app in /opt/clashon, CLI on PATH, TUN capabilities set on install.
root=$(mktemp -d)
mkdir -p "$root/opt" "$root/usr/bin" "$root/usr/share/applications" \
  "$root/usr/share/icons/hicolor/256x256/apps" "$root/DEBIAN"
cp -r "$bundle" "$root/opt/clashon"
ln -s /opt/clashon/clashon "$root/usr/bin/clashon"
cp packaging/linux/dev.clashon.clashon.desktop "$root/usr/share/applications/"
cp app/assets/icon.png "$root/usr/share/icons/hicolor/256x256/apps/dev.clashon.clashon.png"
sed "s/@VERSION@/$v/; s/@SIZE@/$(du -sk "$root/opt" | cut -f1)/" packaging/linux/control > "$root/DEBIAN/control"
install -m 755 packaging/linux/postinst "$root/DEBIAN/postinst"
dpkg-deb --root-owner-group --build "$root" "dist/clashon_${v}_amd64.deb"
ls -l dist
