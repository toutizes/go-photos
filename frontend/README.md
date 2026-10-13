# Toutizes web

React + TypeScript + Vite frontend for the existing Go photo API. The UI is French, follows the system theme, and uses the existing logo. Flutter remains in `../floutizes` for reference and rollback.

## Local development

Use Node 22.12+ and pnpm 10. From the repository root:

```sh
pnpm --dir frontend install --frozen-lockfile
pnpm --dir frontend dev
```

Open `http://127.0.0.1:5173/app/`. Vite proxies `/db` to the Go server at `127.0.0.1:8080`, so there are no production CORS changes. The standard app requires Google/Firebase authentication. Its public Firebase configuration matches `floutizes/lib/firebase_options.dart`; server service-account credentials never belong in this frontend.

The temporary Go, Flutter and React authentication bypasses have been removed. All build and deployment targets use Google/Firebase authentication. Running the backend locally again requires a valid service-account credential file.

## Side-by-side production trial

From the repository root, run:

```sh
make release-react-app
```

This builds with normal Firebase authentication and uploads to `ce1` (override with `SERVER=...`). It uses the existing, unchanged Go static handler: assets go to `/mnt/photos/htdocs/flutter/react-test/` and the entry page to `/mnt/photos/htdocs/flutter/react-test.html`. Open the URL printed by the command, for example `https://toutizes.com/app/react-test.html?v=1234567890`.

The version query is intentional: the existing Go handler caches named HTML files other than `index.html` for one year. Use the freshly printed URL after each release to fetch the updated entry page. Hash-based navigation preserves that version query. Hashed JS/CSS assets from previous builds are retained for open tabs.

Flutter remains at `/app/`; its entry page, assets and service worker are not replaced. The staging build also disables Flutter service-worker retirement. The existing production backend already verifies Firebase tokens, so no backend deployment or credentials upload is needed. Google login on `toutizes.com` uses the same Firebase project and authorized domain as Flutter.

`make build-web-test` builds without uploading. `make push-web-test` is an alias for the same build-and-upload flow as `release-react-app`. To remove the trial later, remove only the `react-test.html` file and `react-test/` directory from the server. **`make push-web-release` replaces Flutter at `/app/`; do not use it for this side-by-side trial.**

## Commands

```sh
pnpm --dir frontend test   # regression tests, including a 70,000-photo grid fixture
pnpm --dir frontend build  # TypeScript checking and production static build
pnpm --dir frontend exec prettier --check src public vite.config.ts
make install-web          # build and copy to the standard local Go static root
make push-web-release     # build and deploy React to the existing production static directory
```

Deployment was prepared but has NOT been run. Production needs only Go and static assets; no Node process is required. The existing backend serves `<static_root>/flutter` at `/app/`, so the destination directory keeps its historical name. New assets have hashed filenames. Deployment copies assets first, then `index.html`, retaining previous assets for open tabs. Old assets can be pruned during later maintenance.

The included `flutter_service_worker.js` retires the old Flutter offline shell, removes only its named caches, unregisters itself and refreshes its app clients. The React app also unregisters an old Flutter worker if the new page has already loaded. There is no offline photo cache or new service worker registration. Check migration from an actual previous deployment before rollout; it has not been exercised against production.

`make push-flutter-release` retains the previous Flutter deployment recipe. Rebuild Flutter with normal auth before rollback. Its historical recipe also copies the Firebase service-account file and requires its existing local path. For a rollback without that path, restore a saved authenticated Flutter static release to `/mnt/photos/htdocs/flutter/`. No production backups have been taken by this task.

## Implemented behavior

- Activité is the landing page, with responsive keyword groups and lazy previews. “Ce jour-là” searches across years.
- Albums are sorted newest first and filterable by name or year.
- Comma-separated search and keyword/date/album links use the existing Lightroom parser.
- The square photo grid renders visible rows plus a small margin, not the whole result. Album results sort oldest first; other results sort newest first.
- A click/tap opens image + metadata; double-tapping from the grid also reaches that level. Double-tapping the image, or using the explicit button, opens fullscreen. Fullscreen uses the browser viewport, including dynamic mobile height; it does not depend on browser Fullscreen API availability.
- Fullscreen hides metadata, supports swipe navigation, pinch/pan (1–5×), reset and zoom buttons. A zoomed pan cannot accidentally navigate. Original-resolution images load when zooming; nearby photos preload at midi resolution.
- Arrow keys navigate, Escape/X closes one level. X retains the currently displayed photo; browser Back/Forward follows the URL. Closing to the grid restores the current photo's row.
- Stereo supports original, parallel, crossed-eye, 300 ms alternating-eye and alignment views. Alignment can be dragged or adjusted with sliders; it is temporary. Existing server offsets are preserved.
- Download the current original or a ZIP of a query in screen/original resolution. ZIPs are explicitly limited to 100 photos. Which 100 are selected is determined by the existing backend.
- Admin is visible only to the existing owner account and uses the server-protected search history endpoint. All Google account email domains remain accepted.

## Data, history and memory

Searches show cached results immediately but always reissue `/db/q` for explicit submissions, including repeat submissions, so server logging sees cache hits. This necessarily repeats the existing query and result transfer: the backend has no logging-only endpoint. Merely advancing the carousel does not issue a new search. The backend still deduplicates consecutive equal searches and shows up to 30 entries per visitor.

The query cache holds three results and clears on account changes. Image blob URLs are shared and reference-counted. Unreferenced entries are evicted above 48 MiB or 160 entries; referenced images may temporarily exceed the budget. In-flight loads without consumers are aborted, and sign-out clears and revokes images. Photo URLs include the file timestamp to distinguish re-exports. This budget measures compressed blobs, not browser-decoded image memory.

Activity previews use single-ID montage endpoints because the API omits album paths and multi-image montage geometry depends on the first source thumbnail. Album covers use the same path. Photo grids use individual mini paths. All fetches carry Firebase bearer tokens; expired tokens are refreshed once. Unauthenticated requests fail before contacting the API.

The only committed backend adjustment for this rewrite is the search-history parser accepting full email addresses across domains. Authentication and photo API behavior remain unchanged. Its dedicated tests can run independently of the pre-existing broken model test compilation:

```sh
GOCACHE=/tmp/go-photos-review-gocache go test backend/model/log_parser.go backend/model/log_parser_test.go
```

## Before production rollout

Verify Google login, logout and expiration with production credentials; exercise Admin with the owner account. Test real iPhone Safari and Android Chrome pinch, swipe, rotation, browser controls and downloads. Check an actual 70k library's payload/latency and extended-session decoded-image memory. The synthetic 70k regression checks grid virtualization, not production throughput. Review the old Flutter service-worker migration and preserve a rollback release.

Panorama height-fill/panning and AI-selected square crops remain deferred. Export dimensions and the largest real query payload are still unknown.
