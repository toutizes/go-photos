import 'dart:convert';
import 'package:logging/logging.dart';
import 'package:http/http.dart' as http;
import '../models/image.dart';
import '../models/directory.dart';
import 'package:firebase_auth/firebase_auth.dart';

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

  List<ImageModel> _toImages(String body) {
    final List<dynamic> jsonList = json.decode(body);
    final results = jsonList.map((json) => ImageModel.fromJson(json)).toList();
    results.sort((a, b) {
      final timeCompare = b.itemTimestamp.compareTo(a.itemTimestamp);
      if (timeCompare != 0) return timeCompare;
      return b.imageName.compareTo(a.imageName);
    });
    return results;
  }

  Future<List<ImageModel>> searchImages(String query) async {
    // Check cache first
    if (_searchCache.containsKey(query)) {
      _logger.fine('Cache hit for query: $query');
      return _searchCache[query]!;
    }

    final headers = await _getHeaders();
    final url = '$baseUrl/db/q?q=${Uri.encodeComponent(query)}';
    _logRequest('GET', url, headers);

    try {
      final response = await _client.get(
        Uri.parse(url),
        headers: headers,
      );
      _logResponse('GET', url, response);

      if (response.statusCode == 200) {
        final results = _toImages(response.body);
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

  String getImageUrl(String relativePath) {
    final url = '$baseUrl$relativePath';
    _logger.finer('Image URL: $url');
    return url;
  }

  List<DirectoryModel> _toDirectories(String body) {
    final List<dynamic> jsonList = json.decode(body);
    final results =
        jsonList.map((json) => DirectoryModel.fromJson(json)).toList();
    results.sort((a, b) {
      final timeCompare = b.directoryTime.compareTo(a.directoryTime);
      if (timeCompare != 0) return timeCompare;
      return a.id.compareTo(b.id);
    });
    return results;
  }

  Future<List<DirectoryModel>> getAlbums() async {
    final url = '$baseUrl/db/q?q=albums:&kind=album';
    final headers = await _getHeaders();
    _logRequest('GET', url, headers);

    try {
      final response = await _client.get(
        Uri.parse(url),
        headers: headers,
      );

      _logResponse('GET', url, response);

      if (response.statusCode == 200) {
        return _toDirectories(response.body);
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

  Future<List<DirectoryModel>> searchAlbums(String query) async {
    final url = '$baseUrl/db/q?q=in:${Uri.encodeComponent(query)}&kind=album';
    final headers = await _getHeaders();
    _logRequest('GET', url, headers);

    try {
      final response = await _client.get(
        Uri.parse(url),
        headers: headers,
      );

      _logResponse('GET', url, response);

      if (response.statusCode == 200) {
        return _toDirectories(response.body);
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

  String getDownloadUrl(
    String albumPath, {
    bool highQuality = false,
  }) {
    return '$baseUrl/db/viewer?command=download&q=${Uri.encodeComponent(albumPath)}&s=${highQuality ? "O" : "M"}';
  }

  String getMontageUrl(List<int> imageIds,
      {required int width, required int height}) {
    final geometry = '${width}x$height';
    final spec = '$geometry-${imageIds.join('-')}';
    return '$baseUrl/db/montage/$spec';
  }

  // Helper method to get the current user's ID token
  Future<String?> _getIdToken() async {
    final user = FirebaseAuth.instance.currentUser;
    if (user == null) {
      _logger.warning('No user is currently signed in');
      return null;
    }
    try {
      // Get current token without forcing refresh
      final token = await user.getIdToken(false);
      _logger.fine('Got ID token');
      return token;
    } catch (e) {
      _logger.warning('Error getting current token, trying refresh: $e');
      try {
        // If current token failed, try forcing a refresh
        final token = await user.getIdToken(true);
        _logger.info('Got fresh token after refresh');
        return token;
      } catch (e) {
        _logger.severe('Error getting ID token even after refresh: $e');
        return null;
      }
    }
  }

  // Helper method to create authenticated headers
  Future<Map<String, String>> _getHeaders() async {
    final token = await _getIdToken();
    final headers = {
      'Accept': 'application/json',
      if (token != null) 'Authorization': 'Bearer $token',
    };
    _logger.fine('Generated headers with token: ${token != null}');
    return headers;
  }

  /// Signs out the current user and clears the authentication state
  Future<void> signOut() async {
    await FirebaseAuth.instance.signOut();
    _logger.info('User signed out successfully');
  }

  Future<Map<String, String>> getImageHeaders() async {
    final token = await _getIdToken();
    return {
      if (token != null) 'Authorization': 'Bearer $token',
    };
  }
}
