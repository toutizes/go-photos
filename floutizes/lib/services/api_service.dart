import 'dart:convert';
import 'package:logging/logging.dart';
import 'package:http/http.dart' as http;
import '../models/image.dart';
import '../models/directory.dart';

class ApiService {
  static ApiService? _instance;
  static final _logger = Logger('ApiService');
  final String baseUrl;
  final http.Client _client;

  // Cache for search results
  final int _maxCacheSize = 3;
  final Map<String, List<ImageModel>> _searchCache = {};
  final List<String> _cacheOrder = [];

  ApiService._({required this.baseUrl}) : _client = http.Client();

  static void _initLogging() {
    Logger.root.level = Level.INFO;
    Logger.root.onRecord.listen((record) {
      print('${record.level.name}: ${record.time}: ${record.message}');
    });
  }

  static void initialize({required String baseUrl}) {
    _initLogging();
    _instance = ApiService._(baseUrl: baseUrl);
  }

  static ApiService get instance {
    if (_instance == null) {
      throw StateError(
          'ApiService has not been initialized. Call initialize() first.');
    }
    return _instance!;
  }

  void _addToCache(String query, List<ImageModel> results) {
    // Remove oldest entry if cache is full
    if (_cacheOrder.length >= _maxCacheSize &&
        !_searchCache.containsKey(query)) {
      final oldestQuery = _cacheOrder.removeAt(0);
      _searchCache.remove(oldestQuery);
    }

    // Add new results to cache
    if (!_searchCache.containsKey(query)) {
      _cacheOrder.add(query);
    }
    _searchCache[query] = results;
    _logger.fine(
        'Cache updated for query: $query (cache size: ${_searchCache.length})');
  }

  void _logRequest(String method, String url, Map<String, String> headers) {
    _logger.info('$method $url');
    _logger.fine('Headers: $headers');
  }

  void _logResponse(String method, String url, http.Response response,
      [String? error]) {
    if (error != null) {
      _logger.warning('$method $url returned ${response.statusCode}: $error');
      _logger.warning('Response headers: ${response.headers}');
      _logger.warning('Response body: ${response.body}');
    } else {
      _logger.fine('$method $url returned ${response.statusCode}');
      _logger.fine('Response headers: ${response.headers}');
    }
  }

  Future<List<ImageModel>> searchImages(String query) async {
    // Check cache first
    if (_searchCache.containsKey(query)) {
      _logger.fine('Cache hit for query: $query');
      return _searchCache[query]!;
    }

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
        final results =
            jsonList.map((json) => ImageModel.fromJson(json)).toList();

        // Sort results by timestamp and name before caching
        results.sort((a, b) {
          final timeCompare = b.itemTimestamp.compareTo(a.itemTimestamp);
          if (timeCompare != 0) return timeCompare;
          return b.imageName.compareTo(a.imageName);
        });

        // Add results to cache
        _addToCache(query, results);

        return results;
      } else {
        final error = 'Failed to search images: ${response.statusCode}';
        _logResponse('GET', url, response, error);
        throw Exception(error);
      }
    } catch (e) {
      _logger.severe('Error searching images: $e');
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
      _logger.severe('Error fetching albums: $e');
      rethrow;
    }
  }

  Future<void> downloadAlbum(
    String albumPath, {
    bool highQuality = false,
  }) async {
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
      _logger.severe('Error downloading album: $e');
      rethrow;
    }
  }

  Future<List<DirectoryModel>> searchAlbums(String query) async {
    final url = '$baseUrl/db/q?q=in:${Uri.encodeComponent(query)}&kind=album';
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
        final error = 'Failed to search albums: ${response.statusCode}';
        _logResponse('GET', url, response, error);
        throw Exception(error);
      }
    } catch (e) {
      _logger.severe('Error searching albums: $e');
      rethrow;
    }
  }

  String getImageUrl(String relativePath) {
    final url = '$baseUrl$relativePath';
    _logger.finer('Image URL: $url');
    return url;
  }
}
