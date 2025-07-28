import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

// Conditional imports for web-only functionality
import 'native_image_stub.dart'
    if (dart.library.html) 'native_image_web.dart' as native_impl;

class NativeImageView extends StatefulWidget {
  final String imageUrl;
  final Map<String, String>? headers;
  final double? width;
  final double? height;
  final BoxFit fit;
  final Widget? fallback;

  const NativeImageView({
    super.key,
    required this.imageUrl,
    this.headers,
    this.width,
    this.height,
    this.fit = BoxFit.contain,
    this.fallback,
  });

  @override
  State<NativeImageView> createState() => _NativeImageViewState();
}

class _NativeImageViewState extends State<NativeImageView> {
  @override
  Widget build(BuildContext context) {
    // Use platform-specific implementation
    if (kIsWeb && native_impl.isWebPlatformViewSupported) {
      return native_impl.createNativeImageElement(
        imageUrl: widget.imageUrl,
        headers: widget.headers,
        width: widget.width,
        height: widget.height,
        fit: widget.fit,
      );
    }

    // Fallback for non-web platforms or when native implementation is not available
    return widget.fallback ?? 
        native_impl.createNativeImageElement(
          imageUrl: widget.imageUrl,
          headers: widget.headers,
          width: widget.width,
          height: widget.height,
          fit: widget.fit,
        );
  }
}