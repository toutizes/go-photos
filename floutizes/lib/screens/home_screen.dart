import 'package:flutter/material.dart';
import 'flow_view.dart';
import 'albums_view.dart';
import '../services/api_service.dart';
import '../models/image.dart';

class HomeScreen extends StatefulWidget {
  final ApiService apiService;

  const HomeScreen({super.key, required this.apiService});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  int _selectedIndex = 0;  // Start with Albums view
  final TextEditingController _searchController = TextEditingController();
  List<ImageModel>? _searchResults;
  bool _isLoading = false;

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _performSearch(String query) async {
    if (query.isEmpty) {
      setState(() {
        _searchResults = null;
      });
      return;
    }

    setState(() {
      _isLoading = true;
    });

    try {
      final results = await widget.apiService.searchImages(query);
      setState(() {
        _searchResults = results;
        _isLoading = false;
        _selectedIndex = 1;  // Switch to Images tab when showing search results
      });
    } catch (e) {
      setState(() {
        _isLoading = false;
      });
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error searching: $e')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: TextField(
          controller: _searchController,
          decoration: InputDecoration(
            hintText: 'Search photos...',
            suffixIcon: IconButton(
              icon: const Icon(Icons.search),
              onPressed: () => _performSearch(_searchController.text),
            ),
            border: InputBorder.none,
          ),
          onSubmitted: _performSearch,
        ),
      ),
      body: _searchResults != null
          ? _buildSearchResults()
          : IndexedStack(
              index: _selectedIndex,
              children: [
                AlbumsView(
                  apiService: widget.apiService,
                  onSearch: (query) {
                    _searchController.text = query;
                    _performSearch(query);
                  },
                ),
                FlowView(apiService: widget.apiService),
              ],
            ),
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _selectedIndex,
        onTap: (index) {
          setState(() {
            _selectedIndex = index;
            // Clear search when switching tabs
            if (_searchResults != null) {
              _searchResults = null;
              _searchController.clear();
            }
          });
        },
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.album),
            label: 'Albums',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.photo_library),
            label: 'Images',
          ),
        ],
      ),
    );
  }

  Widget _buildSearchResults() {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    if (_searchResults!.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Text('No results found'),
            const SizedBox(height: 16),
            TextButton(
              onPressed: () {
                setState(() {
                  _searchResults = null;
                  _searchController.clear();
                });
              },
              child: const Text('Clear Search'),
            ),
          ],
        ),
      );
    }

    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(8.0),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text('${_searchResults!.length} ${_searchResults!.length == 1 ? 'image' : 'images'}'),
              TextButton(
                onPressed: () {
                  setState(() {
                    _searchResults = null;
                    _searchController.clear();
                  });
                },
                child: const Text('Clear Search'),
              ),
            ],
          ),
        ),
        Expanded(
          child: GridView.builder(
            padding: const EdgeInsets.all(8),
            gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
              crossAxisCount: 3,
              mainAxisSpacing: 8,
              crossAxisSpacing: 8,
            ),
            itemCount: _searchResults!.length,
            itemBuilder: (context, index) {
              final image = _searchResults![index];
              return GestureDetector(
                onTap: () {
                  // TODO: Navigate to image detail view
                },
                child: Hero(
                  tag: 'image_${image.id}',
                  child: Image.network(
                    widget.apiService.getImageUrl(image.miniPath),
                    fit: BoxFit.cover,
                    errorBuilder: (context, error, stackTrace) {
                      return const Center(
                        child: Icon(Icons.error_outline),
                      );
                    },
                  ),
                ),
              );
            },
          ),
        ),
      ],
    );
  }
} 
