import 'package:flutter/material.dart';
import 'package:flutter/services.dart' show rootBundle;
import 'package:flutter_markdown/flutter_markdown.dart';
import 'package:go_router/go_router.dart';
import '../services/api_service.dart';
import 'flow_view.dart';
import 'albums_view.dart';
import 'news_view.dart';
import 'activity_view.dart';
import 'admin_queries_view.dart';
import '../models/view_type.dart';
import '../widgets/search_box.dart';

class HomeScreen extends StatefulWidget {
  final ViewType initialView;
  final String? initialSearchString;
  final int? initialImageId;

  const HomeScreen({
    super.key,
    required this.initialView,
    this.initialSearchString,
    this.initialImageId,
  });

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  late ViewType _currentView;
  final TextEditingController _searchController = TextEditingController();
  String _currentSearch = '';
  int? _scrollToImageId; // Add this to track which image to scroll to
  String _helpContent = ''; // Will store the loaded markdown content

  @override
  void initState() {
    super.initState();
    _currentView = widget.initialView;
    if (widget.initialSearchString != null) {
      _currentSearch = widget.initialSearchString!;
      _searchController.text = widget.initialSearchString!;
    } else if (_currentView == ViewType.images) {
      // Default the search term to the current month and day
      final DateTime now = DateTime.now();
      final String month = now.month.toString().padLeft(2, '0');
      final String day = now.day.toString().padLeft(2, '0');
      _currentSearch = '$month-$day';
      _searchController.text = _currentSearch;
    }
    _scrollToImageId = widget.initialImageId;
    _loadHelpContent();
  }

  Future<void> _loadHelpContent() async {
    final String content = await rootBundle.loadString('assets/search_help.md');
    setState(() {
      _helpContent = content;
    });
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  void _performSearch(String query) {
    if (_currentView == ViewType.albums) {
      context.go('/albums?q=${Uri.encodeComponent(query)}');
    } else {
      context.go('/images?q=${Uri.encodeComponent(query)}');
    }
  }

  void _clearSearch() {
    _performSearch('');
  }

  void _showSearchHelp(BuildContext context) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Aide recherche'),
        content: SizedBox(
          width: double.maxFinite,
          child: Markdown(
            data: _helpContent,
            shrinkWrap: true,
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Fermer'),
          ),
        ],
      ),
    );
  }

  Future<void> _handleLogout() async {
    try {
      await ApiService.instance.signOut();
      if (mounted) {
        context.go('/login');
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Erreur lors de la déconnexion: $e')),
        );
      }
    }
  }


  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: SearchBox(
          controller: _searchController,
          hintText: _currentView == ViewType.albums
              ? 'Recherche albums...'
              : 'Recherche photos...',
          onSearch: () => _performSearch(_searchController.text),
          onClear: _clearSearch,
          onHelp: () => _showSearchHelp(context),
          onLogout: _handleLogout,
        ),
      ),
      body: IndexedStack(
        index: _currentView == ViewType.news 
            ? 0 
            : _currentView == ViewType.activity
                ? 1
                : _currentView == ViewType.albums 
                    ? 2 
                    : _currentView == ViewType.images
                        ? 3
                        : 4, // adminQueries
        children: [
          NewsView(
            onAlbumSelected: (albumId) {
              // Quote keywords containing spaces
              var query = albumId.contains(' ') ? '"album:$albumId"' : "album:$albumId";
              context.go('/images?q=${Uri.encodeComponent(query)}');
            },
          ),
          ActivityView(
            onKeywordSearch: (keyword) {
              context.go('/images?q=${Uri.encodeComponent(keyword)}');
            },
            onImageTap: (keyword, imageId) {
              context.go('/images/details/$imageId?q=${Uri.encodeComponent(keyword)}');
            },
          ),
          AlbumsView(
            searchQuery: _currentView == ViewType.albums ? _currentSearch : '',
            onAlbumSelected: (albumId) {
              // Quote keywords containing spaces
              var query = albumId.contains(' ') ? '"album:$albumId"' : "album:$albumId";
              context.go('/images?q=${Uri.encodeComponent(query)}');
            },
          ),
          FlowView(
            searchQuery: _currentView == ViewType.images ? _currentSearch : '',
            scrollToImageId: _scrollToImageId,
            onKeywordSearch: (String keyword, int imageId) => context.go(
                '/images?q=${Uri.encodeComponent(keyword)}&imageId=$imageId'),
          ),
          const AdminQueriesView(),
        ],
      ),
    );
  }
}
