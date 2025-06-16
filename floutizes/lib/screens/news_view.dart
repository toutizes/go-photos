import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../models/directory.dart';
import '../services/api_service.dart';

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

  Widget _buildAlbumItem(DirectoryModel album) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: ListTile(
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        title: Text(
          album.id,
          style: Theme.of(context).textTheme.titleMedium,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 4),
            Text(
              '${album.imageCount} ${album.imageCount == 1 ? 'photo' : 'photos'}',
              style: Theme.of(context).textTheme.bodySmall,
            ),
            const SizedBox(height: 2),
            Text(
              DateFormat('dd/MM/yyyy HH:mm').format(album.directoryTime),
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
          ],
        ),
        trailing: const Icon(Icons.arrow_forward_ios, size: 16),
        onTap: () => widget.onAlbumSelected(album.id),
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
                  Icons.error_outline,
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
                  Icons.folder_outlined,
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
              return _buildAlbumItem(albums[index]);
            },
          ),
        );
      },
    );
  }
}