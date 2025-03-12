class ImageDownloadPlatform {
  static Future<void> downloadSingleImage({
    required String url,
    required String filename,
  }) async {
    throw UnsupportedError(
      'Télechargements non pris en charge sur cette plateforme. Veuillez utiliser un navigateur web.',
    );
  }

  static Future<void> downloadZipFile({
    required String url,
    required void Function(int loaded) onProgress,
  }) async {
    throw UnsupportedError(
      'Télechargements non pris en charge sur cette plateforme. Veuillez utiliser un navigateur web.',
    );
  }
}
