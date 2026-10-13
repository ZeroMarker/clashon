import 'package:flutter/material.dart';

import '../../state/app_state.dart';
import '../home.dart';

class ConnectionsPage extends StatelessWidget {
  const ConnectionsPage({super.key, required this.state});
  final AppState state;

  @override
  Widget build(BuildContext context) {
    final s = state;
    if (!s.running) return CoreStopped(state: s);
    final small = Theme.of(context).textTheme.bodySmall;
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(16, 12, 16, 4),
          child: Row(
            children: [
              Text(
                '${s.connections.length} 个连接 · ↓ ${formatBytes(s.downloadTotal)} ↑ ${formatBytes(s.uploadTotal)}',
              ),
              const Spacer(),
              TextButton.icon(
                onPressed: s.connections.isEmpty ? null : () => s.closeConnection(),
                icon: const Icon(Icons.clear_all),
                label: const Text('全部关闭'),
              ),
            ],
          ),
        ),
        Expanded(
          child: ListView.builder(
            itemCount: s.connections.length,
            itemBuilder: (context, i) {
              final c = s.connections[i];
              return ListTile(
                dense: true,
                title: Text(c.host, maxLines: 1, overflow: TextOverflow.ellipsis),
                subtitle: Text(
                  '${c.rule} · ${c.chain}',
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                trailing: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      '↓ ${formatBytes(c.download)}\n↑ ${formatBytes(c.upload)}',
                      style: small,
                      textAlign: TextAlign.right,
                    ),
                    IconButton(
                      tooltip: '关闭',
                      icon: const Icon(Icons.close, size: 18),
                      onPressed: () => s.closeConnection(c.id),
                    ),
                  ],
                ),
              );
            },
          ),
        ),
      ],
    );
  }
}
