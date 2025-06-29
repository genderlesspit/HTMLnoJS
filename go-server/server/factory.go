package server

import (
	"time"

	"htmlnojs/routebuilder"
)

// ServerBuilder provides a fluent interface for building servers
type ServerBuilder struct {
	server *Server
}

// NewBuilder creates a new server builder
func NewBuilder() *ServerBuilder {
	return &ServerBuilder{
		server: NewServer("localhost", 6968),
	}
}

// Host sets the server host
func (b *ServerBuilder) Host(host string) *ServerBuilder {
	b.server.host = host
	return b
}

// Port sets the server port
func (b *ServerBuilder) Port(port int) *ServerBuilder {
	b.server.port = port
	return b
}

// WithRoutes sets the routes for the server
func (b *ServerBuilder) WithRoutes(routes *routebuilder.RouteCollection) *ServerBuilder {
	b.server.RegisterRoutes(routes)
	return b
}

// WithTimeout sets various timeout configurations
func (b *ServerBuilder) WithTimeout(read, write, idle, shutdown time.Duration) *ServerBuilder {
	b.server.config.ReadTimeout = read
	b.server.config.WriteTimeout = write
	b.server.config.IdleTimeout = idle
	b.server.config.ShutdownTimeout = shutdown
	return b
}

// EnableCORS enables or disables CORS
func (b *ServerBuilder) EnableCORS(enable bool) *ServerBuilder {
	b.server.config.EnableCORS = enable
	return b
}

// EnableLogging enables or disables request logging
func (b *ServerBuilder) EnableLogging(enable bool) *ServerBuilder {
	b.server.config.EnableLogging = enable
	return b
}

// EnableMetrics enables or disables metrics endpoint
func (b *ServerBuilder) EnableMetrics(enable bool) *ServerBuilder {
	b.server.config.EnableMetrics = enable
	return b
}

// WithMiddleware adds middleware to the server
func (b *ServerBuilder) WithMiddleware(mw MiddlewareFunc) *ServerBuilder {
	b.server.AddMiddleware(mw)
	return b
}

// WithLoggingMiddleware adds request logging middleware
func (b *ServerBuilder) WithLoggingMiddleware() *ServerBuilder {
	b.server.AddMiddleware(LoggingMiddleware)
	return b
}

// WithRecoveryMiddleware adds panic recovery middleware
func (b *ServerBuilder) WithRecoveryMiddleware() *ServerBuilder {
	b.server.AddMiddleware(RecoveryMiddleware)
	return b
}

// Build builds and returns the configured server
func (b *ServerBuilder) Build() *Server {
	return b.server
}

// Preset configurations

// Development returns a server configured for development
func Development() *ServerBuilder {
	return NewBuilder().
		Host("localhost").
		Port(6968).
		WithTimeout(10*time.Second, 10*time.Second, 30*time.Second, 10*time.Second).
		EnableCORS(true).
		EnableLogging(true).
		EnableMetrics(true).
		WithLoggingMiddleware().
		WithRecoveryMiddleware()
}

// Production returns a server configured for production
func Production() *ServerBuilder {
	return NewBuilder().
		Host("0.0.0.0").
		Port(80).
		WithTimeout(15*time.Second, 15*time.Second, 60*time.Second, 30*time.Second).
		EnableCORS(false).
		EnableLogging(true).
		EnableMetrics(false).
		WithRecoveryMiddleware()
}

// Testing returns a server configured for testing
func Testing() *ServerBuilder {
	return NewBuilder().
		Host("localhost").
		Port(0). // Random port
		WithTimeout(5*time.Second, 5*time.Second, 10*time.Second, 5*time.Second).
		EnableCORS(true).
		EnableLogging(false).
		EnableMetrics(false)
}

// QuickStart creates a development server with routes
func QuickStart(routes *routebuilder.RouteCollection, port int) *Server {
	return Development().
		Port(port).
		WithRoutes(routes).
		Build()
}