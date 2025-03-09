import 'package:flutter/material.dart';
import '../models/image.dart';
import '../services/api_service.dart';

class ImageDetailView extends StatelessWidget {
  final ImageModel image;
  final ApiService apiService;

  const ImageDetailView({
    super.key,
    required this.image,
    required this.apiService,
  });

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(image.imageName),
        actions: [
          IconButton(
            icon: const Icon(Icons.download),
            onPressed: () {
              // TODO: Implement image download
            },
          ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: InteractiveViewer(
              maxScale: 5.0,
              child: Hero(
                tag: 'image_${image.id}',
                child: Image.network(
                  apiService.getImageUrl(image.maxiPath),
                  fit: BoxFit.contain,
                  errorBuilder: (context, error, stackTrace) {
                    return const Center(
                      child: Icon(Icons.error_outline, size: 48),
                    );
                  },
                ),
              ),
            ),
          ),
          Container(
            padding: const EdgeInsets.all(16),
            color: Theme.of(context).colorScheme.surface,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                if (image.keywords.isNotEmpty) ...[
                  const Text(
                    'Keywords',
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Wrap(
                    spacing: 8,
                    runSpacing: 4,
                    children: image.keywords.map((keyword) => Chip(
                      label: Text(keyword),
                      backgroundColor: Theme.of(context).colorScheme.primaryContainer,
                      labelStyle: TextStyle(
                        color: Theme.of(context).colorScheme.onPrimaryContainer,
                      ),
                    )).toList(),
                  ),
                  const SizedBox(height: 16),
                ],
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text('${image.width} × ${image.height}'),
                    Text(
                      'Taken: ${_formatDate(image.itemTimestamp)}',
                    ),
                  ],
                ),
                if (image.stereo != null) ...[
                  const SizedBox(height: 8),
                  const Text(
                    'Stereo Image',
                    style: TextStyle(fontStyle: FontStyle.italic),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
  }
} 