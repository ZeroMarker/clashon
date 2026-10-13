import 'dart:math';

import 'package:flutter/material.dart';

import '../../core/api.dart';
import '../../state/app_state.dart';

class DashboardPage extends StatelessWidget {
  const DashboardPage({super.key, required this.state});
  final AppState state;

  @override
  Widget build(BuildContext context) {
    final s = state;
    final st = s.status;
    final theme = Theme.of(context);
    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        Row(
          children: [
            Text('ClashOn', style: theme.textTheme.headlineMedium),
            const Spacer(),
            FilledButton.icon(
              key: const Key('toggleCore'),
              onPressed: s.busy ? null : s.toggleCore,
              icon: Icon((st?.running ?? false) ? Icons.stop : Icons.play_arrow),
              label: Text((st?.running ?? false) ? '停止' : '启动'),
            ),
          ],
        ),
        const SizedBox(height: 16),
        Wrap(
          spacing: 16,
          runSpacing: 16,
          children: [
            _Info('状态', s.running ? '运行中' : ((st?.running ?? false) ? '无法连接控制器' : '已停止')),
            _Info('订阅', s.currentProfile.isEmpty ? '-' : s.currentProfile),
            _Info('代理端口', st == null ? '-' : '127.0.0.1:${st.mixedPort}'),
            _Info('内核', st?.version ?? '-'),
          ],
        ),
        const SizedBox(height: 24),
        Text('模式', style: theme.textTheme.titleMedium),
        const SizedBox(height: 8),
        SegmentedButton<String>(
          segments: const [
            ButtonSegment(value: 'rule', label: Text('规则')),
            ButtonSegment(value: 'global', label: Text('全局')),
            ButtonSegment(value: 'direct', label: Text('直连')),
          ],
          selected: {s.mode},
          onSelectionChanged: s.running ? (v) => s.setMode(v.first) : null,
        ),
        const SizedBox(height: 16),
        SwitchListTile(
          contentPadding: EdgeInsets.zero,
          title: const Text('系统代理'),
          subtitle: const Text('将系统 HTTP/HTTPS/SOCKS 代理指向本机代理端口'),
          value: s.systemProxy,
          onChanged: s.setSystemProxy,
        ),
        const SizedBox(height: 24),
        Text('流量', style: theme.textTheme.titleMedium),
        const SizedBox(height: 8),
        Row(children: [
          Icon(Icons.arrow_upward, size: 16, color: theme.colorScheme.tertiary),
          Text(' ${formatBytes(s.traffic.up)}/s    '),
          Icon(Icons.arrow_downward, size: 16, color: theme.colorScheme.primary),
          Text(' ${formatBytes(s.traffic.down)}/s'),
        ]),
        const SizedBox(height: 8),
        SizedBox(
          height: 140,
          child: CustomPaint(
            painter: _TrafficChart(
              s.trafficHistory.toList(),
              up: theme.colorScheme.tertiary,
              down: theme.colorScheme.primary,
              grid: theme.dividerColor,
            ),
            size: Size.infinite,
          ),
        ),
      ],
    );
  }
}

class _Info extends StatelessWidget {
  const _Info(this.label, this.value);
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    final t = Theme.of(context).textTheme;
    return SizedBox(
      width: 200,
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(label, style: t.labelMedium),
              const SizedBox(height: 4),
              Text(value, style: t.titleMedium, overflow: TextOverflow.ellipsis),
            ],
          ),
        ),
      ),
    );
  }
}

class _TrafficChart extends CustomPainter {
  _TrafficChart(this.data, {required this.up, required this.down, required this.grid});
  final List<Traffic> data;
  final Color up, down, grid;

  @override
  void paint(Canvas canvas, Size size) {
    canvas.drawLine(Offset(0, size.height), Offset(size.width, size.height), Paint()..color = grid);
    if (data.length < 2) return;
    final peak = data.fold<int>(1, (m, t) => max(m, max(t.up, t.down)));
    const n = AppState.historyLength;
    Path line(int Function(Traffic) pick) {
      final path = Path();
      for (var i = 0; i < data.length; i++) {
        final x = size.width * (n - data.length + i) / (n - 1);
        final y = size.height * (1 - pick(data[i]) / peak);
        i == 0 ? path.moveTo(x, y) : path.lineTo(x, y);
      }
      return path;
    }

    final stroke = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2;
    canvas.drawPath(line((t) => t.down), stroke..color = down);
    canvas.drawPath(line((t) => t.up), stroke..color = up);
  }

  @override
  bool shouldRepaint(_TrafficChart old) => true;
}
