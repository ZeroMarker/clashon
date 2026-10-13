import 'package:flutter/material.dart';

import '../../state/app_state.dart';
import '../home.dart';

class LogsPage extends StatelessWidget {
  const LogsPage({super.key, required this.state});
  final AppState state;

  @override
  Widget build(BuildContext context) {
    final s = state;
    if (!s.running) return CoreStopped(state: s);
    final logs = s.logs.toList().reversed.toList(); // newest first
    final scheme = Theme.of(context).colorScheme;
    if (logs.isEmpty) return const Center(child: Text('等待日志…'));
    return ListView.builder(
      padding: const EdgeInsets.all(8),
      itemCount: logs.length,
      itemBuilder: (context, i) {
        final l = logs[i];
        final color = switch (l.type) {
          'error' => scheme.error,
          'warning' => Colors.orange,
          _ => scheme.outline,
        };
        final t = l.time;
        String two(int n) => n.toString().padLeft(2, '0');
        return Padding(
          padding: const EdgeInsets.symmetric(vertical: 2, horizontal: 8),
          child: SelectableText.rich(
            TextSpan(
              children: [
                TextSpan(
                  text: '${two(t.hour)}:${two(t.minute)}:${two(t.second)} ',
                  style: TextStyle(color: scheme.outline),
                ),
                TextSpan(
                  text: l.type.padRight(8),
                  style: TextStyle(color: color),
                ),
                TextSpan(text: l.payload),
              ],
            ),
            style: const TextStyle(fontFamily: 'monospace', fontSize: 12),
          ),
        );
      },
    );
  }
}
