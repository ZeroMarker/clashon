import 'package:flutter/material.dart';

import '../../state/app_state.dart';

class ProfilesPage extends StatelessWidget {
  const ProfilesPage({super.key, required this.state});
  final AppState state;

  Future<void> _add(BuildContext context) async {
    final name = TextEditingController(text: state.profiles.isEmpty ? 'default' : '');
    final url = TextEditingController();
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('添加订阅'),
        content: SizedBox(
          width: 420,
          child: Column(mainAxisSize: MainAxisSize.min, children: [
            TextField(controller: name, decoration: const InputDecoration(labelText: '名称')),
            TextField(controller: url, decoration: const InputDecoration(labelText: '订阅链接')),
          ]),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('取消')),
          FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('添加')),
        ],
      ),
    );
    if (ok == true && name.text.trim().isNotEmpty && url.text.trim().isNotEmpty) {
      await state.addProfile(name.text.trim(), url.text.trim());
    }
  }

  @override
  Widget build(BuildContext context) {
    final s = state;
    return Scaffold(
      floatingActionButton: FloatingActionButton.extended(
        onPressed: s.busy ? null : () => _add(context),
        icon: const Icon(Icons.add),
        label: const Text('添加订阅'),
      ),
      body: s.profiles.isEmpty
          ? const Center(child: Text('还没有订阅'))
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                for (final p in s.profiles)
                  Card(
                    child: ListTile(
                      leading: Radio<String>(
                        value: p.name,
                        groupValue: s.currentProfile,
                        onChanged: s.busy ? null : (v) => s.useProfile(v!),
                      ),
                      title: Text(p.name),
                      subtitle: Text([
                        if (p.total != null) '${formatBytes(p.used ?? 0)} / ${formatBytes(p.total!)}',
                        if (p.expire != null) '到期 ${_date(p.expire!)}',
                        if (p.updated != null) '更新于 ${_date(p.updated!)}',
                      ].join(' · ')),
                      trailing: Row(mainAxisSize: MainAxisSize.min, children: [
                        IconButton(
                          tooltip: '更新',
                          icon: const Icon(Icons.refresh),
                          onPressed: s.busy ? null : () => s.updateProfile(p.name),
                        ),
                        IconButton(
                          tooltip: '删除',
                          icon: const Icon(Icons.delete_outline),
                          onPressed: s.busy ? null : () => _confirmRemove(context, p.name),
                        ),
                      ]),
                    ),
                  ),
              ],
            ),
    );
  }

  Future<void> _confirmRemove(BuildContext context, String name) async {
    final ok = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('删除 $name？'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(context, false), child: const Text('取消')),
          FilledButton(onPressed: () => Navigator.pop(context, true), child: const Text('删除')),
        ],
      ),
    );
    if (ok == true) await state.removeProfile(name);
  }
}

String _date(DateTime d) {
  final l = d.toLocal();
  String two(int n) => n.toString().padLeft(2, '0');
  return '${l.year}-${two(l.month)}-${two(l.day)}';
}
