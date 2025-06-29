package main

import (
	"flag"
	"log"
	"path/filepath"
	"os"

	"htmlnojs/routebuilder"
	"htmlnojs/server"
	"htmlnojs/setup"
)

func main() {
	directory := flag.String("directory", ".", "Project directory to serve")
	port := flag.Int("port", 8080, "Server port")
	fastapiPort := flag.Int("fastapi-port", 8081, "FastAPI server port")
	flag.Parse()

	log.SetOutput(os.Stdout)
	log.Printf("Starting HTMLnoJS server for: %s", *directory)

	config := &setup.Config{
		ProjectDir:   *directory,
		PyHTMXDir:    filepath.Join(*directory, "py_htmx"),
		CSSDir:       filepath.Join(*directory, "css"),
		TemplatesDir: filepath.Join(*directory, "templates"),
	}

	fileSet, err := config.GlobFiles()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Discovered %d HTML, %d CSS, %d Python files",
		len(fileSet.TemplateFiles), len(fileSet.CSSFiles), len(fileSet.PyHTMXFiles),
	)

	routeBuilder := routebuilder.NewAllRoutesBuilder(
		config.TemplatesDir,
		config.CSSDir,
		config.PyHTMXDir,
		*fastapiPort,
	)

	routes, err := routeBuilder.BuildAllRoutes(
		fileSet.TemplateFiles,
		fileSet.CSSFiles,
		fileSet.PyHTMXFiles,
	)
	if err != nil {
		log.Fatal(err)
	}

	srv := server.Development().
		Port(*port).
		WithRoutes(routes).
		Build()

	log.Printf("HTMLnoJS server starting at http://localhost:%d", *port)
	log.Printf("FastAPI backend expected at http://localhost:%d", *fastapiPort)
	log.Printf("Route map: http://localhost:%d/_routes", *port)
    log.Printf("Routes.json: http://localhost:%d/_routes.json", *port)
	log.Printf("Health check: http://localhost:%d/health", *port)
	log.Printf("Press Ctrl+C to stop")

	if err := srv.StartWithGracefulShutdown(); err != nil {
		log.Fatal(err)
	}
}
