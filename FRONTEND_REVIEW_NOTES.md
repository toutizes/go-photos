# Toutizes frontend review and rewrite context

Review date: 2026-10-12. Repository revision: `4eb7dad`.

Toutizes is a personal photo archive with about 70,000 deployed photos, according to Matthieu. Lightroom exports remain the source of new photos. The intended next step is a web-only frontend with familiar functionality and improved grid, carousel, and mobile browser behavior, using the unchanged Go backend. The initial review changed no application code. Matthieu subsequently authorized temporary frontend and backend auth bypasses for local browser testing, described below.

Google identification serves three purposes: discourage bots, let Matthieu see who visits, and leave room for stronger access restrictions if unfamiliar visitors arrive. It is currently an identification gate for the archive, not an invitation or per-album permission system.

Matthieu corrected his earlier clarification: **Admin**, not Activité, is used only by him on desktop web. Activité is the key entry point for all visitors to see what is new in the photo collection and must work well on phones in both portrait and landscape. Mobile priorities include Activité, the grid and the carousel; mobile Admin support is not required.

## Initial review and verification status

Read the application Go model and server code, Flutter Dart source, legacy JavaScript and HTML/CSS, sync and maintenance scripts, existing tests, and web/native configuration. Generated protobuf implementation, compiled bundles, dependency internals, and native project scaffolding are not separate sources of product requirements. The pre-existing untracked `.DS_Store` and `floutizes/lib/models/#image.dart#` were left untouched.

During workspace checks, concurrent edits briefly appeared in `floutizes/lib/main.dart`, commenting out auth/API initialization and the router auth listener. A subsequent check showed no tracked changes. The reviewing agent did not make or revert those edits; the editor backup file also disappeared independently. The architecture below describes the source as initially read.

| Check | Result |
| --- | --- |
| Build Go server with Go 1.23.7 | Passed: `GOCACHE=/tmp/go-photos-review-gocache go build -o /tmp/go-photos-review-aserve backend/test/aserve.go`. |
| Run Go model tests | Failed at compilation: `update_test.go:88` and `:107` call the old two-argument `UpdateDirectory`; current function takes four arguments. No passing suite claimed. |
| Flutter static analysis without dependency resolution | Completed with 15 findings: one unused-method warning and fourteen informational findings, mainly deprecated APIs and web imports. |
| Flutter tests without dependency resolution | Failed before test execution with SDK/dependency compatibility errors, including null-aware collection elements and missing `Matrix4` methods. The existing test is also the stock counter test, not a test of this app. |
| Local prerequisites | Go, Flutter, Dart, protoc, ImageMagick and exiftool are installed. The local `go-iptc` replacement and `test-photos` directory exist; the latter contains 804 files. An existing Flutter web build is present. |
| Authenticated local app | Not started. The Firebase credential path in `make serve_small` is missing. Matthieu replied “hold off” when asked for the current path; credential setup and authenticated startup are on hold. |
| Production and mobile browser inspection | Not performed in this pass. No authenticated browser session, performance measurements, or deployed-version equivalence claimed. |

The installed Flutter reports 3.38.3 with Dart 3.10.1. Its initial version check attempted an automatic Git fetch and reported a ref lock conflict in the external Flutter checkout; analysis still ran. No SDK upgrade, lock removal, package upgrade, or application dependency edit was performed. Tests wrote normal ignored build/cache artifacts; the Go executable and build cache were placed under `/tmp`.

## System map

