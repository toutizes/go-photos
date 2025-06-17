# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Go Backend
- `make serve_small` - Start local development server on :8080 with local backend and Firebase auth
- `make generate` - Generate protobuf files (backend/store/db.pb.go from db.proto)
- `make push_serve` - Build and deploy backend server to production
- `make minify` - Process and minify images for web delivery  
- `make build_sync` - Build Lightroom sync tools (gsync, gsync_incr)
- `go test ./backend/model/` - Run backend tests
- `go run backend/test/list_kwds.go <image_path>` - Test keyword extraction on specific image

### Flutter Frontend
- `cd floutizes && flutter run -d chrome --web-port 3000` - Run Flutter web app locally with localhost:8080 backend
- `cd floutizes && flutter run -d chrome --web-port 3000 --dart-define BACKEND=https://toutizes.com` - Run with production backend
- `cd floutizes && flutter build web` - Build for web (local backend)
- `cd floutizes && flutter build web --release --dart-define BACKEND=https://toutizes.com` - Build for production
- `cd floutizes && flutter test` - Run Flutter tests
- `cd floutizes && flutter analyze` - Run static analysis (uses analysis_options.yaml)
- `cd floutizes && flutter emulators --launch Apple` - Launch iOS simulator
- `cd floutizes && flutter run -d "iPhone 15 Pro Max"` - Run on iOS simulator

### Full Stack Development
- `make install_flutter` - Copy built Flutter web app to server htdocs
- `make push_flutter` - Deploy Flutter web app to production server

## Architecture Overview

### System Components
This is a photo management system with three main components:
1. **Go backend**: HTTP API server with image processing, keyword indexing, and Firebase authentication
2. **Flutter frontend**: Cross-platform mobile/web app for browsing and searching photos
3. **Static web interface**: Legacy HTML/JS photo browser in htdocs/

### Backend Architecture (Go)
- **Main server**: `backend/test/aserve.go` - HTTP server with CORS support, Firebase auth, static file serving
- **Database model**: Protobuf-based storage (`backend/store/db.proto`) with Image and Directory messages
- **Core models**: 
  - `backend/model/database.go` - Main database structure with image indexing
  - `backend/model/image.go` - Image metadata with keywords, timestamps, stereoscopic info
  - `backend/model/directory.go` - Photo album/directory structure
  - `backend/model/query.go` - Search query processing with keyword filtering
- **Image processing**: Multi-resolution image pipeline (original → midi → mini → montage)
- **Indexing**: Keyword extraction and search indexing via `backend/model/indexer.go`

### Frontend Architecture (Flutter)
The Flutter app (`floutizes/`) follows the architecture described in `floutizes/CLAUDE.md`:
- **Authentication**: Firebase Auth with Google Sign-In, JWT token management
- **State management**: Provider pattern for global auth state, singleton API service with caching
- **Navigation**: GoRouter with nested navigation shells, bottom tab navigation
- **Image display**: Three-tier sizing (mini/midi/maxi), lazy loading, montage grouping
- **Search**: Query-based search with backend keyword processing

### Data Flow
1. **Image ingestion**: Original photos stored in `orig_root`, processed into multiple sizes
2. **Metadata extraction**: EXIF data, keywords, timestamps indexed in protobuf database
3. **Search queries**: Frontend sends text queries → backend processes keywords → returns matching images
4. **Authentication**: Firebase JWT tokens validate API requests, automatic token refresh

### Firebase Integration
- **Backend**: Uses Firebase Admin SDK for token verification (`firebase_creds` parameter)
- **Frontend**: Firebase Auth with Google Sign-In provider
- **Security**: All API endpoints require valid Firebase JWT tokens

### Directory Structure
- `backend/` - Go backend code (models, tests, utilities)
- `floutizes/` - Flutter mobile/web app
- `htdocs/` - Legacy web interface with JavaScript photo browser
- `sync/` - Lightroom integration tools
- `cmd/` - Utility scripts for deployment and maintenance

### Image Processing Pipeline
- **Original**: Full resolution images in `orig_root`  
- **Midi**: Medium resolution for display
- **Mini**: Thumbnails for grid views
- **Montage**: Grouped similar images for compact display
- All processed images served via HTTP with proper caching headers

### Development Workflow
1. Use `make serve_small` for local development with hot reload
2. Flutter app connects to localhost:8080 backend by default
3. Use `--dart-define BACKEND=https://toutizes.com` to test against production
4. Build and deploy with `make push_serve` (backend) and `make push_flutter` (frontend)

### Testing
- Go backend: Standard Go testing in `backend/model/`
- Flutter: Widget tests in `floutizes/test/`, static analysis via `flutter analyze`
- Integration: Manual testing with real photo libraries via `make list_kwds`