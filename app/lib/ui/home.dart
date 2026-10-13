import 'package:flutter/material.dart';

import '../state/app_state.dart';
import 'pages/connections_page.dart';
import 'pages/dashboard_page.dart';
import 'pages/logs_page.dart';
import 'pages/profiles_page.dart';
import 'pages/proxies_page.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key, required this.state});
  final AppState state;

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int _index = 0;

  static const _destinations = [
    (Icons.dashboard_outlined, Icons.dashboard, '概览'),
    (Icons.hub_outlined, Icons.hub, '代理'),
    (Icons.description_outlined, Icons.description, '订阅'),
    (Icons.swap_vert_outlined, Icons.swap_vert, '连接'),
    (Icons.article_outlined, Icons.article, '日志'),
  ];

  void _select(int i) {
    setState(() => _index = i);
    widget.state.watchConnections(i == 3);
  }

  @override
  Widget build(BuildContext context) {
    final s = widget.state;
    return ListenableBuilder(
      listenable: s,
      builder: (context, _) {
        final page = switch (_index) {
          0 => DashboardPage(state: s),
          1 => ProxiesPage(state: s),
          2 => ProfilesPage(state: s),
          3 => ConnectionsPage(state: s),
          _ => LogsPage(state: s),
        };
        return Scaffold(
          body: Row(
            children: [
              NavigationRail(
                selectedIndex: _index,
                onDestinationSelected: _select,
                labelType: NavigationRailLabelType.all,
                leading: Padding(
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  child: Icon(Icons.bolt,
                      size: 32, color: s.running ? Colors.green : Theme.of(context).disabledColor),
                ),
                destinations: [
                  for (final (icon, selected, label) in _destinations)
                    NavigationRailDestination(
                        icon: Icon(icon), selectedIcon: Icon(selected), label: Text(label)),
                ],
              ),
              const VerticalDivider(width: 1),
              Expanded(
                child: Column(
                  children: [
                    if (s.error != null) _ErrorBanner(state: s),
                    if (s.busy) const LinearProgressIndicator(minHeight: 2),
                    Expanded(child: page),
                  ],
                ),
              ),
            ],
          ),
        );
      },
    );
  }
}

class _ErrorBanner extends StatelessWidget {
  const _ErrorBanner({required this.state});
  final AppState state;

  @override
  Widget build(BuildContext context) {
    return MaterialBanner(
      content: Text(state.error!),
      leading: const Icon(Icons.error_outline),
      backgroundColor: Theme.of(context).colorScheme.errorContainer,
      actions: [TextButton(onPressed: state.clearError, child: const Text('关闭'))],
    );
  }
}

/// Placeholder shown on pages that need a running core.
class CoreStopped extends StatelessWidget {
  const CoreStopped({super.key, required this.state});
  final AppState state;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Text('内核未运行'),
          const SizedBox(height: 12),
          FilledButton.icon(
            onPressed: state.busy ? null : state.toggleCore,
            icon: const Icon(Icons.play_arrow),
            label: const Text('启动'),
          ),
        ],
      ),
    );
  }
}
