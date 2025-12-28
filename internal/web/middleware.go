package web

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// loggingMiddleware logs HTTP requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Log request
		log.Printf("%s %s", r.Method, r.URL.Path)

		// Call next handler
		next.ServeHTTP(w, r)

		// Log completion
		duration := time.Since(start)
		log.Printf("Completed %s %s in %v", r.Method, r.URL.Path, duration)
	})
}

// localhostOnlyMiddleware ensures requests only come from localhost.
func localhostOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract host from RemoteAddr
		host := r.RemoteAddr
		if idx := strings.LastIndex(host, ":"); idx != -1 {
			host = host[:idx]
		}

		// Allow localhost, 127.0.0.1, ::1, and empty (Unix socket)
		if host == "" || host == "127.0.0.1" || host == "localhost" || host == "::1" || host == "[::1]" {
			next.ServeHTTP(w, r)
			return
		}

		// Reject non-localhost requests
		log.Printf("Rejected non-localhost request from %s", r.RemoteAddr)
		http.Error(w, "Forbidden: Only localhost access allowed", http.StatusForbidden)
	})
}

// securityHeadersMiddleware adds security headers.
func securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline'; script-src 'self' 'unsafe-inline'")

		next.ServeHTTP(w, r)
	})
}
