import 'package:flutter/material.dart';
import '../services/api_service.dart';
import 'dart:html' as html;
import 'dart:async';
import 'package:http/http.dart' as http;
import 'dart:typed_data';
import 'dart:js' as js;

enum DownloadQuality {
  medium,
  high,
}

class ImageDownload {
  static Future<void> start(
    BuildContext context,
    String queryString,
    DownloadQuality quality,
  ) async {
    return showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => _DownloadDialog(
        queryString: queryString,
        quality: quality,
      ),
    );
  }
}

class _DownloadDialog extends StatefulWidget {
  final String queryString;
  final DownloadQuality quality;

  const _DownloadDialog({
    required this.queryString,
    required this.quality,
  });

  @override
  State<_DownloadDialog> createState() => _DownloadDialogState();
}

class _DownloadDialogState extends State<_DownloadDialog> {
  bool _isDownloading = true;
  String? _error;
  double _progress = 0.0;
  bool _downloadComplete = false;
  int _downloaded = 0;

  @override
  void initState() {
    super.initState();
    _startDownload();
  }

  Future<void> _startDownload() async {
    final downloadUrl = ApiService.instance.getDownloadUrl(
      widget.queryString,
      highQuality: widget.quality == DownloadQuality.high,
    );

    try {
      final client = http.Client();
      try {
        final xhr = html.HttpRequest();
        xhr.open('GET', downloadUrl);
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
          if (mounted) {
            setState(() {
              _downloaded = event.loaded ?? 0;
              _progress = _downloaded / (1024 * 1024); // Convert to MB
            });
          }
        });

        xhr.onError.listen((event) {
          completer.completeError('XHR download failed');
        });

        // Start the download
        xhr.send();

        // Wait for download to complete
        await completer.future;

        if (!mounted) return;
        setState(() {
          _isDownloading = false;
          _downloadComplete = true;
        });

        // Show success state briefly before closing
        await Future.delayed(const Duration(milliseconds: 800));
        if (!mounted) return;
        Navigator.of(context).pop();
      } finally {
        client.close();
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isDownloading = false;
        _error = e.toString();
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (_isDownloading) ...[
            const CircularProgressIndicator(),
            const SizedBox(height: 16),
            Text(
              _isDownloading
                  ? 'Téléchargement en cours...\n${_progress.toStringAsFixed(1)} MB\n${widget.quality == DownloadQuality.high ? 'Haute' : 'Moyenne'} qualité'
                  : 'Sauvegarde du fichier...',
              textAlign: TextAlign.center,
            ),
          ] else if (_error != null) ...[
            const Icon(
              Icons.error_outline,
              color: Colors.red,
              size: 48,
            ),
            const SizedBox(height: 16),
            Text(
              'Erreur de téléchargement:\n$_error',
              textAlign: TextAlign.center,
              style: const TextStyle(color: Colors.red),
            ),
            const SizedBox(height: 16),
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: const Text('Fermer'),
            ),
          ] else if (_downloadComplete) ...[
            const Icon(
              Icons.check_circle_outline,
              color: Colors.green,
              size: 48,
            ),
            const SizedBox(height: 16),
            const Text(
              'Téléchargement terminé',
              textAlign: TextAlign.center,
              style: TextStyle(color: Colors.green),
            ),
          ],
        ],
      ),
    );
  }
}
