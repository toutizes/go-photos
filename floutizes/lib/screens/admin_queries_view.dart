import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:material_symbols_icons/symbols.dart';
import 'package:go_router/go_router.dart';
import '../models/user_query.dart';
import '../services/api_service.dart';
import '../services/auth_service.dart';

class AdminQueriesView extends StatefulWidget {
  const AdminQueriesView({super.key});

  @override
  State<AdminQueriesView> createState() => _AdminQueriesViewState();
}

class _AdminQueriesViewState extends State<AdminQueriesView> {
  Future<AllUserQueriesModel>? _userQueriesFuture;
  bool _isLoading = false;
  bool _hasInitialized = false;

  @override
  void initState() {
    super.initState();
    // Will be initialized in didChangeDependencies when context is available
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (!_hasInitialized && _isAdmin()) {
      _hasInitialized = true;
      _loadUserQueries();
    }
  }

  bool _isAdmin() {
    final authService = Provider.of<AuthService>(context, listen: false);
    return authService.user?.email == 'matthieu.devin@gmail.com';
  }

  Future<void> _loadUserQueries() async {
    if (!_isAdmin()) {
      return; // Don't make API call if not admin
    }

    setState(() {
      _isLoading = true;
    });

    try {
      _userQueriesFuture = ApiService.instance.getUserQueries();
      await _userQueriesFuture;
    } finally {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    }
  }

  Widget _buildUserQueriesCard(String username, List<UserQueryModel> queries) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: ExpansionTile(
        title: Text(
          username,
          style: Theme.of(context).textTheme.titleMedium?.copyWith(
            fontWeight: FontWeight.bold,
          ),
        ),
        subtitle: Text('${queries.length} queries (max 30, deduped)'),
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'All queries (${queries.length}):',
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 12),
                ...queries.map((query) => _buildQueryItem(query)),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildQueryItem(UserQueryModel query) {
    final dateFormat = DateFormat('MM/dd HH:mm');
    
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 80,
            child: Text(
              dateFormat.format(query.timestamp),
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                color: Theme.of(context).colorScheme.onSurfaceVariant,
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                GestureDetector(
                  onTap: () {
                    // Navigate to photos view with this query
                    context.go('/images?q=${Uri.encodeComponent(query.query)}');
                  },
                  child: Text(
                    query.query,
                    style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: Theme.of(context).colorScheme.primary,
                      decoration: TextDecoration.underline,
                    ),
                  ),
                ),
                if (query.kind.isNotEmpty)
                  Text(
                    'Kind: ${query.kind}',
                    style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: Theme.of(context).colorScheme.onSurfaceVariant,
                      fontStyle: FontStyle.italic,
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    // Check if user is admin
    if (!_isAdmin()) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Symbols.admin_panel_settings,
              size: 48,
              color: Colors.grey,
            ),
            SizedBox(height: 16),
            Text('Accès administrateur requis'),
          ],
        ),
      );
    }

    // Show loading indicator if still loading or future not yet set
    if (_isLoading || _userQueriesFuture == null) {
      return const Center(child: CircularProgressIndicator());
    }

    return FutureBuilder<AllUserQueriesModel>(
      future: _userQueriesFuture,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Center(child: CircularProgressIndicator());
        }

        if (snapshot.hasError) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(
                  Symbols.error_outline,
                  size: 48,
                  color: Colors.grey,
                ),
                const SizedBox(height: 16),
                Text('Erreur requêtes utilisateur: ${snapshot.error}'),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: _loadUserQueries,
                  child: const Text('Réessayer'),
                ),
              ],
            ),
          );
        }

        final userQueries = snapshot.data!;
        
        if (userQueries.users.isEmpty) {
          return const Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  Symbols.search,
                  size: 48,
                  color: Colors.grey,
                ),
                SizedBox(height: 16),
                Text('Aucune requête utilisateur trouvée'),
              ],
            ),
          );
        }

        // Sort users by most recent query timestamp (descending)
        final sortedUsers = userQueries.users.entries.toList()
          ..sort((a, b) {
            // Get the most recent timestamp for each user (first query is most recent)
            final aTimestamp = a.value.isNotEmpty ? a.value.first.timestamp : DateTime(1970);
            final bTimestamp = b.value.isNotEmpty ? b.value.first.timestamp : DateTime(1970);
            return bTimestamp.compareTo(aTimestamp);
          });

        return RefreshIndicator(
          onRefresh: _loadUserQueries,
          child: ListView.builder(
            padding: const EdgeInsets.symmetric(vertical: 8),
            itemCount: sortedUsers.length,
            itemBuilder: (context, index) {
              final userEntry = sortedUsers[index];
              return _buildUserQueriesCard(userEntry.key, userEntry.value);
            },
          ),
        );
      },
    );
  }
}