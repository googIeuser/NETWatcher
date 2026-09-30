import 'package:flutter/material.dart';

import 'app_state.dart';
import 'models.dart';
import 'widgets.dart';

class RestoredSettingsPage extends StatefulWidget {
  const RestoredSettingsPage({super.key, required this.state});

  final AppState state;

  @override
  State<RestoredSettingsPage> createState() => _RestoredSettingsPageState();
}

class _RestoredSettingsPageState extends State<RestoredSettingsPage> {
  late NetWatcherConfig draft;

  @override
  void initState() {
    super.initState();
    draft = widget.state.config;
  }

  @override
  Widget build(BuildContext context) => ListView(
        padding: const EdgeInsets.fromLTRB(28, 27, 28, 36),
        children: [
          LayoutBuilder(builder: (context, constraints) {
            const heading = _SettingsHeading();
            final save = FilledButton.icon(
              onPressed: () => widget.state.saveConfig(draft),
              icon: const Icon(Icons.save_outlined),
              label: const Text('Save settings'),
            );
            if (constraints.maxWidth < 650) {
              return Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [heading, const SizedBox(height: 12), save],
              );
            }
            return Row(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                const Expanded(child: heading),
                const SizedBox(width: 20),
                save,
              ],
            );
          }),
          const SizedBox(height: 18),
          LayoutBuilder(builder: (context, constraints) {
            final monitoring = Panel(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const _SettingGroupTitle('01', 'Appearance & monitoring'),
                  const SizedBox(height: 22),
                  _SettingField(
                    label: 'Theme',
                    child: DropdownButtonFormField<String>(
                      initialValue: draft.theme,
                      dropdownColor: Theme.of(context).colorScheme.surface,
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.onSurface,
                        fontFamily: 'Bahnschrift',
                        fontSize: 16,
                      ),
                      items: const [
                        DropdownMenuItem(value: 'dark', child: Text('Dark')),
                        DropdownMenuItem(value: 'light', child: Text('Light')),
                      ],
                      onChanged: (value) =>
                          setState(() => draft = draft.copyWith(theme: value)),
                    ),
                  ),
                  _SettingField(
                    label: 'Monitoring interval',
                    child: TextFormField(
                      initialValue: draft.intervalSeconds.toString(),
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(suffixText: 'seconds'),
                      onChanged: (value) => draft = draft.copyWith(
                        intervalSeconds:
                            double.tryParse(value) ?? draft.intervalSeconds,
                      ),
                    ),
                  ),
                  SwitchListTile(
                    contentPadding: EdgeInsets.zero,
                    value: draft.startMonitoringAutomatically,
                    title: const Text('Start monitoring automatically'),
                    onChanged: (value) => setState(() => draft =
                        draft.copyWith(startMonitoringAutomatically: value)),
                  ),
                ],
              ),
            );
            final startup = Panel(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  const _SettingGroupTitle('02', 'Startup & alerts'),
                  const SizedBox(height: 12),
                  SwitchListTile(
                    contentPadding: EdgeInsets.zero,
                    value: draft.startWithWindows,
                    title: const Text('Start with Windows'),
                    subtitle:
                        const Text('Adds NetWatcher to your startup list.'),
                    onChanged: (value) => setState(
                        () => draft = draft.copyWith(startWithWindows: value)),
                  ),
                  SwitchListTile(
                    contentPadding: EdgeInsets.zero,
                    value: draft.startMinimizedToNotificationArea,
                    title: const Text('Start minimized to tray'),
                    onChanged: draft.startWithWindows
                        ? (value) => setState(() => draft = draft.copyWith(
                            startMinimizedToNotificationArea: value))
                        : null,
                  ),
                  SwitchListTile(
                    contentPadding: EdgeInsets.zero,
                    value: draft.keepRunningInTrayOnClose,
                    title: const Text('Keep running in tray on close'),
                    onChanged: (value) => setState(() => draft =
                        draft.copyWith(keepRunningInTrayOnClose: value)),
                  ),
                  SwitchListTile(
                    contentPadding: EdgeInsets.zero,
                    value: draft.showOutageNotifications,
                    title: const Text('Outage and recovery notifications'),
                    onChanged: (value) => setState(() =>
                        draft = draft.copyWith(showOutageNotifications: value)),
                  ),
                ],
              ),
            );
            if (constraints.maxWidth < 900) {
              return Column(
                children: [monitoring, const SizedBox(height: 14), startup],
              );
            }
            return Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(child: monitoring),
                const SizedBox(width: 14),
                Expanded(child: startup),
              ],
            );
          }),
        ],
      );
}

class _SettingsHeading extends StatelessWidget {
  const _SettingsHeading();

  @override
  Widget build(BuildContext context) => Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text('PREFERENCES  /  06',
              style: Theme.of(context).textTheme.labelSmall?.copyWith(
                    fontFamily: 'Consolas',
                    color: Theme.of(context).colorScheme.onSurfaceVariant,
                    fontWeight: FontWeight.w800,
                    letterSpacing: 1.8,
                  )),
          const SizedBox(height: 7),
          Text('Settings',
              style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.w900,
                    letterSpacing: -.6,
                  )),
          const SizedBox(height: 4),
          Text('Tune how NetWatcher runs on this computer.',
              style: TextStyle(
                  color: Theme.of(context).colorScheme.onSurfaceVariant)),
        ],
      );
}

class _SettingGroupTitle extends StatelessWidget {
  const _SettingGroupTitle(this.number, this.title);
  final String number;
  final String title;

  @override
  Widget build(BuildContext context) => Row(
        children: [
          Text(number,
              style: TextStyle(
                fontFamily: 'Consolas',
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              )),
          const SizedBox(width: 12),
          Expanded(
            child: Text(title,
                style: Theme.of(context)
                    .textTheme
                    .titleLarge
                    ?.copyWith(fontWeight: FontWeight.w900)),
          ),
        ],
      );
}

class _SettingField extends StatelessWidget {
  const _SettingField({required this.label, required this.child});
  final String label;
  final Widget child;

  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.only(bottom: 18),
        child: LayoutBuilder(builder: (context, constraints) {
          if (constraints.maxWidth < 600) {
            return Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(label,
                    style: const TextStyle(fontWeight: FontWeight.w700)),
                const SizedBox(height: 8),
                child,
              ],
            );
          }
          return Row(children: [
            Expanded(
                child: Text(label,
                    style: const TextStyle(fontWeight: FontWeight.w700))),
            const SizedBox(width: 20),
            SizedBox(width: 260, child: child),
          ]);
        }),
      );
}
