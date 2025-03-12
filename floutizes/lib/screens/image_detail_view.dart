import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:intl/intl.dart';
import '../models/image.dart';
import '../services/api_service.dart';
import '../utils/image_download.dart';
import 'photo_view.dart';
import 'package:go_router/go_router.dart';

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
    // Initialize the controller immediately with page 0
    _pageController = PageController(initialPage: 0);
    _pageController.addListener(_onPageChanged);
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

    setState(() {
      _images = images;
      final initialIndex = images.indexWhere((img) => img.id == widget.imageId);
      if (initialIndex == -1) {
        // Defaults to first image
        _currentIndex = 0;
      } else {
        _currentIndex = initialIndex;
      }
    });

    // Wait for the PageView to be built and controller to be attached
    if (_currentIndex > 0) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (mounted && _pageController.hasClients) {
          _pageController.jumpToPage(_currentIndex);
        }
      });
    }
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

      // Update the browser URL without triggering navigation
      if (_images != null &&
          _currentIndex >= 0 &&
          _currentIndex < _images!.length) {
        final currentImage = _images![_currentIndex];
        context.go(
            '/images/details/${currentImage.id}?q=${Uri.encodeComponent(widget.searchQuery)}');
      }
    }
  }

  void _precacheNearbyImages(int currentIndex) {
    var images = _images;
    if (images == null) {
      return;
    }
    // Pre-cache images before and after the current one
    const int preCacheWidth = 3;
    for (var i = -preCacheWidth; i <= preCacheWidth; i++) {
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
            'Utilisez les flèches ou faites glisser de gauche à droite ou de haut en bas pour voir les images.'),
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
      } else if ((event.logicalKey == LogicalKeyboardKey.arrowUp ||
              event.logicalKey == LogicalKeyboardKey.arrowLeft) &&
          _currentIndex > 0) {
        _pageController.previousPage(
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeInOut,
        );
      } else if ((event.logicalKey == LogicalKeyboardKey.arrowDown ||
              event.logicalKey == LogicalKeyboardKey.arrowRight) &&
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
          Text('${image.width} × ${image.height}'),
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
    return GestureDetector(
      onDoubleTap: () {
        Navigator.of(context).push(
          MaterialPageRoute(
            builder: (context) => PhotoView(
              imageUrl: ApiService.instance.getImageUrl(image.midiPath),
              width: image.width,
              height: image.height,
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
            frameBuilder: (context, child, frame, wasSynchronouslyLoaded) {
              if (wasSynchronouslyLoaded) {
                return child;
              }
              final isDark = Theme.of(context).brightness == Brightness.dark;
              return AnimatedSwitcher(
                duration: const Duration(milliseconds: 200),
                child: frame != null 
                    ? child 
                    : LayoutBuilder(
                        builder: (context, constraints) {
                          final aspectRatio = image.width / image.height;
                          double width, height;
                          
                          if (constraints.maxWidth / constraints.maxHeight > aspectRatio) {
                            // Height constrained
                            height = constraints.maxHeight;
                            width = height * aspectRatio;
                          } else {
                            // Width constrained
                            width = constraints.maxWidth;
                            height = width / aspectRatio;
                          }
                          
                          return Center(
                            child: Container(
                              width: width,
                              height: height,
                              color: isDark ? Colors.grey.shade800 : Colors.grey.shade200,
                            ),
                          );
                        },
                      ),
              );
            },
            errorBuilder: (context, error, stackTrace) {
              return const Center(
                child: Icon(Icons.error_outline, size: 48),
              );
            },
          ),
        ),
      ),
    );
  }

  Widget _pageView(List<ImageModel> images) {
    // On web, always scroll horizontally.
    // On mobile, scroll based on orientation
    final isLandscape =
        MediaQuery.of(context).size.width > MediaQuery.of(context).size.height;

    return PageView.builder(
      controller: _pageController,
      scrollDirection:
          (isLandscape && !kIsWeb) ? Axis.vertical : Axis.horizontal,
      itemCount: images.length,
      itemBuilder: (context, index) {
        final image = images[index];

        if (isLandscape) {
          // Landscape layout: image on the left, info on the right
          return Row(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Image takes 70% of the width in landscape
              Expanded(
                flex: 7, // 70% of the space
                child: _photo(image),
              ),
              // Keywords and info take 30% of the width
              Expanded(
                flex: 3, // 30% of the space
                child: SingleChildScrollView(
                  child: _keywords(image),
                ),
              ),
            ],
          );
        } else {
          // Portrait layout: image on top, info below
          return Column(
            children: [
              Expanded(
                child: _photo(image),
              ),
              _keywords(image),
            ],
          );
        }
      },
    );
  }

  bool _hasPrev() {
    // If controller is attached, use it
    if (_pageController.hasClients) {
      var pos = _pageController.page?.round();
      if (pos != null) return pos > 0;
    }
    // Otherwise use our known state
    return _currentIndex > 0;
  }

  bool _hasNext(List<ImageModel> images) {
    // If controller is attached, use it
    if (_pageController.hasClients) {
      var pos = _pageController.page?.round();
      if (pos != null) return pos < images.length - 1;
    }
    // Otherwise use our known state
    return _currentIndex < images.length - 1;
  }

  void _prevPage() {
    if (_currentIndex > 0) {
      if (_pageController.hasClients) {
        _pageController.previousPage(
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeInOut,
        );
      } else {
        setState(() {
          _currentIndex--;
          _pageController.jumpToPage(_currentIndex);
        });
      }
    }
  }

  void _nextPage() {
    if (_images != null && _currentIndex < _images!.length - 1) {
      if (_pageController.hasClients) {
        _pageController.nextPage(
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeInOut,
        );
      } else {
        setState(() {
          _currentIndex++;
          _pageController.jumpToPage(_currentIndex);
        });
      }
    }
  }

  Widget _pageNav(List<ImageModel> images) {
    bool hasPrev = _hasPrev();
    bool hasNext = _hasNext(images);
    // On web, always use horizontal navigation
    // On mobile, use navigation based on orientation
    final isLandscape = !kIsWeb &&
        MediaQuery.of(context).size.width > MediaQuery.of(context).size.height;

    // On web, arrows are always visible if navigation is possible
    // On mobile, arrows fade in/out
    final prevOpacity = kIsWeb ? (hasPrev ? 1.0 : 0.3) : (hasPrev ? 0.7 : 0.0);
    final nextOpacity = kIsWeb ? (hasNext ? 1.0 : 0.3) : (hasNext ? 0.7 : 0.0);

    if (isLandscape) {
      // Vertical navigation for landscape mode on mobile
      return Positioned.fill(
        child: Column(
          children: [
            Material(
              color: Colors.transparent,
              child: InkWell(
                onTap: hasPrev ? _prevPage : null,
                child: SizedBox(
                  height: 60,
                  child: Opacity(
                    opacity: prevOpacity,
                    child: const Icon(
                      Icons.keyboard_arrow_up,
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
                onTap: hasNext ? _nextPage : null,
                child: SizedBox(
                  height: 60,
                  child: Opacity(
                    opacity: nextOpacity,
                    child: const Icon(
                      Icons.keyboard_arrow_down,
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
    } else {
      // Horizontal navigation for portrait mode and web
      return Positioned.fill(
        child: Row(
          children: [
            Material(
              color: Colors.transparent,
              child: InkWell(
                onTap: hasPrev ? _prevPage : null,
                child: SizedBox(
                  width: 60,
                  child: Opacity(
                    opacity: prevOpacity,
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
                onTap: hasNext ? _nextPage : null,
                child: SizedBox(
                  width: 60,
                  child: Opacity(
                    opacity: nextOpacity,
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

  Widget _buildDownloadButton(ImageModel? image) {
    final bool isEnabled = image != null;
    
    return PopupMenuButton<String>(
      icon: Icon(
        Icons.download,
        color: isEnabled ? null : Theme.of(context).disabledColor,
      ),
      tooltip: 'Télécharger',
      enabled: isEnabled,
      itemBuilder: (context) => [
        const PopupMenuItem(
          value: 'current_high',
          child: Text('Cette image (haute qualité)'),
        ),
        const PopupMenuItem(
          value: 'all_medium',
          child: Text('Toutes les images (qualité moyenne)'),
        ),
        const PopupMenuItem(
          value: 'all_high',
          child: Text('Toutes les images (haute qualité)'),
        ),
      ],
      onSelected: (value) {
        if (image == null) return;
        
        switch (value) {
          case 'current_high':
            // Download current image in high quality
            final url = ApiService.instance.getImageUrl(image.maxiPath);
            final filename = '${image.albumDir}_${image.id}.jpg'.replaceAll('/', '_');
            ImageDownload.downloadSingleImage(
              url: url,
              filename: filename,
            );
            break;
          case 'all_medium':
            // Download all images in medium quality
            ImageDownload.start(
              context,
              widget.searchQuery,
              DownloadQuality.medium,
            );
            break;
          case 'all_high':
            // Download all images in high quality
            ImageDownload.start(
              context,
              widget.searchQuery,
              DownloadQuality.high,
            );
            break;
        }
      },
    );
  }

  AppBar _appBar() {
    final currentImage = _images != null && _currentIndex >= 0 && _currentIndex < _images!.length 
        ? _images![_currentIndex]
        : null;

    return AppBar(
      title: Row(
        children: [
          Expanded(
            child: Text(widget.searchQuery),
          ),
          if (_images != null)
            Text(
              '${_currentIndex + 1}/${_images!.length}',
              style: Theme.of(context).textTheme.titleMedium,
            ),
        ],
      ),
      actions: [
        if (kIsWeb) _buildDownloadButton(currentImage),
        IconButton(
          icon: const Icon(Icons.help_outline),
          onPressed: _showNavigationButtons,
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
