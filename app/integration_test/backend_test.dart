// Runs the GUI state layer against the real clashon CLI. Needs CLASHON_BIN,
// CLASHON_HOME (a scratch dir) and CLASHON_TEST_SUB_URL; skipped otherwise.
import 'dart:io';

import 'package:clashon/core/backend.dart';
import 'package:clashon/state/app_state.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();
  final env = Platform.environment;
  final url = env['CLASHON_TEST_SUB_URL'] ?? '';
  final skip = url.isEmpty || (env['CLASHON_BIN'] ?? '').isEmpty || (env['CLASHON_HOME'] ?? '').isEmpty
      ? 'needs CLASHON_BIN, CLASHON_HOME and CLASHON_TEST_SUB_URL'
      : null;

  testWidgets('add, import and switch subscriptions with Chinese names', (tester) async {
    final s = AppState(Backend());
    await tester.runAsync(() async {
      await s.init();
      expect(s.error, isNull, reason: 'init');

      await s.addProfile('一分机场 测试', url);
      expect(s.error, isNull, reason: 'add');
      expect(s.profiles.map((p) => p.name), contains('一分机场 测试'));
      expect(s.profiles.firstWhere((p) => p.name == '一分机场 测试').total, isNotNull);

      final link = 'clash://install-config?url=${Uri.encodeComponent(url)}&name=${Uri.encodeComponent('链接导入')}';
      await s.importLink(link);
      expect(s.error, isNull, reason: 'import');
      expect(s.notice, contains('链接导入'));

      await s.useProfile('链接导入');
      expect(s.error, isNull, reason: 'use');
      expect(s.currentProfile, '链接导入');

      await s.toggleCore();
      expect(s.error, isNull, reason: 'up');
      expect(s.running, isTrue);
      expect(s.groups, isNotEmpty);
      await s.toggleCore();
      expect(s.running, isFalse);
    });
    s.dispose();
  }, skip: skip != null);
}
