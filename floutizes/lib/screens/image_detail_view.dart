import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:intl/intl.dart';
import '../models/image.dart';
import '../services/api_service.dart';
import '../utils/image_download.dart';
import 'package:go_router/go_router.dart';
import '../main.dart';  // Import ImmersiveModeScope from main.dart

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
  bool _isImmersiveMode = false;

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

    // Start pre-caching immediately
    _precacheNearbyImages(_currentIndex);

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

  void _precacheNearbyImages(int currentIndex) async {
    var images = _images;
    if (images == null) {
      return;
    }
    var headers = await ApiService.instance.getImageHeaders();
    if (!mounted) return;

    const int preCacheWidth = 2;
    for (var i = preCacheWidth; i >= -preCacheWidth; i--) {
      final targetIndex = currentIndex + i;
      if (targetIndex >= 0 && targetIndex < images.length && i != 0) {
        final imageUrl =
            ApiService.instance.getImageUrl(images[targetIndex].midiPath);
        precacheImage(NetworkImage(imageUrl, headers: headers), context);
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
      SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text('${image.width} × ${image.height}'),
            const SizedBox(width: 16),
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
                          widget.onKeywordSearch!(keyword, image.id);
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

  void _setImmersiveMode(bool value) {
    setState(() {
      _isImmersiveMode = value;
    });
    ImmersiveModeScope.of(context).value = value;
  }

  Widget _photo(ImageModel image) {
    return GestureDetector(
      onDoubleTap: () {
        _setImmersiveMode(true);
      },
      onTap: _isImmersiveMode ? () {
        _setImmersiveMode(false);
      } : null,
      child: TweenAnimationBuilder(
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeInOut,
        tween: Tween<double>(
          begin: _isImmersiveMode ? 0.9 : 1.0,
          end: _isImmersiveMode ? 1.0 : 0.9,
        ),
        builder: (context, double value, child) {
          return Transform.scale(
            scale: value,
            child: child,
          );
        },
        child: InteractiveViewer(
          maxScale: 5.0,
          child: Hero(
            tag: 'image_${image.id}',
            child: FutureBuilder<Map<String, String>>(
              future: ApiService.instance.getImageHeaders(),
              builder: (context, snapshot) {
                if (!snapshot.hasData) return const CircularProgressIndicator();
                return LayoutBuilder(
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

                    return Container(
                      width: width,
                      height: height,
                      child: Image.network(
                        ApiService.instance.getImageUrl(image.midiPath),
                        headers: snapshot.data,
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
                                : Container(
                                    width: width,
                                    height: height,
                                    color: isDark
                                        ? Colors.grey.shade800
                                        : Colors.grey.shade200,
                                  ),
                          );
                        },
                        errorBuilder: (context, error, stackTrace) {
                          return const Center(
                            child: Icon(Icons.error_outline, size: 48),
                          );
                        },
                      ),
                    );
                  },
                );
              },
            ),
          ),
        ),
      ),
    );
  }

  Widget _pageView(List<ImageModel> images) {
    final mediaQuery = MediaQuery.of(context);
    final isLandscape = mediaQuery.size.width > mediaQuery.size.height;
    final topPadding = mediaQuery.padding.top;
    final bottomPadding = mediaQuery.padding.bottom;

    return PageView.builder(
      controller: _pageController,
      scrollDirection: Axis.horizontal,
      itemCount: images.length,
      itemBuilder: (context, index) {
        final image = images[index];

        if (_isImmersiveMode) {
          return Stack(
            children: [
              AnimatedPositioned(
                duration: const Duration(milliseconds: 300),
                curve: Curves.easeInOut,
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                child: _photo(image),
              ),
            ],
          );
        }

        if (isLandscape) {
          return Row(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Expanded(
                flex: 7,
                child: Padding(
                  padding: EdgeInsets.only(
                    top: kToolbarHeight + topPadding,
                    bottom: kBottomNavigationBarHeight + bottomPadding,
                  ),
                  child: _photo(image),
                ),
              ),
              if (!_isImmersiveMode) Expanded(
                flex: 3,
                child: Padding(
                  padding: EdgeInsets.only(
                    top: kToolbarHeight + topPadding,
                    bottom: kBottomNavigationBarHeight + bottomPadding,
                  ),
                  child: SingleChildScrollView(
                    child: _keywords(image),
                  ),
                ),
              ),
            ],
          );
        } else {
          return Column(
            children: [
              Expanded(
                child: Padding(
                  padding: EdgeInsets.only(
                    top: kToolbarHeight + topPadding,
                    bottom: kBottomNavigationBarHeight + bottomPadding,
                  ),
                  child: _photo(image),
                ),
              ),
              AnimatedSize(
                duration: const Duration(milliseconds: 300),
                child: AnimatedOpacity(
                  duration: const Duration(milliseconds: 300),
                  opacity: _isImmersiveMode ? 0.0 : 1.0,
                  child: Container(
                    height: _isImmersiveMode ? 0 : MediaQuery.of(context).size.height * 0.2,
                    child: SingleChildScrollView(
                      child: _keywords(image),
                    ),
                  ),
                ),
              ),
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

    // Arrows are always visible if navigation is possible, semi-transparent if disabled
    final prevOpacity = hasPrev ? 1.0 : 0.3;
    final nextOpacity = hasNext ? 1.0 : 0.3;

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
            final filename =
                '${image.albumDir}_${image.id}.jpg'.replaceAll('/', '_');
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

  @override
  Widget build(BuildContext context) {
    final mediaQuery = MediaQuery.of(context);
    final topPadding = mediaQuery.padding.top;
    final bottomPadding = mediaQuery.padding.bottom;

    return Scaffold(
      body: Stack(
        children: [
          // Main content
          Positioned.fill(
            child: FutureBuilder<List<ImageModel>>(
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
          ),

          // AppBar
          AnimatedPositioned(
            duration: const Duration(milliseconds: 300),
            curve: Curves.easeInOut,
            top: _isImmersiveMode ? -kToolbarHeight - topPadding : 0,
            left: 0,
            right: 0,
            height: kToolbarHeight + topPadding,
            child: Material(
              elevation: 4,
              child: Container(
                padding: EdgeInsets.only(top: topPadding),
                color: Theme.of(context).appBarTheme.backgroundColor ?? Theme.of(context).colorScheme.surface,
                child: AppBar(
                  backgroundColor: Colors.transparent,
                  elevation: 0,
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
                    if (kIsWeb) _buildDownloadButton(_images?[_currentIndex]),
                    IconButton(
                      icon: const Icon(Icons.help_outline),
                      onPressed: _showNavigationButtons,
                    ),
                  ],
                ),
              ),
            ),
          ),

          // Bottom Navigation Bar from parent
          AnimatedPositioned(
            duration: const Duration(milliseconds: 300),
            curve: Curves.easeInOut,
            bottom: _isImmersiveMode ? -kBottomNavigationBarHeight - bottomPadding : 0,
            left: 0,
            right: 0,
            child: Container(
              color: Theme.of(context).bottomNavigationBarTheme.backgroundColor ?? Theme.of(context).colorScheme.surface,
              child: SafeArea(
                top: false,
                child: SizedBox(
                  height: kBottomNavigationBarHeight,
                  child: const SizedBox.shrink(),  // Placeholder for bottom nav
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    return DateFormat('yyyy-MM-dd').format(date);
  }
}
