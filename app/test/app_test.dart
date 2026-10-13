import 'dart:convert';

import 'package:clashon/core/api.dart';
import 'package:clashon/core/backend.dart';
import 'package:clashon/state/app_state.dart';
import 'package:clashon/ui/home.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';

class FakeBackend implements Backend {
  bool running = false;
  bool sysproxy = false;
  final calls = <String>[];

  @override
  String get binary => 'fake';

  @override
  Future<CoreStatus> status() async => CoreStatus.fromJson({
        'running': running,
        'reachable': running,
        'profile': 'test',
        'mode': 'rule',
        'mixedPort': 7890,
        'controller': '127.0.0.1:9090',
        'secret': 's',
      });

  @override
  Future<void> up() async => running = true;
  @override
  Future<void> down() async => running = false;

  @override
  Future<(List<Profile>, String)> profiles() async => (
        [
          Profile.fromJson({
            'name': 'test',
            'url': 'https://x',
            'updated': '2026-10-13T06:00:00Z',
            'userinfo': {'upload': 1024, 'download': 1024, 'total': 1 << 30}
          })
        ],
        'test'
      );

  @override
  Future<void> addProfile(String name, String url) async => calls.add('add $name');
  @override
  Future<void> updateProfile(String name) async => calls.add('update $name');
  @override
  Future<void> useProfile(String name) async => calls.add('use $name');
  @override
  Future<void> removeProfile(String name) async => calls.add('rm $name');
  @override
  Future<void> setSystemProxy(bool on) async => sysproxy = on;
  @override
  Future<bool> systemProxy() async => sysproxy;
}

MihomoApi fakeApi(String controller, String secret) {
  final client = MockClient((req) async {
    expect(req.headers['Authorization'], 'Bearer s');
    switch (req.url.path) {
      case '/proxies':
        return http.Response(
            jsonEncode({
              'proxies': {
                'GLOBAL': {'name': 'GLOBAL', 'type': 'Selector', 'now': 'Proxy', 'all': ['Proxy']},
                'Proxy': {'name': 'Proxy', 'type': 'Selector', 'now': 'HK', 'all': ['HK', 'US']},
                'HK': {'name': 'HK', 'type': 'Trojan', 'history': [{'delay': 88}]},
                'US': {'name': 'US', 'type': 'Shadowsocks', 'history': []},
              }
            }),
            200);
      case '/proxies/Proxy':
        return http.Response('', 204);
      case '/traffic':
        return http.Response('{"up":1,"down":2048}\n', 200);
      case '/logs':
        return http.Response('{"type":"info","payload":"hello"}\n', 200);
    }
    return http.Response('{"message":"not found"}', 404);
  });
  return MihomoApi(controller, secret, client: client);
}

void main() {
  test('Connection parses chain and target', () {
    final c = Connection.fromJson({
      'id': '1',
      'metadata': {'host': '', 'destinationIP': '1.1.1.1', 'destinationPort': '443'},
      'chains': ['HK', 'Proxy'],
      'rule': 'Match',
      'rulePayload': '',
    });
    expect(c.host, '1.1.1.1:443');
    expect(c.chain, 'Proxy → HK');
    expect(c.rule, 'Match');
  });

  test('formatBytes', () {
    expect(formatBytes(512), '512 B');
    expect(formatBytes(1536), '1.5 KiB');
  });

  test('AppState starts core, loads groups and streams', () async {
    final b = FakeBackend();
    final s = AppState(b, apiFactory: fakeApi);
    await s.init();
    expect(s.running, isFalse);
    expect(s.profiles.single.total, 1 << 30);

    await s.toggleCore();
    expect(s.running, isTrue);
    expect(s.groups, ['Proxy', 'GLOBAL']);
    expect(s.proxies['HK']!.delay, 88);
    await Future<void>.delayed(const Duration(milliseconds: 50));
    expect(s.traffic.down, 2048);
    expect(s.logs.single.payload, 'hello');

    await s.select('Proxy', 'US');
    expect(s.error, isNull);
    await s.setSystemProxy(true);
    expect(b.sysproxy, isTrue);

    await s.toggleCore();
    expect(s.running, isFalse);
    expect(s.groups, isEmpty);
    s.dispose();
  });

  testWidgets('pages render', (tester) async {
    tester.view.physicalSize = const Size(1200, 800);
    tester.view.devicePixelRatio = 1;
    addTearDown(tester.view.reset);

    final s = AppState(FakeBackend(), apiFactory: fakeApi);
    await tester.runAsync(s.init);
    await tester.pumpWidget(MaterialApp(home: HomePage(state: s)));
    expect(find.text('已停止'), findsOneWidget);

    await tester.runAsync(s.toggleCore);
    await tester.pump();
    expect(find.text('运行中'), findsOneWidget);

    await tester.tap(find.text('代理'));
    await tester.pump();
    expect(find.text('HK'), findsOneWidget);
    expect(find.text('88ms'), findsOneWidget);

    await tester.tap(find.text('订阅'));
    await tester.pump();
    expect(find.textContaining('2.0 KiB / 1.0 GiB'), findsOneWidget);

    await tester.tap(find.text('日志'));
    await tester.pump();
    expect(find.textContaining('hello', findRichText: true), findsOneWidget);

    await tester.runAsync(s.toggleCore);
    s.dispose();
  });
}
