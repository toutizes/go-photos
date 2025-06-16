class DirectoryModel {
  // Album path (rel_pat in backend)
  final String id;
  // Timestamp of first image in album (ats in backend)
  final DateTime albumTime;
  // Timestamp of album directory (dts in backend)
  final DateTime directoryTime;
  // Number of images in album (nimgs in backend)
  final int imageCount;
  // Cover image id (cov in backend)
  final int coverId;
  // Cover image name (CovName in backend)
  final String coverName;
  // First 4 photo IDs for previews (PreviewIds in backend)
  final List<int> previewIds;
  // First 4 photo names for previews (PreviewNames in backend)
  final List<String> previewNames;

  DirectoryModel({
    required this.id,
    required this.albumTime,
    required this.directoryTime,
    required this.imageCount,
    required this.coverId,
    required this.coverName,
    required this.previewIds,
    required this.previewNames,
  });

  factory DirectoryModel.fromJson(Map<String, dynamic> json) {
    return DirectoryModel(
      id: json['Id'] as String,
      albumTime: DateTime.fromMillisecondsSinceEpoch(json['Ats'] * 1000),
      directoryTime: DateTime.fromMillisecondsSinceEpoch(json['Dts'] * 1000),
      imageCount: json['Nimgs'] as int,
      coverId: json['Cov'] as int,
      coverName: json['CovName'] as String,
      previewIds: List<int>.from(json['PreviewIds'] ?? []),
      previewNames: List<String>.from(json['PreviewNames'] ?? []),
    );
  }

  // Helper methods to get paths for this album
  String get miniPath => '/db/mini/$id';
  String get midiPath => '/db/midi/$id';
  String get maxiPath => '/db/maxi/$id';

  // Get the path for the cover image - using the first image in the album
  String get coverMiniPath => '$miniPath/$coverName';
  String get coverMidiPath => '$midiPath/$coverName';
  String get coverMaxiPath => '$maxiPath/$coverName';

  // Get preview image paths
  List<String> get previewMiniPaths => previewNames.map((name) => '$miniPath/$name').toList();
  List<String> get previewMidiPaths => previewNames.map((name) => '$midiPath/$name').toList();
}