| Area | Main files | Role |
| --- | --- | --- |
| Server | `backend/test/aserve.go` | Production-style HTTP server despite its location under `test`; Firebase verification, routing, CORS, static frontend serving. |
| Persistent metadata | `backend/store/db.proto`, `backend/model/database.go` | Per-directory binary protobuf and readable protobuf text indexes on disk; in-memory search indexes. No SQL database. |
| Photo metadata | `load_image.go`, `update.go`, `image.go`, `directory.go` | JPEG/IPTC keywords, EXIF dates, dimensions, directory grouping, legacy metadata preservation. |
| Search | `query.go`, `indexer.go`, `synonyms.go`, `synonyms.txt`, `accents.go` | Date, album and keyword query language, names and aliases, keyword/subword indexes. |
| Images | `minifier.go`, `montage.go`, `commands.go` | Resizing, thumbnail strips, image serving, ZIP downloads and database reload. |
| Activity and visitors | `database.go`, `json.go`, `log_parser.go` | Recent keyword groups and admin search history. |
| Frontend shell | `floutizes/lib/main.dart`, `screens/home_screen.dart` | Firebase setup, Provider auth state, GoRouter branches, search, navigation and themes. |
| Frontend data | `services/`, `models/` | API requests, auth refresh, client sorting, three-query cache, JSON adaptation. |
| Photo browsing | `screens/flow_view.dart`, `albums_view.dart`, `image_detail_view.dart` | Grid, albums, carousel, metadata, stereo and immersive mode. |
| Web images and downloads | `widgets/native_image*.dart`, `utils/montaged_images.dart`, `utils/image_download*.dart` | Authenticated image requests, native DOM images, sprite cropping, blob downloads. |
| Ingestion | `sync/gsync.go`, `cmd/`, root `Makefile` | Lightroom export actions, Google Drive transfer, server pull and deployment. |
| Earlier browser app | `htdocs/` | jQuery photo browser, stereo UI, justified rows, experimental scrolling and old video UI. |

## Lightroom to browser

1. Lightroom exports photos to `final` folders under configured local photo roots.
2. `gsync` maps each file to Google Drive's Photos tree, removing the `final` path component. It **moves** exported files with `os.Rename`.
3. Build-time `Type=F` deletes each destination directory before moving its exports; `Type=I` preserves existing contents. These are operationally different export actions.
4. `cmd/cron-pull.sh` pulls selected directories from Drive on the server, updates directory modification times, then requests a database reload. This script appears older than the current auth setup: its reload request has no bearer token, although `/db/viewer` is now authenticated on both server ports. Actual production scheduling needs confirmation.
5. The Go loader scans originals, compares directory/file times with stored metadata, extracts changed JPEG metadata and builds its in-memory index. Database loading can update indexes and generate resized copies.
6. The browser requests query metadata and authenticated photo bytes. There is no upload or Lightroom editing UI in Flutter.

`make test_sync` passes `--dry_run`, but `gsync.go` has no argument parser for that flag and leaves `dry_run=false`. Do not use that target as a harmless inspection command. No sync, deployment, reload, or photo-moving command was run.

EXIF date preference is original, digitized, generic date, then file modification time. Directory names can correct dates that appear too far in the future, useful for older scanned material. Keywords come from IPTC record 2/dataset 25, with UTF-8 or Latin-1 decoding and semicolon splitting. Existing metadata supplies fallbacks when fresh extraction lacks it. The exported “original” served by `maxi` is the archive file from this pipeline, not necessarily a camera RAW file.

## Backend contract to preserve

All registered `/db` endpoints require `Authorization: Bearer <Firebase ID token>`. Plain image tags cannot supply that header, so a replacement browser frontend needs an explicit authenticated image-loading strategy. The current frontend uses header-bearing image requests and fetched blobs.

| Endpoint | Contract and caveats |
| --- | --- |
| `/db/q?q=...` | Returns the entire deduplicated photo array. No pagination, cursor, server sort parameter, or total-only endpoint. Empty query returns an empty array. |
| `/db/q?q=albums:&kind=album` | Returns albums containing photos. Frontend sorts them. |
| `/db/q?q=in:<text>&kind=album` | Album-name substring search used by the Albums tab. |
| `/db/mini/<album>/<file>` | JPEG thumbnail, centered square crop up to 360 pixels; small inputs are not upscaled. |
| `/db/midi/<album>/<file>` | JPEG preserving aspect ratio, longest side up to 2048 pixels, quality 90. |
| `/db/maxi/<album>/<file>` | Exported original file. |
| `/db/montage/360x360-<id>-<id>...` | Horizontal strip of thumbnails, cached on disk. The geometry string is parsed but current generation uses actual thumbnail dimensions. |
| `/db/recent-keywords` | Keyword counts and up to four preview IDs/names; single-keyword endpoint has a load-time cache. |
| `/db/recent-keyword-groups` | Activity groups, up to five keywords and four preview images per group. Computed on request. |
| `/db/user-queries` | Admin-only map of usernames to recent searches. Optional `user` parameter returns one user's history. |
| `/db/viewer?command=download&q=...&s=M` | ZIP of at most 100 matching midi images. `s=O` selects originals; implementation treats any non-`M` value as originals. |
| `/db/viewer?command=reload` | Rebuilds metadata/index and derivatives, then swaps active data. Not a browsing operation. |

