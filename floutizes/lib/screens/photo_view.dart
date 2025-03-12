import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../services/api_service.dart';

class PhotoView extends StatefulWidget {
  final String imageUrl;
  final int width;
  final int height;

  const PhotoView({
    super.key,
    required this.imageUrl,
    required this.width,
    required this.height,
  });

  @override
  State<PhotoView> createState() => _PhotoViewState();
}

class _PhotoViewState extends State<PhotoView> {
  final FocusNode _focusNode = FocusNode();

  @override
  void initState() {
    super.initState();
    // Request focus when the view is created for keyboard handling
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _focusNode.requestFocus();
    });
  }

  @override
  void dispose() {
    _focusNode.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      body: KeyboardListener(
        focusNode: _focusNode,
        onKeyEvent: (event) {
          if (event is KeyDownEvent &&
              event.logicalKey == LogicalKeyboardKey.escape) {
            Navigator.of(context).pop();
          }
        },
        child: Stack(
          children: [
            // Main image with zoom and pan
            GestureDetector(
              onTap: () => Navigator.of(context).pop(),
              child: InteractiveViewer(
                maxScale: 10.0, // Allow more zoom than in detail view
                child: Center(
                  child: FutureBuilder<Map<String, String>>(
                    future: ApiService.instance.getImageHeaders(),
                    builder: (context, snapshot) {
                      if (!snapshot.hasData)
                        return const CircularProgressIndicator();
                      return Image.network(
                        widget.imageUrl,
                        headers: snapshot.data,
                        fit: BoxFit.contain,
                        frameBuilder:
                            (context, child, frame, wasSynchronouslyLoaded) {
                          if (wasSynchronouslyLoaded) {
                            return child;
                          }
                          final isDark =
                              Theme.of(context).brightness == Brightness.dark;
                          return AnimatedSwitcher(
                            duration: const Duration(milliseconds: 200),
                            child: frame != null
                                ? child
                                : LayoutBuilder(
                                    builder: (context, constraints) {
                                      final aspectRatio =
                                          widget.width / widget.height;
                                      double width, height;

                                      if (constraints.maxWidth /
                                              constraints.maxHeight >
                                          aspectRatio) {
                                        // Height constrained
                                        height = constraints.maxHeight;
                                        width = height * aspectRatio;
                                      } else {
                                        // Width constrained
                                        width = constraints.maxWidth;
                                        height = width / aspectRatio;
                                      }
                                      return Center(
                                        child: Container(
                                          width: width,
                                          height: height,
                                          color: isDark
                                              ? Colors.grey.shade800
                                              : Colors.grey.shade200,
                                        ),
                                      );
                                    },
                                  ),
                          );
                        },
                        errorBuilder: (context, error, stackTrace) {
                          return const Center(
                            child: Icon(Icons.error_outline,
                                size: 48, color: Colors.white),
                          );
                        },
                      );
                    },
                  ),
                ),
              ),
            ),
            // Close button in top-right corner
            Positioned(
              top: 0,
              right: 0,
              child: SafeArea(
                child: IconButton(
                  icon: const Icon(Icons.close, color: Colors.white),
                  onPressed: () => Navigator.of(context).pop(),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
