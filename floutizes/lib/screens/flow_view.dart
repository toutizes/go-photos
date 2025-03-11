import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../models/image.dart';
import '../services/api_service.dart';
import '../utils/layout_utils.dart';
import '../utils/montaged_images.dart';

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
    
    // Calculate number of columns using the same logic as in _images
    final numColumns = calculateOptimalColumns(
      screenWidth: width,
      minItemWidth: 200.0,
      maxItemWidth: 300.0,
      maxAllowedColumns: 6,
    );

    // Calculate item width and height
    final itemWidth = (width - 16) / numColumns; // Account for padding (8 on each side)
    final rowHeight = itemWidth + 8; // Square items + spacing
    final row = index ~/ numColumns;

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

    // Calculate item size based on available width
    final screenWidth = MediaQuery.of(context).size.width;
    final itemWidth = (screenWidth - 16 - (numColumns - 1) * 8) / numColumns;

    // Create montaged images handler
    final montaged = MontagedImages.fromImageModels(images);

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
            child: montaged.buildImage(image),
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
