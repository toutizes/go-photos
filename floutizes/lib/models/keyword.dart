class KeywordImageInfo {
  final int id;
  final String name;

  KeywordImageInfo({
    required this.id,
    required this.name,
  });

  factory KeywordImageInfo.fromJson(Map<String, dynamic> json) {
    return KeywordImageInfo(
      id: json['id'] as int,
      name: json['name'] as String,
    );
  }
}

class KeywordModel {
  final String keyword;
  final int count;
  final List<KeywordImageInfo> recentImages;

  KeywordModel({
    required this.keyword,
    required this.count,
    required this.recentImages,
  });

  factory KeywordModel.fromJson(Map<String, dynamic> json) {
    final recentImagesJson = json['recent_images'] as List<dynamic>? ?? [];
    final recentImages = recentImagesJson
        .map((imageJson) => KeywordImageInfo.fromJson(imageJson as Map<String, dynamic>))
        .toList();

    return KeywordModel(
      keyword: json['keyword'] as String,
      count: json['count'] as int,
      recentImages: recentImages,
    );
  }
}

class KeywordGroupModel {
  final List<KeywordModel> keywords;
  final List<KeywordImageInfo> recentImages;
  final double totalWeight;
  final int totalCount;

  KeywordGroupModel({
    required this.keywords,
    required this.recentImages,
    required this.totalWeight,
    required this.totalCount,
  });

  factory KeywordGroupModel.fromJson(Map<String, dynamic> json) {
    final keywordsJson = json['keywords'] as List<dynamic>? ?? [];
    final keywords = keywordsJson
        .map((keywordJson) => KeywordModel.fromJson(keywordJson as Map<String, dynamic>))
        .toList();

    final recentImagesJson = json['recent_images'] as List<dynamic>? ?? [];
    final recentImages = recentImagesJson
        .map((imageJson) => KeywordImageInfo.fromJson(imageJson as Map<String, dynamic>))
        .toList();

    return KeywordGroupModel(
      keywords: keywords,
      recentImages: recentImages,
      totalWeight: (json['total_weight'] as num?)?.toDouble() ?? 0.0,
      totalCount: json['total_count'] as int? ?? 0,
    );
  }
}