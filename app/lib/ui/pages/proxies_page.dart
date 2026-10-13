import 'package:flutter/material.dart';

import '../../core/api.dart';
import '../../state/app_state.dart';
import '../home.dart';

class ProxiesPage extends StatelessWidget {
  const ProxiesPage({super.key, required this.state});
  final AppState state;

  @override
  Widget build(BuildContext context) {
    final s = state;
    if (!s.running) return CoreStopped(state: s);
    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: s.groups.length,
      itemBuilder: (context, i) {
        final g = s.proxies[s.groups[i]];
        if (g == null) return const SizedBox.shrink();
        return _GroupCard(state: s, group: g, initiallyExpanded: i == 0);
      },
    );
  }
}

class _GroupCard extends StatelessWidget {
  const _GroupCard({required this.state, required this.group, required this.initiallyExpanded});
  final AppState state;
  final Proxy group;
  final bool initiallyExpanded;

  @override
  Widget build(BuildContext context) {
    final testing = state.testing.contains(group.name);
    return Card(
      clipBehavior: Clip.antiAlias,
      child: ExpansionTile(
        initiallyExpanded: initiallyExpanded,
        expandedAlignment: Alignment.topLeft,
        title: Text(group.name),
        subtitle: Text('${group.type} · ${group.now ?? '-'}'),
        trailing: testing
            ? const SizedBox(
                width: 24,
                height: 24,
                child: CircularProgressIndicator(strokeWidth: 2),
              )
            : IconButton(
                tooltip: '测速',
                icon: const Icon(Icons.speed),
                onPressed: () => state.testGroup(group.name),
              ),
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(12, 0, 12, 12),
            child: Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                for (final name in group.all)
                  _ProxyChip(
                    name: name,
                    proxy: state.proxies[name],
                    selected: name == group.now,
                    onTap: group.selectable ? () => state.select(group.name, name) : null,
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

class _ProxyChip extends StatelessWidget {
  const _ProxyChip({required this.name, required this.proxy, required this.selected, this.onTap});
  final String name;
  final Proxy? proxy;
  final bool selected;
  final VoidCallback? onTap;

  @override
  Widget build(BuildContext context) {
    final scheme = Theme.of(context).colorScheme;
    final d = proxy?.delay ?? 0;
    final color = d <= 0
        ? scheme.outline
        : d < 300
        ? Colors.green
        : d < 800
        ? Colors.orange
        : scheme.error;
    return SizedBox(
      width: 220,
      child: Material(
        color: selected ? scheme.primaryContainer : scheme.surfaceContainerHighest,
        borderRadius: BorderRadius.circular(8),
        child: InkWell(
          borderRadius: BorderRadius.circular(8),
          onTap: onTap,
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(name, maxLines: 1, overflow: TextOverflow.ellipsis),
                const SizedBox(height: 2),
                Row(
                  children: [
                    Text(proxy?.type ?? '', style: Theme.of(context).textTheme.bodySmall),
                    const Spacer(),
                    Text(
                      d > 0 ? '${d}ms' : '-',
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(color: color),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