Photo JSON fields: `Id`, `Ad` (album path), `In` (filename), `Its` (taken time), `Fts` (file time), `H`, `W`, `Kwd`, nullable `Stereo` with `Dx`, `Dy`, `AnaDx`, `AnaDy`. API timestamps are Unix seconds; protobuf timestamps use milliseconds. Album fields: `Id`, `Ats`, `Dts`, `Nimgs`, `Cov`, `CovName`, `PreviewIds`, `PreviewNames`. `Dts` is actually the latest contained photo file time, despite some comments calling it directory time. The previews exclude the cover and include up to three additional files.

Image IDs are 32-bit FNV hashes of album path, filename and file modification time. Re-exporting a file can change its ID. A stale detail link currently falls back to the first matching photo; IDs are not permanent archival identifiers.

The server redirects `/` to `/app/` and serves an arbitrary static application from `<static_root>/flutter`, falling back to `index.html` for missing paths. This provides a feasible deployment location for a replacement static frontend without changing Go. The directory name need not dictate the frontend framework. Assets and image files receive one-year public caching headers; `index.html` does not. Asset versioning and re-exported photo freshness therefore need explicit handling in the replacement.

## Authentication and visitor monitoring

The web login uses Google Sign-In, exchanges its credentials with Firebase, and lets router redirects restore the requested location. `AuthService` stores the Firebase user and ID token. JSON API requests refresh the token and retry once after a 401. Grid image requests, native-image fetches and download requests do not share that retry path.

The backend verifies Firebase tokens and extracts email. The admin history endpoint checks a fixed owner email server-side; its tab is also hidden from other users. General photo access has no per-person allowlist. The reload command shares the general auth gate rather than the history endpoint's owner-only check. A future restriction to approved people would require server-side enforcement beyond a frontend rewrite.

Search logging happens on `/db/q`. The admin parser reads `aserve.log*`, recognizes only `@gmail.com` addresses, ignores empty searches and `albums:`, removes consecutive duplicate searches and returns up to 30 per user. Ordinary Google accounts using other email domains would not appear in that history. Photo byte requests log URL and remote address without attaching the user's email, and advancing the carousel does not issue a dedicated view event. API cache hits also skip server query logging. Preserve this distinction when discussing “who looks at photos.”

`web/index.html` initializes a JavaScript Firebase app using a `toutizes` configuration, while Dart initializes `floutizes`. The HTML startup waits for its JavaScript initialization before starting Flutter. Confirm the intended production configuration before reusing this bootstrap. No private credentials were read or stored in these notes.

## Current user journeys

The interface is mostly French and follows system light/dark appearance. The primary navigation is Activité, Albums, Photos, plus Admin for the owner. GoRouter maintains branch navigation state. Initial location is Activité; Photos without an explicit query defaults to today's month/day across years.

Albums show square cover cards, directory paths and photo counts. They sort by the first photo's capture time descending, then album ID. Selecting one opens an `album:` query. General photo queries sort capture time descending; recognized album queries sort ascending. Equal-time photo ties use descending filenames. Album detection in the Flutter sorter simply splits on spaces and checks for a token starting `album:`, which is fragile for quoted album paths.

