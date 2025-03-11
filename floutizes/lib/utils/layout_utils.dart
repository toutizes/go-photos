import 'package:flutter/material.dart';

/// Calculates the optimal number of columns for a grid based on screen width and desired item width.
/// 
/// Parameters:
/// - screenWidth: The available width of the screen
/// - minItemWidth: The minimum desired width for each item
/// - maxItemWidth: The maximum desired width for each item
/// - maxAllowedColumns: The maximum number of columns allowed
/// - padding: The total horizontal padding of the grid (default: 16.0)
/// - spacing: The spacing between items (default: 8.0)
int calculateOptimalColumns({
  required double screenWidth,
  required double minItemWidth,
  required double maxItemWidth,
  required int maxAllowedColumns,
  double padding = 16.0,
  double spacing = 8.0,
}) {
  // Calculate number of columns that would fit with minItemWidth
  int maxColumns = ((screenWidth - padding) / (minItemWidth + spacing)).floor();
  // Calculate number of columns that would fit with maxItemWidth
  int minColumns = ((screenWidth - padding) / (maxItemWidth + spacing)).ceil();
  
  // Ensure minColumns doesn't exceed our maximum allowed columns
  minColumns = minColumns.clamp(1, maxAllowedColumns);
  // Ensure maxColumns is at least as large as minColumns
  maxColumns = maxColumns.clamp(minColumns, maxAllowedColumns);
  
  return maxColumns;
} 