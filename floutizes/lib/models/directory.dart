class DirectoryModel {
  final String id;           // Album path (rel_pat in backend)
  final DateTime albumTime;  // Timestamp of first image in album (ats in backend)
  final DateTime directoryTime; // Timestamp of album directory (dts in backend)
  final int imageCount;     // Number of images in album (nimgs in backend)
  final int coverId;        // Cover image id (cov in backend)
  final String coverName;    // Cover image name (CovName in backend)

  DirectoryModel({
    required this.id,
    required this.albumTime,
    required this.directoryTime,
    required this.imageCount,
    required this.coverId,
    required this.coverName,
  });

  factory DirectoryModel.fromJson(Map<String, dynamic> json) {
    return DirectoryModel(
      id: json['Id'] as String,
      albumTime: DateTime.fromMillisecondsSinceEpoch(json['Ats'] * 1000),
      directoryTime: DateTime.fromMillisecondsSinceEpoch(json['Dts'] * 1000),
      imageCount: json['Nimgs'] as int,
      coverId: json['Cov'] as int,
      coverName: json['CovName'] as String,
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
} 