The photo grid uses square tiles, 8-pixel gaps, 8-pixel outer padding, a width-based column calculator and a maximum of six columns. Widgets are built lazily, but the complete metadata result is fetched and sorted first. Eight consecutive photos are combined into each montage request; this is request batching, not grouping visually similar photos. Album covers use the same mechanism. Two empty rows pad the bottom of the photo grid.

Tapping a tile opens `/images/details/<id>?q=...`. The carousel provides horizontal paging, arrows, arrow-key navigation, a position counter, keyword/album/date links, download actions, pinch/pan through `InteractiveViewer` with maximum scale 5, and immersive mode. Double-tap toggles immersion rather than a conventional zoom toggle. Immersion hides app navigation; it is not an implementation of the browser Fullscreen API. The normal viewer always uses midi images, including when zoomed. There is no ordinary-photo timed slideshow in this implementation.

Landscape places image and metadata in a 70/30 split; portrait gives metadata 20% of viewport height. Returning to the grid includes the current image ID and estimates a scroll offset to center its row. Changing pages updates the route with `context.go`, despite a comment describing an update without navigation.

Stereo images have parallel, crossed-eye, animated alternating-eye, and alignment modes. Alternation runs every 300 ms. Stereo keywords can create zero-offset stereo metadata and overwrite server-provided offsets in the Dart model. Alignment drag changes the in-memory model; saving to the backend is commented out. `HandleSet` exists in Go but is not registered in the active server. Legacy JavaScript supports additional stereo shortcuts and attempted persistence, which should not be assumed functional today.

Activité ranks keywords drawn from the 20 albums with greatest `ItemTime` (the first photo time, or index time for empty directories). It excludes keywords occurring only once, weights the first five occurrences by relative recency, and groups keywords sharing their newest preview image. This is photo-content activity, not visitor activity. It is not a “last month” filter despite an older comment. The separate News view orders albums by file-update time; it exists and is constructed, but has no current navigation branch.

## Grid and carousel investigation priorities

These are findings from code inspection. Browser testing is needed to quantify their user-visible effect.

- Each HomeScreen constructs all five views in an `IndexedStack`. Hidden News, Activity and Albums views initiate requests, and each visited navigation branch can create another HomeScreen. Initial work and repeated requests may exceed what the visible tab needs.
- Every grid build recreates the montage grouping for the complete result. Each rendered tile then searches the item list with `indexWhere`. This makes large-result CPU costs worth measuring even though widgets are lazy.
- Montage cropping relies on a horizontal strip, fixed nominal thumbnail size and alignment. Albums use `BoxFit.none`; some small inputs produce smaller thumbnails. Test cropping at every tile width before preserving this implementation.
- Native carousel images register a new HTML platform-view factory during builds and create authenticated blob URLs without revoking them. Nearby-image preloading instead uses Flutter `NetworkImage`. Cache reuse across those two mechanisms is unverified; repeated downloads/decodes and retained memory are plausible.
- Mobile gesture ownership crosses native DOM images, `InteractiveViewer`, `PageView`, and overlay controls. Test swipe versus pan while zoomed, double-tap, pinch, long-press and browser back gestures explicitly.
- Returning from immersion does not have an obvious shared-notifier reset in `dispose`. Test route changes and back navigation while immersive. Keyword links also mix route changes with `Navigator.pop`; history/scroll restoration needs browser verification.
- Stereo controls use an unwrapped horizontal row. Search actions and the carousel count compete for narrow header space. Activity also overflows narrow screens and needs responsive previews for portrait and landscape phones.
- Photo detail has both the parent bottom navigation and an internal bottom placeholder/padding. Review the resulting image area, especially landscape phones, keyboard opening and dynamic browser toolbars.
- Empty detail results are handled in the body, but the download toolbar still indexes `_images[_currentIndex]`. Empty/stale links deserve an explicit regression case.
- Search cache is FIFO for three query strings, has no expiry or auth-user key, and is not cleared by sign-out. New photos may not appear for a cached query during a long visit.
- Download XHR handlers do not check HTTP status before creating the file, and buffer the complete response in a blob. Error responses can appear as downloads; large original ZIPs can consume substantial mobile memory.
- Activity preview metadata lacks album paths. Its montage fallback constructs an individual mini URL with an empty album directory and is unlikely to recover correctly.

