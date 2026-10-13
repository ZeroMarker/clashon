import 'dart:convert';
import 'dart:io';

import 'package:path/path.dart' as p;

/// Error reported by the clashon CLI.
class BackendException implements Exception {
  BackendException(this.message);
  final String message;
  @override
  String toString() => message;
}

/// Status as printed by `clashon status --json`.
class CoreStatus {
  CoreStatus.fromJson(Map<String, dynamic> j)
      : running = j['running'] as bool? ?? false,
        reachable = j['reachable'] as bool? ?? false,
        pid = j['pid'] as int?,
        version = j['version'] as String?,
        profile = j['profile'] as String?,
        mode = j['mode'] as String?,
        mixedPort = j['mixedPort'] as int? ?? 0,
        controller = j['controller'] as String? ?? '',
        secret = j['secret'] as String? ?? '';

  final bool running;
  final bool reachable;
  final int? pid;
  final String? version;
  final String? profile;
  final String? mode;
  final int mixedPort;
  final String controller;
  final String secret;
}

class Profile {
  Profile.fromJson(Map<String, dynamic> j)
      : name = j['name'] as String,
        url = j['url'] as String? ?? '',
        updated = DateTime.tryParse(j['updated'] as String? ?? ''),
        used = _used(j['userinfo']),
        total = (j['userinfo'] as Map?)?['total'] as int?,
        expire = DateTime.tryParse((j['userinfo'] as Map?)?['expire'] as String? ?? '');

  final String name;
  final String url;
  final DateTime? updated;
  final int? used;
  final int? total;
  final DateTime? expire;

  static int? _used(Object? u) {
    if (u is! Map) return null;
    return (u['upload'] as int? ?? 0) + (u['download'] as int? ?? 0);
  }
}

/// Runs the bundled `clashon` CLI, which owns the core process and profiles.
class Backend {
  Backend({String? binary}) : binary = binary ?? locate();

  final String binary;

  /// Finds the CLI: $CLASHON_BIN, next to the app executable, then PATH.
  static String locate() {
    final env = Platform.environment['CLASHON_BIN'];
    if (env != null && env.isNotEmpty) return env;
    final name = Platform.isWindows ? 'clashon.exe' : 'clashon';
    final beside = p.join(p.dirname(Platform.resolvedExecutable), name);
    if (File(beside).existsSync()) return beside;
    return name;
  }

  Future<String> _run(List<String> args) async {
    final ProcessResult r;
    try {
      r = await Process.run(binary, [...args, '--json']);
    } on ProcessException catch (e) {
      throw BackendException('cannot run $binary: ${e.message}');
    }
    if (r.exitCode != 0) {
      final msg = (r.stderr as String).trim().replaceFirst(RegExp(r'^Error: '), '');
      throw BackendException(msg.isEmpty ? 'clashon ${args.join(' ')} failed' : msg);
    }
    return r.stdout as String;
  }

  Future<CoreStatus> status() async =>
      CoreStatus.fromJson(jsonDecode(await _run(['status'])) as Map<String, dynamic>);

  Future<void> up() => _run(['up']);
  Future<void> down() => _run(['down']);

  Future<(List<Profile>, String)> profiles() async {
    final j = jsonDecode(await _run(['sub', 'ls'])) as Map<String, dynamic>;
    final list = (j['profiles'] as List? ?? [])
        .map((e) => Profile.fromJson(e as Map<String, dynamic>))
        .toList();
    return (list, j['current'] as String? ?? '');
  }

  Future<void> addProfile(String name, String url) => _run(['sub', 'add', url, '-n', name]);
  Future<void> updateProfile(String name) => _run(['sub', 'update', name]);
  Future<void> useProfile(String name) => _run(['sub', 'use', name]);
  Future<void> removeProfile(String name) => _run(['sub', 'rm', name]);

  Future<void> setSystemProxy(bool on) => _run(['sysproxy', on ? 'on' : 'off']);
  Future<bool> systemProxy() async =>
      (jsonDecode(await _run(['sysproxy'])) as Map<String, dynamic>)['enabled'] as bool? ?? false;
}
