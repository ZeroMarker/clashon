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
      secret = j['secret'] as String? ?? '',
      tun = j['tun'] as bool? ?? false,
      tunReady = j['tunReady'] as bool? ?? false;

  final bool running;
  final bool reachable;
  final int? pid;
  final String? version;
  final String? profile;
  final String? mode;
  final int mixedPort;
  final String controller;
  final String secret;
  final bool tun;
  final bool tunReady;
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

  /// Finds the CLI: $CLASHON_BIN, bundled with the app, then PATH.
  /// It sits next to the executable, except on macOS where it lives in
  /// Contents/Resources (the case-insensitive filesystem would clash with
  /// the ClashOn executable).
  static String locate() {
    final env = Platform.environment['CLASHON_BIN'];
    if (env != null && env.isNotEmpty) return env;
    final name = Platform.isWindows ? 'clashon.exe' : 'clashon';
    final dir = p.dirname(Platform.resolvedExecutable);
    final bundled = Platform.isMacOS ? p.join(dir, '..', 'Resources', name) : p.join(dir, name);
    if (File(bundled).existsSync()) return bundled;
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

  Future<void> setTun(bool on) => _run(['tun', on ? 'on' : 'off']);

  /// Adds a subscription from a clash:// link and returns its name.
  Future<String> importLink(String link) async =>
      (jsonDecode(await _run(['import', link])) as Map<String, dynamic>)['name'] as String;

  Future<bool> autostart() async =>
      (jsonDecode(await _run(['desktop', 'autostart'])) as Map<String, dynamic>)['enabled']
          as bool? ??
      false;

  Future<void> setAutostart(bool on, String exe) =>
      _run(['desktop', 'autostart', on ? 'on' : 'off', '--exec', exe]);

  Future<void> registerSchemes(String exe) => _run(['desktop', 'register', '--exec', exe]);

  /// Returns the newer release (version, url), or null if up to date.
  Future<(String, String)?> checkUpdate() async {
    final j = jsonDecode(await _run(['update', 'check'])) as Map<String, dynamic>;
    if (j['newer'] != true) return null;
    return (j['latest'] as String, j['url'] as String);
  }
}

/// Opens [url] in the default browser.
Future<void> openUrl(String url) async {
  final (cmd, args) = Platform.isWindows
      ? ('rundll32', ['url.dll,FileProtocolHandler', url])
      : Platform.isMacOS
      ? ('open', [url])
      : ('xdg-open', [url]);
  await Process.run(cmd, args);
}
