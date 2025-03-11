import 'package:flutter/material.dart';
import '../models/directory.dart';
import '../services/api_service.dart';
import '../utils/layout_utils.dart';

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
    // Calculate optimal number of columns based on screen width
    final numColumns = calculateOptimalColumns(
      screenWidth: MediaQuery.of(context).size.width,
      minItemWidth: 200.0,
      maxItemWidth: 300.0,
      maxAllowedColumns: 6,
    );

    // Calculate item size based on available width
    final screenWidth = MediaQuery.of(context).size.width;
    final itemWidth = (screenWidth - 16 - (numColumns - 1) * 8) / numColumns;
    final itemHeight = itemWidth; // Square items

    // Group albums into montage groups of 8, padding the last group if needed
    const montageGroupSize = 8;
    final List<List<int>> montageGroups = [];
    
    for (var i = 0; i < albums.length; i += montageGroupSize) {
      final end = (i + montageGroupSize <= albums.length) ? i + montageGroupSize : albums.length;
      final group = albums.sublist(i, end).map((album) => album.coverId).toList();
      
      // Pad the last group with repeated last cover ID if needed
      if (group.length < montageGroupSize && group.isNotEmpty) {
        final lastId = group.last;
        while (group.length < montageGroupSize) {
          group.add(lastId);
        }
      }
      
      montageGroups.add(group);
    }

    // Create montage URLs for each group
    final montageUrls = montageGroups.map((group) => ApiService.instance.getMontageUrl(
      group,
      width: itemWidth.round(),
      height: itemHeight.round(),
    )).toList();

    return GridView.builder(
      padding: const EdgeInsets.all(8),
      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: numColumns,
        childAspectRatio: 1.0,
        mainAxisSpacing: 8.0,
        crossAxisSpacing: 8.0,
      ),
      itemCount: albums.length,
      itemBuilder: (context, index) {
        final album = albums[index];
        final montageGroupIndex = index ~/ montageGroupSize;
        final positionInGroup = index % montageGroupSize;
        final montageUrl = montageUrls[montageGroupIndex];

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
                      montageUrl,
                      fit: BoxFit.none,
                      alignment: Alignment(-1 + 2 * positionInGroup / (montageGroupSize - 1), 0),
                      width: itemWidth,
                      height: itemHeight,
                      errorBuilder: (context, error, stackTrace) {
                        // Fallback to individual image if montage fails
                        return Image.network(
                          ApiService.instance.getImageUrl(album.coverMiniPath),
                          fit: BoxFit.cover,
                          errorBuilder: (context, error, stackTrace) {
                            return const Center(
                              child: Icon(Icons.error_outline),
                            );
                          },
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
