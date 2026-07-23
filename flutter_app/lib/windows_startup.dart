import 'dart:io';

class WindowsStartup {
  static const _runKey =
      r'HKCU\Software\Microsoft\Windows\CurrentVersion\Run';
  static const _valueName = 'NetWatcher';

  static Future<void> sync(bool enabled) async {
    if (!Platform.isWindows) return;

    final result = enabled
        ? await Process.run(
            'reg.exe',
            [
              'add',
              _runKey,
              '/v',
              _valueName,
              '/t',
              'REG_SZ',
              '/d',
              '"${Platform.resolvedExecutable}" --autostart',
              '/f',
            ],
          )
        : await Process.run(
            'reg.exe',
            ['delete', _runKey, '/v', _valueName, '/f'],
          );

    // reg delete returns exit code 1 when the value is already absent.
    if (result.exitCode != 0 && (enabled || result.exitCode != 1)) {
      final details = result.stderr.toString().trim();
      throw StateError(
        details.isEmpty
            ? 'Windows startup setting could not be updated.'
            : 'Windows startup setting could not be updated: $details',
      );
    }
  }
}
