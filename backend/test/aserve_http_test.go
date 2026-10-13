package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReactTrialAlongsideFlutter(t *testing.T) {
	oldRoot := *static_root
	*static_root = t.TempDir()
	t.Cleanup(func() { *static_root = oldRoot })
	for name, body := range map[string]string{
		"index.html":                "flutter entry",
		"react-test.html":           "react entry",
		"react-test/assets/test.js": "react asset",
	} {
		path := filepath.Join(*static_root, "flutter", name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	for _, tc := range []struct{ url, body, mime string }{
		{"/app/", "flutter entry", "text/html"},
		{"/app/react-test.html?v=123", "react entry", "text/html"},
		{"/app/react-test/assets/test.js", "react asset", "application/javascript"},
	} {
		w := httptest.NewRecorder()
		handleFlutterApp(w, httptest.NewRequest(http.MethodGet, tc.url, nil))
		if w.Code != http.StatusOK || w.Body.String() != tc.body || !strings.HasPrefix(w.Header().Get("Content-Type"), tc.mime) {
			t.Fatalf("%s: status=%d type=%s body=%q", tc.url, w.Code, w.Header().Get("Content-Type"), w.Body.String())
		}
	}
}

func TestPhotoAPIRequiresAuthentication(t *testing.T) {
	called := false
	handler := AuthMiddleware(func(w http.ResponseWriter, r *http.Request) { called = true })
	w := httptest.NewRecorder()
	handler(w, httptest.NewRequest(http.MethodGet, "/db/q?q=albums:&kind=album", nil))
	if w.Code != http.StatusUnauthorized || called {
		t.Fatalf("unauthenticated request: status=%d handler called=%v", w.Code, called)
	}
}
