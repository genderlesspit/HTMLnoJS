package setup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type PyHTMXHandler struct {
	Name       string
	FilePath   string
	Functions  []string
	APIRoute   string
}

type PyHTMXValidator struct {
	config   *Config
	handlers []PyHTMXHandler
	pyCount  int
}

// NewPyHTMXValidator creates a new Python HTMX validator
func NewPyHTMXValidator(config *Config) *PyHTMXValidator {
	return &PyHTMXValidator{
		config:   config,
		handlers: make([]PyHTMXHandler, 0),
	}
}

// GetHandlers returns the discovered Python HTMX handlers
func (p *PyHTMXValidator) GetHandlers() []PyHTMXHandler {
	return p.handlers
}

// GetPyCount returns the number of valid Python files found
func (p *PyHTMXValidator) GetPyCount() int {
	return p.pyCount
}

// ValidatePyHTMXDir ensures py_htmx directory contains valid Python files
// and provisions a README if it doesn't exist
func (p *PyHTMXValidator) ValidatePyHTMXDir() error {
	if err := p.provisionReadme(); err != nil {
		return err
	}

	if err := p.scanPythonFiles(); err != nil {
		return err
	}

	p.logResults()
	return nil
}

func (p *PyHTMXValidator) provisionReadme() error {
	readmePath := filepath.Join(p.config.PyHTMXDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		content, err := GetPyHTMXREADME()
		if err != nil {
			return fmt.Errorf("failed to load Python HTMX README template: %w", err)
		}

		if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create py_htmx README.md: %w", err)
		}
		log.Printf("Created README.md in py_htmx directory")
	}
	return nil
}

func (p *PyHTMXValidator) scanPythonFiles() error {
	return filepath.Walk(p.config.PyHTMXDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".py") {
			return nil
		}

		// Skip README files
		if strings.Contains(strings.ToLower(info.Name()), "readme") {
			return nil
		}

		p.addPythonHandler(path, info.Name())
		return nil
	})
}

func (p *PyHTMXValidator) addPythonHandler(filePath, filename string) {
	name := strings.TrimSuffix(filename, ".py")

	// Create API route based on file path relative to py_htmx directory
	relPath, _ := filepath.Rel(p.config.PyHTMXDir, filePath)
	apiRoute := "/api/" + strings.TrimSuffix(relPath, ".py")
	apiRoute = strings.ReplaceAll(apiRoute, "\\", "/") // Handle Windows paths

	handler := PyHTMXHandler{
		Name:      name,
		FilePath:  filePath,
		Functions: []string{}, // TODO: Parse file to extract htmx_ functions
		APIRoute:  apiRoute,
	}

	p.handlers = append(p.handlers, handler)
	p.pyCount++
}

func (p *PyHTMXValidator) logResults() {
	log.Printf("PyHTMX directory validated: %d Python files found", p.pyCount)

	if p.pyCount == 0 {
		log.Printf("WARNING: No Python files found in py_htmx directory")
	}

	for _, handler := range p.handlers {
		log.Printf("Discovered Python handler: %s -> %s", handler.APIRoute, handler.Name)
	}
}