## Search compatibility

The backend has two parsers. Local `serve_small` explicitly enables the newer comma-separated Lightroom parser; the server flag defaults to the older parser. Production startup flags have not been verified.

| Syntax | Meaning |
| --- | --- |
| `person, place, 2025` | Lightroom parser ANDs comma-separated criteria. Multiword terms use keyword/name matching; single words use substring matching. |
| `2025`, `2025-06`, `2025-06-15` | Capture year, month or day, with matching date-like keywords also considered. |
| `2020--2022` | Year range. Arbitrary full-date ranges described in help are not implemented by the current regexes. |
| `10-12` | Month/day across all years. |
| `album:<exact path>` | All photos from exactly that album. |
| `in:<text>` or `titre:<text>` | Photos from albums whose paths contain the text. |
| `count:<n>` | Photos with exactly n keywords. |
| `stereo:` | Photos with backend stereo metadata; keyword-based frontend stereo detection is separate. |

Name aliases live in `synonyms.txt`. Search indexes also contain subwords, filenames and French accent-normalized variants. The frontend should preserve the backend's semantics rather than promise broader linguistic matching or an invented OR syntax.

The search hint tells users to use commas, but bundled help still teaches quotes and space-separated combinations. Album selection still wraps paths containing spaces in quotes, which the Lightroom parser does not strip. These mismatches are frontend compatibility concerns with an unchanged backend.

## Earlier implementation and scope boundaries

The legacy `htdocs` app contains a justified row layout based on photo aspect ratios, a thumbnail rail, URL state encoded as base64, bounded image preloading and stereo keyboard shortcuts. It is useful evidence of earlier interaction intent. Its direct image URLs and unauthenticated AJAX are not compatible with the current auth gate as written.

Legacy video UI and RSS helpers exist, but the active server does not register an RSS route, the current Directory model omits videos, and Flutter has no video screen. These are historical capabilities to clarify, not automatic rewrite requirements. Android/iOS wrappers contain Flutter hosting and platform configuration; the user explicitly makes web the target.

## Requirements clarified by Matthieu

Responses to questions 1–12 (2026-10-12):

- The main reported viewer problem is that photos are much too small on a landscape phone, including in the current zoomed/immersive mode. In the photo-with-metadata view, the image should fill as much of its reserved area as possible.
- Keep square grid crops. AI-assisted selection of a better square crop is a possible later enhancement, not part of the current rewrite requirements.
- Provide two successive viewing levels: grid → image with metadata → fullscreen image. Double-tap opens the next level in both transitions. Single-tap behavior has not yet been specified.
- Fullscreen shows the photo without metadata, supports swipes to navigate between photos and pinch to zoom, and offers a visible close control. An X is suitable; keeping the current magnifier-minus control is also acceptable. Whether fullscreen uses the browser Fullscreen API or fills the available browser viewport remains an implementation decision to validate on mobile browsers.
- Pinch zoom may be unnecessary in the image-with-metadata view; this is a tentative preference. Double-tap should enter fullscreen from that view.
- Activité remains the main entry point for all visitors and must work well in portrait and landscape on phones. The earlier desktop-only statement applied to Admin instead.
- Preserve all stereo modes present in the Flutter app, including temporary alignment controls. Matthieu uses them; persistence of adjustments is not required.
- Keep the existing 100-photo ZIP limit and make the limit clear in the UI.
- Compatibility with old photo URLs is not required. Links are generally shared briefly in chats rather than bookmarked; breaking older URLs is acceptable.
- Search history is sufficient; per-photo viewing history is not needed. User searches must still be logged when results come from the frontend cache. Google accounts must not be restricted to Gmail addresses, and search history should include other email domains.
- Comma-separated search is canonical and production uses `--use_lr_parser`. Visitors typically start by typing text (often a single word), then follow keyword chips in photo metadata. Preserve both paths and make search help consistent with comma-separated syntax.
- All application UI text should be French. Preserve the system light/dark theme and current logo. Ordinary browser tabs are sufficient; an installed home-screen app is not a requirement.
- All Lightroom exports are JPEGs, with varying filename extensions/case such as `.jpg`, `.JPG` and `.jpeg`. New photos are typically added once or twice a month. Panoramas are very wide JPEGs and must remain viewable.
- A fullscreen panorama mode that fills the available height and allows horizontal panning is a possible enhancement **after** the rewrite, not an initial requirement.

