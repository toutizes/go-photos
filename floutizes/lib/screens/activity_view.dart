import 'package:flutter/material.dart';
import 'package:material_symbols_icons/symbols.dart';
import '../models/keyword.dart';
import '../models/image.dart';
import '../services/api_service.dart';
import '../utils/montaged_images.dart';

class ActivityView extends StatefulWidget {
  final Function(String) onKeywordSearch;

  const ActivityView({
    super.key,
    required this.onKeywordSearch,
  });

  @override
  State<ActivityView> createState() => _ActivityViewState();
}

class _ActivityViewState extends State<ActivityView> {
  Future<List<KeywordGroupModel>>? _keywordGroupsFuture;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _loadRecentKeywords();
  }

  Future<void> _loadRecentKeywords() async {
    setState(() {
      _isLoading = true;
    });

    try {
      _keywordGroupsFuture = ApiService.instance.getRecentKeywordGroups();
      await _keywordGroupsFuture;
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  Widget _buildKeywordGroupCard(KeywordGroupModel group) {
    // Create fake ImageModel objects for montage from the recent images
    final List<ImageModel> previewImages = [];

    for (final recentImage in group.recentImages) {
      previewImages.add(ImageModel(
        id: recentImage.id,
        albumDir: '', // We don't have album dir in the simplified response
        imageName: recentImage.name,
        itemTimestamp: DateTime
            .now(), // We don't have timestamp in the simplified response
        fileTimestamp: DateTime.now(),
        height: 1,
        width: 1,
        keywords: [],
      ));
    }

    // Create montaged images handler
    final montaged =
        MontagedImages.fromImageModels(previewImages, groupSize: 4);

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Keywords row
            Wrap(
              spacing: 8,
              children: group.keywords
                  .map((keyword) => GestureDetector(
                        onTap: () {
                          // Quote keywords containing spaces
                          final searchTerm = keyword.keyword.contains(' ')
                              ? '"${keyword.keyword}"'
                              : keyword.keyword;
                          widget.onKeywordSearch(searchTerm);
                        },
                        child: Chip(
                          label: Text(
                            '${keyword.keyword} (${keyword.count} récentes)',
                            style: Theme.of(context).textTheme.bodySmall,
                          ),
                          backgroundColor:
                              Theme.of(context).colorScheme.secondaryContainer,
                        ),
                      ))
                  .toList(),
            ),
            const SizedBox(height: 12),
            // Second line: Photo previews
            if (previewImages.isNotEmpty)
              Row(
                children: [
                  for (int i = 0; i < previewImages.length; i++) ...[
                    Container(
                      width: 80,
                      height: 80,
                      decoration: BoxDecoration(
                        borderRadius: BorderRadius.circular(8),
                        color: Theme.of(context)
                            .colorScheme
                            .surfaceContainerHighest,
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
    );
  }

  Widget _buildKeywordCard(KeywordModel keyword) {
    // Create fake ImageModel objects for montage from the recent images
    final List<ImageModel> previewImages = [];

    for (final recentImage in keyword.recentImages) {
      previewImages.add(ImageModel(
        id: recentImage.id,
        albumDir: '', // We don't have album dir in the simplified response
        imageName: recentImage.name,
        itemTimestamp: DateTime
            .now(), // We don't have timestamp in the simplified response
        fileTimestamp: DateTime.now(),
        height: 1,
        width: 1,
        keywords: [],
      ));
    }

    // Create montaged images handler
    final montaged =
        MontagedImages.fromImageModels(previewImages, groupSize: 4);

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: InkWell(
        onTap: () {
          // Quote keywords containing spaces
          final searchTerm = keyword.keyword.contains(' ')
              ? '"${keyword.keyword}"'
              : keyword.keyword;
          widget.onKeywordSearch(searchTerm);
        },
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // First line: Keyword and count
              Row(
                children: [
                  Expanded(
                    child: Text(
                      keyword.keyword,
                      style: Theme.of(context).textTheme.titleMedium,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  Text(
                    '${keyword.count} ${keyword.count == 1 ? 'photo récente' : 'photos récentes'}',
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
              if (previewImages.isNotEmpty)
                Row(
                  children: [
                    for (int i = 0; i < previewImages.length; i++) ...[
                      Container(
                        width: 80,
                        height: 80,
                        decoration: BoxDecoration(
                          borderRadius: BorderRadius.circular(8),
                          color: Theme.of(context)
                              .colorScheme
                              .surfaceContainerHighest,
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
                      if (i < previewImages.length - 1)
                        const SizedBox(width: 8),
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

    return FutureBuilder<List<KeywordGroupModel>>(
      future: _keywordGroupsFuture,
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
                Text('Erreur activité: ${snapshot.error}'),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: _loadRecentKeywords,
                  child: const Text('Réessayer'),
                ),
              ],
            ),
          );
        }

        final groups = snapshot.data!;

        if (groups.isEmpty) {
          return const Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  Symbols.trending_up,
                  size: 48,
                  color: Colors.grey,
                ),
                SizedBox(height: 16),
                Text('Aucune activité récente trouvée'),
              ],
            ),
          );
        }

        return RefreshIndicator(
          onRefresh: _loadRecentKeywords,
          child: ListView.builder(
            padding: const EdgeInsets.symmetric(vertical: 8),
            itemCount: groups.length,
            itemBuilder: (context, index) {
              return _buildKeywordGroupCard(groups[index]);
            },
          ),
        );
      },
    );
  }
}
