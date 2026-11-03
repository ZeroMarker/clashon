#!/usr/bin/env bash
# Windows end-to-end (Git Bash on a GitHub runner): install with the real
# installer, then exercise subscriptions from the CLI and via clash:// links
# opened through the shell, the way a browser click does.
set -euo pipefail
: "${CLASHON_TEST_SUB_URL:?set CLASHON_TEST_SUB_URL}"

dir="$LOCALAPPDATA/Programs/ClashOn"
C="$dir/clashon.exe"

echo "== install"
./dist/ClashOn-*-setup.exe //VERYSILENT //SUPPRESSMSGBOXES //NORESTART
for _ in $(seq 60); do [ -f "$dir/clashon-gui.exe" ] && break; sleep 1; done
"$C" version
reg query 'HKCU\Software\Classes\clash\shell\open\command'

has() { "$C" sub ls --json | grep -qF "\"name\": \"$1\""; }
wait_for() {
  for _ in $(seq 60); do has "$1" && return 0; sleep 1; done
  echo "profile '$1' did not appear"; "$C" sub ls --json; return 1
}
gui_count() { tasklist //FI "IMAGENAME eq clashon-gui.exe" //NH | grep -c clashon-gui.exe || true; }
enc=$(python -c 'import os,urllib.parse;print(urllib.parse.quote(os.environ["CLASHON_TEST_SUB_URL"],safe=""))')
open_link() { # ShellExecute, like clicking the link in a browser
  powershell -NoProfile -Command "Start-Process 'clash://install-config?url=$enc&name=$1'"
}

echo "== CLI add with a Chinese name"
"$C" sub add "$CLASHON_TEST_SUB_URL" -n "命令行 测试"
has "命令行 测试"

echo "== link while the GUI is not running (cold start)"
[ "$(gui_count)" = 0 ]
open_link cold-link
wait_for cold-link
sleep 2
[ "$(gui_count)" = 1 ]

echo "== link while the GUI is running (forwarded to the running instance)"
open_link warm-link
wait_for warm-link
sleep 3
echo "gui processes: $(gui_count)"
[ "$(gui_count)" = 1 ]

echo "== link with a Chinese name"
open_link '%E4%B8%AD%E6%96%87%E8%AE%A2%E9%98%85'
wait_for "中文订阅"

echo "== core through the imported profile"
"$C" sub use warm-link
"$C" up
code=$(curl -s -m 20 -x http://127.0.0.1:7890 -o /dev/null -w '%{http_code}' https://www.gstatic.com/generate_204)
echo "via proxy: HTTP $code"
[ "$code" = 204 ]
"$C" sysproxy on
reg query 'HKCU\Software\Microsoft\Windows\CurrentVersion\Internet Settings' //v ProxyServer
"$C" sysproxy off
"$C" down
taskkill //IM clashon-gui.exe //F >/dev/null 2>&1 || true
echo "Windows e2e ok"
