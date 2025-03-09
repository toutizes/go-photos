class ImageModel {
  final int id;           // Image id
  final String albumDir;  // Album directory
  final String imageName; // Image filename in the album dir
  final DateTime itemTimestamp;  // Image taken timestamp
  final DateTime fileTimestamp;  // Image file timestamp
  final int height;
  final int width;
  final List<String> keywords;
  final StereoInfo? stereo;

  ImageModel({
    required this.id,
    required this.albumDir,
    required this.imageName,
    required this.itemTimestamp,
    required this.fileTimestamp,
    required this.height,
    required this.width,
    required this.keywords,
    this.stereo,
  });

  String get miniPath => '/db/mini/$albumDir/$imageName';
  String get midiPath => '/db/midi/$albumDir/$imageName';
  String get maxiPath => '/db/maxi/$albumDir/$imageName';

  factory ImageModel.fromJson(Map<String, dynamic> json) {
    return ImageModel(
      id: json['Id'] as int,
      albumDir: json['Ad'] as String,
      imageName: json['In'] as String,
      itemTimestamp: DateTime.fromMillisecondsSinceEpoch(json['Its'] * 1000),
      fileTimestamp: DateTime.fromMillisecondsSinceEpoch(json['Fts'] * 1000),
      height: json['H'] as int,
      width: json['W'] as int,
      keywords: List<String>.from(json['Kwd'] ?? []),
      stereo: json['Stereo'] != null ? StereoInfo.fromJson(json['Stereo']) : null,
    );
  }
}

class StereoInfo {
  final double dx;
  final double dy;
  final double anaDx;
  final double anaDy;

  StereoInfo({
    required this.dx,
    required this.dy,
    required this.anaDx,
    required this.anaDy,
  });

  factory StereoInfo.fromJson(Map<String, dynamic> json) {
    return StereoInfo(
      dx: (json['Dx'] as num).toDouble(),
      dy: (json['Dy'] as num).toDouble(),
      anaDx: (json['AnaDx'] as num).toDouble(),
      anaDy: (json['AnaDy'] as num).toDouble(),
    );
  }
} 