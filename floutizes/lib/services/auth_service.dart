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
      _idToken = await user?.getIdToken(false);
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

  User? get user => _user;
  String? get idToken => _idToken;
  bool get initialized => _initialized;
  bool get isAuthenticated => _user != null;

  Future<void> signOut() async {
    await _auth.signOut();
  }
}
