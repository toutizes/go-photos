import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import '../models/image.dart';
import '../services/api_service.dart';
import 'photo_view.dart';

class ImageDetailView extends StatefulWidget {
  final String searchQuery;
  final int imageId;
  final Function(String, int)? onKeywordSearch;

  const ImageDetailView({
    super.key,
    required this.searchQuery,
    required this.imageId,
    this.onKeywordSearch,
  });

  @override
  State<ImageDetailView> createState() => _ImageDetailViewState();
}

class _ImageDetailViewState extends State<ImageDetailView> {
  late PageController _pageController;
  late int _currentIndex;
  final FocusNode _focusNode = FocusNode();
  Future<List<ImageModel>>? _imagesFuture;
  List<ImageModel>? _images;

  @override
  void initState() {
    super.initState();
    _loadImages();
    // Request focus when the view is created
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _focusNode.requestFocus();
    });
  }

  void _loadImages() async {
    _imagesFuture = ApiService.instance.searchImages(widget.searchQuery);
    // Initialize the page controller once we have the images
    final images = await _imagesFuture;
    if (!mounted || images == null) return;

    final initialIndex = images.indexWhere((img) => img.id == widget.imageId);
    if (initialIndex == -1) {
      // Defaults to first image
      _currentIndex = 0;
    } else {
      _currentIndex = initialIndex;
    }
    _pageController = PageController(initialPage: _currentIndex);
    _pageController.addListener(_onPageChanged);
    setState(() {}); // Trigger rebuild with initialized controller
  }

  @override
  void dispose() {
    _pageController.removeListener(_onPageChanged);
    _pageController.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _onPageChanged() {
    if (_currentIndex != _pageController.page?.round()) {
      setState(() {
        _currentIndex = _pageController.page?.round() ?? _currentIndex;
      });
      _precacheNearbyImages(_currentIndex);
    }
  }

  void _precacheNearbyImages(int currentIndex) {
    var images = _images;
    if (images == null) {
      return;
    }
    // Pre-cache 2 images before and after the current one
    for (var i = -2; i <= 2; i++) {
      final targetIndex = currentIndex + i;
      if (targetIndex >= 0 && targetIndex < images.length && i != 0) {
        final imageUrl =
            ApiService.instance.getImageUrl(images[targetIndex].midiPath);
        precacheImage(NetworkImage(imageUrl), context);
      }
    }
  }

  void _showNavigationButtons() {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text(
            'Utilisez les flèches ou faites glisser de gauche à droite pour voir les images.'),
        duration: Duration(seconds: 2),
        behavior: SnackBarBehavior.floating,
      ),
    );
  }

  void _onKeyEvent(KeyEvent event) {
    if (_images == null) {
      return;
    }
    if (event is KeyDownEvent) {
      if (event.logicalKey == LogicalKeyboardKey.escape) {
        Navigator.of(context).pop();
      } else if (event.logicalKey == LogicalKeyboardKey.arrowLeft &&
          _currentIndex > 0) {
        _pageController.previousPage(
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeInOut,
        );
      } else if (event.logicalKey == LogicalKeyboardKey.arrowRight &&
          _currentIndex < _images!.length - 1) {
        _pageController.nextPage(
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeInOut,
        );
      }
    }
  }

  List<Widget> _legend(ImageModel image) {
    return [
      const SizedBox(height: 16),
      Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              Text('${image.width} × ${image.height}'),
              const SizedBox(width: 8),
              SelectableText(
                'ID: ${image.id}',
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: Theme.of(context).colorScheme.secondary,
                    ),
              ),
            ],
          ),
          GestureDetector(
            onTap: () {
              if (widget.onKeywordSearch != null) {
                final date = _formatDate(image.itemTimestamp);
                widget.onKeywordSearch!(date, image.id);
                Navigator.of(context).pop();
              }
            },
            child: Text(
              _formatDate(image.itemTimestamp),
              style: TextStyle(
                color: Theme.of(context).colorScheme.primary,
                decoration: TextDecoration.underline,
              ),
            ),
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
    ];
  }

  Widget _keywords(ImageModel image) {
    return Container(
      padding: const EdgeInsets.all(16),
      color: Theme.of(context).colorScheme.surface,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Wrap(
            spacing: 8,
            runSpacing: 4,
            children: [
              // Album chip first
              GestureDetector(
                onTap: () {
                  if (widget.onKeywordSearch != null) {
                    final albumQuery = 'album:${image.albumDir}';
                    widget.onKeywordSearch!(albumQuery, image.id);
                    Navigator.of(context).pop();
                  }
                },
                child: Chip(
                  avatar: const Icon(Icons.album, size: 18),
                  label: Text(image.albumDir),
                  backgroundColor:
                      Theme.of(context).colorScheme.secondaryContainer,
                  labelStyle: TextStyle(
                    color: Theme.of(context).colorScheme.onSecondaryContainer,
                  ),
                ),
              ),
              // Then all keywords
              if (image.keywords.isNotEmpty)
                ...image.keywords.map((keyword) => GestureDetector(
                      onTap: () {
                        if (widget.onKeywordSearch != null) {
                          // Quote keywords containing spaces
                          final searchQuery =
                              keyword.contains(' ') ? '"$keyword"' : keyword;
                          widget.onKeywordSearch!(searchQuery, image.id);
                          Navigator.of(context).pop();
                        }
                      },
                      child: Chip(
                        label: Text(keyword),
                        backgroundColor:
                            Theme.of(context).colorScheme.primaryContainer,
                        labelStyle: TextStyle(
                          color:
                              Theme.of(context).colorScheme.onPrimaryContainer,
                        ),
                      ),
                    )),
            ],
          ),
          ..._legend(image),
        ],
      ),
    );
  }

  Widget _photo(ImageModel image) {
    return Expanded(
      child: GestureDetector(
        onDoubleTap: () {
          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => PhotoView(
                imageUrl: ApiService.instance.getImageUrl(image.maxiPath),
              ),
            ),
          );
        },
        child: InteractiveViewer(
          maxScale: 5.0,
          child: Hero(
            tag: 'image_${image.id}',
            child: Image.network(
              ApiService.instance.getImageUrl(image.midiPath),
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
    );
  }

  Widget _pageView(List<ImageModel> images) {
    return PageView.builder(
      controller: _pageController,
      itemCount: images.length,
      itemBuilder: (context, index) {
        final image = images[index];
        return Column(
          children: [
            _photo(image),
            _keywords(image),
          ],
        );
      },
    );
  }

  bool _hasPrev() {
    if (!_pageController.hasClients) return false;
    var pos = _pageController.page?.round();
    if (pos == null) return false;
    return pos > 0;
  }

  bool _hasNext(List<ImageModel> images) {
    if (!_pageController.hasClients) return false;
    var pos = _pageController.page?.round();
    if (pos == null) return false;
    return pos < images.length - 1;
  }

  Widget _pageNav(List<ImageModel> images) {
    bool hasPrev = _hasPrev();
    bool hasNext = _hasNext(images);
    return Positioned.fill(
      child: Row(
        children: [
          Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: hasPrev
                  ? () => _pageController.previousPage(
                        duration: const Duration(milliseconds: 300),
                        curve: Curves.easeInOut,
                      )
                  : null,
              child: SizedBox(
                width: 60,
                child: Opacity(
                  opacity: hasPrev ? 0.7 : 0,
                  child: const Icon(
                    Icons.chevron_left,
                    size: 40,
                    color: Colors.white,
                  ),
                ),
              ),
            ),
          ),
          const Spacer(),
          Material(
            color: Colors.transparent,
            child: InkWell(
              onTap: hasNext
                  ? () => _pageController.nextPage(
                        duration: const Duration(milliseconds: 300),
                        curve: Curves.easeInOut,
                      )
                  : null,
              child: SizedBox(
                width: 60,
                child: Opacity(
                  opacity: hasNext ? 0.7 : 0,
                  child: const Icon(
                    Icons.chevron_right,
                    size: 40,
                    color: Colors.white,
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _currentImageView() {
    var images = _images;
    if (images!.isEmpty) {
      return const Center(child: Text('Pas d\'images disponibles'));
    }
    return KeyboardListener(
      focusNode: _focusNode,
      onKeyEvent: _onKeyEvent,
      child: Stack(
        children: [
          _pageView(images),
          _pageNav(images),
        ],
      ),
    );
  }

  AppBar _appBar() {
    return AppBar(
      title: Text(widget.searchQuery),
      actions: [
        IconButton(
          icon: const Icon(Icons.help_outline),
          onPressed: _showNavigationButtons,
        ),
        IconButton(
          icon: const Icon(Icons.download),
          onPressed: () {
            // TODO: Implement download functionality
          },
        ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: _appBar(),
      body: FutureBuilder<List<ImageModel>>(
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
                    child: const Text('Réessayer'),
                  ),
                ],
              ),
            );
          }

          _images = snapshot.data!;
          return _currentImageView();
        },
      ),
    );
  }

  String _formatDate(DateTime date) {
    return DateFormat('yyyy-MM-dd').format(date);
  }
}
