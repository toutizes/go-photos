import 'package:flutter/material.dart';

// Stub implementation for non-web platforms
Widget createNativeImageElement({
  required String imageUrl,
  Map<String, String>? headers,
  double? width,
  double? height,
  BoxFit fit = BoxFit.contain,
}) {
  // Return a regular Flutter Image widget for non-web platforms
  return Image.network(
    imageUrl,
    headers: headers,
    width: width,
    height: height,
    fit: fit,
    errorBuilder: (context, error, stackTrace) {
      return const Center(
        child: Icon(Icons.error_outline, size: 48),
      );
    },
  );
}

bool get isWebPlatformViewSupported => false;