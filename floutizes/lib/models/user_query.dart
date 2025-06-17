class UserQueryModel {
  final String query;
  final DateTime timestamp;
  final String kind;

  UserQueryModel({
    required this.query,
    required this.timestamp,
    required this.kind,
  });

  factory UserQueryModel.fromJson(Map<String, dynamic> json) {
    return UserQueryModel(
      query: json['query'] as String,
      timestamp: DateTime.parse(json['timestamp'] as String),
      kind: json['kind'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'query': query,
      'timestamp': timestamp.toIso8601String(),
      'kind': kind,
    };
  }
}

class AllUserQueriesModel {
  final Map<String, List<UserQueryModel>> users;

  AllUserQueriesModel({required this.users});

  factory AllUserQueriesModel.fromJson(Map<String, dynamic> json) {
    final Map<String, List<UserQueryModel>> users = {};
    
    final usersJson = json['users'] as Map<String, dynamic>;
    for (final entry in usersJson.entries) {
      final username = entry.key;
      final queriesJson = entry.value as List<dynamic>;
      users[username] = queriesJson
          .map((queryJson) => UserQueryModel.fromJson(queryJson as Map<String, dynamic>))
          .toList();
    }
    
    return AllUserQueriesModel(users: users);
  }

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> usersJson = {};
    for (final entry in users.entries) {
      usersJson[entry.key] = entry.value.map((query) => query.toJson()).toList();
    }
    return {'users': usersJson};
  }
}