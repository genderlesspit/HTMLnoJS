package setup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type HTMLRoute struct {
	Name     string
	FilePath string
	Route    string
}

type HTMLValidator struct {
	config    *Config
	routes    []HTMLRoute
	htmlCount int
}

// NewHTMLValidator creates a new HTML validator for the given config
func NewHTMLValidator(config *Config) *HTMLValidator {
	return &HTMLValidator{
		config: config,
		routes: make([]HTMLRoute, 0),
	}
}

// GetRoutes returns the discovered HTML routes
func (h *HTMLValidator) GetRoutes() []HTMLRoute {
	return h.routes
}

// GetHTMLCount returns the number of valid HTML files found
func (h *HTMLValidator) GetHTMLCount() int {
	return h.htmlCount
}

// ValidateTemplatesDir ensures templates directory only contains HTML files
// and provisions a README if it doesn't exist
func (h *HTMLValidator) ValidateTemplatesDir() error {
	if err := h.provisionReadme(); err != nil {
		return err
	}

	if err := h.scanTemplates(); err != nil {
		return err
	}

	h.logResults()
	return nil
}

func (h *HTMLValidator) provisionReadme() error {
	readmePath := filepath.Join(h.config.TemplatesDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		content, err := GetTemplatesREADME()
		if err != nil {
			return fmt.Errorf("failed to load templates README template: %w", err)
		}

		if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create README.md: %w", err)
		}
		log.Printf("Created README.md in project directory")
	}
	return nil
}

func (h *HTMLValidator) scanTemplates() error {
	files, err := filepath.Glob(filepath.Join(h.config.TemplatesDir, "*"))
	if err != nil {
		return fmt.Errorf("failed to scan templates directory: %w", err)
	}

	var invalidFiles []string

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil || info.IsDir() {
			continue
		}

		filename := filepath.Base(file)
		ext := strings.ToLower(filepath.Ext(filename))

		if ext == ".html" || ext == ".htm" {
			h.addHTMLRoute(file, filename)
		} else {
			invalidFiles = append(invalidFiles, filename)
		}
	}

	h.logInvalidFiles(invalidFiles)
	return nil
}

func (h *HTMLValidator) addHTMLRoute(filePath, filename string) {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	route := "/" + name

	if name == "index" {
		route = "/"
	}

	htmlRoute := HTMLRoute{
		Name:     name,
		FilePath: filePath,
		Route:    route,
	}

	h.routes = append(h.routes, htmlRoute)
	h.htmlCount++
}

func (h *HTMLValidator) logInvalidFiles(invalidFiles []string) {
	for _, file := range invalidFiles {
		log.Printf("WARNING: Non-HTML file found in templates directory: %s", file)
	}

	if len(invalidFiles) > 0 {
		log.Printf("WARNING: Templates directory should only contain HTML files. Found %d invalid files.", len(invalidFiles))
	}
}

func (h *HTMLValidator) logResults() {
	log.Printf("Templates directory validated: %d HTML files found", h.htmlCount)

	if h.htmlCount == 0 {
		log.Printf("WARNING: No HTML files found in templates directory")
	}

	for _, route := range h.routes {
		log.Printf("Discovered HTML route: %s -> %s", route.Route, route.Name)
	}
}