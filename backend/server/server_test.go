package server

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"toutizes.com/go-photwo/backend/catalog"
)

type fakeVerifier struct {
	tokens map[string]string
}

func (fv *fakeVerifier) VerifyEmail(_ context.Context, idToken string) (string, error) {
	if email, ok := fv.tokens[idToken]; ok {
		return email, nil
	}
	return "", errors.New("invalid token")
}

func writeTestJPEG(t *testing.T, path string, w, h int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 80, G: 120, B: 160, A: 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatal(err)
	}
}

func TestServerEndToEnd(t *testing.T) {
	origRoot := t.TempDir()
	root := t.TempDir()
	staticRoot := t.TempDir()
	logDir := t.TempDir()

	// Populate static web files
	for name, body := range map[string]string{
		"index.html":                "react entry",
		"react-test.html":           "react test entry",
		"react-test/assets/test.js": "react asset",
	} {
		p := filepath.Join(staticRoot, "flutter", name)
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Populate test photos
	writeTestJPEG(t, filepath.Join(origRoot, "2025/2025-06-01/photo1.jpg"), 800, 600)
	writeTestJPEG(t, filepath.Join(origRoot, "2025/2025-06-01/photo2.jpg"), 800, 600)

	catOpts := catalog.Options{
		OrigRoot:        origRoot,
		Root:            root,
		StaticRoot:      staticRoot,
		UpdateDisk:      true,
		Minify:          true,
		MinifierThreads: 2,
	}
	cat, err := catalog.LoadCatalog(catOpts)
	if err != nil {
		t.Fatal(err)
	}

	verifier := &fakeVerifier{
		tokens: map[string]string{
			"user-tok":  "visitor@example.org",
			"admin-tok": "matthieu.devin@gmail.com",
		},
	}

	srv := New(cat, Config{
		CatalogOpts: catOpts,
		LogDir:      logDir,
		Verifier:    verifier,
	})
	handler := srv.Handler()

	// 1. Static /app/ routes
	for _, tc := range []struct{ url, body, mime string }{
		{"/app/", "react entry", "text/html"},
		{"/app/react-test.html?v=123", "react test entry", "text/html"},
		{"/app/react-test/assets/test.js", "react asset", "application/javascript"},
		{"/app/nonexistent/route", "react entry", "text/html"},
	} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.url, nil))
		if w.Code != http.StatusOK || w.Body.String() != tc.body || !strings.HasPrefix(w.Header().Get("Content-Type"), tc.mime) {
			t.Fatalf("%s: status=%d type=%s body=%q", tc.url, w.Code, w.Header().Get("Content-Type"), w.Body.String())
		}
	}

	// 2. Unauthenticated API request rejected
	{
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/db/q?q=albums:&kind=album", nil))
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for unauthenticated request, got %d", w.Code)
		}
	}

	// 3. Authenticated photo search & album search
	var photos []catalog.JsonImage
	{
		req := httptest.NewRequest(http.MethodGet, "/db/q?q=2025", nil)
		req.Header.Set("Authorization", "Bearer user-tok")
		req.Header.Set("Origin", "https://toutizes.com")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on /db/q, got %d", w.Code)
		}
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "https://toutizes.com" {
			t.Fatalf("expected CORS origin https://toutizes.com, got %q", got)
		}
		if err := json.Unmarshal(w.Body.Bytes(), &photos); err != nil {
			t.Fatal(err)
		}
		if len(photos) != 2 {
			t.Fatalf("expected 2 photos, got %d", len(photos))
		}
	}

	var albums []catalog.JsonDirectory
	{
		req := httptest.NewRequest(http.MethodGet, "/db/q?q=albums:&kind=album", nil)
		req.Header.Set("Authorization", "Bearer user-tok")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on /db/q?kind=album, got %d", w.Code)
		}
		if err := json.Unmarshal(w.Body.Bytes(), &albums); err != nil {
			t.Fatal(err)
		}
		if len(albums) != 1 || albums[0].Id != "2025/2025-06-01" || albums[0].Nimgs != 2 {
			t.Fatalf("unexpected albums response: %+v", albums)
		}
	}

	// 4. Thumbnail by ID serves mini directly
	{
		thumbURL := fmt.Sprintf("/db/thumb/%d", photos[0].Id)
		req := httptest.NewRequest(http.MethodGet, thumbURL, nil)
		req.Header.Set("Authorization", "Bearer user-tok")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "image/jpeg" {
			t.Fatalf("expected 200 image/jpeg on %s, got %d (%s)", thumbURL, w.Code, w.Header().Get("Content-Type"))
		}
	}

	// 5. Mini, Midi, Maxi serving
	for _, size := range []string{"mini", "midi", "maxi"} {
		imgURL := fmt.Sprintf("/db/%s/%s/%s", size, photos[0].Ad, photos[0].In)
		req := httptest.NewRequest(http.MethodGet, imgURL, nil)
		req.Header.Set("Authorization", "Bearer user-tok")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK || w.Header().Get("Content-Type") != "image/jpeg" {
			t.Fatalf("expected 200 image/jpeg on %s, got %d", imgURL, w.Code)
		}
	}

	// 6. Zip download
	{
		req := httptest.NewRequest(http.MethodGet, "/db/viewer?command=download&q=2025&s=M", nil)
		req.Header.Set("Authorization", "Bearer user-tok")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 on zip download, got %d", w.Code)
		}
		zr, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
		if err != nil || len(zr.File) != 2 {
			t.Fatalf("expected valid zip with 2 files, got %v (err=%v)", zr, err)
		}
	}

	// 7. Admin user-queries endpoint (403 for non-admin, 200 for admin)
	logContent := `2025/06/17 12:00:00 Query from alex@gmail.com: "beach" (kind: )
2025/06/17 12:01:00 Query from alex@example.org: "mountains" (kind: )
2025/06/17 12:02:00 Query from alex@example.org: "mountains" (kind: )
2025/06/17 12:03:00 Query from family+photos@example.co.uk: "vacances" (kind: album)
2025/06/17 12:04:00 Query from alex@gmail.com: "albums:" (kind: album)
`
	if err := os.WriteFile(filepath.Join(logDir, "aserve.log"), []byte(logContent), 0644); err != nil {
		t.Fatal(err)
	}
	{
		req := httptest.NewRequest(http.MethodGet, "/db/user-queries", nil)
		req.Header.Set("Authorization", "Bearer user-tok")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 for non-admin user-queries, got %d", w.Code)
		}
	}
	{
		req := httptest.NewRequest(http.MethodGet, "/db/user-queries", nil)
		req.Header.Set("Authorization", "Bearer admin-tok")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for admin user-queries, got %d", w.Code)
		}
		var res AllUserQueriesResult
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatal(err)
		}
		if len(res.Users) != 3 || len(res.Users["alex@example.org"]) != 1 {
			t.Fatalf("unexpected user queries result: %+v", res.Users)
		}
	}

	// 8. LocalHandler allows unauthenticated reload on localhost :8081
	{
		writeTestJPEG(t, filepath.Join(origRoot, "2025/2025-06-01/photo3.jpg"), 800, 600)
		future := time.Now().Add(2 * time.Second)
		_ = os.Chtimes(filepath.Join(origRoot, "2025/2025-06-01"), future, future)

		localHandler := srv.LocalHandler()
		w := httptest.NewRecorder()
		localHandler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/db/viewer?command=reload", nil))
		if w.Code != http.StatusOK || w.Body.String() != "Reloaded" {
			t.Fatalf("expected Reloaded on local handler, got %d %q", w.Code, w.Body.String())
		}
		if got := len(srv.Catalog().QueryPhotos("2025")); got != 3 {
			t.Fatalf("expected 3 photos after reload, got %d", got)
		}
	}
}
