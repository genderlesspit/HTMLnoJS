package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"htmlnojs/routebuilder"
)

type Server struct {
	host           string
	port           int
	mux            *http.ServeMux
	server         *http.Server
	routes         *routebuilder.RouteCollection
	middleware     []MiddlewareFunc
	config         ServerConfig
}

type ServerConfig struct {
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	EnableCORS      bool
	EnableLogging   bool
	EnableMetrics   bool
}

type MiddlewareFunc func(http.Handler) http.Handler

// NewServer creates a new HTMLnoJS server
func NewServer(host string, port int) *Server {
	return &Server{
		host:       host,
		port:       port,
		mux:        http.NewServeMux(),
		middleware: make([]MiddlewareFunc, 0),
		config: ServerConfig{
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 30 * time.Second,
			EnableCORS:      true,
			EnableLogging:   true,
			EnableMetrics:   false,
		},
	}
}

// SetConfig updates server configuration
func (s *Server) SetConfig(config ServerConfig) {
	s.config = config
}

// AddMiddleware adds middleware to the server
func (s *Server) AddMiddleware(mw MiddlewareFunc) {
	s.middleware = append(s.middleware, mw)
}

// RegisterRoutes registers all routes from a RouteCollection
func (s *Server) RegisterRoutes(routes *routebuilder.RouteCollection) error {
	log.Printf("Registering %d routes with HTTP server...", routes.Metadata.TotalRoutes)

	s.routes = routes

	// Register HTML routes
	for _, route := range routes.HTMLRoutes {
		handler := s.wrapHandler(route.Handler, route.RequiresAuth)
		s.mux.HandleFunc(route.Route, handler)
		log.Printf("Registered HTML route: %s %s", route.Method, route.Route)
	}

	// Register CSS routes
	for _, route := range routes.CSSRoutes {
		handler := s.wrapStaticHandler(route.Handler)
		s.mux.HandleFunc(route.Route, handler)
		log.Printf("Registered CSS route: %s %s", route.Method, route.Route)
	}

	// Register Python API routes
	for _, route := range routes.PythonRoutes {
		handler := s.wrapAPIHandler(route.Handler, route.RequiresAuth, route.RateLimit, route.CacheTimeout)
		s.mux.HandleFunc(route.Route, handler)
		log.Printf("Registered Python route: %s %s", route.Method, route.Route)
	}

	// Register built-in routes
	s.registerBuiltinRoutes()

	log.Printf("All routes registered successfully!")
	return nil
}

func (s *Server) registerBuiltinRoutes() {
	// Health check endpoint
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","routes":%d}`, s.routes.Metadata.TotalRoutes)
	})

	// Route map endpoint
	s.mux.HandleFunc("/_routes", func(w http.ResponseWriter, r *http.Request) {
		if s.routes == nil {
			http.Error(w, "No routes loaded", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "=== HTMLnoJS Route Map ===\n\n")

		// HTML Routes
		fmt.Fprintf(w, "HTML ROUTES:\n")
		for _, route := range s.routes.HTMLRoutes {
			auth := ""
			if route.RequiresAuth {
				auth = " [AUTH]"
			}
			fmt.Fprintf(w, "  %s %s -> %s%s\n", route.Method, route.Route, route.Name, auth)
		}

		// CSS Routes
		fmt.Fprintf(w, "\nCSS ROUTES:\n")
		for _, route := range s.routes.CSSRoutes {
			fmt.Fprintf(w, "  %s %s -> %s [%s]\n", route.Method, route.Route, route.Name, route.Category)
		}

		// Python Routes
		fmt.Fprintf(w, "\nPYTHON API ROUTES:\n")
		for _, route := range s.routes.PythonRoutes {
			auth := ""
			if route.RequiresAuth {
				auth = " [AUTH]"
			}
			fmt.Fprintf(w, "  %s %s -> %s%s\n", route.Method, route.Route, route.Function, auth)
		}

		// Summary
		fmt.Fprintf(w, "\nSUMMARY: %d total routes\n", s.routes.Metadata.TotalRoutes)
	})

	// Metrics endpoint (if enabled)
	if s.config.EnableMetrics {
		s.mux.HandleFunc("/_metrics", s.handleMetrics)
	}
}

