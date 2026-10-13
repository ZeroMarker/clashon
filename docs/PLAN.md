# ClashOn 技术选型与开发计划

> 目标：基于 Clash 内核，提供 **CLI / TUI / 桌面 GUI / 移动端** 全平台客户端。

## 0. 约束与前提

| 项 | 决定 |
|---|---|
| iOS | **仅开发与自用**（Xcode 直装 / 开发者证书侧载），**不考虑 App Store 上架**及审核、国区限制 |
| 许可证 | 可灵活调整。因内嵌链接 mihomo（GPL-3.0），**项目整体改为 GPL-3.0**（当前为 MIT，需在 M0 替换 LICENSE） |
| 内核 | 不自研，使用 mihomo（Clash.Meta） |

## 1. 总体架构

```
┌───────────────┬──────────────┬─────────────────────────────┐
│   CLI         │   TUI        │  GUI (Desktop + Mobile)     │
│  (Go/cobra)   │ (Go/Bubble   │  (Flutter)                  │
│               │   Tea)       │                             │
├───────────────┴──────────────┼─────────────────────────────┤
│   clashon-sdk (Go)           │  Dart client / FFI bindings │
│   订阅、配置合并、REST API    │                             │
├──────────────────────────────┴─────────────────────────────┤
│   mihomo core（桌面：子进程/服务；移动：库形式内嵌）          │
│   系统层：TUN / 系统代理 / VpnService / NetworkExtension     │
└────────────────────────────────────────────────────────────┘
```

- **桌面端**：mihomo 以独立进程 / 系统服务运行，GUI、CLI、TUI 都通过 external-controller（REST + WebSocket）控制，三者可同时连接同一内核。
- **移动端**：mihomo 编译为库，嵌入 VPN 扩展进程；App 与扩展通过 IPC（Android AIDL/Binder，iOS App Group + 本地 socket）通信。

## 2. 技术选型

| 模块 | 选型 | 理由 | 备选 |
|---|---|---|---|
| 内核 | **mihomo** (Go) | 协议最全、社区活跃、REST/WS API 完整 | sing-box |
| 公共 SDK | **Go** | 与内核同语言，CLI/TUI 直接复用 | — |
| CLI | **Go + cobra + viper** | 单二进制，交叉编译简单 | — |
| TUI | **Go + Bubble Tea + Lip Gloss + Bubbles** | 生态成熟，复用 SDK | Rust ratatui |
| GUI（桌面 + 移动） | **Flutter 3.x** | 一套代码覆盖 Win/macOS/Linux/Android/iOS，FlClash 已验证可行 | 桌面 Tauri 2 + React（clash-verge-rev 路线） |
| 状态管理 / 存储 | Riverpod + drift(SQLite) | 类型安全，跨平台 | Isar |
| Android 接入 | cgo `.so`（或 gomobile `.aar`）+ **VpnService** | 内核进程内运行 | — |
| iOS / macOS 接入 | gomobile `.xcframework` + **NetworkExtension (PacketTunnelProvider)** | Apple 平台唯一方案 | — |
| 桌面 TUN / 提权 | Win：Windows Service + wintun；macOS：SMAppService 特权 helper；Linux：systemd 服务 + `cap_net_admin` | 避免每次启动弹授权 | — |
| 构建发布 | GitHub Actions 矩阵；GoReleaser（CLI/TUI）；Flutter 各平台打包 | — | — |

### 为什么 GUI 选 Flutter 而不是 Tauri

需要同时覆盖移动与桌面。Flutter 让小团队维护 5 个平台成为可能；Tauri 2 的移动端在 VPN 扩展等深度原生集成上仍不成熟。若将来只保留桌面端，可再评估 Tauri。

## 3. iOS 方案（仅开发 / 自用）

由于不上架，方案大幅简化：

