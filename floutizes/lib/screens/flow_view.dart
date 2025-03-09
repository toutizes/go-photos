import 'package:flutter/material.dart';
import '../models/image.dart';
import '../services/api_service.dart';
import 'image_detail_view.dart';

class FlowView extends StatefulWidget {
  final ApiService apiService;
  final String searchQuery;
  final int? scrollToImageId;  // ID of the image to scroll to after loading
  final Function(String, int)? onKeywordSearch;  // Callback for keyword search with current image ID

  const FlowView({
    super.key,
    required this.apiService,
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

  Future<void> _loadImages() async {
    setState(() {
      _isLoading = true;
    });

    try {
      if (widget.searchQuery.isEmpty) {
        _imagesFuture = widget.apiService.searchImages('all:');
      } else {
        _imagesFuture = widget.apiService.searchImages(widget.searchQuery);
      }
      
      // Wait for the images to load, then scroll if needed
      final images = await _imagesFuture;
      if (images == null) return;  // Don't attempt to scroll if images is null
      
      if (widget.scrollToImageId != null && mounted) {
        final index = images.indexWhere((img) => img.id == widget.scrollToImageId);
        if (index != -1) {
          // Wait for the grid to be built
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted) return;
            
            // Calculate the grid metrics
            final width = MediaQuery.of(context).size.width;
            final height = MediaQuery.of(context).size.height;
            final itemWidth = (width - 32) / 3;  // Account for padding and spacing
            final rowHeight = itemWidth;  // Square items
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
          });
        }
      }
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
                  child: const Text('Retry'),
                ),
              ],
            ),
          );
        }

        final images = snapshot.data!;
        if (images.isEmpty) {
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
                  const Text('No images available')
                else
                  Column(
                    children: [
                      Text('No images found for "${widget.searchQuery}"'),
                      const SizedBox(height: 8),
                      if (widget.onKeywordSearch != null)
                        TextButton.icon(
                          onPressed: () => widget.onKeywordSearch!('', -1),
                          icon: const Icon(Icons.clear),
                          label: const Text('Clear search'),
                        ),
                    ],
                  ),
              ],
            ),
          );
        }

        return GridView.builder(
          controller: _scrollController,
          padding: const EdgeInsets.all(8),
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: 3,
            mainAxisSpacing: 8,
            crossAxisSpacing: 8,
          ),
          itemCount: images.length,
          itemBuilder: (context, index) {
            final image = images[index];
            return GestureDetector(
              onTap: () {
                Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (context) => ImageDetailView(
                      image: image,
                      apiService: widget.apiService,
                      allImages: images,
                      currentIndex: index,
                      onKeywordSearch: widget.onKeywordSearch,
                    ),
                  ),
                );
              },
              child: Hero(
                tag: 'image_${image.id}',
                child: Image.network(
                  widget.apiService.getImageUrl(image.miniPath),
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
      },
    );
  }
} 
