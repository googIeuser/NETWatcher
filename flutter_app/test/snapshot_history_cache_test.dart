import 'package:flutter_test/flutter_test.dart';
import 'package:netwatcher/process_core_service.dart';

Map<String, dynamic> _sample(int epochMs, {bool success = true}) => {
      'time': DateTime.fromMillisecondsSinceEpoch(epochMs, isUtc: true)
          .toIso8601String(),
      'latency': success ? 10.0 : 0.0,
      'success': success,
    };

Map<String, dynamic> _snapshot({
  required int epoch,
  required bool delta,
  required List<Map<String, dynamic>> history,
  int cutoffMs = 0,
  int bucketSeconds = 1,
}) =>
    {
      'monitoring': true,
      'historyEpoch': epoch,
      'historyDelta': delta,
      'historyBucketSeconds': bucketSeconds,
      'historyCutoffMs': cutoffMs,
      'targets': [
        {
          'target': {'id': 'cloudflare', 'name': 'Cloudflare'},
          'state': 'online',
          'history': history,
        },
      ],
    };

void main() {
  const start = 1700000000000;

  test('asks for full history until a snapshot has been received', () {
    expect(SnapshotHistoryCache().requestFields(), isEmpty);
  });

  test('appends new chart buckets to the cached history', () {
    final cache = SnapshotHistoryCache();
    cache.apply(_snapshot(epoch: 3, delta: false, history: [
      _sample(start),
      _sample(start + 2000),
    ]));

    expect(cache.requestFields(),
        {'history_after': start + 2000, 'history_epoch': 3});

    final snapshot = cache.apply(_snapshot(epoch: 3, delta: true, history: [
      _sample(start + 2000, success: false),
      _sample(start + 4000),
    ]));

    final history = snapshot.targets.single.history;
    expect(history.map((sample) => sample.time.millisecondsSinceEpoch),
        [start, start + 2000, start + 4000]);
    expect(history[1].success, isFalse);
  });

  test('drops samples older than the chart range', () {
    final cache = SnapshotHistoryCache();
    cache.apply(_snapshot(epoch: 1, delta: false, history: [
      _sample(start),
      _sample(start + 2000),
    ]));

    final snapshot = cache.apply(_snapshot(
      epoch: 1,
      delta: true,
      cutoffMs: start + 1000,
      history: [_sample(start + 4000)],
    ));

    expect(
        snapshot.targets.single.history
            .map((sample) => sample.time.millisecondsSinceEpoch),
        [start + 2000, start + 4000]);
  });

  test('replaces the cache when the history epoch changes', () {
    final cache = SnapshotHistoryCache();
    cache.apply(_snapshot(epoch: 1, delta: false, history: [_sample(start)]));

    final snapshot = cache.apply(
        _snapshot(epoch: 2, delta: true, history: [_sample(start + 60000)]));

    expect(snapshot.targets.single.history, hasLength(1));
    expect(cache.requestFields()['history_epoch'], 2);
  });

  Map<String, dynamic> event(String message) =>
      {'time': message, 'level': 'info', 'category': 'monitor', 'message': message};

  test('adds new events in front of the cached events', () {
    final cache = SnapshotHistoryCache();
    cache.apply({
      'eventCount': 2,
      'recentEvents': [event('second'), event('first')],
    });
    expect(cache.requestFields()['events_seen'], 2);

    final snapshot = cache.apply({
      'eventCount': 3,
      'eventsDelta': true,
      'recentEvents': [event('third')],
    });

    expect(snapshot.recentEvents.map((event) => event.message),
        ['third', 'second', 'first']);
    expect(cache.requestFields()['events_seen'], 3);
  });

  test('keeps at most fifty events', () {
    final cache = SnapshotHistoryCache();
    cache.apply({
      'eventCount': 50,
      'recentEvents': [for (var i = 50; i > 0; i--) event('event $i')],
    });

    final snapshot = cache.apply({
      'eventCount': 51,
      'eventsDelta': true,
      'recentEvents': [event('event 51')],
    });

    expect(snapshot.recentEvents, hasLength(50));
    expect(snapshot.recentEvents.first.message, 'event 51');
    expect(snapshot.recentEvents.last.message, 'event 2');
  });
}
