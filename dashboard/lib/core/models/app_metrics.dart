class AppMetrics {
  final String appId;
  final double cpu;
  final int memory;
  final String status;
  final DateTime timestamp;

  AppMetrics({
    required this.appId,
    required this.cpu,
    required this.memory,
    required this.status,
    required this.timestamp,
  });

  factory AppMetrics.fromJson(Map<String, dynamic> json) {
    return AppMetrics(
      appId: json['app_id'] as String,
      cpu: (json['cpu'] as num).toDouble(),
      memory: json['memory'] as int,
      status: json['status'] as String,
      timestamp: DateTime.fromMillisecondsSinceEpoch((json['timestamp'] as int) * 1000),
    );
  }
}
