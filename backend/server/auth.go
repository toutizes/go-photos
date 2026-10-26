package server

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
)

type contextKey string

const userEmailKey contextKey = "userEmail"

// TokenVerifier verifies a bearer token and returns the authenticated user's email.
type TokenVerifier interface {
	VerifyEmail(ctx context.Context, idToken string) (string, error)
}

// FirebaseVerifier adapts a Firebase *auth.Client to TokenVerifier.
type FirebaseVerifier struct {
	Client *auth.Client
}

// VerifyEmail verifies the Firebase ID token and extracts the "email" claim.
func (fv *FirebaseVerifier) VerifyEmail(ctx context.Context, idToken string) (string, error) {
	if fv == nil || fv.Client == nil {
		return "", errors.New("auth client not configured")
	}
	tok, err := fv.Client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", err
	}
	email, _ := tok.Claims["email"].(string)
	if email == "" {
		return "", errors.New("token missing email claim")
	}
	return email, nil
}

// UserEmail returns the authenticated user email stored in ctx, if any.
func UserEmail(ctx context.Context) string {
	if v, ok := ctx.Value(userEmailKey).(string); ok {
		return v
	}
	if v, ok := ctx.Value("userEmail").(string); ok {
		return v
	}
	return ""
}

// WithUserEmail returns a copy of ctx carrying the authenticated user's email.
func WithUserEmail(ctx context.Context, email string) context.Context {
	ctx = context.WithValue(ctx, userEmailKey, email)
	return context.WithValue(ctx, "userEmail", email)
}

var allowedOrigins = []string{
	"http://localhost",
	"http://localhost:3000",
	"https://toutizes.com",
}

// AddCorsHeaders writes CORS response headers for allowed origins.
func AddCorsHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" {
		origin = r.Header.Get("Referer")
	}
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

// AuthMiddleware validates the Firebase Bearer token on incoming requests.
func AuthMiddleware(verifier TokenVerifier, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			AddCorsHeaders(w, r)
			w.WriteHeader(http.StatusOK)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No authorization token provided", http.StatusUnauthorized)
			return
		}
		idToken := strings.TrimPrefix(authHeader, "Bearer ")
		if verifier == nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		email, err := verifier.VerifyEmail(r.Context(), idToken)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(WithUserEmail(r.Context(), email)))
	}
}
