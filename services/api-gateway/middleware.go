package main

import (
	"context"
	"domino/shared/jwt"
	"fmt"
	"net/http"
	"strings"
)

// enableCORS wraps a handler with CORS headers and handles OPTIONS preflight
// The browser sends an OPTIONS preflight before every cross-origin POST.
// Without a 200 response to OPTIONS, the actual POST is never sent
func enableCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		handler(w, r)
	}
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// (Authorization, "Bearer <token>")
		authHeader := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(authHeader, "Bearer ")
		if !ok || token == "" {
			fmt.Println("couldnt extract jwt token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// <token> -> claims
		claims, err := jwt.Parse(token)
		if err != nil {
			fmt.Println("couldnt parse jwt token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Add to context
		// Note r.WithContext(ctx), not mutating r.Context() in place request contexts are immutable
		ctx := context.WithValue(r.Context(), "userID", claims.UserID)
		ctx = context.WithValue(ctx, "userType", claims.UserType)
		next(w, r.WithContext(ctx))
	}
}
