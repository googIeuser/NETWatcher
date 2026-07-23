import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/app.dart';
import 'package:netwatcher/app_state.dart';
import 'package:netwatcher/mock_core_service.dart';
import 'package:netwatcher/models.dart';

class _ImmediateReportService extends MockCoreService {
  @override
  Future<ReportResult> generateEvidenceReport(int days) async {
    return ReportResult(
      kind: 'evidence_${days}d',
      path: 'C:\\Temp\\netwatcher_evidence_${days}d.html',
      createdAt: DateTime.now().toIso8601String(),
      message: 'Preview report created.',
    );
  }
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  testWidgets('reports page exposes working report actions', (tester) async {
    await tester.binding.setSurfaceSize(const Size(1366, 900));
    final state = await AppState.create(
      service: _ImmediateReportService(),
      pollSnapshots: false,
      manageWindowsStartup: false,
    );

    addTearDown(() async {
      state.dispose();
      await tester.binding.setSurfaceSize(null);
    });

    await tester.pumpWidget(NetWatcherApp(state: state));
    await tester.pumpAndSettle();

    await tester.tap(find.byKey(const ValueKey<String>('nav-3')));
    await tester.pumpAndSettle();

    expect(find.text('ISP Evidence Report'), findsOneWidget);
    expect(find.text('Create and open evidence report'), findsOneWidget);
    expect(find.text('Coming in core integration'), findsNothing);

    await tester.tap(find.text('Create and open evidence report'));
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 500));

    expect(find.text('Open file'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });
}
