import 'package:flutter/material.dart';
import '../models/image.dart';
import '../models/directory.dart';
import '../services/api_service.dart';

abstract class MontagedItem {
  int get id;
  String get fallbackPath;
  Object get object;
}

class ImageModelAdapter implements MontagedItem {
  final ImageModel image;

  ImageModelAdapter(this.image);

  @override
  int get id => image.id;

  @override
  String get fallbackPath => image.miniPath;

  @override
  Object get object => image;
}

class DirectoryModelAdapter implements MontagedItem {
  final DirectoryModel directory;

  DirectoryModelAdapter(this.directory);

  @override
  int get id => directory.coverId;

  @override
  String get fallbackPath => directory.coverMiniPath;

  @override
  Object get object => directory;
}

class MontagedImages<T> {
  static const int _imageSize = 300;
  static const int _defaultGroupSize = 8;

  final List<MontagedItem> _items;
  final List<String> _montageUrls;
  final int groupSize;

  MontagedImages._({
    required List<MontagedItem> items,
    required this.groupSize,
  })  : _items = items,
        _montageUrls = _createMontageUrls(items, groupSize);

  factory MontagedImages.fromImageModels(List<ImageModel> images,
      {int groupSize = _defaultGroupSize}) {
    return MontagedImages._(
      items: images.map((img) => ImageModelAdapter(img)).toList(),
      groupSize: groupSize,
    );
  }

  factory MontagedImages.fromDirectoryModels(List<DirectoryModel> directories,
      {int groupSize = _defaultGroupSize}) {
    return MontagedImages._(
      items: directories.map((dir) => DirectoryModelAdapter(dir)).toList(),
      groupSize: groupSize,
    );
  }

  static List<String> _createMontageUrls(
      List<MontagedItem> items, int groupSize) {
    final List<List<int>> montageGroups = [];

    // Group items into groups of size groupSize
    for (var i = 0; i < items.length; i += groupSize) {
      final end =
          (i + groupSize <= items.length) ? i + groupSize : items.length;
      final group = items.sublist(i, end).map((item) => item.id).toList();

      // Pad the last group with repeated last ID if needed
      if (group.length < groupSize && group.isNotEmpty) {
        final lastId = group.last;
        while (group.length < groupSize) {
          group.add(lastId);
        }
      }

      montageGroups.add(group);
    }

    // Create montage URLs for each group
    return montageGroups
        .map((group) => ApiService.instance.getMontageUrl(
              group,
              width: _imageSize,
              height: _imageSize,
            ))
        .toList();
  }

  Widget buildImage(T item) {
    final index = _items.indexWhere((i) => i.object == item);

    if (index == -1) {
      throw ArgumentError('Item not found in montage collection');
    }

    final montageGroupIndex = index ~/ groupSize;
    final positionInGroup = index % groupSize;
    final montageUrl = _montageUrls[montageGroupIndex];
    final adaptedItem = _items[index];

    return FutureBuilder<Map<String, String>>(
      future: ApiService.instance.getImageHeaders(),
      builder: (context, snapshot) {
        if (!snapshot.hasData) return const CircularProgressIndicator();

        return Image.network(
          montageUrl,
          headers: snapshot.data,
          fit: BoxFit.none,
          alignment: Alignment(-1 + 2 * positionInGroup / (groupSize - 1), 0),
          width: _imageSize.toDouble(),
          height: _imageSize.toDouble(),
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
                      width: _imageSize.toDouble(),
                      height: _imageSize.toDouble(),
                      color:
                          isDark ? Colors.grey.shade800 : Colors.grey.shade200,
                    ),
            );
          },
          errorBuilder: (context, error, stackTrace) {
            // Fallback to individual image if montage fails
            return Image.network(
              ApiService.instance.getImageUrl(adaptedItem.fallbackPath),
              headers: snapshot.data,
              fit: BoxFit.cover,
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
                          width: _imageSize.toDouble(),
                          height: _imageSize.toDouble(),
                          color: isDark
                              ? Colors.grey.shade800
                              : Colors.grey.shade200,
                        ),
                );
              },
              errorBuilder: (context, error, stackTrace) {
                return const Center(
                  child: Icon(Icons.error_outline),
                );
              },
            );
          },
        );
      },
    );
  }
}
