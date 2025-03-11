import 'package:flutter/material.dart';
import '../models/directory.dart';
import '../services/api_service.dart';

class AlbumsView extends StatefulWidget {
  final String searchQuery;
  final Function(String) onAlbumSelected;

  const AlbumsView({
    super.key,
    required this.searchQuery,
    required this.onAlbumSelected,
  });

  @override
  State<AlbumsView> createState() => _AlbumsViewState();
}

class _AlbumsViewState extends State<AlbumsView> {
  Future<List<DirectoryModel>>? _albumsFuture;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _loadAlbums();
  }

  @override
  void didUpdateWidget(AlbumsView oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.searchQuery != oldWidget.searchQuery) {
      _loadAlbums();
    }
  }

  Future<void> _loadAlbums() async {
    setState(() {
      _isLoading = true;
    });

    try {
      if (widget.searchQuery.isEmpty) {
        _albumsFuture = ApiService.instance.getAlbums();
      } else {
        _albumsFuture = ApiService.instance.searchAlbums(widget.searchQuery);
      }
      await _albumsFuture;
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  Widget _albumView(List<DirectoryModel> albums) {
    return GridView.builder(
      padding: const EdgeInsets.all(8),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        childAspectRatio: 1.0,
        mainAxisSpacing: 8.0,
        crossAxisSpacing: 8.0,
      ),
      itemCount: albums.length,
      itemBuilder: (context, index) {
        final album = albums[index];
        return GestureDetector(
          onTap: () => widget.onAlbumSelected(album.id),
          child: Card(
            clipBehavior: Clip.antiAlias,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Expanded(
                  child: Hero(
                    tag: 'album_${album.id}',
                    child: Image.network(
                      ApiService.instance.getImageUrl(album.coverMidiPath),
                      fit: BoxFit.cover,
                      errorBuilder: (context, error, stackTrace) {
                        return const Center(
                          child: Icon(Icons.error_outline),
                        );
                      },
                    ),
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        album.id,
                        style: Theme.of(context).textTheme.titleMedium,
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                      Text(
                        '${album.imageCount} ${album.imageCount == 1 ? 'photo' : 'photos'}',
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    return FutureBuilder<List<DirectoryModel>>(
      future: _albumsFuture,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Center(child: CircularProgressIndicator());
        }

        if (snapshot.hasError) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text('Erreur albums: ${snapshot.error}'),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: _loadAlbums,
                  child: const Text('Réessayer'),
                ),
              ],
            ),
          );
        }

        final albums = snapshot.data!;
        if (albums.isEmpty) {
          return const Center(child: Text('Pas d\'albums'));
        }
        return _albumView(albums);
      },
    );
  }
}
