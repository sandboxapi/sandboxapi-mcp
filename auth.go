package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/server"
)

// contextKey is the type for context keys in this package.
type contextKey string

// userContextKey is used to store the authenticated user in the request context.
const userContextKey contextKey = "mcp_user"

// APIKeyAuthFunc returns an HTTPContextFunc that validates Bearer tokens
// against the provided API key.
func APIKeyAuthFunc(apiKey string) server.HTTPContextFunc {
	return func(ctx context.Context, r *http.Request) context.Context {
		if apiKey == "" {
			return context.WithValue(ctx, userContextKey, "anonymous")
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			return ctx
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token == auth || token != apiKey {
			return ctx
		}

		return context.WithValue(ctx, userContextKey, "api_key_user")
	}
}

// AuthMiddleware returns an http.Handler that rejects requests without a
// valid Bearer token before they reach the MCP server.
func AuthMiddleware(apiKey string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiKey == "" {
			next.ServeHTTP(w, r)
			return
		}

		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		if token == auth || token != apiKey {
			http.Error(w, `{"error":"invalid API key"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
