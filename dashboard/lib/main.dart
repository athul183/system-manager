import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:sys_dashboard/core/theme.dart';
import 'package:sys_dashboard/features/monitor/dashboard_screen.dart';

void main() {
  runApp(const ProviderScope(child: SysDashboardApp()));
}

class SysDashboardApp extends StatelessWidget {
  const SysDashboardApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'System Manager',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.darkTheme,
      home: const DashboardScreen(),
    );
  }
}
