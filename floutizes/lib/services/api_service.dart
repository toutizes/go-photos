import 'dart:convert';

import 'package:logging/logging.dart';
import 'package:http/http.dart' as http;
import 'package:firebase_auth/firebase_auth.dart';

import '../models/image.dart';
import '../models/directory.dart';
import '../models/keyword.dart';
import '../models/user_query.dart';
import 'auth_service.dart';

class ApiService {
  static ApiService? _instance;
  static final _logger = Logger('ApiService');
  final AuthService authService;
  final String baseUrl;
  final http.Client _client;

  // Cache for search results
  final int _maxCacheSize = 3;
  final Map<String, List<ImageModel>> _searchCache = {};
  final List<String> _cacheOrder = [];

  ApiService._({required this.baseUrl, required this.authService})
      : _client = http.Client();

  static void _initLogging() {
    Logger.root.level = Level.OFF;
    Logger.root.onRecord.listen((record) {
      print('${record.level.name}: ${record.time}: ${record.message}');
    });
  }

  static void initialize({required String baseUrl, required authService}) {
    _initLogging();
    _instance = ApiService._(baseUrl: baseUrl, authService: authService);
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

  List<ImageModel> _toImages(String body, String query) {
    final List<dynamic> jsonList = json.decode(body);
    final results = jsonList.map((json) => ImageModel.fromJson(json)).toList();
    
    // Check if this is an album query (any term starts with "album:")
    final queryTerms = query.toLowerCase().split(' ');
    final isAlbumQuery = queryTerms.any((term) => term.startsWith('album:'));
    
    results.sort((a, b) {
      final timeCompare = isAlbumQuery 
          ? a.itemTimestamp.compareTo(b.itemTimestamp)  // Increasing for album queries
          : b.itemTimestamp.compareTo(a.itemTimestamp); // Decreasing for other queries
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

    final url = '$baseUrl/db/q?q=${Uri.encodeComponent(query)}';

    try {
      final response = await _makeAuthenticatedRequest(
        (headers) => _client.get(Uri.parse(url), headers: headers),
        'GET',
        url,
      );
      _logResponse('GET', url, response);

      if (response.statusCode == 200) {
        final results = _toImages(response.body, query);
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
      final timeCompare = b.albumTime.compareTo(a.albumTime);
      if (timeCompare != 0) return timeCompare;
      return a.id.compareTo(b.id);
    });
    return results;
  }

  Future<List<DirectoryModel>> getAlbums() async {
    final url = '$baseUrl/db/q?q=albums:&kind=album';

    try {
      final response = await _makeAuthenticatedRequest(
        (headers) => _client.get(Uri.parse(url), headers: headers),
        'GET',
        url,
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

    try {
      final response = await _makeAuthenticatedRequest(
        (headers) => _client.get(Uri.parse(url), headers: headers),
        'GET',
        url,
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
  String? _getIdToken() => authService.idToken;

  // Helper method to create authenticated headers
  Map<String, String> _getHeaders() {
    final token = _getIdToken();
    final headers = {
      'Accept': 'application/json',
      if (token != null) 'Authorization': 'Bearer $token',
    };
    _logger.fine('Generated headers with token: ${token != null}');
    return headers;
  }

  /// Make an authenticated HTTP request with automatic token refresh on 401
  Future<http.Response> _makeAuthenticatedRequest(
    Future<http.Response> Function(Map<String, String> headers) requestFn,
    String method,
    String url,
  ) async {
    var headers = _getHeaders();
    _logRequest(method, url, headers);

    var response = await requestFn(headers);
    
    // If we get a 401, try refreshing the token and retry once
    if (response.statusCode == 401) {
      _logger.info('Got 401, attempting token refresh...');
      
      final newToken = await authService.refreshToken();
      if (newToken != null) {
        headers = _getHeaders();
        _logger.info('Retrying request with refreshed token');
        response = await requestFn(headers);
      }
    }
    
    return response;
  }

  Future<List<KeywordModel>> getRecentKeywords() async {
    final url = '$baseUrl/db/recent-keywords';

    try {
      final response = await _makeAuthenticatedRequest(
        (headers) => _client.get(Uri.parse(url), headers: headers),
        'GET',
        url,
      );

      _logResponse('GET', url, response);

      if (response.statusCode == 200) {
        final Map<String, dynamic> responseData = json.decode(response.body);
        final List<dynamic> keywordsJson = responseData['keywords'] as List<dynamic>;
        return keywordsJson
            .map((json) => KeywordModel.fromJson(json as Map<String, dynamic>))
            .toList();
      } else {
        final error = 'Failed to fetch recent keywords: ${response.statusCode}';
        _logResponse('GET', url, response, error);
        throw Exception(error);
      }
    } catch (e) {
      _logger.severe('Error fetching recent keywords: $e');
      rethrow;
    }
  }

  Future<List<KeywordGroupModel>> getRecentKeywordGroups() async {
    final url = '$baseUrl/db/recent-keyword-groups';

    try {
      final response = await _makeAuthenticatedRequest(
        (headers) => _client.get(Uri.parse(url), headers: headers),
        'GET',
        url,
      );

      _logResponse('GET', url, response);

      if (response.statusCode == 200) {
        final Map<String, dynamic> responseData = json.decode(response.body);
        final List<dynamic> groupsJson = responseData['groups'] as List<dynamic>;
        return groupsJson
            .map((json) => KeywordGroupModel.fromJson(json as Map<String, dynamic>))
            .toList();
      } else {
        final error = 'Failed to fetch recent keyword groups: ${response.statusCode}';
        _logResponse('GET', url, response, error);
        throw Exception(error);
      }
    } catch (e) {
      _logger.severe('Error fetching recent keyword groups: $e');
      rethrow;
    }
  }

  /// Signs out the current user and clears the authentication state
  Future<void> signOut() async {
    await FirebaseAuth.instance.signOut();
    _logger.info('User signed out successfully');
  }

  Map<String, String> getImageHeaders() {
    final token = _getIdToken();
    return {
      if (token != null) 'Authorization': 'Bearer $token',
    };
  }

  Future<AllUserQueriesModel> getUserQueries() async {
    final url = '$baseUrl/db/user-queries';

    try {
      final response = await _makeAuthenticatedRequest(
        (headers) => _client.get(Uri.parse(url), headers: headers),
        'GET',
        url,
      );

      _logResponse('GET', url, response);

      if (response.statusCode == 200) {
        final Map<String, dynamic> responseData = json.decode(response.body);
        return AllUserQueriesModel.fromJson(responseData);
      } else {
        final error = 'Failed to fetch user queries: ${response.statusCode}';
        _logResponse('GET', url, response, error);
        throw Exception(error);
      }
    } catch (e) {
      _logger.severe('Error fetching user queries: $e');
      rethrow;
    }
  }
}
