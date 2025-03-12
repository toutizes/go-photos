import 'package:flutter/material.dart';
import '../services/api_service.dart';
import 'image_download_stub.dart'
    if (dart.library.html) 'image_download_web.dart';

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

  /// Downloads a single image directly to disk
  static Future<void> downloadSingleImage({
    required String url,
    required String filename,
  }) {
    return ImageDownloadPlatform.downloadSingleImage(
      url: url,
      filename: filename,
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
      await ImageDownloadPlatform.downloadZipFile(
        url: downloadUrl,
        onProgress: (loaded) {
          if (mounted) {
            setState(() {
              _downloaded = loaded;
              _progress = _downloaded / (1024 * 1024); // Convert to MB
            });
          }
        },
      );

      if (!mounted) return;
      setState(() {
        _isDownloading = false;
        _downloadComplete = true;
      });

      // Show success state briefly before closing
      await Future.delayed(const Duration(milliseconds: 800));
      if (!mounted) return;
      Navigator.of(context).pop();
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
