import 'package:flutter/material.dart';
import 'package:material_symbols_icons/symbols.dart';
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

class MontageGroup {
  final List<MontagedItem> items;
  final String montageUrl;

  MontageGroup(this.items, this.montageUrl);
}

class MontagedImages<T> {
  static const int _imageSize = 360;
  static const int _defaultGroupSize = 8;

  final List<MontagedItem> _items;
  final List<MontageGroup> _montageGroups;
  final int groupSize;

  MontagedImages._({
    required List<MontagedItem> items,
    required this.groupSize,
  })  : _items = items,
        _montageGroups = _createMontageGroups(items, groupSize);

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

  // Group items into groups of size groupSize, the last group may be smaller
  static List<MontageGroup> _createMontageGroups(
      List<MontagedItem> items, int groupSize) {
    final List<MontageGroup> montageGroups = [];
    for (var i = 0; i < items.length; i += groupSize) {
      final end =
          (i + groupSize <= items.length) ? i + groupSize : items.length;
      final groupItems = items.sublist(i, end).toList();
      final url = ApiService.instance.getMontageUrl(
        groupItems.map((item) => item.id).toList(),
        width: _imageSize,
        height: _imageSize,
      );
      montageGroups.add(MontageGroup(groupItems, url));
    }
    return montageGroups;
  }

  Widget buildImage(T item, {BoxFit fit = BoxFit.cover}) {
    final index = _items.indexWhere((i) => i.object == item);

    if (index == -1) {
      throw ArgumentError('Item not found in montage collection');
    }

    final montageGroupIndex = index ~/ groupSize;
    final positionInGroup = index % groupSize;
    final montageGroup = _montageGroups[montageGroupIndex];
    final adaptedItem = _items[index];
    final headers = ApiService.instance.getImageHeaders();

    double dx;
    if (montageGroup.items.length == 1) {
      dx = -1.0;
    } else {
      dx = -1 + 2 * positionInGroup / (montageGroup.items.length - 1);
    }
    return Image.network(
      montageGroup.montageUrl,
      headers: headers,
      fit: fit,
      alignment: Alignment(dx, 0),
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
                  color: isDark ? Colors.grey.shade800 : Colors.grey.shade200,
                ),
        );
      },
      errorBuilder: (context, error, stackTrace) {
        // Fallback to individual image if montage fails
        return Image.network(
          ApiService.instance.getImageUrl(adaptedItem.fallbackPath),
          headers: headers,
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
                      color:
                          isDark ? Colors.grey.shade800 : Colors.grey.shade200,
                    ),
            );
          },
          errorBuilder: (context, error, stackTrace) {
            return const Center(
              child: Icon(Symbols.error_outline),
            );
          },
        );
      },
    );
  }
}
