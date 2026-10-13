import 'dart:io';

import 'package:flutter/material.dart';

import '../../state/app_state.dart';

class SettingsPage extends StatelessWidget {
  const SettingsPage({super.key, required this.state});
  final AppState state;

  @override
  Widget build(BuildContext context) {
    final s = state;
    final st = s.status;
    return ListView(
      padding: const EdgeInsets.all(16),
      children: [
        SwitchListTile(
          key: const Key('tun'),
          title: const Text('TUN 模式'),
          subtitle: Text(
            s.tunReady
                ? '接管所有应用的流量，包括不支持代理设置的程序'
                : '需要权限：在终端运行 sudo clashon service install（Linux），'
                      '或以管理员身份运行（Windows/macOS）',
          ),
          value: s.tun,
          onChanged: s.busy || (!s.tunReady && !s.tun) ? null : s.setTun,
        ),
        SwitchListTile(
          title: const Text('系统代理'),
          subtitle: const Text('将系统 HTTP/HTTPS/SOCKS 代理指向本机代理端口'),
          value: s.systemProxy,
          onChanged: s.setSystemProxy,
        ),
        SwitchListTile(
          title: const Text('开机自启'),
          subtitle: const Text('登录后在后台启动，只显示托盘图标'),
          value: s.autostart,
          onChanged: (v) => s.setAutostart(v, Platform.resolvedExecutable),
        ),
        const Divider(),
        ListTile(title: const Text('代理端口'), trailing: Text('${st?.mixedPort ?? '-'}')),
        ListTile(title: const Text('控制接口'), trailing: Text(st?.controller ?? '-')),
        ListTile(title: const Text('内核版本'), trailing: Text(st?.version ?? '-')),
        ListTile(title: const Text('CLI'), subtitle: Text(s.backend.binary)),
      ],
    );
  }
}
