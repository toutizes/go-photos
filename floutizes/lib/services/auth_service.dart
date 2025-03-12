import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter/foundation.dart';

class AuthService extends ChangeNotifier {
  final FirebaseAuth _auth = FirebaseAuth.instance;
  User? _user;
  bool _initialized = false;
  late final Future<void> initializationDone;

  AuthService() {
    // Set up the initialization future
    initializationDone = _initialize();

    // Set up continuous auth state monitoring
    _auth.authStateChanges().listen((User? user) {
      print("TT AuthService. user changed: $_user");
      _user = user;
      notifyListeners();
    });
  }

  Future<void> _initialize() async {
    // Wait for the first auth state change
    await _auth.authStateChanges().first;
    _initialized = true;
    _user = _auth.currentUser;
    print("TT AuthService._initialize: $_user");
    notifyListeners();
  }

  User? get user => _user;
  bool get initialized => _initialized;
  bool get isAuthenticated => _user != null;

  Future<void> signOut() async {
    await _auth.signOut();
  }

  Future<String?> get idToken async => await _user?.getIdToken();
}
