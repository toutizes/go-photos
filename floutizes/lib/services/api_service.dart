import 'dart:convert';
import 'package:logging/logging.dart';
import 'package:http/http.dart' as http;
import '../models/image.dart';
import '../models/directory.dart';

class ApiService {
  static final _log = Logger('ApiService');
  final String baseUrl;
  final http.Client _client;

  static void initLogging() {
    Logger.root.level = Level.ALL;
    Logger.root.onRecord.listen((record) {
      print('${record.level.name}: ${record.time}: ${record.message}');
    });
  }

  ApiService({required this.baseUrl, http.Client? client})
      : _client = client ?? http.Client();

  void _logRequest(String method, String url, Map<String, String> headers) {
    _log.info('$method $url');
    _log.fine('Headers: $headers');
  }

  void _logResponse(String method, String url, http.Response response,
      [String? error]) {
    if (error != null) {
      _log.warning('$method $url returned ${response.statusCode}: $error');
      _log.warning('Response headers: ${response.headers}');
      _log.warning('Response body: ${response.body}');
    } else {
      _log.fine('$method $url returned ${response.statusCode}');
      _log.fine('Response headers: ${response.headers}');
    }
  }

  Future<List<ImageModel>> searchImages(String query) async {
    final url = '$baseUrl/db/q?q=${Uri.encodeComponent(query)}';
    final headers = {
      'Accept': 'application/json',
    };
    _logRequest('GET', url, headers);

    try {
      final response = await _client.get(
        Uri.parse(url),
        headers: headers,
      );
      _logResponse('GET', url, response);

      if (response.statusCode == 200) {
        final List<dynamic> jsonList = json.decode(response.body);
        return jsonList.map((json) => ImageModel.fromJson(json)).toList();
      } else {
        final error = 'Failed to search images: ${response.statusCode}';
        _logResponse('GET', url, response, error);
        throw Exception(error);
      }
    } catch (e) {
      _log.severe('Error searching images: $e');
      rethrow;
    }
  }

  Future<List<DirectoryModel>> getAlbums() async {
    final url = '$baseUrl/db/q?q=albums:&kind=album';
    final headers = {
      'Accept': 'application/json',
    };
    _logRequest('GET', url, headers);

    try {
      final response = await _client.get(
        Uri.parse(url),
        headers: headers,
      );

      _logResponse('GET', url, response);

      if (response.statusCode == 200) {
        final List<dynamic> jsonList = json.decode(response.body);
        return jsonList.map((json) => DirectoryModel.fromJson(json)).toList();
      } else {
        final error = 'Failed to fetch albums: ${response.statusCode}';
        _logResponse('GET', url, response, error);
        throw Exception(error);
      }
    } catch (e) {
      _log.severe('Error fetching albums: $e');
      rethrow;
    }
  }

  Future<List<ImageModel>> getAlbumImages(String albumPath) async {
    final url = '$baseUrl/db/q?q=album:${Uri.encodeComponent(albumPath)}';
    final headers = {
      'Accept': 'application/json',
      'Origin': baseUrl,
    };
    _logRequest('GET', url, headers);

    try {
      final response = await _client.get(
        Uri.parse(url),
        headers: headers,
      );

      _logResponse('GET', url, response);

      if (response.statusCode == 200) {
        final List<dynamic> jsonList = json.decode(response.body);
        return jsonList.map((json) => ImageModel.fromJson(json)).toList();
      } else {
        final error = 'Failed to fetch album images: ${response.statusCode}';
        _logResponse('GET', url, response, error);
        throw Exception(error);
      }
    } catch (e) {
      _log.severe('Error fetching album images: $e');
      rethrow;
    }
  }

  Future<void> downloadAlbum(String albumPath,
      {bool highQuality = false}) async {
    final url =
        '$baseUrl/db/viewer?command=download&q=${Uri.encodeComponent(albumPath)}&s=${highQuality ? "O" : "M"}';
    final headers = {
      'Accept': 'application/octet-stream',
      'Origin': baseUrl,
    };
    _logRequest('GET', url, headers);

    try {
      final response = await _client.get(
        Uri.parse(url),
        headers: headers,
      );

      _logResponse('GET', url, response);

      if (response.statusCode != 200) {
        final error = 'Failed to download album: ${response.statusCode}';
        _logResponse('GET', url, response, error);
        throw Exception(error);
      }

      // TODO: Handle the downloaded ZIP file
      // For web: Create a download link
      // For mobile: Save to device storage
    } catch (e) {
      _log.severe('Error downloading album: $e');
      rethrow;
    }
  }

  String getImageUrl(String relativePath) {
    final url = '$baseUrl$relativePath';
    _log.finer('Image URL: $url');
    return url;
  }
}
