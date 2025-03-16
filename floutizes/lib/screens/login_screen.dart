import 'package:flutter/material.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:go_router/go_router.dart';

class LoginScreen extends StatefulWidget {
  final String? from;

  const LoginScreen({
    super.key,
    this.from,
  });

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  bool _isLoading = false;

  Future<void> _signInWithGoogle() async {
    setState(() {
      _isLoading = true;
    });

    try {
      // Initialize Google Sign In with web client id
      final GoogleSignIn googleSignIn = kIsWeb
          ? GoogleSignIn(
              clientId:
                  '347191427262-hbqb3r32te4vct21qujvgjit0m16t4gb.apps.googleusercontent.com',
            )
          : GoogleSignIn();
      final GoogleSignInAccount? googleUser = await googleSignIn.signIn();

      if (googleUser == null) {
        // User canceled the sign-in flow
        return;
      }

      // Obtain auth details from request
      final GoogleSignInAuthentication googleAuth =
          await googleUser.authentication;

      // Create a new credential
      final credential = GoogleAuthProvider.credential(
        accessToken: googleAuth.accessToken,
        idToken: googleAuth.idToken,
      );

      // Sign in to Firebase with the Google credential
      await FirebaseAuth.instance.signInWithCredential(credential);
      
      // Navigate to the original page or albums page after successful sign in
      if (mounted) {
        if (widget.from != null) {
          final decodedPath = Uri.decodeComponent(widget.from!);
          final uri = Uri.parse(decodedPath);
          context.go('${uri.path}${uri.query.isEmpty ? '' : '?${uri.query}'}');
        } else {
          context.go('/albums');
        }
      }
    } catch (e) {
      // Show error message
      print('Erreur de connexion: ${e.toString()}');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Erreur de connexion: ${e.toString()}'),
            backgroundColor: Colors.red,
          ),
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // App logo
            Image.asset(
              'assets/logo.png',
              width: 120,
              height: 120,
            ),
            const SizedBox(height: 24),
            // App title
            const Text(
              'Photos Toutizes',
              style: TextStyle(
                fontSize: 32,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 16),
            const Text(
              'Il faut maintenant s\'authentifier pour voir les photos',
              style: TextStyle(
                fontSize: 18,
              ),
            ),
            const SizedBox(height: 48),
            // Sign in button
            if (_isLoading)
              const CircularProgressIndicator()
            else
              ElevatedButton.icon(
                onPressed: _signInWithGoogle,
                icon: SvgPicture.asset(
                  'assets/google_logo.svg',
                  height: 24,
                ),
                label: const Text('Se connecter avec Google'),
                style: ElevatedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 12,
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}
