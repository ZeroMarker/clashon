import 'dart:async';
import 'dart:collection';

import 'package:flutter/foundation.dart';

import '../core/api.dart';
import '../core/backend.dart';

/// Single source of truth for the UI. Lifecycle and profiles go through the
/// CLI backend; live data comes from the mihomo controller.
class AppState extends ChangeNotifier {
  AppState(this.backend, {MihomoApi Function(String controller, String secret)? apiFactory})
      : _apiFactory = apiFactory ?? MihomoApi.new;

  final Backend backend;
  final MihomoApi Function(String, String) _apiFactory;

  CoreStatus? status;
  MihomoApi? api;
  String? error;
  bool busy = false;

  List<Profile> profiles = [];
  String currentProfile = '';

  Map<String, Proxy> proxies = {};
  List<String> groups = [];
  final Set<String> testing = {};

  String mode = 'rule';
  bool systemProxy = false;

  static const historyLength = 60;
  final ListQueue<Traffic> trafficHistory = ListQueue();
  Traffic get traffic => trafficHistory.isEmpty ? const Traffic(0, 0) : trafficHistory.last;

  static const maxLogs = 500;
  final ListQueue<LogLine> logs = ListQueue();

  List<Connection> connections = [];
  int uploadTotal = 0, downloadTotal = 0;

  StreamSubscription<Traffic>? _trafficSub;
  StreamSubscription<LogLine>? _logSub;
  Timer? _connTimer;

  bool get running => status?.reachable ?? false;

  Future<void> init() async {
    await refresh();
    await _guard(() async => systemProxy = await backend.systemProxy());
  }

  /// Reloads status and profiles, (re)connecting to the core if it is up.
  Future<void> refresh() => _guard(() async {
        final st = await backend.status();
        final (list, current) = await backend.profiles();
        profiles = list;
        currentProfile = current;
        final wasRunning = running;
        status = st;
        if (st.reachable && !wasRunning) {
          _connect(st);
        } else if (!st.reachable && wasRunning) {
          _disconnect();
        }
        if (running) {
          mode = st.mode ?? mode;
          await _loadProxies();
        }
      });

  Future<void> toggleCore() => _guard(() async {
        if (status?.running ?? false) {
          await backend.down();
        } else {
          await backend.up();
        }
        await refresh();
      }, busyFlag: true);

  Future<void> addProfile(String name, String url) =>
      _guard(() async {
        await backend.addProfile(name, url);
        await refresh();
      }, busyFlag: true);

  Future<void> updateProfile(String name) => _guard(() async {
        await backend.updateProfile(name);
        await refresh();
      }, busyFlag: true);

  Future<void> useProfile(String name) => _guard(() async {
        await backend.useProfile(name);
        await refresh();
      }, busyFlag: true);

  Future<void> removeProfile(String name) => _guard(() async {
        await backend.removeProfile(name);
        await refresh();
      });

  Future<void> setSystemProxy(bool on) => _guard(() async {
        await backend.setSystemProxy(on);
        systemProxy = on;
      });

  Future<void> setMode(String m) => _guard(() async {
        await api!.setMode(m);
        mode = m;
      });

  Future<void> select(String group, String proxy) => _guard(() async {
        await api!.select(group, proxy);
        await _loadProxies();
      });

  Future<void> testGroup(String group) async {
    if (api == null || testing.contains(group)) return;
    testing.add(group);
    notifyListeners();
    await _guard(() async {
      try {
        await api!.testGroup(group);
      } finally {
        testing.remove(group);
      }
      await _loadProxies();
    });
  }

  Future<void> closeConnection([String? id]) => _guard(() async {
        await api!.closeConnection(id);
        await _loadConnections();
      });

  /// Polls connections while a view needs them.
  void watchConnections(bool on) {
    _connTimer?.cancel();
    _connTimer = null;
    if (on && running) {
      _loadConnections();
      _connTimer = Timer.periodic(const Duration(seconds: 1), (_) => _loadConnections());
    }
  }

  void clearError() {
    error = null;
    notifyListeners();
  }

  void _connect(CoreStatus st) {
    final a = api = _apiFactory(st.controller, st.secret);
    _trafficSub = a.traffic().listen((t) {
      trafficHistory.addLast(t);
      while (trafficHistory.length > historyLength) {
        trafficHistory.removeFirst();
      }
      notifyListeners();
    }, onError: (_) {});
    _logSub = a.logs().listen((l) {
      logs.addLast(l);
      while (logs.length > maxLogs) {
        logs.removeFirst();
      }
      notifyListeners();
    }, onError: (_) {});
  }

  void _disconnect() {
    _trafficSub?.cancel();
    _logSub?.cancel();
    _connTimer?.cancel();
    api?.close();
    api = null;
    proxies = {};
    groups = [];
    connections = [];
    trafficHistory.clear();
  }

  Future<void> _loadProxies() async {
    final p = await api!.proxies();
    proxies = p;
    // GLOBAL lists groups in config order.
    groups = [
      ...?p['GLOBAL']?.all.where((n) => p[n]?.isGroup ?? false),
      if (p.containsKey('GLOBAL')) 'GLOBAL',
    ];
  }

  Future<void> _loadConnections() async {
    if (api == null) return;
    try {
      final (list, up, down) = await api!.connections();
      connections = list;
      uploadTotal = up;
      downloadTotal = down;
      notifyListeners();
    } catch (_) {
      // transient; the next tick retries
    }
  }

  Future<void> _guard(Future<void> Function() fn, {bool busyFlag = false}) async {
    if (busyFlag) {
      busy = true;
      notifyListeners();
    }
    try {
      await fn();
      error = null;
    } catch (e) {
      error = e.toString();
    } finally {
      if (busyFlag) busy = false;
      notifyListeners();
    }
  }

  @override
  void dispose() {
    _disconnect();
    super.dispose();
  }
}

String formatBytes(int n) {
  const units = ['B', 'KiB', 'MiB', 'GiB', 'TiB'];
  var v = n.toDouble();
  var i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return i == 0 ? '$n B' : '${v.toStringAsFixed(1)} ${units[i]}';
}
