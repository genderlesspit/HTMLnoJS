package routebuilder

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type HTMLRoute struct {
	Name         string
	FilePath     string
	Route        string
	Method       string
	Handler      http.HandlerFunc
	Template     string
	CSSFiles     []string
	RequiresAuth bool
	Metadata     map[string]interface{}
}

type HTMLRouteBuilder struct {
	templatesDir string
	cssFiles     []string
	routes       []HTMLRoute
}

// NewHTMLRouteBuilder creates a new HTML route builder
func NewHTMLRouteBuilder(templatesDir string, cssFiles []string) *HTMLRouteBuilder {
	return &HTMLRouteBuilder{
		templatesDir: templatesDir,
		cssFiles:     cssFiles,
		routes:       make([]HTMLRoute, 0),
	}
}

// BuildRoutes discovers and builds HTML template routes
func (h *HTMLRouteBuilder) BuildRoutes(htmlFiles []string) ([]HTMLRoute, error) {
	for _, filePath := range htmlFiles {
		if !strings.HasSuffix(strings.ToLower(filePath), ".html") {
			continue
		}

		route, err := h.buildHTMLRoute(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to build route for %s: %w", filePath, err)
		}

		h.routes = append(h.routes, route)
	}

	return h.routes, nil
}

func (h *HTMLRouteBuilder) buildHTMLRoute(filePath string) (HTMLRoute, error) {
	filename := filepath.Base(filePath)
	name := strings.TrimSuffix(filename, ".html")

	// Determine route path
	routePath := "/" + name
	if name == "index" {
		routePath = "/"
	}

	// Check for special route patterns
	method := "GET"
	requiresAuth := false
	metadata := make(map[string]interface{})

	// Parse special naming conventions
	if strings.Contains(name, "_auth") {
		requiresAuth = true
		metadata["auth_required"] = true
	}

	if strings.Contains(name, "_admin") {
		requiresAuth = true
		metadata["admin_required"] = true
	}

	if strings.HasPrefix(name, "api_") {
		// API endpoints might need different handling
		metadata["is_api"] = true
	}

	// Determine CSS dependencies based on template name
	cssFiles := h.determineCSSFiles(name)

	route := HTMLRoute{
		Name:         name,
		FilePath:     filePath,
		Route:        routePath,
		Method:       method,
		Handler:      h.createTemplateHandler(filePath, cssFiles),
		Template:     filePath,
		CSSFiles:     cssFiles,
		RequiresAuth: requiresAuth,
		Metadata:     metadata,
	}

	return route, nil
}

func (h *HTMLRouteBuilder) determineCSSFiles(templateName string) []string {
	var relevantCSS []string

	// Always include global CSS
	for _, cssFile := range h.cssFiles {
		cssName := strings.ToLower(filepath.Base(cssFile))

		// Include global CSS files
		if strings.Contains(cssName, "global") ||
		   strings.Contains(cssName, "main") ||
		   strings.Contains(cssName, "reset") ||
		   strings.Contains(cssName, "variables") {
			relevantCSS = append(relevantCSS, cssFile)
		}

		// Include template-specific CSS
		if strings.Contains(cssName, templateName) {
			relevantCSS = append(relevantCSS, cssFile)
		}

		// Include component CSS based on template name
		if strings.Contains(templateName, "form") && strings.Contains(cssName, "form") {
			relevantCSS = append(relevantCSS, cssFile)
		}
		if strings.Contains(templateName, "button") && strings.Contains(cssName, "button") {
			relevantCSS = append(relevantCSS, cssFile)
		}
		if strings.Contains(templateName, "nav") && strings.Contains(cssName, "nav") {
			relevantCSS = append(relevantCSS, cssFile)
		}
	}

	return relevantCSS
}

func (h *HTMLRouteBuilder) createTemplateHandler(templatePath string, cssFiles []string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Read the HTML template file
		content, err := os.ReadFile(templatePath)
		if err != nil {
			http.Error(w, "Template not found", http.StatusNotFound)
			return
		}

		// Convert to string for processing
		html := string(content)

		// Inject CSS files into the head section
		cssLinks := h.generateCSSLinks(cssFiles)
		if cssLinks != "" {
			// Try to inject after <head> tag
			if strings.Contains(html, "<head>") {
				html = strings.Replace(html, "<head>", "<head>\n    "+cssLinks, 1)
			} else if strings.Contains(html, "<html>") {
				// If no head tag, inject after <html>
				html = strings.Replace(html, "<html>", "<html>\n<head>\n    "+cssLinks+"\n</head>", 1)
			} else {
				// If no html structure, prepend CSS
				html = cssLinks + "\n" + html
			}
		}

		// Set content type and serve
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(html))
	}
}

func (h *HTMLRouteBuilder) generateCSSLinks(cssFiles []string) string {
	if len(cssFiles) == 0 {
		return ""
	}

	var links []string
	for _, cssFile := range cssFiles {
		// Convert file path to URL path
		cssName := filepath.Base(cssFile)
		cssURL := "/css/" + cssName
		links = append(links, fmt.Sprintf(`<link rel="stylesheet" href="%s">`, cssURL))
	}

	return strings.Join(links, "\n    ")
}

// GetRoutes returns all built routes
func (h *HTMLRouteBuilder) GetRoutes() []HTMLRoute {
	return h.routes
}

// GetRouteByName finds a route by its name
func (h *HTMLRouteBuilder) GetRouteByName(name string) (*HTMLRoute, bool) {
	for _, route := range h.routes {
		if route.Name == name {
			return &route, true
		}
	}
	return nil, false
}

// GetRoutesByPattern finds routes matching a pattern
func (h *HTMLRouteBuilder) GetRoutesByPattern(pattern string) []HTMLRoute {
	var matches []HTMLRoute
	for _, route := range h.routes {
		if strings.Contains(route.Name, pattern) {
			matches = append(matches, route)
		}
	}
	return matches
}