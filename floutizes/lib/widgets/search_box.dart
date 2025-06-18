import 'package:flutter/material.dart';
import 'package:material_symbols_icons/symbols.dart';

class SearchBox extends StatefulWidget {
  final TextEditingController controller;
  final String hintText;
  final VoidCallback? onSearch;
  final VoidCallback? onClear;
  final VoidCallback? onHelp;
  final VoidCallback? onLogout;
  final bool showHelpButton;
  final bool showLogoutButton;

  const SearchBox({
    super.key,
    required this.controller,
    required this.hintText,
    this.onSearch,
    this.onClear,
    this.onHelp,
    this.onLogout,
    this.showHelpButton = true,
    this.showLogoutButton = true,
  });

  @override
  State<SearchBox> createState() => _SearchBoxState();
}

class _SearchBoxState extends State<SearchBox> {

  @override
  void initState() {
    super.initState();
    widget.controller.addListener(() {
      setState(() {}); // Rebuild when text changes
    });
  }

  Widget _buildSuffixIcons(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (widget.controller.text.isNotEmpty && widget.onClear != null)
          IconButton(
            icon: Icon(
              Symbols.clear,
              color: Theme.of(context).colorScheme.onSurfaceVariant,
            ),
            onPressed: widget.onClear,
            tooltip: 'Effacer la recherche',
          ),
        if (widget.onSearch != null)
          IconButton(
            icon: Icon(
              Symbols.search,
              color: Theme.of(context).colorScheme.primary,
            ),
            onPressed: widget.onSearch,
            tooltip: 'Rechercher',
          ),
        if (widget.showHelpButton && widget.onHelp != null)
          IconButton(
            icon: Icon(
              Symbols.help_outline,
              color: Theme.of(context).colorScheme.onSurfaceVariant,
            ),
            onPressed: widget.onHelp,
            tooltip: 'Aide recherche',
          ),
        if (widget.showLogoutButton && widget.onLogout != null)
          IconButton(
            icon: Icon(
              Symbols.logout,
              color: Theme.of(context).colorScheme.onSurfaceVariant,
            ),
            onPressed: widget.onLogout,
            tooltip: 'Déconnexion',
          ),
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      height: 48,
      decoration: BoxDecoration(
        color: Theme.of(context).colorScheme.surfaceVariant.withOpacity(0.3),
        borderRadius: BorderRadius.circular(24),
        border: Border.all(
          color: Theme.of(context).colorScheme.outline.withOpacity(0.2),
          width: 1,
        ),
      ),
      child: TextField(
        controller: widget.controller,
        decoration: InputDecoration(
          hintText: widget.hintText,
          hintStyle: TextStyle(
            color: Theme.of(context).colorScheme.onSurfaceVariant,
            fontSize: 16,
          ),
          suffixIcon: _buildSuffixIcons(context),
          border: InputBorder.none,
          contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
        ),
        style: TextStyle(
          color: Theme.of(context).colorScheme.onSurface,
          fontSize: 16,
        ),
        onSubmitted: (_) => widget.onSearch?.call(),
      ),
    );
  }
}