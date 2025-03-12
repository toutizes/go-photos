import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'screens/home_screen.dart';
import 'services/api_service.dart';
import 'models/view_type.dart';
import 'screens/image_detail_view.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:provider/provider.dart';
import 'firebase_options.dart';
import 'services/auth_service.dart';
import 'screens/login_screen.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Firebase.initializeApp(
    options: DefaultFirebaseOptions.currentPlatform,
  );
  const backendUrl =
      String.fromEnvironment('BACKEND', defaultValue: 'http://localhost:8080');
  ApiService.initialize(baseUrl: backendUrl);
  // await ApiService.instance.signOut();
  runApp(const MyApp());
}

final _router = GoRouter(
  // debugLogDiagnostics: true,
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
                  path: 'details/:imageId',
                  builder: (context, state) {
                    final imageId = int.tryParse(state.pathParameters['imageId'] ?? '');
                    if (imageId == null) {
                      return const Center(child: Text('Invalid image ID'));
                    }
                    final q = state.uri.queryParameters['q'];
                    return ImageDetailView(
                      key: ValueKey(q),
                      searchQuery: state.uri.queryParameters['q'] ?? '',
                      imageId: imageId,
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
    return ChangeNotifierProvider(
      create: (_) => AuthService(),
      child: MaterialApp(
        title: 'Floutizes',
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
        home: Consumer<AuthService>(
          builder: (context, auth, _) {
            // Show splash screen while Firebase initializes
            if (!auth.initialized) {
              return const Scaffold(
                body: Center(
                  child: CircularProgressIndicator(),
                ),
              );
            }

            // Show login screen if not authenticated
            if (!auth.isAuthenticated) {
              return const LoginScreen();
            }

            // Show main app if authenticated
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
          },
        ),
      ),
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
