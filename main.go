package main

import (
	"flag"
	"log"
	"path/filepath"

	"htmlnojs/routebuilder"
	"htmlnojs/server"
	"htmlnojs/setup"
)

func main() {
	directory := flag.String("directory", ".", "Project directory to serve")
	port := flag.Int("port", 8080, "Server port")
	flag.Parse()

	log.Printf("Starting HTMLnoJS server for: %s", *directory)

	// Get project configuration
	config := &setup.Config{
		ProjectDir:   *directory,
		PyHTMXDir:    filepath.Join(*directory, "py_htmx"),
		CSSDir:       filepath.Join(*directory, "css"),
		TemplatesDir: filepath.Join(*directory, "templates"),
	}

	// Discover files
	fileSet, err := config.GlobFiles()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Discovered %d HTML, %d CSS, %d Python files",
		len(fileSet.TemplateFiles), len(fileSet.CSSFiles), len(fileSet.PyHTMXFiles))

	// Build all routes
	routeBuilder := routebuilder.NewAllRoutesBuilder(
		config.TemplatesDir,
		config.CSSDir,
		config.PyHTMXDir,
	)

	routes, err := routeBuilder.BuildAllRoutes(
		fileSet.TemplateFiles,
		fileSet.CSSFiles,
		fileSet.PyHTMXFiles,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create and start server
	srv := server.Development().
		Port(*port).
		WithRoutes(routes).
		Build()

	log.Printf("🚀 HTMLnoJS server starting at http://localhost:%d", *port)
	log.Printf("📊 Route map: http://localhost:%d/_routes", *port)
	log.Printf("💚 Health check: http://localhost:%d/health", *port)
	log.Printf("Press Ctrl+C to stop")

	// Start server with graceful shutdown
	if err := srv.StartWithGracefulShutdown(); err != nil {
		log.Fatal(err)
	}
}