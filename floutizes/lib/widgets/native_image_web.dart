import 'dart:html' as html;
import 'dart:ui_web' as ui_web;
import 'dart:typed_data';
import 'package:flutter/material.dart';

// Web-specific implementation
Widget createNativeImageElement({
  required String imageUrl,
  Map<String, String>? headers,
  double? width,
  double? height,
  BoxFit fit = BoxFit.contain,
}) {
  final viewId = 'native-image-${DateTime.now().millisecondsSinceEpoch}';
  
  // Register the platform view factory
  ui_web.platformViewRegistry.registerViewFactory(viewId, (int viewId) {
    final img = html.ImageElement()
      ..style.width = '100%'
      ..style.height = '100%'
      ..style.objectFit = _getObjectFit(fit)
      ..style.display = 'block'
      ..style.margin = '0'
      ..style.padding = '0';

    // If headers are provided, fetch with auth and create blob URL
    if (headers != null && headers.isNotEmpty) {
      _fetchImageWithAuth(imageUrl, headers).then((blobUrl) {
        if (blobUrl != null) {
          img.src = blobUrl;
        } else {
          // Fallback to direct URL if fetch fails
          img.src = imageUrl;
        }
      }).catchError((error) {
        // Fallback to direct URL on error
        img.src = imageUrl;
      });
    } else {
      // No auth needed, use direct URL
      img.src = imageUrl;
    }
    
    return img;
  });

  return SizedBox(
    width: width,
    height: height,
    child: HtmlElementView(viewType: viewId),
  );
}

// Fetch image with authentication headers and create blob URL
Future<String?> _fetchImageWithAuth(String imageUrl, Map<String, String> headers) async {
  try {
    final response = await html.HttpRequest.request(
      imageUrl,
      method: 'GET',
      requestHeaders: headers,
      responseType: 'arraybuffer',
    );
    
    if (response.status == 200) {
      final Uint8List bytes = Uint8List.view(response.response);
      final blob = html.Blob([bytes]);
      return html.Url.createObjectUrl(blob);
    }
  } catch (e) {
    print('Error fetching image with auth: $e');
  }
  return null;
}

String _getObjectFit(BoxFit fit) {
  switch (fit) {
    case BoxFit.contain:
      return 'contain';
    case BoxFit.cover:
      return 'cover';
    case BoxFit.fill:
      return 'fill';
    case BoxFit.fitHeight:
      return 'scale-down';
    case BoxFit.fitWidth:
      return 'scale-down';
    case BoxFit.scaleDown:
      return 'scale-down';
    case BoxFit.none:
      return 'none';
  }
}

bool get isWebPlatformViewSupported => true;