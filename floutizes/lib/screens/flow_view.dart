import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../models/image.dart';
import '../services/api_service.dart';
import '../utils/layout_utils.dart';

/// Calculates the optimal number of columns for a grid based on screen width and desired item width.
/// 
/// Parameters:
/// - screenWidth: The available width of the screen
/// - minItemWidth: The minimum desired width for each item
/// - maxItemWidth: The maximum desired width for each item
/// - maxAllowedColumns: The maximum number of columns allowed
/// - padding: The total horizontal padding of the grid (default: 16.0)
/// - spacing: The spacing between items (default: 8.0)
int calculateOptimalColumns({
  required double screenWidth,
  required double minItemWidth,
  required double maxItemWidth,
  required int maxAllowedColumns,
  double padding = 16.0,
  double spacing = 8.0,
}) {
  // Calculate number of columns that would fit with minItemWidth
  int maxColumns = ((screenWidth - padding) / (minItemWidth + spacing)).floor();
  // Calculate number of columns that would fit with maxItemWidth
  int minColumns = ((screenWidth - padding) / (maxItemWidth + spacing)).ceil();
  
  // Ensure minColumns doesn't exceed our maximum allowed columns
  minColumns = minColumns.clamp(1, maxAllowedColumns);
  // Ensure maxColumns is at least as large as minColumns
  maxColumns = maxColumns.clamp(minColumns, maxAllowedColumns);
  
  return maxColumns;
}

class FlowView extends StatefulWidget {
  final String searchQuery;
  // ID of the image to scroll to after loading
  final int? scrollToImageId;
  // Callback for keyword search with current image ID
  final Function(String, int)? onKeywordSearch;

  const FlowView({
    super.key,
    required this.searchQuery,
    this.scrollToImageId,
    this.onKeywordSearch,
  });

  @override
  State<FlowView> createState() => _FlowViewState();
}

class _FlowViewState extends State<FlowView> {
  Future<List<ImageModel>>? _imagesFuture;
  bool _isLoading = false;
  final ScrollController _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    _loadImages();
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  @override
  void didUpdateWidget(FlowView oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.searchQuery != oldWidget.searchQuery) {
      _loadImages();
    }
  }

  void _scrollToImage(Duration timeStamp) async {
    if (widget.scrollToImageId == null) return;
    if (!mounted) return;

    final images = await _imagesFuture;
    if (images == null) return;

    final index = images.indexWhere((img) => img.id == widget.scrollToImageId);
    if (index == -1) return;

    // Add a small delay to ensure the ScrollController is attached
    await Future.delayed(const Duration(milliseconds: 100));
    if (!mounted) return;

    // Calculate the grid metrics
    final width = MediaQuery.of(context).size.width;
    final height = MediaQuery.of(context).size.height;
    final itemWidth = (width - 32) / 3; // Account for padding and spacing
    final rowHeight = itemWidth + 8; // Square items + spacing
    final row = index ~/ 3;

    // Calculate target offset to center the item
    final itemOffset = row * rowHeight;
    final targetOffset = itemOffset - (height - rowHeight) / 2;

    // Ensure the scroll controller is ready
    if (_scrollController.hasClients) {
      _scrollController.animateTo(
        targetOffset.clamp(0.0, _scrollController.position.maxScrollExtent),
        duration: const Duration(milliseconds: 500),
        curve: Curves.easeInOut,
      );
    }
  }

  Future<void> _loadImages() async {
    setState(() {
      _isLoading = true;
      if (widget.searchQuery.isEmpty) {
        _imagesFuture = ApiService.instance.searchImages('all:');
      } else {
        _imagesFuture = ApiService.instance.searchImages(widget.searchQuery);
      }
    });

    try {
      WidgetsBinding.instance.addPostFrameCallback(_scrollToImage);
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  Widget _noImages() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(
            Icons.search_off,
            size: 48,
            color: Colors.grey,
          ),
          const SizedBox(height: 16),
          if (widget.searchQuery.isEmpty)
            const Text('Saissisez un mot clé pour commencer')
          else
            Column(
              children: [
                Text('Pas d\'images pour "${widget.searchQuery}"'),
                const SizedBox(height: 8),
                if (widget.onKeywordSearch != null)
                  TextButton.icon(
                    onPressed: () => widget.onKeywordSearch!('', -1),
                    icon: const Icon(Icons.clear),
                    label: const Text('Annuler'),
                  ),
              ],
            ),
        ],
      ),
    );
  }

  PreferredSizeWidget _buildAppBar(List<ImageModel>? images) {
    return AppBar(
      title: Row(
        children: [
          if (widget.searchQuery.isNotEmpty) ...[
            Expanded(
              child: Text(widget.searchQuery),
            ),
            if (images != null)
              Text(
                '${images.length} ${images.length == 1 ? 'photo' : 'photos'}',
                style: Theme.of(context).textTheme.titleMedium,
              ),
          ],
        ],
      ),
    );
  }

  Widget _images(List<ImageModel> images) {
    // Calculate optimal number of columns based on screen width
    final numColumns = calculateOptimalColumns(
      screenWidth: MediaQuery.of(context).size.width,
      minItemWidth: 200.0,
      maxItemWidth: 300.0,
      maxAllowedColumns: 6,
    );

    // Add two rows of padding items for better scrolling of last rows
    final int paddingItemCount = numColumns * 2; // 2 rows × numColumns

    return GridView.builder(
      controller: _scrollController,
      padding: const EdgeInsets.all(8),
      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: numColumns,
        mainAxisSpacing: 8,
        crossAxisSpacing: 8,
      ),
      itemCount: images.length + paddingItemCount,
      itemBuilder: (context, index) {
        if (index >= images.length) {
          // Return an empty, transparent container for padding items
          return const SizedBox();
        }

        final image = images[index];
        return GestureDetector(
          onTap: () {
            context.go(
                '/images/details/${image.id}?q=${Uri.encodeComponent(widget.searchQuery)}');
          },
          child: Hero(
            tag: 'image_${image.id}',
            child: Image.network(
              ApiService.instance.getImageUrl(image.miniPath),
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
    );
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return const Center(child: CircularProgressIndicator());
    }

    return FutureBuilder<List<ImageModel>>(
      future: _imagesFuture,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Center(child: CircularProgressIndicator());
        }

        if (snapshot.hasError) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Text('Error loading images: ${snapshot.error}'),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: _loadImages,
                  child: const Text('Réssayer'),
                ),
              ],
            ),
          );
        }

        final images = snapshot.data!;
        return Scaffold(
          appBar: _buildAppBar(images),
          body: images.isEmpty ? _noImages() : _images(images),
        );
      },
    );
  }
}
