import 'package:flutter/material.dart';
import 'screens/home_screen.dart';
import 'services/api_service.dart';
import 'models/view_type.dart';

void main() {
  // Initialize logging
  ApiService.initLogging();
  
  // Initialize ApiService singleton
  final backendUrl = const String.fromEnvironment('BACKEND', defaultValue: 'http://localhost:8080');
  ApiService.initialize(baseUrl: backendUrl);
  
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Toutizes Photos',
      theme: ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(
          seedColor: Colors.blue,
          brightness: Brightness.light,
        ),
      ),
      darkTheme: ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(
          seedColor: Colors.blue,
          brightness: Brightness.dark,
        ),
      ),
      home: const HomeScreen(
        initialView: ViewType.albums,
      ),
    );
  }
} 
