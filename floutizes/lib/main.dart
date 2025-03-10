import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'screens/home_screen.dart';
import 'services/api_service.dart';
import 'models/view_type.dart';
import 'screens/image_detail_view.dart';

void main() {
  // Initialize logging
  ApiService.initLogging();

  // Initialize ApiService singleton
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
          ],
        ),
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/images',
              builder: (context, state) => const HomeScreen(
                initialView: ViewType.images,
                initialSearchString: null,
              ),
              routes: [
                GoRoute(
                  path: ':searchQuery',
                  builder: (context, state) {
                    final searchQuery =
                        state.pathParameters['searchQuery'] ?? '';
                    final imageId = state.uri.queryParameters['imageId'];
                    return HomeScreen(
                      initialView: ViewType.images,
                      initialSearchString: searchQuery,
                      initialImageId:
                          imageId != null ? int.tryParse(imageId) : null,
                    );
                  },
                  routes: [
                    GoRoute(
                      path: 'details/:index',
                      builder: (context, state) {
                        final queryString =
                            state.pathParameters['searchQuery'] ?? '';
                        final index =
                            int.tryParse(state.pathParameters['index'] ?? '');
                        if (index == null) {
                          return const Center(child: Text('Invalid index'));
                        }
                        return ImageDetailView(
                          searchQuery: queryString,
                          currentIndex: index,
                          onKeywordSearch: (query, imageId) {
                            context.go('/images/$query?imageId=$imageId');
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