func (s *Server) wrapHandler(handler http.HandlerFunc, requiresAuth bool) http.HandlerFunc {
	wrapped := http.Handler(handler)

	// Apply authentication if required
	if requiresAuth {
		wrapped = http.HandlerFunc(s.authMiddleware(handler))
	}

	// Apply common middleware
	for i := len(s.middleware) - 1; i >= 0; i-- {
		wrapped = s.middleware[i](wrapped)
	}

	// Convert back to HandlerFunc
	return func(w http.ResponseWriter, r *http.Request) {
		wrapped.ServeHTTP(w, r)
	}
}

func (s *Server) wrapStaticHandler(handler http.HandlerFunc) http.HandlerFunc {
	wrapped := http.Handler(handler)

	// Apply static file middleware (caching, compression)
	wrapped = http.HandlerFunc(s.staticMiddleware(handler))

	// Apply CORS if enabled
	if s.config.EnableCORS {
		wrapped = http.HandlerFunc(s.corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
			wrapped.ServeHTTP(w, r)
		}))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		wrapped.ServeHTTP(w, r)
	}
}

func (s *Server) wrapAPIHandler(handler http.HandlerFunc, requiresAuth bool, rateLimit int, cacheTimeout int) http.HandlerFunc {
	wrapped := http.Handler(handler)

	// Apply caching if configured
	if cacheTimeout > 0 {
		wrapped = http.HandlerFunc(s.cacheMiddleware(handler, cacheTimeout))
	}

	// Apply rate limiting if configured
	if rateLimit > 0 {
		wrapped = http.HandlerFunc(s.rateLimitMiddleware(handler, rateLimit))
	}

	// Apply authentication if required
	if requiresAuth {
		wrapped = http.HandlerFunc(s.authMiddleware(handler))
	}

	// Apply API middleware (JSON handling, CORS, etc.)
	wrapped = http.HandlerFunc(s.apiMiddleware(handler))

	return func(w http.ResponseWriter, r *http.Request) {
		wrapped.ServeHTTP(w, r)
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	addr := fmt.Sprintf("%s:%d", s.host, s.port)

	s.server = &http.Server{
		Addr:         addr,
		Handler:      s.mux,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}

	log.Printf("HTMLnoJS server starting on %s", addr)
	log.Printf("Server configuration:")
	log.Printf("  - Read timeout: %v", s.config.ReadTimeout)
	log.Printf("  - Write timeout: %v", s.config.WriteTimeout)
	log.Printf("  - CORS enabled: %v", s.config.EnableCORS)
	log.Printf("  - Logging enabled: %v", s.config.EnableLogging)

	return s.server.ListenAndServe()
}

// StartWithGracefulShutdown starts the server with graceful shutdown handling
func (s *Server) StartWithGracefulShutdown() error {
	// Start server in goroutine
	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Server shutting down...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	// Shutdown server
	if err := s.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	log.Println("Server exited")
	return nil
}

// Stop stops the HTTP server
func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	return s.server.Shutdown(ctx)
}

// GetRoutes returns the registered routes
func (s *Server) GetRoutes() *routebuilder.RouteCollection {
	return s.routes
}

// GetAddr returns the server address
func (s *Server) GetAddr() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

// Middleware implementations

func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement actual authentication
		// For now, just check for a simple auth header
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func (s *Server) staticMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set cache headers for static files
		w.Header().Set("Cache-Control", "public, max-age=31536000")
		next(w, r)
	}
}

func (s *Server) apiMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set API-specific headers
		w.Header().Set("Content-Type", "text/html") // HTMX returns HTML fragments

		if s.config.EnableCORS {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		next(w, r)
	}
}

func (s *Server) cacheMiddleware(next http.HandlerFunc, timeout int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement actual caching logic
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", timeout))
		next(w, r)
	}
}

func (s *Server) rateLimitMiddleware(next http.HandlerFunc, limit int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement actual rate limiting
		// For now, just add a header indicating the limit
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
		next(w, r)
	}
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "# HTMLnoJS Server Metrics\n")
	fmt.Fprintf(w, "total_routes %d\n", s.routes.Metadata.TotalRoutes)
	fmt.Fprintf(w, "html_routes %d\n", s.routes.Metadata.HTMLCount)
	fmt.Fprintf(w, "css_routes %d\n", s.routes.Metadata.CSSCount)
	fmt.Fprintf(w, "python_routes %d\n", s.routes.Metadata.PythonCount)
	fmt.Fprintf(w, "auth_required_routes %d\n", s.routes.Metadata.AuthRequired)
}