Search-history implementation implications: `searchImages` currently returns immediately on a cache hit without calling `/db/q`, so the backend never sees that search. With the backend unchanged, logging requires another ordinary query request (and its normal query/result cost), even if cached results are shown immediately. A lightweight logging-only request would require a backend addition. Separately, `/db/q` already logs the full email for all domains, but `backend/model/log_parser.go` only parses Gmail addresses and keys history by the email prefix. Supporting non-Gmail accounts in Admin history requires a small backend parser change, with full email identities to avoid cross-domain collisions. Record this as a needed exception to the unchanged-backend scope for later implementation; no backend changes are being made now.

Specific phone/browser models, metadata placement, thumbnail rail and timed slideshow preferences remain unanswered, as do typical export dimensions, largest album/result sizes and re-export frequency (distinct from the confirmed frequency of new additions). No viewer implementation changes have been requested yet.

## Production deployment context supplied by Matthieu

Matthieu confirmed `make push_serve` as the backend installation command and `make push-web-release` as the production frontend deployment command. In this checkout, `push_serve` builds a Linux amd64 binary and copies it to `ce1:/mnt/photos/bin/`; it does not restart the process. `push-web-release` builds and uploads the frontend to the server. `install_flutter` is only for copying an existing web build and synonyms into the local `/tmp/aserve/db/htdocs/` tree. How the backend is stopped/restarted remains to clarify. No deployment commands were run.

The supplied production launcher is recorded below (whitespace normalized). This confirms the production paths and flags; the credential path is on the server and does not resolve the missing local development credential file.

```bash
#!/bin/bash
LOG_FILE="/mnt/photos/daemon/aserve.log.$(date '+%Y-%m-%d:%H:%m:%S')"
export MAGICK_TMPDIR=/mnt/photos/tmp

/mnt/photos/bin/aserve \
  --bin_root=/usr/bin/ \
  --orig_root=/mnt/photos/originals \
  --root=/mnt/photos \
  --static_root=/mnt/photos/htdocs \
  --use_lr_parser \
  --letsencrypt_root=/mnt/photos/letsencrypt/config \
  --firebase_creds=/mnt/photos/floutizes-firebase-adminsdk-fbsvc-584c308c82.json \
  --use_https=true \
  --num_cpu=4 \
  --minifier_threads=2 \
  --log_dir=/mnt/photos/daemon \
  "$@" \
  &> "$LOG_FILE" &

echo $! > /mnt/photos/daemon/aserve.pid
```

## Questions for the next discussion

Responses to all 12 questions are recorded above; retain the questions below for reference. Remaining details are noted above. Question 4 has been corrected to remove the mistaken premise about Activity's audience.

1. Which current carousel failures bother you most: missed swipes, zoom/pan conflicts, loading flashes, image sharpness, lost position, browser back behavior, or something else? Which phones and browsers reproduce them?
2. Should the grid retain square crops, show complete aspect ratios in justified rows like the earlier app, or offer density/layout choices? Are faces being cropped the main issue?
3. What should tap, double-tap, swipe and pinch do in the viewer? Should metadata normally be hidden, below the image, or beside it? Would a thumbnail rail or timed slideshow be useful?
4. Which landing experience should visitors see? **Answered: Activité for everyone, including phones in both orientations; Admin is the owner-only desktop view.**
5. Are all stereo modes used, and is alignment editing important? Is persistence of adjustments expected, or is temporary alignment sufficient?
6. Is the existing 100-photo ZIP ceiling acceptable if clearly stated? Is downloading whole large albums essential? The present endpoint cannot express arbitrary selections or offsets.
7. How often are photo links shared or bookmarked? Must existing Flutter URLs and older `pic.html` links keep working? What should happen after a photo is re-exported and its ID changes?
8. Is visitor search history sufficient, or do you eventually want per-photo viewing history? The latter requires a backend capability beyond this frontend-only scope. Do visitors use Google accounts with non-Gmail addresses?
9. Is comma-separated search now canonical, and does production use `--use_lr_parser`? Should people mostly type free text, use keyword chips, or combine both?
10. Should the replacement preserve the mostly French text, system theme and current logo? Is an installed home-screen web app important, or just use in ordinary browser tabs?
11. What export sizes, largest album/result sizes and re-export frequency are typical? Are there panoramas or non-JPEG files that the browser must handle?
12. When local work resumes, what are the current credential path, production startup flags and server pull/reload mechanism? The repository scripts appear to span several generations of the deployment.

