import 'package:flutter/material.dart';
import '../models/image.dart';
import '../services/api_service.dart';
import 'image_detail_view.dart';

class FlowView extends StatefulWidget {
  final ApiService apiService;
  final String searchQuery;

  const FlowView({
    super.key,
    required this.apiService,
    required this.searchQuery,
  });

  @override
  State<FlowView> createState() => _FlowViewState();
}

class _FlowViewState extends State<FlowView> {
  Future<List<ImageModel>>? _imagesFuture;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    _loadImages();
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
      await _imagesFuture;
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
          return const Center(child: Text('No images found'));
        }

        return GridView.builder(
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