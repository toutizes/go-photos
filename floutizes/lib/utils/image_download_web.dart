import 'dart:async';
import 'dart:html' as html;

class ImageDownloadPlatform {
  static Future<void> downloadSingleImage({
    required String url,
    required String filename,
  }) async {
    final xhr = html.HttpRequest();
    xhr.open('GET', url);
    xhr.responseType = 'blob';

    final completer = Completer<void>();

    xhr.onLoad.listen((event) {
      final blob = xhr.response as html.Blob;
      final url = html.Url.createObjectUrlFromBlob(blob);
      
      final anchor = html.AnchorElement(href: url)
        ..download = filename
        ..style.display = 'none';
      html.document.body!.children.add(anchor);
      anchor.click();
      
      // Cleanup
      html.document.body!.children.remove(anchor);
      html.Url.revokeObjectUrl(url);
      
      completer.complete();
    });

    xhr.onError.listen((event) {
      completer.completeError('Failed to download image');
    });

    xhr.send();
    return completer.future;
  }

  static Future<void> downloadZipFile({
    required String url,
    required void Function(int loaded) onProgress,
  }) async {
    final xhr = html.HttpRequest();
    xhr.open('GET', url);
    xhr.responseType = 'blob';

    final completer = Completer<void>();
    
    xhr.onLoad.listen((event) {
      final blob = xhr.response as html.Blob;
      final url = html.Url.createObjectUrlFromBlob(blob);
      
      // Create and trigger download
      final anchor = html.AnchorElement(href: url)
        ..download = 'images.zip'
        ..style.display = 'none';
      html.document.body!.children.add(anchor);
      anchor.click();

      // Cleanup
      html.document.body!.children.remove(anchor);
      html.Url.revokeObjectUrl(url);
      
      completer.complete();
    });

    xhr.onProgress.listen((event) {
      onProgress(event.loaded ?? 0);
    });

    xhr.onError.listen((event) {
      completer.completeError('XHR download failed');
    });

    xhr.send();
    return completer.future;
  }
} 