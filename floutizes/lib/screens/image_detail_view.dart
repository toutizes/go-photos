import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:intl/intl.dart';
import '../models/image.dart';
import '../services/api_service.dart';
import '../utils/image_download.dart';
import 'package:go_router/go_router.dart';
import '../main.dart'; // Import ImmersiveModeScope from main.dart

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

/// Custom clipper to show only half of the image
class _HalfImageClipper extends CustomClipper<Rect> {
  final Rect clipRect;

  _HalfImageClipper(this.clipRect);

  @override
  Rect getClip(Size size) => clipRect;

  @override
  bool shouldReclip(_HalfImageClipper oldClipper) =>
      oldClipper.clipRect != clipRect;
}

class _ImageDetailViewState extends State<ImageDetailView>
    with SingleTickerProviderStateMixin {
  late PageController _pageController;
  late int _currentIndex;
  final FocusNode _focusNode = FocusNode();
  Future<List<ImageModel>>? _imagesFuture;
  List<ImageModel>? _images;
  bool _isImmersiveMode = false;
  StereoViewMode _stereoViewMode = StereoViewMode.none;
  // Track the user's preferred stereo view mode
  StereoViewMode _preferredStereoViewMode = StereoViewMode.parallel;

  // For stereo animation
  AnimationController? _stereoAnimationController;
  bool _showLeftImage = true;
  
  // Horizontal offset for fine-tuning the stereo image alignment
  double _stereoHorizontalOffset = 0.0;

  @override
  void initState() {
    super.initState();
    // Initialize the controller immediately with page 0
    _pageController = PageController(initialPage: 0);
    _pageController.addListener(_onPageChanged);

    // Initialize the animation controller for animated stereo mode
    _stereoAnimationController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 300),
    );

    _stereoAnimationController!.addStatusListener((status) {
      if (status == AnimationStatus.completed) {
        if (_stereoViewMode == StereoViewMode.animated && mounted) {
          setState(() {
            _showLeftImage = !_showLeftImage;
          });
          _stereoAnimationController!.reset();
          _stereoAnimationController!.forward();
        }
      }
    });

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

      // Check if the current image has the stereo attribute
      if (_currentIndex >= 0 &&
          _currentIndex < images.length &&
          images[_currentIndex].stereo != null) {
        // Use the preferred stereo view mode (default is parallel)
        _stereoViewMode = _preferredStereoViewMode;

        // Start animation if stereo mode is animated
        if (_stereoViewMode == StereoViewMode.animated) {
          _showLeftImage = true;
          _stereoAnimationController?.reset();
          _stereoAnimationController?.forward();
        }
      } else {
        // Make sure animation is stopped for non-stereo images
        _stereoAnimationController?.stop();
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
    _stereoAnimationController?.dispose();
    _pageController.removeListener(_onPageChanged);
    _pageController.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _onPageChanged() {
    if (_currentIndex != _pageController.page?.round()) {
      setState(() {
        _currentIndex = _pageController.page?.round() ?? _currentIndex;

        // Check if the current image has the stereo attribute
        if (_images != null &&
            _currentIndex >= 0 &&
            _currentIndex < _images!.length &&
            _images![_currentIndex].stereo != null) {
          // Use the preferred stereo view mode
          _stereoViewMode = _preferredStereoViewMode;

          // Restart animation if mode is animated
          if (_stereoViewMode == StereoViewMode.animated) {
            _showLeftImage = true;
            _stereoAnimationController?.reset();
            _stereoAnimationController?.forward();
          }
        } else {
          // Reset stereo view mode for non-stereo images
          _stereoViewMode = StereoViewMode.none;
          _stereoAnimationController?.stop();
        }
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
          'Image Stéréo',
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

  void _setStereoViewMode(StereoViewMode mode) {
    setState(() {
      if (_stereoViewMode == mode) {
        _stereoViewMode = StereoViewMode.none;
        _stereoAnimationController?.stop();
      } else {
        _stereoViewMode = mode;
        _preferredStereoViewMode = mode;

        if (mode == StereoViewMode.animated) {
          _showLeftImage = true;
          _stereoAnimationController?.reset();
          _stereoAnimationController?.forward();
        } else {
          _stereoAnimationController?.stop();
        }
      }
    });
  }

  Widget _photo(ImageModel image) {
    return GestureDetector(
      onDoubleTap: () {
        _setImmersiveMode(true);
      },
      onTap: _isImmersiveMode
          ? () {
              _setImmersiveMode(false);
            }
          : null,
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
        child: Column(
          children: [
            if (image.stereo != null) _buildStereoControls(image),
            Expanded(
              child: InteractiveViewer(
                maxScale: 5.0,
                child: Hero(
                  tag: 'image_${image.id}',
                  child: FutureBuilder<Map<String, String>>(
                    future: ApiService.instance.getImageHeaders(),
                    builder: (context, snapshot) {
                      if (!snapshot.hasData)
                        return const CircularProgressIndicator();
                      return LayoutBuilder(
                        builder: (context, constraints) {
                          final aspectRatio = image.width / image.height;
                          double width, height;

                          if (constraints.maxWidth / constraints.maxHeight >
                              aspectRatio) {
                            // Height constrained
                            height = constraints.maxHeight;
                            width = height * aspectRatio;
                          } else {
                            // Width constrained
                            width = constraints.maxWidth;
                            height = width / aspectRatio;
                          }

                          // Check if this is a stereo image with an active stereo view mode
                          if (image.stereo != null &&
                              _stereoViewMode != StereoViewMode.none) {
                            // For stereo images, display the image based on the selected view mode
                            return _buildStereoImageView(
                              image: image,
                              width: width,
                              height: height,
                              headers: snapshot.data,
                            );
                          }

                          return SizedBox(
                            width: width,
                            height: height,
                            child: Image.network(
                              ApiService.instance.getImageUrl(image.midiPath),
                              headers: snapshot.data,
                              fit: BoxFit.contain,
                              frameBuilder: (context, child, frame,
                                  wasSynchronouslyLoaded) {
                                if (wasSynchronouslyLoaded) {
                                  return child;
                                }
                                final isDark = Theme.of(context).brightness ==
                                    Brightness.dark;
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
          ],
        ),
      ),
    );
  }

  /// Utility method to create an image that shows either left or right half
  Widget _imageHalf({
    required String imageUrl,
    required double width,
    required double height,
    required Map<String, String>? headers,
    required bool showLeftSide,
    double horizontalOffset = 0, // Horizontal offset for fine-tuning image alignment
  }) {
    // Calculate the alignment based on side and offset
    // Alignment is in range -1 to 1, where -1 is far left, 1 is far right
    // Convert offset to alignment scale (divide by width to get a value between -1 and 1)
    final double offsetAlignment = horizontalOffset / width;
    final double baseAlignment = showLeftSide ? -1.0 : 1.0; // Left or right edge
    final Alignment alignment = Alignment(baseAlignment + offsetAlignment, 0);
    
    return SizedBox(
      width: width,
      height: height,
      // We make the image double the width to show only half of the image in the sizedbox
      child: Image.network(
        imageUrl,
        headers: headers,
        fit: BoxFit.cover,
        alignment: alignment,
        width: width * 2,
        height: height,
      ),
    );
  }

  Widget _buildStereoImageView({
    required ImageModel image,
    required double width,
    required double height,
    required Map<String, String>? headers,
  }) {
    final imageUrl = ApiService.instance.getImageUrl(image.midiPath);

    // For animated mode in fullscreen, we use a slightly different approach
    if (_stereoViewMode == StereoViewMode.animated) {
      // Display either left half or right half based on animation state
      return Column(
        children: [
          SizedBox(
            width: width / 2,
            child: _imageHalf(
              imageUrl: imageUrl,
              width: width,
              height: height * 0.9, // Make room for slider
              headers: headers,
              showLeftSide: _showLeftImage,
              horizontalOffset: _showLeftImage ? 0 : _stereoHorizontalOffset, // Apply offset only to right image
            ),
          ),
          // Add a slider to control the horizontal offset
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Text('Alignement:'),
                Expanded(
                  child: Slider(
                    value: _stereoHorizontalOffset,
                    min: 0, // Only positive values
                    max: width / 2,
                    divisions: 100, // Reduced divisions since range is halved
                    label: _stereoHorizontalOffset.toStringAsFixed(0),
                    onChanged: (value) {
                      setState(() {
                        _stereoHorizontalOffset = value;
                      });
                    },
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.restart_alt),
                  onPressed: () {
                    setState(() {
                      _stereoHorizontalOffset = 0;
                    });
                  },
                  tooltip: 'Réinitialiser le décalage',
                ),
              ],
            ),
          ),
        ],
      );
    }

    // For parallel and cross-eyed modes, determine which side goes where
    bool leftSideOnLeft, rightSideOnRight;
    if (_stereoViewMode == StereoViewMode.parallel) {
      leftSideOnLeft = true; // Left eye sees left half
      rightSideOnRight = true; // Right eye sees right half
    } else {
      // StereoViewMode.crossEyed
      leftSideOnLeft = false; // Left eye sees right half
      rightSideOnRight = false; // Right eye sees left half
    }

    // For other modes, we load the image twice side by side
    // Use a fixed size container to prevent overflow
    return SizedBox(
      width: width,
      height: height,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          // Left eye image (always on the left side of screen)
          SizedBox(
            width: width / 2,
            child: _imageHalf(
              imageUrl: imageUrl,
              width: width / 2,
              height: height,
              headers: headers,
              showLeftSide: leftSideOnLeft, // Determined by view mode
              horizontalOffset: 0, // No offset for parallel/cross-eyed modes
            ),
          ),
          // Right eye image (always on the right side of screen)
          SizedBox(
            width: width / 2,
            child: _imageHalf(
              imageUrl: imageUrl,
              width: width / 2,
              height: height,
              headers: headers,
              showLeftSide: !rightSideOnRight, // Determined by view mode
              horizontalOffset: 0, // No offset for parallel/cross-eyed modes
            ),
          ),
        ],
      ),
    );
  }

  /// Builds the stereo control buttons for switching between parallel and
  /// cross-eyed viewing modes
  Widget _buildStereoControls(ImageModel image) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Tooltip(
            message:
                'Pour image stéréo - regardez avec les yeux parallèles (fixez l\'horizon)',
            child: ElevatedButton.icon(
              icon: const Icon(Icons.arrow_upward_rounded),
              label: const Text('Vue Parallèle'),
              style: ElevatedButton.styleFrom(
                backgroundColor: _stereoViewMode == StereoViewMode.parallel
                    ? Theme.of(context).colorScheme.primaryContainer
                    : null,
                foregroundColor: _stereoViewMode == StereoViewMode.parallel
                    ? Theme.of(context).colorScheme.onPrimaryContainer
                    : null,
              ),
              onPressed: () {
                _setStereoViewMode(StereoViewMode.parallel);
              },
            ),
          ),
          const SizedBox(width: 16),
          Tooltip(
            message:
                'Pour image stéréo - utilisez la technique des yeux croisés (focus devant l\'image)',
            child: ElevatedButton.icon(
              icon: const Icon(Icons.compare_arrows_rounded),
              label: const Text('Vue Croisée'),
              style: ElevatedButton.styleFrom(
                backgroundColor: _stereoViewMode == StereoViewMode.crossEyed
                    ? Theme.of(context).colorScheme.primaryContainer
                    : null,
                foregroundColor: _stereoViewMode == StereoViewMode.crossEyed
                    ? Theme.of(context).colorScheme.onPrimaryContainer
                    : null,
              ),
              onPressed: () {
                _setStereoViewMode(StereoViewMode.crossEyed);
              },
            ),
          ),
          const SizedBox(width: 16),
          Tooltip(
            message:
                'Pour image stéréo - alternance automatique des vues gauche et droite',
            child: ElevatedButton.icon(
              icon: const Icon(Icons.animation_rounded),
              label: const Text('Vue Animée'),
              style: ElevatedButton.styleFrom(
                backgroundColor: _stereoViewMode == StereoViewMode.animated
                    ? Theme.of(context).colorScheme.primaryContainer
                    : null,
                foregroundColor: _stereoViewMode == StereoViewMode.animated
                    ? Theme.of(context).colorScheme.onPrimaryContainer
                    : null,
              ),
              onPressed: () {
                _setStereoViewMode(StereoViewMode.animated);
              },
            ),
          ),
        ],
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
              if (!_isImmersiveMode)
                Expanded(
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
                    height: _isImmersiveMode
                        ? 0
                        : MediaQuery.of(context).size.height * 0.2,
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
                color: Theme.of(context).appBarTheme.backgroundColor ??
                    Theme.of(context).colorScheme.surface,
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
            bottom: _isImmersiveMode
                ? -kBottomNavigationBarHeight - bottomPadding
                : 0,
            left: 0,
            right: 0,
            child: Container(
              color:
                  Theme.of(context).bottomNavigationBarTheme.backgroundColor ??
                      Theme.of(context).colorScheme.surface,
              child: SafeArea(
                top: false,
                child: SizedBox(
                  height: kBottomNavigationBarHeight,
                  child: const SizedBox.shrink(), // Placeholder for bottom nav
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

enum StereoViewMode {
  none,
  parallel,
  crossEyed,
  animated,
}
