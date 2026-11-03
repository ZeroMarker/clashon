import 'dart:io';

import 'package:path/path.dart' as p;

/// Diagnostic log in the system temp dir, for bug reports: links received,
/// errors, startup arguments.
final logFile = File(p.join(Directory.systemTemp.path, 'clashon-gui.log'));

void log(String message) {
  try {
    logFile.writeAsStringSync(
      '${DateTime.now().toIso8601String()} $message\n',
      mode: FileMode.append,
      flush: true,
    );
  } catch (_) {
    // Logging must never break the app.
  }
}
