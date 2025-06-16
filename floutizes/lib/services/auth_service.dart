import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter/foundation.dart';

class AuthService extends ChangeNotifier {
  final FirebaseAuth _auth = FirebaseAuth.instance;
  User? _user;
  String? _idToken;
  bool _initialized = false;
  late final Future<void> initializationDone;

  AuthService() {
    // Set up the initialization future
    initializationDone = _initialize();

    // Set up continuous auth state monitoring
    _auth.authStateChanges().listen((User? user) async {
      _user = user;
      await _refreshToken();
      notifyListeners();
    });
  }

  Future<void> _initialize() async {
    // Wait for the first auth state change
    await _auth.authStateChanges().first;
    _initialized = true;
    _user = _auth.currentUser;
    notifyListeners();
  }

  /// Refresh the ID token from Firebase
  Future<void> _refreshToken() async {
    if (_user != null) {
      try {
        _idToken = await _user!.getIdToken(false); // Use cached token initially
      } catch (e) {
        debugPrint('Error refreshing token: $e');
        _idToken = null;
      }
    } else {
      _idToken = null;
    }
  }

  /// Manually refresh the token (for use by ApiService on 401 errors)
  Future<String?> refreshToken() async {
    if (_user != null) {
      try {
        _idToken = await _user!.getIdToken(true); // Force refresh on demand
        notifyListeners();
        return _idToken;
      } catch (e) {
        debugPrint('Error force refreshing token: $e');
        _idToken = null;
        notifyListeners();
        return null;
      }
    }
    return null;
  }

  User? get user => _user;
  String? get idToken => _idToken;
  bool get initialized => _initialized;
  bool get isAuthenticated => _user != null;

  Future<void> signOut() async {
    await _auth.signOut();
  }
}
