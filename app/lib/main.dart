import 'dart:async';
import 'dart:io';

import 'package:app_links/app_links.dart';
import 'package:flutter/material.dart';
import 'package:tray_manager/tray_manager.dart';
import 'package:window_manager/window_manager.dart';

import 'core/backend.dart';
import 'state/app_state.dart';
import 'ui/home.dart';

bool get isDesktop => Platform.isLinux || Platform.isMacOS || Platform.isWindows;

Future<void> main(List<String> args) async {
  WidgetsFlutterBinding.ensureInitialized();
  final state = AppState(Backend());
  // Started at login: stay in the tray.
  final minimized = args.contains('--minimized');
  if (isDesktop) {
    await windowManager.ensureInitialized();
    await windowManager.waitUntilReadyToShow(
      const WindowOptions(size: Size(1000, 700), minimumSize: Size(720, 480), title: 'ClashOn'),
      () async {
        if (minimized) return;
        await windowManager.show();
        await windowManager.focus();
      },
    );
  }
  runApp(ClashOnApp(state: state));
  unawaited(state.init());
  // Keep clash:// pointing at this build; macOS uses Info.plist instead.
  if (Platform.isLinux || Platform.isWindows) {
    unawaited(state.backend.registerSchemes(Platform.resolvedExecutable).catchError((_) {}));
  }
}

class ClashOnApp extends StatefulWidget {
  const ClashOnApp({super.key, required this.state});
  final AppState state;

  @override
  State<ClashOnApp> createState() => _ClashOnAppState();
}

class _ClashOnAppState extends State<ClashOnApp> with WindowListener, TrayListener {
  Timer? _poll;
  StreamSubscription<Uri>? _links;

  @override
  void initState() {
    super.initState();
    // Pick up core started/stopped from the CLI or TUI.
    _poll = Timer.periodic(const Duration(seconds: 5), (_) => widget.state.refresh());
    // Covers the launch link and links forwarded from later launches.
    _links = AppLinks().uriLinkStream.listen((uri) async {
      if (isDesktop) await _show();
      await widget.state.importLink(uri.toString());
    });
    if (isDesktop) {
      windowManager.addListener(this);
      windowManager.setPreventClose(true);
      trayManager.addListener(this);
      _initTray();
      widget.state.addListener(_updateTrayMenu);
    }
  }

  Future<void> _initTray() async {
    try {
      await trayManager.setIcon(Platform.isWindows ? 'assets/tray.ico' : 'assets/tray.png');
      await _updateTrayMenu();
    } catch (_) {
      // No tray on this desktop (e.g. GNOME without AppIndicator); closing
      // quits instead, and a minimized start must not leave the app invisible.
      await _show();
    }
  }

  bool? _trayRunning;
  bool? _trayProxy;

  Future<void> _updateTrayMenu() async {
    final s = widget.state;
    if (_trayRunning == s.running && _trayProxy == s.systemProxy) return;
    _trayRunning = s.running;
    _trayProxy = s.systemProxy;
    await trayManager.setContextMenu(
      Menu(
        items: [
          MenuItem(key: 'show', label: '显示窗口'),
          MenuItem.separator(),
          MenuItem.checkbox(key: 'core', label: '内核运行', checked: s.running),
          MenuItem.checkbox(key: 'sysproxy', label: '系统代理', checked: s.systemProxy),
          MenuItem.separator(),
          MenuItem(key: 'quit', label: '退出'),
        ],
      ),
    );
  }

  @override
  void onTrayIconMouseDown() => _show();

  @override
  void onTrayIconRightMouseDown() => trayManager.popUpContextMenu();

  @override
  void onTrayMenuItemClick(MenuItem item) {
    final s = widget.state;
    switch (item.key) {
      case 'show':
        _show();
      case 'core':
        s.toggleCore();
      case 'sysproxy':
        s.setSystemProxy(!s.systemProxy);
      case 'quit':
        _quit();
    }
  }

  Future<void> _show() async {
    await windowManager.show();
    await windowManager.focus();
  }

  /// Closing hides to the tray; the core keeps running in the background.
  @override
  Future<void> onWindowClose() async {
    if (_trayRunning != null) {
      await windowManager.hide();
    } else {
      await _quit();
    }
  }

  Future<void> _quit() async {
    // Do not leave the system pointing at a proxy that may go away.
    if (widget.state.systemProxy) await widget.state.setSystemProxy(false);
    await trayManager.destroy();
    await windowManager.destroy();
  }

  @override
  void dispose() {
    _poll?.cancel();
    _links?.cancel();
    windowManager.removeListener(this);
    trayManager.removeListener(this);
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'ClashOn',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(colorSchemeSeed: Colors.indigo, useMaterial3: true),
      darkTheme: ThemeData(
        colorSchemeSeed: Colors.indigo,
        brightness: Brightness.dark,
        useMaterial3: true,
      ),
      home: HomePage(state: widget.state),
    );
  }
}
