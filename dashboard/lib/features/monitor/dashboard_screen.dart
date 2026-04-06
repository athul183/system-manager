import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sys_dashboard/core/models/app_metrics.dart';
import 'package:sys_dashboard/core/services/websocket_service.dart';
import 'package:sys_dashboard/features/monitor/widgets/app_card.dart';

class DashboardScreen extends ConsumerWidget {
  const DashboardScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final metricsAsync = ref.watch(metricsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('System Manager'),
        backgroundColor: Colors.transparent,
        elevation: 0,
        actions: [
          IconButton(
            icon: const Icon(Icons.settings),
            onPressed: () {},
          )
        ],
      ),
      body: metricsAsync.when(
        data: (metrics) {
          if (metrics.isEmpty) {
            return const Center(child: Text("No applications discovered yet."));
          }

          return Padding(
            padding: const EdgeInsets.all(24.0),
            child: GridView.builder(
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 3, // Responsive count can be added later
                crossAxisSpacing: 24,
                mainAxisSpacing: 24,
                childAspectRatio: 1.2,
              ),
              itemCount: metrics.length,
              itemBuilder: (context, index) {
                return AppCard(metric: metrics[index]);
              },
            ),
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, stack) => Center(child: Text('Error: \$err')),
      ),
    );
  }
}