- **签名**：免费 Apple ID 可真机调试，但 **NetworkExtension 需要付费开发者账号**（$99/年）才能拿到 `packet-tunnel-provider` entitlement；无账号时仅能在 **macOS 版**（同一套 NE 代码）上验证大部分逻辑。
- **分发**：Xcode 直装或 Ad-hoc；也可导出 IPA 供 TrollStore / AltStore / SideStore 侧载自用。
- **无需**：App Store 审核、隐私标签、国区合规、TestFlight 流程。
- **仍需处理的技术点**：
  - NE 扩展内存上限约 **50 MB**：mihomo 裁剪编译 tag，GeoIP/GeoSite 改用 MMDB lite / 规则集按需加载，`debug.SetMemoryLimit(40MB)` + 调低 `GOGC`。
  - App 与扩展共享 App Group 容器存放配置与日志。
  - 与 macOS 共用同一 PacketTunnel target 代码。

## 4. 许可证方案

- mihomo 为 **GPL-3.0**，移动端/桌面端以库形式静态链接 → 衍生作品须为 GPL-3.0。
- 决定：**全仓库采用 GPL-3.0**（M0 阶段将 `LICENSE` 由 MIT 改为 GPL-3.0，README 注明）。
- 由于 iOS 不上架，GPL 与 App Store 条款的冲突不再是问题。
- 若将来希望 SDK 部分宽松授权：可将 `sdk/` 设计为只通过 REST API 与内核通信（不链接 mihomo），单独以 MIT 发布。

## 5. 开发计划（约 6 个月，2–3 人）

| 阶段 | 周期 | 交付物 |
|---|---|---|
| **M0 基础** | 第 1–2 周 | monorepo 骨架、LICENSE 改 GPL-3.0、CI 矩阵、mihomo submodule 与全平台交叉编译（含 gomobile） |
| **M1 SDK + CLI** | 第 3–5 周 | 订阅导入/更新、配置合并（覆写/脚本）、启停、切换节点、测延迟、日志；`clashon up/down/sub/proxy/rule/log` |
| **M2 TUI** | 第 6–8 周 | 节点面板、连接列表、实时流量图、日志、规则命中；适合服务器 / SSH 场景 |
| **M3 桌面 GUI** | 第 9–14 周 | Win/macOS/Linux：托盘、系统代理、TUN 服务、开机自启、自动更新、深链 `clash://install-config` |
| **M4 Android** | 第 15–18 周 | VpnService、分应用代理、Quick Settings Tile、通知栏流量 |
| **M5 iOS / macOS NE** | 第 19–21 周 | PacketTunnel、内存优化、侧载/直装构建脚本（不含上架流程，较原计划缩短 1 周） |
| **M6 打磨发布** | 第 22–25 周 | i18n（中/英）、主题、崩溃日志、文档、1.0 发布 |

**首个可用版本**：M1 结束（约第 5 周）的 CLI，面向 Linux 服务器用户。

### 里程碑验收标准

- M1：CLI 在 Linux/macOS/Windows 上可导入订阅并稳定代理 24h。
- M2：TUI 在 80×24 终端下完整可用，CPU 占用 < 2%。
- M3：三平台 TUN 模式一键开启，无需每次输入管理员密码。
- M4：Android 10+ 真机运行，后台保活 24h 不断连。
- M5：iOS 真机内存常驻 < 45 MB。

## 6. 仓库结构

```
clashon/
├── core/        # mihomo submodule + 适配层 + gomobile / cgo 导出
├── sdk/         # Go：订阅、配置、API 客户端
├── cli/         # cobra
├── tui/         # bubbletea
├── app/         # Flutter（android/ ios/ macos/ windows/ linux/）
│   └── native/  # PacketTunnel、VpnService 原生代码
├── service/     # 桌面特权助手 / 系统服务
├── docs/
└── .github/workflows/
```

## 7. 风险

| 风险 | 应对 |
|---|---|
| iOS NE 内存超限被杀 | 裁剪内核、限制规则集、早期（M0）即做内存基准测试 |
| 桌面 TUN 提权三平台差异大 | M3 预留 2 周专门处理，参考 clash-verge-rev 的 service 实现 |
| mihomo API 变动 | 锁定 submodule 版本，`core/` 适配层隔离 |
| 无付费 Apple 账号 | 先在 macOS NE 上开发验证，真机阶段再购买 |