## Validation remaining for the rewrite

Resume authenticated local browsing only when requested. Use a temporary derivative/index root so original photos remain untouched. Keep the backend's existing Firebase verification. Production flags and deployment commands are now confirmed above; the deployed build revision remains unverified.

Then exercise login and deep-link restoration, activity, album search, a large photo query, carousel transitions, keyword navigation, grid restoration, stereo, downloads, sign-out and token expiry. Repeat the core photo flows on a narrow iPhone Safari viewport and Android Chrome, including rotation, soft keyboard and browser back/forward. Measure first usable content, query payload size, image request counts, and memory after a long browsing session.

Matthieu accepted **TypeScript + React DOM + Vite** for the web-only replacement, using pnpm to align with his existing Aleppo tooling. Build static assets for the existing `/app/` deployment and Go API; no additional production Node server is needed. React Native Web is not the selected UI layer. Package versions and viewer libraries remain to be selected during implementation.

## Proposed implementation milestones

1. Create a separate `frontend/` application alongside `floutizes/`, with French UI, the existing logo and system theme. Connect it to the real local Go API and isolate its build output from the Flutter build. Preserve Google/Firebase authentication for production, with an explicit local-only development bypass for current testing. Keep the Flutter app available for comparison and rollback.
2. Build one complete photo-browsing flow with real data: comma-separated search → square grid → image with metadata → fullscreen. Include metadata chips, double-tap transitions, swipe navigation, pinch/pan in fullscreen, a close control, browser history and return-to-grid position. Maximize image area on landscape phones. Bound mounted grid items and image/blob caches, and preload only nearby viewer images.
3. Review this first working flow in desktop and real mobile browsers, in portrait and landscape, before extending the app. Check gesture conflicts, long-query scrolling, loading behavior and memory. Viewport emulation alone does not verify phone touch behavior.
4. Complete Activité as the responsive landing page, Albums, all existing stereo modes with temporary alignment, downloads with the 100-photo limit, and desktop Admin. Address cached-search logging and the separately identified non-Gmail history parser requirement.
5. Verify authenticated production builds, feature parity, large-library behavior and deployment/rollback steps. Update deployment tooling to publish the new static build only when ready; deploying the rewrite is a separate action from local implementation.

First reviewable milestone: the real local photo collection can be searched and browsed through both viewer levels with correct landscape sizing and browser Back behavior. Framework scaffolding by itself is not that milestone. Panorama panning and AI-selected crops remain later enhancements.

## Local browser testing after auth bypass authorization

On 2026-10-12, Matthieu authorized temporary local auth changes. The app now runs at `http://localhost:8080/app/`, with 804 local test photos indexed across 23 directories (including parent directories). Originals remain under `/Users/matthieu/projects/test-photos`; generated indexes, mini/midi files, montage cache, static frontend and logs are isolated under `/tmp/go-photos-local`.

The backend flag `--dev_no_auth` skips Firebase initialization and supplies `local-dev@example.test` to request handlers. Both listeners bind to `127.0.0.1` in this mode, and combining it with HTTPS is rejected. Normal startup still requires Firebase credentials, verified by launching without the bypass and observing `Must pass --firebase_creds`. The local identity is not the owner; admin access is intentionally not impersonated.

