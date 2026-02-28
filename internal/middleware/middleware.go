package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

var k keyfunc.Keyfunc

type apiKeyRoundTripper struct {
	apiKey string
	next   http.RoundTripper
}

func (t *apiKeyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone request to avoid modifying original
	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("apikey", t.apiKey)
	reqClone.Header.Set("Authorization", "Bearer "+t.apiKey) // Some Supabase configs require both
	return t.next.RoundTrip(reqClone)
}

func InitJWKS() {
	supabaseURL := os.Getenv("SUPABASE_URL")
	anonKey := os.Getenv("SUPABASE_PUBLISHABLE")
	if supabaseURL == "" || anonKey == "" {
		log.Fatalf("Environment variables not set: SUPABASE_URL=%q, SUPABASE_PUBLISHABLE=%q", supabaseURL, anonKey)
	}

	// Sanitize URL: Remove trailing slash and /auth/v1 if present
	baseURL := strings.TrimSuffix(supabaseURL, "/")
	baseURL = strings.TrimSuffix(baseURL, "/auth/v1")
	jwksURL := baseURL + "/auth/v1/jwks"

	log.Printf("Initializing JWKS. Original URL: %q, Sanitized Base: %q, Final JWKS URL: %q", supabaseURL, baseURL, jwksURL)

	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &apiKeyRoundTripper{
			apiKey: anonKey,
			next:   http.DefaultTransport,
		},
	}

	storage, err := jwkset.NewStorageFromHTTP(jwksURL, jwkset.HTTPClientStorageOptions{
		Client: client,
	})
	if err != nil {
		log.Fatalf("Failed to create JWK storage for URL %q: %v", jwksURL, err)
	}

	k, err = keyfunc.New(keyfunc.Options{
		Storage: storage,
	})
	if err != nil {
		log.Fatalf("Failed to create keyfunc: %v", err)
	}
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "unauthorized", "message": "Missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"error": "unauthorized", "message": "Invalid Authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, k.Keyfunc)

		if err != nil {
			fmt.Printf("JWT Error: %v\n", err)
			http.Error(w, `{"error": "unauthorized", "message": "Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			http.Error(w, `{"error": "unauthorized", "message": "Invalid token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error": "unauthorized", "message": "Invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		userID, ok := claims["sub"].(string)
		if !ok {
			http.Error(w, `{"error": "unauthorized", "message": "Missing 'sub' claim in token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
