import 'package:flutter/material.dart';
import 'package:flutter/services.dart' show rootBundle;
import 'package:flutter_markdown/flutter_markdown.dart';
import 'flow_view.dart';
import 'albums_view.dart';
import '../models/view_type.dart';

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
  int? _scrollToImageId;  // Add this to track which image to scroll to
  String _helpContent = '';  // Will store the loaded markdown content

  @override
  void initState() {
    super.initState();
    _currentView = widget.initialView;
    if (widget.initialSearchString != null) {
      _currentSearch = widget.initialSearchString!;
      _searchController.text = widget.initialSearchString!;
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
    setState(() {
      _currentSearch = query;
      _scrollToImageId = null;  // Reset scroll target on manual search
    });
  }

  void _clearSearch() {
    setState(() {
      _currentSearch = '';
      _searchController.clear();
      _scrollToImageId = null;
    });
  }

  void _handleKeywordSearch(String keyword, int imageId) {
    setState(() {
      _currentView = ViewType.images;  // Switch to Images tab
      _currentSearch = keyword;
      _searchController.text = keyword;
      _scrollToImageId = imageId;  // Set the image to scroll to
    });
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

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: TextField(
          controller: _searchController,
          decoration: InputDecoration(
            hintText: _currentView == ViewType.albums ? 'Recherche albums...' : 'Recherche photos...',
            suffixIcon: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (_currentSearch.isNotEmpty)
                  IconButton(
                    icon: const Icon(Icons.clear),
                    onPressed: _clearSearch,
                  ),
                IconButton(
                  icon: const Icon(Icons.search),
                  onPressed: () => _performSearch(_searchController.text),
                ),
                IconButton(
                  icon: const Icon(Icons.help_outline),
                  onPressed: () => _showSearchHelp(context),
                  tooltip: 'Aide recherche',
                ),
              ],
            ),
            border: InputBorder.none,
          ),
          onSubmitted: _performSearch,
        ),
      ),
      body: IndexedStack(
        index: _currentView == ViewType.albums ? 0 : 1,
        children: [
          AlbumsView(
            searchQuery: _currentView == ViewType.albums ? _currentSearch : '',
            onAlbumSelected: (albumId) {
              setState(() {
                _currentView = ViewType.images;
                _currentSearch = 'album:$albumId';
                _searchController.text = _currentSearch;
              });
            },
          ),
          FlowView(
            searchQuery: _currentView == ViewType.images ? _currentSearch : '',
            scrollToImageId: _scrollToImageId,
            onKeywordSearch: _handleKeywordSearch,
          ),
        ],
      ),
    );
  }
} 
