#!/usr/bin/env bash
# End-to-end TUN test on a Linux VM with sudo: grant privileges, enable TUN,
# then check that traffic without any proxy settings goes through the core.
set -euo pipefail

: "${CLASHON_TEST_SUB_URL:?set CLASHON_TEST_SUB_URL}"
B="$PWD/bin/clashon"
export CLASHON_HOME="$(mktemp -d)"
trap '"$B" down 2>/dev/null || true; echo "--- core.log"; tail -30 "$CLASHON_HOME/core.log" 2>/dev/null || true' EXIT

"$B" sub add "$CLASHON_TEST_SUB_URL" -n e2e >/dev/null
if "$B" tun on 2>/dev/null; then
  echo "expected tun on to fail without privileges" >&2
  exit 1
fi
sudo "$B" service install
"$B" tun on
"$B" status --json | grep -q '"tunReady": true'
"$B" up

for _ in $(seq 20); do
  ip -o link | grep -qiE 'meta|utun' && break
  sleep 0.5
done
ip -br link | grep -iE 'meta|utun'

# No proxy flags: the request must be captured by the TUN device.
code=$(curl -s -m 20 -o /dev/null -w '%{http_code}' https://www.gstatic.com/generate_204)
echo "direct curl through TUN: HTTP $code"
[ "$code" = 204 ]
# The connection must have entered via the TUN gateway, not the mixed port.
grep -q '\[TCP\] 198\.18\.0\.1:[0-9]* --> .*:443 match' "$CLASHON_HOME/core.log"
echo "TUN e2e ok"
