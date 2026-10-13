import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;

class ApiException implements Exception {
  ApiException(this.message);
  final String message;
  @override
  String toString() => message;
}

class Proxy {
  Proxy.fromJson(Map<String, dynamic> j)
      : name = j['name'] as String,
        type = j['type'] as String? ?? '',
        now = j['now'] as String?,
        all = (j['all'] as List?)?.cast<String>() ?? const [],
        delay = _lastDelay(j['history']);

  final String name;
  final String type;
  final String? now;
  final List<String> all;
  final int delay; // 0 = unknown or timeout

  bool get isGroup => all.isNotEmpty;
  bool get selectable => type == 'Selector';

  static int _lastDelay(Object? h) {
    if (h is! List || h.isEmpty) return 0;
    return (h.last as Map)['delay'] as int? ?? 0;
  }
}

class Connection {
  Connection.fromJson(Map<String, dynamic> j)
      : id = j['id'] as String,
        host = _target(j['metadata'] as Map? ?? const {}),
        rule = [j['rule'], j['rulePayload']].where((e) => e != null && e != '').join(','),
        chain = ((j['chains'] as List?)?.cast<String>() ?? const []).reversed.join(' → '),
        upload = j['upload'] as int? ?? 0,
        download = j['download'] as int? ?? 0,
        start = DateTime.tryParse(j['start'] as String? ?? '') ?? DateTime.now();

  final String id;
  final String host;
  final String rule;
  final String chain;
  final int upload;
  final int download;
  final DateTime start;

  static String _target(Map m) {
    final h = (m['host'] as String?)?.isNotEmpty == true ? m['host'] : m['destinationIP'];
    return '$h:${m['destinationPort']}';
  }
}

class Traffic {
  const Traffic(this.up, this.down);
  final int up;
  final int down;
}

class LogLine {
  LogLine(this.type, this.payload) : time = DateTime.now();
  final String type;
  final String payload;
  final DateTime time;
}

/// Client for the mihomo external-controller REST API.
class MihomoApi {
  MihomoApi(String controller, this.secret, {http.Client? client})
      : base = controller.contains('://') ? controller : 'http://$controller',
        _http = client ?? http.Client();

  final String base;
  final String secret;
  final http.Client _http;

  Map<String, String> get _headers => {
        if (secret.isNotEmpty) 'Authorization': 'Bearer $secret',
        'Content-Type': 'application/json',
      };

  Uri _uri(String path, [Map<String, String>? q]) =>
      Uri.parse(base).replace(path: path, queryParameters: q);

  Future<dynamic> _send(String method, String path, {Object? body, Map<String, String>? query}) async {
    final req = http.Request(method, _uri(path, query))..headers.addAll(_headers);
    if (body != null) req.body = jsonEncode(body);
    final res = await http.Response.fromStream(await _http.send(req).timeout(const Duration(seconds: 15)));
    if (res.statusCode >= 300) {
      var msg = 'HTTP ${res.statusCode}';
      try {
        msg = (jsonDecode(res.body) as Map)['message'] as String? ?? msg;
      } catch (_) {}
      throw ApiException(msg);
    }
    return res.body.isEmpty ? null : jsonDecode(utf8.decode(res.bodyBytes));
  }

  Future<Map<String, Proxy>> proxies() async {
    final j = await _send('GET', '/proxies') as Map<String, dynamic>;
    return (j['proxies'] as Map<String, dynamic>)
        .map((k, v) => MapEntry(k, Proxy.fromJson(v as Map<String, dynamic>)));
  }

  Future<void> select(String group, String proxy) =>
      _send('PUT', '/proxies/${Uri.encodeComponent(group)}', body: {'name': proxy});

  Future<Map<String, int>> testGroup(String group,
      {String url = 'https://www.gstatic.com/generate_204', int timeoutMs = 5000}) async {
    final j = await _send('GET', '/group/${Uri.encodeComponent(group)}/delay',
        query: {'url': url, 'timeout': '$timeoutMs'}) as Map<String, dynamic>;
    return j.map((k, v) => MapEntry(k, v as int));
  }

  Future<String> mode() async => ((await _send('GET', '/configs')) as Map)['mode'] as String;

  Future<void> setMode(String mode) => _send('PATCH', '/configs', body: {'mode': mode});

  Future<(List<Connection>, int, int)> connections() async {
    final j = await _send('GET', '/connections') as Map<String, dynamic>;
    final list = (j['connections'] as List? ?? [])
        .map((e) => Connection.fromJson(e as Map<String, dynamic>))
        .toList()
      ..sort((a, b) => b.start.compareTo(a.start));
    return (list, j['uploadTotal'] as int? ?? 0, j['downloadTotal'] as int? ?? 0);
  }

  Future<void> closeConnection([String? id]) =>
      _send('DELETE', id == null ? '/connections' : '/connections/${Uri.encodeComponent(id)}');

  Stream<Traffic> traffic() => _stream('/traffic').map(
      (j) => Traffic(j['up'] as int? ?? 0, j['down'] as int? ?? 0));

  Stream<LogLine> logs({String level = 'info'}) => _stream('/logs', {'level': level}).map(
      (j) => LogLine(j['type'] as String? ?? '', j['payload'] as String? ?? ''));

  /// Reads newline-delimited JSON from a streaming endpoint.
  Stream<Map<String, dynamic>> _stream(String path, [Map<String, String>? q]) async* {
    final req = http.Request('GET', _uri(path, q))..headers.addAll(_headers);
    final res = await _http.send(req);
    if (res.statusCode >= 300) throw ApiException('$path: HTTP ${res.statusCode}');
    await for (final line in res.stream.transform(utf8.decoder).transform(const LineSplitter())) {
      if (line.isEmpty) continue;
      try {
        yield jsonDecode(line) as Map<String, dynamic>;
      } on FormatException {
        continue;
      }
    }
  }

  void close() => _http.close();
}
