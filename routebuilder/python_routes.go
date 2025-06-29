package routebuilder

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CSSRoute struct {
	Name         string
	FilePath     string
	Route        string
	Method       string
	Handler      http.HandlerFunc
	Category     string
	LoadOrder    int
	Minified     bool
	Dependencies []string
	MediaQuery   string
	Metadata     map[string]interface{}
}

type CSSRouteBuilder struct {
	cssDir string
	routes []CSSRoute
}

// NewCSSRouteBuilder creates a new CSS route builder
func NewCSSRouteBuilder(cssDir string) *CSSRouteBuilder {
	return &CSSRouteBuilder{
		cssDir: cssDir,
		routes: make([]CSSRoute, 0),
	}
}

// BuildRoutes discovers and builds CSS file routes
func (c *CSSRouteBuilder) BuildRoutes(cssFiles []string) ([]CSSRoute, error) {
	for _, filePath := range cssFiles {
		if !strings.HasSuffix(strings.ToLower(filePath), ".css") {
			continue
		}

		route, err := c.buildCSSRoute(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to build CSS route for %s: %w", filePath, err)
		}

		c.routes = append(c.routes, route)
	}

	// Sort routes by load order
	c.sortByLoadOrder()

	return c.routes, nil
}

func (c *CSSRouteBuilder) buildCSSRoute(filePath string) (CSSRoute, error) {
	filename := filepath.Base(filePath)
	name := strings.TrimSuffix(filename, ".css")

	// Create route path
	routePath := "/css/" + filename

	// Determine category and load order
	category := c.categorizeCSS(name)
	loadOrder := c.determineLoadOrder(name, category)

	// Check if minified
	isMinified := strings.Contains(name, ".min") || strings.Contains(name, "-min")

	// Determine media query if any
	mediaQuery := c.determineMediaQuery(name)

	// Find dependencies
	dependencies := c.findDependencies(name, category)

	metadata := map[string]interface{}{
		"file_size": c.getFileSize(filePath),
		"category":  category,
	}

	route := CSSRoute{
		Name:         name,
		FilePath:     filePath,
		Route:        routePath,
		Method:       "GET",
		Handler:      c.createCSSHandler(filePath),
		Category:     category,
		LoadOrder:    loadOrder,
		Minified:     isMinified,
		Dependencies: dependencies,
		MediaQuery:   mediaQuery,
		Metadata:     metadata,
	}

	return route, nil
}

func (c *CSSRouteBuilder) categorizeCSS(name string) string {
	name = strings.ToLower(name)

	switch {
	case strings.Contains(name, "reset") || strings.Contains(name, "normalize"):
		return "reset"
	case strings.Contains(name, "variables") || strings.Contains(name, "custom-properties"):
		return "variables"
	case strings.Contains(name, "main") || strings.Contains(name, "global") || strings.Contains(name, "base"):
		return "global"
	case strings.Contains(name, "theme") || strings.Contains(name, "dark") || strings.Contains(name, "light"):
		return "theme"
	case strings.Contains(name, "util") || strings.Contains(name, "helper") || strings.Contains(name, "atomic"):
		return "utility"
	case strings.Contains(name, "responsive") || strings.Contains(name, "breakpoint"):
		return "responsive"
	case strings.Contains(name, "print"):
		return "print"
	default:
		return "component"
	}
}

func (c *CSSRouteBuilder) determineLoadOrder(name, category string) int {
	// Lower numbers load first
	switch category {
	case "reset":
		return 10
	case "variables":
		return 20
	case "global":
		return 30
	case "theme":
		return 40
	case "component":
		return 50
	case "utility":
		return 60
	case "responsive":
		return 70
	case "print":
		return 80
	default:
		return 100
	}
}

func (c *CSSRouteBuilder) determineMediaQuery(name string) string {
	name = strings.ToLower(name)

	if strings.Contains(name, "print") {
		return "print"
	}
	if strings.Contains(name, "mobile") {
		return "screen and (max-width: 768px)"
	}
	if strings.Contains(name, "tablet") {
		return "screen and (min-width: 769px) and (max-width: 1024px)"
	}
	if strings.Contains(name, "desktop") {
		return "screen and (min-width: 1025px)"
	}
	if strings.Contains(name, "dark") {
		return "screen and (prefers-color-scheme: dark)"
	}

	return "screen" // default
}

func (c *CSSRouteBuilder) findDependencies(name, category string) []string {
	var deps []string

	// Components might depend on variables and base styles
	if category == "component" {
		deps = append(deps, "variables", "global")
	}

	// Themes depend on variables
	if category == "theme" {
		deps = append(deps, "variables")
	}

	// Utilities can depend on variables
	if category == "utility" {
		deps = append(deps, "variables")
	}

	return deps
}

func (c *CSSRouteBuilder) getFileSize(filePath string) int64 {
	if info, err := os.Stat(filePath); err == nil {
		return info.Size()
	}
	return 0
}

func (c *CSSRouteBuilder) createCSSHandler(cssPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CSS headers
		w.Header().Set("Content-Type", "text/css")
		w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year cache

		// Read and serve CSS file
		content, err := os.ReadFile(cssPath)
		if err != nil {
			http.Error(w, "CSS file not found", http.StatusNotFound)
			return
		}

		// Set Last-Modified header
		if info, err := os.Stat(cssPath); err == nil {
			w.Header().Set("Last-Modified", info.ModTime().UTC().Format(http.TimeFormat))
		}

		w.Write(content)
	}
}

func (c *CSSRouteBuilder) sortByLoadOrder() {
	// Simple bubble sort by load order
	for i := 0; i < len(c.routes)-1; i++ {
		for j := 0; j < len(c.routes)-i-1; j++ {
			if c.routes[j].LoadOrder > c.routes[j+1].LoadOrder {
				c.routes[j], c.routes[j+1] = c.routes[j+1], c.routes[j]
			}
		}
	}
}

// GetRoutes returns all built CSS routes
func (c *CSSRouteBuilder) GetRoutes() []CSSRoute {
	return c.routes
}

// GetRoutesByCategory returns routes filtered by category
func (c *CSSRouteBuilder) GetRoutesByCategory(category string) []CSSRoute {
	var matches []CSSRoute
	for _, route := range c.routes {
		if route.Category == category {
			matches = append(matches, route)
		}
	}
	return matches
}

// GetLoadOrderedRoutes returns routes in loading order
func (c *CSSRouteBuilder) GetLoadOrderedRoutes() []CSSRoute {
	return c.routes // Already sorted by load order
}

// GenerateLinkTags generates HTML link tags for CSS files
func (c *CSSRouteBuilder) GenerateLinkTags() string {
	var tags strings.Builder

	for _, route := range c.routes {
		media := route.MediaQuery
		if media == "" {
			media = "all"
		}

		tags.WriteString(fmt.Sprintf(`<link rel="stylesheet" href="%s" media="%s">`,
			route.Route, media))
		tags.WriteString("\n    ")
	}

	return tags.String()
}