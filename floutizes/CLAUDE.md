# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Building and Running
- `flutter build web` - Build for web
- `flutter build web --release --dart-define BACKEND=https://toutizes.com` - Build web release with production backend
- `flutter run -d chrome --web-port 3000` - Run in Chrome with local backend (localhost:8080)
- `flutter run -d chrome --web-port 3000 --dart-define BACKEND=https://toutizes.com` - Run in Chrome with production backend
- `flutter run -d "iPhone 15 Pro Max"` - Run on iOS simulator
- `flutter run -d "iPhone 15 Pro Max" --dart-define BACKEND=https://toutizes.com` - Run on iOS simulator with production backend

### Testing and Analysis
- `flutter test` - Run unit tests
- `flutter analyze` - Run static analysis (configured in analysis_options.yaml)
- `flutter doctor` - Check Flutter environment status

### Emulator Management
- `flutter emulators --launch Apple` - Launch iOS emulator
- `flutter devices` - List available devices

## Architecture Overview

### Backend Communication
The app connects to a photo backend service via HTTP API:
- **Local development**: `http://localhost:8080` (default)
- **Production**: `https://toutizes.com` (set via `--dart-define BACKEND=https://toutizes.com`)
- Backend URL is configured in main.dart using `String.fromEnvironment('BACKEND')`

### Authentication Flow
- Firebase Auth integration with Google Sign-In
- `AuthService` manages authentication state globally via ChangeNotifier
- JWT tokens are automatically included in API requests via Authorization headers
- **Reactive token renewal**: API requests automatically retry with refreshed tokens on 401 authentication failures
- Router-level authentication guards redirect unauthenticated users to `/login`
- Authentication state is preserved across app restarts

### State Management Pattern
- **Global state**: Provider pattern with ChangeNotifier (AuthService)
- **API service**: Singleton pattern with caching (ApiService)
- **Navigation**: GoRouter with nested navigation shells
- **Local state**: StatefulWidget with setState for component-specific state

### Key Data Models
- `ImageModel`: Represents individual photos with metadata, keywords, and stereoscopic info
- `DirectoryModel`: Represents photo albums/directories with cover images and counts
- Both models provide path helpers for different image sizes (mini/midi/maxi)

### Image Display Architecture
- **Three image sizes**: mini (thumbnails), midi (medium), maxi (full resolution)
- **Caching**: CachedNetworkImage for performance
- **Responsive layout**: Dynamic grid columns based on screen width (calculateOptimalColumns)
- **Montage support**: Automatic grouping of similar images into montages
- **Lazy loading**: Images load on-demand in scrollable grids

### Navigation Structure
- **Bottom navigation**: Albums tab and Photos tab
- **Nested routing**: Each tab maintains its own navigation stack
- **Deep linking**: Support for search queries and image IDs in URLs
- **Immersive mode**: Hide navigation bars for full-screen image viewing

### Search and Filtering
- **Query-based search**: Backend handles complex search queries
- **Caching**: ApiService caches search results (max 3 queries)
- **Keyword navigation**: Clickable keywords in image details
- **Album browsing**: Separate album search with `in:` syntax

### Platform-Specific Features
- **Web**: Download functionality for search results
- **Mobile**: Native navigation and touch interactions
- **Cross-platform**: Material 3 design with light/dark themes

### Error Handling
- Network errors are caught and displayed with retry options
- Authentication errors automatically redirect to login
- Invalid routes show appropriate error messages
- Loading states and empty states are handled consistently

## Development Notes

### Adding New Screens
- Add route to GoRouter configuration in main.dart
- Consider whether it needs authentication guards
- Use appropriate navigation shell (albums vs images branch)

### API Integration
- Extend ApiService for new endpoints
- Maintain authentication headers pattern
- Consider caching strategy for new data types
- Follow existing error handling patterns

### Image Processing
- Use the three-tier image size system (mini/midi/maxi)
- Consider montage grouping for similar images
- Implement proper loading states and error handling
- Use Hero widgets for smooth transitions between screens