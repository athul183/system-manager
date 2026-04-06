import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import '../models/app_metrics.dart';

// Provides the parsed metrics stream globally
final metricsProvider = StreamProvider<List<AppMetrics>>((ref) {
  return ref.watch(webSocketServiceProvider).metricsStream;
});

final webSocketServiceProvider = Provider<WebSocketService>((ref) {
  final service = WebSocketService();
  ref.onDispose(() => service.dispose());
  return service;
});

class WebSocketService {
  WebSocketChannel? _channel;
  
  Stream<List<AppMetrics>> get metricsStream async* {
    const wsUrl = 'ws://127.0.0.1:8080/ws?token=mock_jwt_for_ui_since_no_login_screen_yet';
    _channel = WebSocketChannel.connect(Uri.parse(wsUrl));

    await for (final message in _channel!.stream) {
      if (message is String) {
        // Use Isolate (compute) to parse JSON off the main UI thread to guarantee 60 FPS
        final metrics = await compute(_parseMetrics, message);
        yield metrics;
      }
    }
  }

  void sendCommand(String command, String appId) {
    if (_channel != null) {
      _channel!.sink.add(jsonEncode({
        'command': command,
        'app_id': appId,
      }));
    }
  }

  void dispose() {
    _channel?.sink.close();
  }
}

// Global top-level function for compute Isolate
List<AppMetrics> _parseMetrics(String message) {
  try {
    final decoded = jsonDecode(message) as Map<String, dynamic>;
    if (decoded['type'] == 'metrics') {
      final payload = decoded['payload'] as List<dynamic>;
      return payload.map((e) => AppMetrics.fromJson(e as Map<String, dynamic>)).toList();
    }
  } catch (e) {
    debugPrint("Failed to parse metrics in Isolate: \$e");
  }
  return [];
}
