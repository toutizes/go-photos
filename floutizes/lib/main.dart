import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'screens/home_screen.dart';
import 'services/api_service.dart';
import 'models/view_type.dart';

void main() {
  // Initialize logging
  ApiService.initLogging();

  // Initialize ApiService singleton
  final backendUrl = const String.fromEnvironment('BACKEND',
      defaultValue: 'http://localhost:8080');
  ApiService.initialize(baseUrl: backendUrl);

  runApp(const MyApp());
}

final _router = GoRouter(
  initialLocation: '/albums',
  routes: [
    GoRoute(
      path: '/albums',
      builder: (context, state) => const HomeScreen(
        initialView: ViewType.albums,
        initialSearchString: null,
      ),
      routes: [
        GoRoute(
          path: ':searchQuery',
          builder: (context, state) => HomeScreen(
            initialView: ViewType.albums,
            initialSearchString: state.pathParameters['searchQuery'],
          ),
        ),
      ],
    ),
    GoRoute(
      path: '/images',
      builder: (context, state) => const HomeScreen(
        initialView: ViewType.images,
        initialSearchString: null,
      ),
      routes: [
        GoRoute(
          path: ':searchQuery',
          builder: (context, state) => HomeScreen(
            initialView: ViewType.images,
            initialSearchString: state.pathParameters['searchQuery'],
          ),
        ),
      ],
    ),
  ],
);

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
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
      routerConfig: _router,
    );
  }
}
