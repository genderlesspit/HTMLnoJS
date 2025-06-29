package server

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %v %s",
			r.Method,
			r.URL.Path,
			wrapped.statusCode,
			duration,
			r.UserAgent(),
		)
	})
}

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds common security headers
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// CompressionMiddleware adds gzip compression (placeholder)
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement actual compression
		// This would check Accept-Encoding header and compress responses
		next.ServeHTTP(w, r)
	})
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		w.Header().Set("X-Request-ID", requestID)
		r.Header.Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r)
	})
}

// HTMXMiddleware adds HTMX-specific headers and handling
func HTMXMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			// Add HTMX-specific headers
			w.Header().Set("Vary", "HX-Request")

			// Set cache policy for HTMX requests
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	// Simple timestamp-based ID
	// In production, you'd want something more robust like UUID
	return time.Now().Format("20060102150405.000000")
}

// MethodOverrideMiddleware allows method override via _method form parameter
func MethodOverrideMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			if method := r.FormValue("_method"); method != "" {
				switch method {
				case "PUT", "PATCH", "DELETE":
					r.Method = method
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

// CSRFMiddleware provides basic CSRF protection (placeholder)
func CSRFMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement actual CSRF protection
		// This would generate and validate CSRF tokens
		next.ServeHTTP(w, r)
	})
}

// TimeoutMiddleware wraps requests with a timeout
func TimeoutMiddleware(timeout time.Duration) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.TimeoutHandler(next, timeout, "Request Timeout")
	}
}

// ConditionalMiddleware applies middleware only if condition is met
func ConditionalMiddleware(condition func(*http.Request) bool, middleware MiddlewareFunc) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if condition(r) {
				middleware(next).ServeHTTP(w, r)
			} else {
				next.ServeHTTP(w, r)
			}
		})
	}
}

// Common condition functions

// IsHTMXRequest returns true if the request is from HTMX
func IsHTMXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// IsJSONRequest returns true if the request accepts JSON
func IsJSONRequest(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return accept == "application/json" ||
		   r.Header.Get("Content-Type") == "application/json"
}

// IsStaticRequest returns true if the request is for static assets
func IsStaticRequest(r *http.Request) bool {
	path := r.URL.Path
	return path == "/css/" || path == "/js/" || path == "/images/" || path == "/favicon.ico"
}