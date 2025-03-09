import 'package:flutter/material.dart';
import 'package:infinite_scroll_pagination/infinite_scroll_pagination.dart';
import '../models/image.dart';
import '../services/api_service.dart';

class FlowView extends StatefulWidget {
  final ApiService apiService;
  final String? initialQuery;

  const FlowView({
    super.key,
    required this.apiService,
    this.initialQuery,
  });

  @override
  State<FlowView> createState() => _FlowViewState();
}

class _FlowViewState extends State<FlowView> {
  static const _pageSize = 20;
  
  final PagingController<int, ImageModel> _pagingController =
      PagingController(firstPageKey: 0);
  
  String _currentQuery = '';
  List<ImageModel> _allImages = [];

  @override
  void initState() {
    super.initState();
    _currentQuery = widget.initialQuery ?? '';
    _pagingController.addPageRequestListener(_fetchPage);
  }

  @override
  void dispose() {
    _pagingController.dispose();
    super.dispose();
  }

  Future<void> _fetchPage(int pageKey) async {
    try {
      if (_allImages.isEmpty || pageKey == 0) {
        final newItems = await widget.apiService.searchImages(_currentQuery);
        _allImages = newItems;
      }

      final startIndex = pageKey * _pageSize;
      final isLastPage = startIndex + _pageSize >= _allImages.length;
      
      if (isLastPage) {
        _pagingController.appendLastPage(
          _allImages.sublist(startIndex),
        );
      } else {
        final nextPageKey = pageKey + 1;
        _pagingController.appendPage(
          _allImages.sublist(startIndex, startIndex + _pageSize),
          nextPageKey,
        );
      }
    } catch (error) {
      _pagingController.error = error;
    }
  }

  @override
  Widget build(BuildContext context) {
    return PagedGridView<int, ImageModel>(
      pagingController: _pagingController,
      builderDelegate: PagedChildBuilderDelegate<ImageModel>(
        itemBuilder: (context, item, index) => GestureDetector(
          onTap: () {
            // TODO: Navigate to image detail view
          },
          child: Hero(
            tag: 'image_${item.id}',
            child: Image.network(
              widget.apiService.getImageUrl(item.miniPath),
              fit: BoxFit.cover,
              errorBuilder: (context, error, stackTrace) {
                return const Center(
                  child: Icon(Icons.error_outline),
                );
              },
            ),
          ),
        ),
        firstPageProgressIndicatorBuilder: (_) => const Center(
          child: CircularProgressIndicator(),
        ),
        noItemsFoundIndicatorBuilder: (_) => const Center(
          child: Text('No images found'),
        ),
        firstPageErrorIndicatorBuilder: (context) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text('Error: ${_pagingController.error}'),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () => _pagingController.refresh(),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
      gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 3,
        childAspectRatio: 1.0,
        mainAxisSpacing: 8.0,
        crossAxisSpacing: 8.0,
      ),
    );
  }
} 