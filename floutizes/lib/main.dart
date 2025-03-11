import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'screens/home_screen.dart';
import 'services/api_service.dart';
import 'models/view_type.dart';
import 'screens/image_detail_view.dart';

void main() {
  const backendUrl =
      String.fromEnvironment('BACKEND', defaultValue: 'http://localhost:8080');
  ApiService.initialize(baseUrl: backendUrl);

  runApp(const MyApp());
}

final _router = GoRouter(
  debugLogDiagnostics: true,
  initialLocation: '/albums',
  routes: [
    StatefulShellRoute.indexedStack(
      builder: (context, state, navigationShell) {
        return ScaffoldWithNestedNavigation(navigationShell: navigationShell);
      },
      branches: [
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/albums',
              builder: (context, state) {
                final q = state.uri.queryParameters['q'];
                return HomeScreen(
                  key: ValueKey(q),
                  initialView: ViewType.albums,
                  initialSearchString: state.uri.queryParameters['q'],
                );
              },
            ),
          ],
        ),
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/images',
              builder: (context, state) {
                final imageId = state.uri.queryParameters['imageId'];
                final q = state.uri.queryParameters['q'];
                return HomeScreen(
                  key: ValueKey("$imageId $q"),
                  initialView: ViewType.images,
                  initialSearchString: q,
                  initialImageId:
                      imageId != null ? int.tryParse(imageId) : null,
                );
              },
              routes: [
                GoRoute(
                  path: 'details/:index',
                  builder: (context, state) {
                    final index =
                        int.tryParse(state.pathParameters['index'] ?? '');
                    if (index == null) {
                      return const Center(child: Text('Invalid index'));
                    }
                    final q = state.uri.queryParameters['q'];
                    return ImageDetailView(
                      key: ValueKey(q),
                      searchQuery: state.uri.queryParameters['q'] ?? '',
                      currentIndex: index,
                      onKeywordSearch: (query, imageId) {
                        context.go(
                            '/images?q=${Uri.encodeComponent(query)}&imageId=$imageId');
                      },
                    );
                  },
                ),
              ],
            ),
          ],
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

class ScaffoldWithNestedNavigation extends StatelessWidget {
  const ScaffoldWithNestedNavigation({
    super.key,
    required this.navigationShell,
  });

  final StatefulNavigationShell navigationShell;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: navigationShell,
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: navigationShell.currentIndex,
        items: const [
          BottomNavigationBarItem(icon: Icon(Icons.album), label: 'Albums'),
          BottomNavigationBarItem(
              icon: Icon(Icons.photo_library), label: 'Photos'),
        ],
        onTap: (index) => navigationShell.goBranch(
          index,
          initialLocation: index == navigationShell.currentIndex,
        ),
      ),
    );
  }
}
