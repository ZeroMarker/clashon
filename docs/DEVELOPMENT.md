# 开发环境

## 磁盘占用

Linux 上实测（Go 1.27.2、Flutter 3.47.7，项目已完整编译一次）：

| 部分 | 大小 | 说明 |
|---|---|---|
| Go 工具链 | 284 MB | |
| Go 依赖缓存（`go env GOMODCACHE`） | 659 MB | mihomo 依赖较多 |
| Go 编译缓存（`go env GOCACHE`） | 3.3 GB | 随编译增长，`go clean -cache` 可清空，下次编译重建 |
| Flutter SDK（含引擎缓存） | 1.2 GB | |
| Dart 依赖缓存（`~/.pub-cache`） | 199 MB | |
| 项目编译产物（`app/build`、`app/.dart_tool`） | 约 190 MB | Linux 版 |
| **合计** | **约 5.8 GB** | 清空 Go 编译缓存后约 2.5 GB |

编译其他平台还需要对应的系统工具链（估计值，未实测）：

| 目标平台 | 额外需要 | 约占用 |
|---|---|---|
| Linux 桌面 | clang、cmake、ninja、GTK 3 和 AppIndicator 开发包 | < 0.5 GB |
| Windows 桌面 | Visual Studio 2022 Build Tools，"使用 C++ 的桌面开发" | 7–9 GB |
| Android | Android SDK + NDK + Gradle 缓存 | 6–10 GB |
| macOS / iOS | Xcode | 15 GB 以上 |

在 Windows 上编译调试（Go + Flutter + VS Build Tools + 缓存）建议预留 15 GB 左右。只使用不编译的话，装 [Releases](https://github.com/ZeroMarker/clashon/releases) 里的安装包即可。

## 环境准备

### GitHub Codespaces（推荐）

在仓库页面 Code → Codespaces 新建。`.devcontainer/` 会自动安装 Go、gomobile、Flutter 和 Linux 桌面编译依赖，首次创建约 5–10 分钟。

- 可做：Go 内核/SDK/CLI/TUI 全部开发；Flutter Linux 桌面版编译和测试；Android APK 编译。
- 不可做：TUN（容器内没有 `/dev/net/tun`，也拿不到 `CAP_NET_ADMIN`，TUN 测试在 CI 的 Linux 虚拟机上跑）；Windows/macOS/iOS 编译。
- 测试订阅：仓库的 Codespaces secret `CLASHON_TEST_SUB_URL` 会作为环境变量注入。
- 在虚拟显示器里跑 GUI（截图检查用）：

  ```sh
  sudo apt-get install -y xvfb imagemagick xdotool libgl1-mesa-dri
  Xvfb :99 +extension GLX -screen 0 1200x800x24 &
  DISPLAY=:99 LIBGL_ALWAYS_SOFTWARE=1 app/build/linux/x64/release/bundle/clashon-gui
  ```

### 本机

1. 克隆时带上子模块：`git clone --recursive https://github.com/ZeroMarker/clashon.git`
2. 安装 [Go](https://go.dev/dl/)（stable）和 [Flutter](https://docs.flutter.dev/get-started/install)（stable）。
3. 按目标平台安装上表中的工具链。Linux（Debian/Ubuntu）：

   ```sh
   sudo apt-get install -y clang cmake ninja-build pkg-config libgtk-3-dev liblzma-dev libayatana-appindicator3-dev
   ```

## 常用命令

```sh
make build            # bin/clashon（CLI + TUI + 内嵌 mihomo）和 bin/clashon-tui
make test             # go test ./...
go vet ./...

cd app
flutter pub get
flutter analyze
flutter test                          # 单元和界面测试（不需要真实内核）
flutter build linux --release         # 或 windows / macos
cp ../bin/clashon build/linux/x64/release/bundle/   # GUI 从自身目录找 CLI
```

GUI 依赖 `clashon` 命令行程序：先找环境变量 `CLASHON_BIN`，再找 GUI 可执行文件旁边（macOS 为 `Contents/Resources`），最后找 PATH。

数据目录默认是系统配置目录下的 `clashon/`，可用 `CLASHON_HOME` 指向临时目录，避免影响自己的订阅：

```sh
export CLASHON_HOME=/tmp/clashon-dev
bin/clashon sub add "$CLASHON_TEST_SUB_URL" -n test
bin/clashon up && bin/clashon tui
```

GUI 的诊断日志在系统临时目录的 `clashon-gui.log`（链接、错误、启动参数），报告问题时请附上。

## 测试

| 测试 | 位置 | 运行环境 |
|---|---|---|
| Go 单元测试 | `go test ./...` | CI：Linux / macOS / Windows |
| Flutter 单元和界面测试 | `app/test/` | CI：Linux |
| GUI 状态层对接真实 CLI | `app/integration_test/` | CI：Windows（`flutter test integration_test -d windows`） |
| TUN 端到端 | `scripts/e2e-tun.sh` | CI：Linux 虚拟机（需要 sudo） |
| Windows 端到端：安装包、中文名称、`clash://` 链接冷/热启动、代理、系统代理 | `scripts/e2e-windows.sh` | CI：Windows |

需要真实订阅的测试读取 `CLASHON_TEST_SUB_URL`，未设置时跳过（包括来自 fork 的 PR）。

## 发布

推送 `v*` 标签触发 `.github/workflows/release.yml`，构建 Linux（deb、tar.gz）、Windows（安装程序、免安装 zip）、macOS（dmg）和各平台 CLI，安装冒烟测试通过后发布到 GitHub Releases；版本号含 `-` 的标记为预发布。手动运行（workflow_dispatch）只生成构建产物，不发布。
