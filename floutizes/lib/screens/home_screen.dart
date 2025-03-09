import 'package:flutter/material.dart';
import 'flow_view.dart';
import 'albums_view.dart';
import '../services/api_service.dart';

class HomeScreen extends StatefulWidget {
  final ApiService apiService;

  const HomeScreen({super.key, required this.apiService});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  int _selectedIndex = 0;
  final TextEditingController _searchController = TextEditingController();
  String _currentSearch = '';

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  void _performSearch(String query) {
    setState(() {
      _currentSearch = query;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: TextField(
          controller: _searchController,
          decoration: InputDecoration(
            hintText: _selectedIndex == 0 ? 'Search albums...' : 'Search photos...',
            suffixIcon: IconButton(
              icon: const Icon(Icons.search),
              onPressed: () => _performSearch(_searchController.text),
            ),
            border: InputBorder.none,
          ),
          onSubmitted: _performSearch,
        ),
      ),
      body: IndexedStack(
        index: _selectedIndex,
        children: [
          AlbumsView(
            apiService: widget.apiService,
            searchQuery: _selectedIndex == 0 ? _currentSearch : '',
            onAlbumSelected: (albumId) {
              setState(() {
                _selectedIndex = 1;
                _currentSearch = 'album:$albumId';
                _searchController.text = _currentSearch;
              });
            },
          ),
          FlowView(
            apiService: widget.apiService,
            searchQuery: _selectedIndex == 1 ? _currentSearch : '',
          ),
        ],
      ),
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _selectedIndex,
        onTap: (index) {
          setState(() {
            _selectedIndex = index;
            _currentSearch = '';
            _searchController.clear();
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
} 
