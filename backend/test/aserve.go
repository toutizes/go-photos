package main

import (
	"context"
	"flag"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
	// "github.com/alexedwards/scs/v2"
)

import (
	"toutizes.com/go-photwo/backend/model"
)

var static_root = flag.String("static_root", "", "path to the static files")
var orig_root = flag.String("orig_root", "", "path to the original images")
var root = flag.String("root", "", "path to the database index, mini, etc")
var url_prefix = flag.String("url_prefix", "/db", "prefix for the urls.")
var num_cpu = flag.Int("num_cpu", 0, "Number of CPUs to use.  0 means MAXPROC.")
var update_db = flag.Bool("update_db", true, "If true update the database files.")
var force_reload = flag.Bool("force_reload", false, "If true force a reload of images.")
var use_https = flag.Bool("use_https", false, "If true listen for HTTPS in 443.")
var firebase_creds = flag.String("firebase_creds", "", "Path to the Firebase service account credentials JSON file")

var authClient *auth.Client

// var sessionManager *scs.SessionManager
// var cookieSalt = "da89HIuneDMBa8eThg-9VYcDScApDUKIXaiFXcbvMys"

func Log(n string, r *http.Request) {
	log.Printf("%s: %s %s\n", n, r.RemoteAddr, r.URL)
}

func LogHandler(n string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Log(n, r)
		handler.ServeHTTP(w, r)
	})
}

// Add CORS headers to all responses
func AddCorsHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// If no Origin header, fall back to the Referer
		origin = r.Header.Get("Referer")
	}
	
	// Allow both localhost and toutizes.com
	allowedOrigins := []string{
		"http://localhost",
		"http://localhost:3000",
		"https://toutizes.com",
	}
	
	// Check if the origin is allowed
	for _, allowed := range allowedOrigins {
		if strings.HasPrefix(origin, allowed) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			break
		}
	}
	
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Origin, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

// Add function to get content type
func getContentType(path string) string {
	ext := filepath.Ext(path)
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".svg":
		return "image/svg+xml"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".ico":
		return "image/x-icon"
	default:
		if ct := mime.TypeByExtension(ext); ct != "" {
			return ct
		}
		return "application/octet-stream"
	}
}

// Flutter web app handler
func handleFlutterApp(w http.ResponseWriter, r *http.Request) {
	// The path to your built Flutter web files
	webRoot := *static_root + "/flutter"
	// Get the requested path and remove /app/ prefix
	path := strings.TrimPrefix(r.URL.Path, "/app/")

	if path == "" {
		path = "index.html"
	}

	// Create the full file path
	filePath := filepath.Join(webRoot, path)

	// Prevent directory traversal
	if !strings.HasPrefix(filepath.Clean(filePath), webRoot) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// For SPA routing, serve index.html for non-existent files
		filePath = filepath.Join(webRoot, "index.html")
	}

	// Set content type and other headers
	w.Header().Set("Content-Type", getContentType(filePath))
	AddCorsHeaders(w, r)

	// Cache static assets but not index.html
	if !strings.HasSuffix(filePath, "index.html") {
		w.Header().Set("Cache-Control", "public, max-age=31536000")
	} else {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	}

	http.ServeFile(w, r, filePath)
}

