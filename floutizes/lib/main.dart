import 'package:flutter/material.dart';
import 'screens/home_screen.dart';
import 'services/api_service.dart';

void main() {
  // Initialize logging
  ApiService.initLogging();
  
  runApp(const MyApp());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    // Get the backend URL from environment variable or use default
    final backendUrl = const String.fromEnvironment('BACKEND', defaultValue: 'http://localhost:8080');
    final apiService = ApiService(baseUrl: backendUrl);

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
      home: HomeScreen(apiService: apiService),
    );
  }
} 
