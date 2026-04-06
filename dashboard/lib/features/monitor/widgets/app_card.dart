import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:sys_dashboard/core/models/app_metrics.dart';
import 'package:sys_dashboard/core/services/websocket_service.dart';
import 'package:sys_dashboard/core/theme.dart';
import 'package:sys_dashboard/features/logs/log_viewer.dart';

class AppCard extends ConsumerStatefulWidget {
  final AppMetrics metric;
  const AppCard({super.key, required this.metric});

  @override
  ConsumerState<AppCard> createState() => _AppCardState();
}

class _AppCardState extends ConsumerState<AppCard> {
  final List<FlSpot> _cpuData = [];
  double _timeCount = 0;

  @override
  void didUpdateWidget(AppCard oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.metric.timestamp != widget.metric.timestamp) {
      _timeCount++;
      _cpuData.add(FlSpot(_timeCount, widget.metric.cpu));
      if (_cpuData.length > 30) {
        _cpuData.removeAt(0); 
      }
    }
  }

  void _sendCommand(String command) {
    ref.read(webSocketServiceProvider).sendCommand(command, widget.metric.appId);
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text("\$command sent to \${widget.metric.appId}")),
    );
  }

  @override
  Widget build(BuildContext context) {
    final statusColor = widget.metric.status == 'running' ? AppTheme.success : AppTheme.error;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Expanded(
                  child: Text(
                    widget.metric.appId,
                    style: Theme.of(context).textTheme.titleLarge,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                  decoration: BoxDecoration(
                    color: statusColor.withOpacity(0.2),
                    borderRadius: BorderRadius.circular(12),
                    border: Border.all(color: statusColor.withOpacity(0.5)),
                  ),
                  child: Text(
                    widget.metric.status.toUpperCase(),
                    style: TextStyle(color: statusColor, fontSize: 12, fontWeight: FontWeight.bold),
                  ),
                )
              ],
            ),
            const SizedBox(height: 16),
            
            Row(
              children: [
                _buildStatColumn('CPU', '\${widget.metric.cpu.toStringAsFixed(1)}%'),
                const SizedBox(width: 24),
                _buildStatColumn('RAM', '\${widget.metric.memory} MB'),
              ],
            ),
            
            const SizedBox(height: 20),
            
            Expanded(
              child: _cpuData.isEmpty
                  ? const Center(child: Text("Waiting for data..."))
                  : LineChart(
                      LineChartData(
                        gridData: const FlGridData(show: false),
                        titlesData: const FlTitlesData(show: false),
                        borderData: FlBorderData(show: false),
                        lineBarsData: [
                          LineChartBarData(
                            spots: _cpuData,
                            isCurved: true,
                            color: AppTheme.primary,
                            barWidth: 3,
                            isStrokeCapRound: true,
                            dotData: const FlDotData(show: false),
                            belowBarData: BarAreaData(
                              show: true,
                              color: AppTheme.primary.withOpacity(0.1),
                            ),
                          ),
                        ],
                      ),
                    ),
            ),
            
            const SizedBox(height: 16),
            
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              children: [
                _buildControlButton(Icons.play_arrow, 'Start', () => _sendCommand('START_SERVICE'), AppTheme.success),
                _buildControlButton(Icons.stop, 'Stop', () => _sendCommand('STOP_SERVICE'), AppTheme.error),
                _buildControlButton(Icons.refresh, 'Restart', () => _sendCommand('RESTART_SERVICE'), AppTheme.warning),
                _buildControlButton(Icons.receipt_long, 'Logs', () {
                  showDialog(
                    context: context,
                    builder: (context) => LogViewer(appId: widget.metric.appId),
                  );
                }, AppTheme.accent),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatColumn(String label, String value) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: Theme.of(context).textTheme.bodyMedium),
        const SizedBox(height: 4),
        Text(value, style: Theme.of(context).textTheme.titleLarge?.copyWith(fontSize: 20)),
      ],
    );
  }

  Widget _buildControlButton(IconData icon, String tooltip, VoidCallback onTap, Color color) {
    return Tooltip(
      message: tooltip,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(8),
        child: Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: color.withOpacity(0.1),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, color: color, size: 20),
        ),
      ),
    );
  }
}