// AuthMiddleware verifies the Firebase ID token
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle CORS preflight requests
		if r.Method == "OPTIONS" {
			AddCorsHeaders(w, r)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		// log.Printf("Auth header received: %s", authHeader)
		if authHeader == "" {
			// log.Printf("No authorization header found. All headers: %v", r.Header)
			http.Error(w, "No authorization token provided", http.StatusUnauthorized)
			return
		}

		// Remove "Bearer " prefix if present
		idToken := strings.TrimPrefix(authHeader, "Bearer ")

		// Verify the Firebase ID token
		token, err := authClient.VerifyIDToken(r.Context(), idToken)
		if err != nil {
			// log.Printf("Error verifying ID token: %v", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user email to context
		userEmail := token.Claims["email"].(string)
		ctx := context.WithValue(r.Context(), "userEmail", userEmail)
		
		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func main() {
	flag.Parse()

	if *num_cpu == 0 {
		*num_cpu = runtime.NumCPU()
	}
	log.Printf("Cpus: %d\n", runtime.GOMAXPROCS(*num_cpu))

	if *static_root == "" {
		log.Fatal("Must pass --static_root")
	}
	if *orig_root == "" {
		log.Fatal("Must pass --orig_root")
	}
	if *root == "" {
		log.Fatal("Must pass --root")
	}
	if *firebase_creds == "" {
		log.Fatal("Must pass --firebase_creds")
	}

	// Initialize Firebase Admin SDK
	opt := option.WithCredentialsFile(*firebase_creds)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
	}

	// Initialize Firebase Auth client
	authClient, err = app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Error initializing Firebase Auth client: %v", err)
	}

	db := model.NewDatabase2(*orig_root, *root, *static_root)
	db.Load(*update_db, *update_db, *force_reload)
	log.Printf("Serving...")

	// // Initialize the session manager
	// sessionManager = scs.New()
	// sessionManager.Store = scs.NewCookieStore([]byte(cookieSalt))

	var mux = http.NewServeMux()
	
	// Wrap API endpoints with AuthMiddleware
	mux.HandleFunc(*url_prefix+"/q",
		func(w http.ResponseWriter, r *http.Request) {
			// Log("TT db/q", r)
			AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				AddCorsHeaders(w, r)
				model.HandleQuery(w, r, db)
			})(w, r)
		})
	mux.HandleFunc(*url_prefix+"/montage/",
		func(w http.ResponseWriter, r *http.Request) {
			// Log("/montage", r)
			AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				AddCorsHeaders(w, r)
				model.HandleMontage2(w, r, db)
			})(w, r)
		})
	mux.HandleFunc(*url_prefix+"/viewer",
		func(w http.ResponseWriter, r *http.Request) {
			Log("/viewer", r)
			AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				AddCorsHeaders(w, r)
				model.HandleCommands(w, r, db)
			})(w, r)
		})
	mux.HandleFunc(*url_prefix+"/mini/",
		func(w http.ResponseWriter, r *http.Request) {
			Log("/mini", r)
			AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				AddCorsHeaders(w, r)
				model.HandleFile(w, r, *url_prefix, *root)
			})(w, r)
		})
	mux.HandleFunc(*url_prefix+"/midi/",
		func(w http.ResponseWriter, r *http.Request) {
			// Log("/midi", r)
			AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				AddCorsHeaders(w, r)
				model.HandleFile(w, r, *url_prefix, *root)
			})(w, r)
		})
	mux.HandleFunc(*url_prefix+"/maxi/",
		func(w http.ResponseWriter, r *http.Request) {
			Log("/maxi", r)
			AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				AddCorsHeaders(w, r)
				model.HandleFile(w, r, *url_prefix+"/maxi/", *orig_root)
			})(w, r)
		})
	// mux.Handle("/",
	//   LogHandler(
	//     "/ default",
	//     http.FileServer(http.Dir(*static_root + "/"))))
	// mux.Handle(
	// 	"/db/",
	// 	LogHandler(
	// 		"/db files",
	// 		http.FileServer(http.Dir(*static_root))))
	mux.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			// http.Redirect(w, r, "/db/pic.html", 301)
			http.Redirect(w, r, "/app/", 301)
		})

	// Add the Flutter web app handler
	mux.HandleFunc("/app/",
		func(w http.ResponseWriter, r *http.Request) {
			Log("/app", r)
			handleFlutterApp(w, r)
		})

	// Always serve on http:8081, useful for access from localhost.
	go http.ListenAndServe(":8081", mux)

	// Serve images on https:8443/http:8080 (redirect)
	if *use_https {
		// Redirect HTTP to HTTPS
		var http_mux = http.NewServeMux()
		http_mux.HandleFunc("/",
			func(w http.ResponseWriter, r *http.Request) {
				http.Redirect(w, r, "https://toutizes.com"+r.RequestURI, 301)
			})
		go http.ListenAndServe(":8080", http_mux)

		// Listen on HTTPS.
		err := http.ListenAndServeTLS(":8443",
			"/etc/letsencrypt/live/toutizes.com/fullchain.pem",
			"/etc/letsencrypt/live/toutizes.com/privkey.pem",
			mux)
		log.Printf("Web server error: %s\n", err)
	} else {
		err := http.ListenAndServe(":8080", mux)
		log.Printf("Web server error: %s\n", err)
	}
}
