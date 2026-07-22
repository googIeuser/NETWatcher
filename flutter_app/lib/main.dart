import 'package:flutter/material.dart';

import 'app.dart';
import 'app_state.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final state = await AppState.create();
  runApp(NetWatcherApp(state: state));
}
