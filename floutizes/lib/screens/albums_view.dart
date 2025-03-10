import 'package:flutter/material.dart';
import '../models/image.dart';
import '../models/directory.dart';
import '../services/api_service.dart';

class AlbumsView extends StatefulWidget {
  final ApiService apiService;
  final String searchQuery;
  final Function(String) onAlbumSelected;

  const AlbumsView({
    super.key,
    required this.apiService,
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
        _albumsFuture = widget.apiService.getAlbums();
      } else {
        _albumsFuture = widget.apiService.searchAlbums(widget.searchQuery);
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
                Text('Error loading albums: ${snapshot.error}'),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: _loadAlbums,
                  child: const Text('Retry'),
                ),
              ],
            ),
          );
        }

        final albums = snapshot.data!;
        if (albums.isEmpty) {
          return const Center(child: Text('No albums found'));
        }

        // Sort albums by directoryTime (descending) and id (ascending)
        final sortedAlbums = List<DirectoryModel>.from(albums)
          ..sort((a, b) {
            final timeCompare = b.directoryTime.compareTo(a.directoryTime);
            if (timeCompare != 0) return timeCompare;
            return a.id.compareTo(b.id);
          });

        return GridView.builder(
          padding: const EdgeInsets.all(8),
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: 2,
            childAspectRatio: 1.0,
            mainAxisSpacing: 8.0,
            crossAxisSpacing: 8.0,
          ),
          itemCount: sortedAlbums.length,
          itemBuilder: (context, index) {
            final album = sortedAlbums[index];
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
                          widget.apiService.getImageUrl(album.coverMidiPath),
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
      },
    );
  }
}

class AlbumDetailView extends StatelessWidget {
  final ApiService apiService;
  final String albumPath;
  final List<ImageModel> images;

  const AlbumDetailView({
    super.key,
    required this.apiService,
    required this.albumPath,
    required this.images,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(albumPath),
        actions: [
          PopupMenuButton<String>(
            onSelected: (value) async {
              try {
                await apiService.downloadAlbum(albumPath, highQuality: value == 'download_hq');
                if (!context.mounted) return;
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Download started')),
                );
              } catch (e) {
                if (!context.mounted) return;
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text('Download failed: $e')),
                );
              }
            },
            itemBuilder: (context) => [
              const PopupMenuItem(
                value: 'download_normal',
                child: Text('Download Album'),
              ),
              const PopupMenuItem(
                value: 'download_hq',
                child: Text('Download Album (HQ)'),
              ),
            ],
          ),
        ],
      ),
      body: GridView.builder(
        padding: const EdgeInsets.all(8),
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 3,
          childAspectRatio: 1.0,
          mainAxisSpacing: 8.0,
          crossAxisSpacing: 8.0,
        ),
        itemCount: images.length,
        itemBuilder: (context, index) {
          final image = images[index];
          return GestureDetector(
            onTap: () {
              // TODO: Navigate to image detail view
            },
            child: Hero(
              tag: 'image_${image.id}',
              child: Image.network(
                apiService.getImageUrl(image.miniPath),
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
    );
  }
} 
