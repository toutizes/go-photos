import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:material_symbols_icons/symbols.dart';
import 'screens/home_screen.dart';
import 'services/api_service.dart';
import 'models/view_type.dart';
import 'screens/image_detail_view.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:provider/provider.dart';
import 'firebase_options.dart';
import 'services/auth_service.dart';
import 'screens/login_screen.dart';

late final AuthService authService;

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize Firebase
  await Firebase.initializeApp(
    options: DefaultFirebaseOptions.currentPlatform,
  );

  // Initialize and wait for AuthService
  authService = AuthService();
  await authService.initializationDone;

  const backendUrl =
      String.fromEnvironment('BACKEND', defaultValue: 'http://localhost:8080');
  ApiService.initialize(authService: authService, baseUrl: backendUrl);
  // await ApiService.instance.signOut();
  runApp(const MyApp());
}

final _router = GoRouter(
  // debugLogDiagnostics: true,
  initialLocation: '/activity',
  refreshListenable: authService,
  redirect: (context, state) {
    final auth = Provider.of<AuthService>(context, listen: false);
    final isLoginRoute = state.matchedLocation == '/login';

    // If not authenticated and not on login page, redirect to login
    if (!auth.isAuthenticated && !isLoginRoute) {
      // Store the attempted location in the query parameters
      return '/login?from=${Uri.encodeComponent(state.uri.toString())}';
    }

    // If authenticated and on login page, redirect to stored location or default
    if (auth.isAuthenticated && isLoginRoute) {
      final from = state.uri.queryParameters['from'];
      if (from != null) {
        final decodedPath = Uri.decodeComponent(from);
        // Extract the path and query from the full URL
        final uri = Uri.parse(decodedPath);
        return '${uri.path}${uri.query.isEmpty ? '' : '?${uri.query}'}';
      }
      return '/images';
    }

    return null;
  },
  routes: [
    GoRoute(
      path: '/login',
      builder: (context, state) => LoginScreen(
        from: state.uri.queryParameters['from'],
      ),
    ),
    StatefulShellRoute.indexedStack(
      builder: (context, state, navigationShell) {
        return ScaffoldWithNestedNavigation(navigationShell: navigationShell);
      },
      branches: [
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/activity',
              builder: (context, state) {
                return const HomeScreen(
                  key: ValueKey('activity'),
                  initialView: ViewType.activity,
                );
              },
            ),
          ],
        ),
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
                    final imageId =
                        int.tryParse(state.pathParameters['imageId'] ?? '');
                    if (imageId == null) {
                      return const Center(child: Text('Invalid image ID'));
                    }
                    final q = state.uri.queryParameters['q'];
                    return ImageDetailView(
                      key: ValueKey(q),
                      searchQuery: state.uri.queryParameters['q'] ?? '',
                      imageId: imageId,
                      onKeywordSearch: (query, imageId) {
                        // Quote keywords containing spaces
                        var query2 = query.contains(' ') ? '"$query"' : query;
                        context.go(
                            '/images?q=${Uri.encodeComponent(query2)}&imageId=$imageId');
                      },
                      onSearch: (query, imageId) {
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
        // Admin branch - conditionally added based on user email
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/admin',
              builder: (context, state) {
                return const HomeScreen(
                  key: ValueKey('admin'),
                  initialView: ViewType.adminQueries,
                );
              },
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
    return ChangeNotifierProvider.value(
      value: authService,
      child: MaterialApp.router(
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
      ),
    );
  }
}

class ImmersiveModeNotifier extends ValueNotifier<bool> {
  ImmersiveModeNotifier() : super(false);
}

class ImmersiveModeScope extends InheritedNotifier<ImmersiveModeNotifier> {
  const ImmersiveModeScope({
    super.key,
    required ImmersiveModeNotifier notifier,
    required super.child,
  }) : super(notifier: notifier);

  static ImmersiveModeNotifier of(BuildContext context) {
    return context
        .dependOnInheritedWidgetOfExactType<ImmersiveModeScope>()!
        .notifier!;
  }
}

class ScaffoldWithNestedNavigation extends StatefulWidget {
  const ScaffoldWithNestedNavigation({
    super.key,
    required this.navigationShell,
  });

  final StatefulNavigationShell navigationShell;

  @override
  State<ScaffoldWithNestedNavigation> createState() =>
      _ScaffoldWithNestedNavigationState();
}

class _ScaffoldWithNestedNavigationState
    extends State<ScaffoldWithNestedNavigation> {
  final _immersiveModeNotifier = ImmersiveModeNotifier();

  @override
  void dispose() {
    _immersiveModeNotifier.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return ImmersiveModeScope(
      notifier: _immersiveModeNotifier,
      child: ValueListenableBuilder<bool>(
        valueListenable: _immersiveModeNotifier,
        builder: (context, isImmersive, child) {
          return Consumer<AuthService>(
            builder: (context, authService, child) {
              // Check if current user is admin
              final isAdmin = authService.user?.email == 'matthieu.devin@gmail.com';
              
              // Build navigation items conditionally
              final navigationItems = [
                BottomNavigationBarItem(
                  icon: Icon(Symbols.trending_up, color: Theme.of(context).colorScheme.onSurfaceVariant),
                  activeIcon: Icon(Symbols.trending_up, color: Theme.of(context).colorScheme.primary),
                  label: 'Activité',
                ),
                BottomNavigationBarItem(
                  icon: Icon(Symbols.photo_album, color: Theme.of(context).colorScheme.onSurfaceVariant),
                  activeIcon: Icon(Symbols.photo_album, color: Theme.of(context).colorScheme.primary),
                  label: 'Albums',
                ),
                BottomNavigationBarItem(
                  icon: Icon(Symbols.photo_library, color: Theme.of(context).colorScheme.onSurfaceVariant),
                  activeIcon: Icon(Symbols.photo_library, color: Theme.of(context).colorScheme.primary),
                  label: 'Photos',
                ),
                if (isAdmin)
                  BottomNavigationBarItem(
                    icon: Icon(Symbols.admin_panel_settings, color: Theme.of(context).colorScheme.onSurfaceVariant),
                    activeIcon: Icon(Symbols.admin_panel_settings, color: Theme.of(context).colorScheme.primary),
                    label: 'Admin',
                  ),
              ];

              return Scaffold(
                body: widget.navigationShell,
                bottomNavigationBar: isImmersive
                    ? null
                    : BottomNavigationBar(
                        type: BottomNavigationBarType.fixed,
                        currentIndex: widget.navigationShell.currentIndex,
                        selectedItemColor: Theme.of(context).colorScheme.primary,
                        unselectedItemColor: Theme.of(context).colorScheme.onSurfaceVariant,
                        backgroundColor: Theme.of(context).colorScheme.surface,
                        items: navigationItems,
                        onTap: (index) => widget.navigationShell.goBranch(
                          index,
                          initialLocation:
                              index == widget.navigationShell.currentIndex,
                        ),
                      ),
              );
            },
          );
        },
      ),
    );
  }
}
