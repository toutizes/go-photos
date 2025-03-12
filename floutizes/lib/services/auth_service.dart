import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter/foundation.dart';

class AuthService extends ChangeNotifier {
  final FirebaseAuth _auth = FirebaseAuth.instance;
  User? _user;
  bool _initialized = false;

  AuthService() {
    _auth.authStateChanges().listen((User? user) {
      _user = user;
      _initialized = true;
      notifyListeners();
    });
  }

  User? get user => _user;
  bool get initialized => _initialized;
  bool get isAuthenticated => _user != null;

  Future<void> signOut() async {
    await _auth.signOut();
  }

  String? get idToken => _user?.getIdToken() as String?;
} 