The frontend build flag `--dart-define=DEV_NO_AUTH=true` skips Dart Firebase initialization and makes AuthService locally authenticated. Sign-out is a no-op in this mode. Normal builds retain Google/Firebase authentication. The existing HTML Firebase bootstrap is unchanged, so the page still loads external Firebase JavaScript.

`flutter pub get --offline` refreshed the stale generated package configuration and updated six SDK-related transitive packages in `pubspec.lock`: leak_tracker, leak_tracker_flutter_testing, leak_tracker_testing, meta, test_api and vector_math. The previous lockfile was saved to `/tmp/go-photos-local/pubspec.lock.before`. No major app dependency upgrade was performed. The web release build passed; Flutter analysis still reports the same 15 pre-existing findings. `git diff --check` passed.

Commands used from the repository root:

```bash
mkdir -p /tmp/go-photos-local/htdocs/flutter /tmp/go-photos-local/log
cp backend/model/synonyms.txt /tmp/go-photos-local/htdocs/synonyms.txt
cd floutizes
flutter --no-version-check pub get --offline
flutter --no-version-check build web --no-pub --dart-define=DEV_NO_AUTH=true
cd ..
cp -R floutizes/build/web/. /tmp/go-photos-local/htdocs/flutter/
GOCACHE=/tmp/go-photos-review-gocache go run backend/test/aserve.go \
  --dev_no_auth --use_lr_parser \
  --orig_root=/Users/matthieu/projects/test-photos \
  --root=/tmp/go-photos-local/db \
  --static_root=/tmp/go-photos-local/htdocs \
  --log_dir=/tmp/go-photos-local/log \
  > /tmp/go-photos-local/log/aserve.log 2>&1
```

The generated `floutizes/build/web` currently contains the local bypass build. It is for local testing only; make a normal authenticated build before any eventual deployment. No production deployment was performed.

### Verified browser behavior

Browser: Codex in-app browser, desktop 1280×720 and responsive viewport overrides at 390×844 and 320×740. Viewport resizing is not a real iPhone Safari or Android touch-device test.

| Flow | Observation |
| --- | --- |
| Startup | Activity feed renders without login, with real local thumbnails. |
| Albums to grid | Selecting `2005/2005-01-30` opens its 13-photo grid successfully. |
| Grid to detail | Opens selected photo at position 9/13. |
| Keyboard and arrow navigation | Advances to 10/13 and 11/13, updates photo and URL. |
| Immersive mode | Visible zoom control hides app chrome and enlarges the image; exit restores it. |
| Browser Back | **Confirmed mismatch:** from photo 11, Back changes URL from ID `506431947` to prior ID `3733024446`, while the same photo remains visible. Exiting immersion still shows 11/13. The state does not reconcile a changed route ID for the same query. |
| Return to grid | Returns to the same album query with current `imageId`; grid scroll restoration occurs. Exact centering accuracy was not measured. |
| Search and scrolling | Searching `1979` yields 358 photos; scrolling several screens renders additional thumbnails. |
| Narrow activity feed | **Confirmed overflow; in scope to fix:** at 320 pixels, fourth previews extend outside cards and are cut off by the viewport. At 390 pixels, the preview row consumes the intended right padding and extends slightly beyond the card. Activité is the main entry point for all visitors and must work in both phone orientations. |
| Narrow carousel | At 390 pixels the search text is heavily truncated; landscape photos leave substantial vertical space, and date metadata sits partly below the visible metadata area. |

Screenshots: `/tmp/go-photos-local/carousel-mobile.png`, `/tmp/go-photos-local/grid-mobile.png`, `/tmp/go-photos-local/activity-320.png`.

Still unverified: actual touch pinch/pan/swipe ownership, real Safari/Chrome mobile behavior, stereo with representative sample images, long-session memory growth, large-library performance at 70k photos, authenticated expiry/sign-out, and download completion. The discovered UI defects were recorded rather than fixed; the temporary code changes serve local startup only.
