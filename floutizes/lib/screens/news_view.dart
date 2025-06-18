import 'package:flutter/material.dart';
import 'package:material_symbols_icons/symbols.dart';
import '../models/directory.dart';
import '../models/image.dart';
import '../services/api_service.dart';
import '../utils/montaged_images.dart';

class NewsView extends StatefulWidget {
  final Function(String) onAlbumSelected;

  const NewsView({
    super.key,
    required this.onAlbumSelected,
  });

  @override
  State<NewsView> createState() => _NewsViewState();
}

class _NewsViewState extends State<NewsView> {
  Future<List<DirectoryModel>>? _albumsFuture;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _loadRecentAlbums();
  }

  Future<void> _loadRecentAlbums() async {
    setState(() {
      _isLoading = true;
    });

    try {
      // Get all albums and sort by directoryTime (most recent first)
      _albumsFuture = ApiService.instance.getAlbums().then((albums) {
        albums.sort((a, b) => b.directoryTime.compareTo(a.directoryTime));
        return albums;
      });
      await _albumsFuture;
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  Widget _buildAlbumCard(DirectoryModel album) {
    // Create fake ImageModel objects for montage
    final List<ImageModel> previewImages = [];
    
    // Add cover image
    previewImages.add(ImageModel(
      id: album.coverId,
      albumDir: album.id,
      imageName: album.coverName,
      itemTimestamp: album.albumTime,
      fileTimestamp: album.directoryTime,
      height: 1,
      width: 1,
      keywords: [],
    ));
    
    // Add preview images (only up to 3 additional ones)
    for (int i = 0; i < album.previewIds.length && i < 3; i++) {
      previewImages.add(ImageModel(
        id: album.previewIds[i],
        albumDir: album.id,
        imageName: album.previewNames[i],
        itemTimestamp: album.albumTime,
        fileTimestamp: album.directoryTime,
        height: 1,
        width: 1,
        keywords: [],
      ));
    }
    
    // Create montaged images handler
    final montaged = MontagedImages.fromImageModels(previewImages, groupSize: 4);
    
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: InkWell(
        onTap: () => widget.onAlbumSelected(album.id),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // First line: Album title and photo count
              Row(
                children: [
                  Expanded(
                    child: Text(
                      album.id,
                      style: Theme.of(context).textTheme.titleMedium,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  Text(
                    '${album.imageCount} ${album.imageCount == 1 ? 'photo' : 'photos'}',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: Theme.of(context).colorScheme.onSurfaceVariant,
                    ),
                  ),
                  const SizedBox(width: 8),
                  const Icon(Symbols.arrow_forward_ios, size: 16),
                ],
              ),
              const SizedBox(height: 12),
              // Second line: Photo previews
              Row(
                children: [
                  for (int i = 0; i < previewImages.length; i++) ...[
                    Container(
                      width: 80,
                      height: 80,
                      decoration: BoxDecoration(
                        borderRadius: BorderRadius.circular(8),
                        color: Theme.of(context).colorScheme.surfaceVariant,
                      ),
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: SizedBox(
                          width: 80,
                          height: 80,
                          child: FittedBox(
                            fit: BoxFit.cover,
                            child: SizedBox(
                              width: 360,
                              height: 360,
                              child: montaged.buildImage(previewImages[i]),
                            ),
                          ),
                        ),
                      ),
                    ),
                    if (i < previewImages.length - 1) const SizedBox(width: 8),
                  ],
                ],
              ),
            ],
          ),
        ),
      ),
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
                const Icon(
                  Symbols.error_outline,
                  size: 48,
                  color: Colors.grey,
                ),
                const SizedBox(height: 16),
                Text('Erreur actualités: ${snapshot.error}'),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: _loadRecentAlbums,
                  child: const Text('Réessayer'),
                ),
              ],
            ),
          );
        }

        final albums = snapshot.data!;
        
        if (albums.isEmpty) {
          return const Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  Symbols.folder,
                  size: 48,
                  color: Colors.grey,
                ),
                SizedBox(height: 16),
                Text('Aucun album trouvé'),
              ],
            ),
          );
        }

        return RefreshIndicator(
          onRefresh: _loadRecentAlbums,
          child: ListView.builder(
            padding: const EdgeInsets.symmetric(vertical: 8),
            itemCount: albums.length,
            itemBuilder: (context, index) {
              return _buildAlbumCard(albums[index]);
            },
          ),
        );
      },
    );
  }
}