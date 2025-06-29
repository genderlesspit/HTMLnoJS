package routebuilder

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

type RouteCollection struct {
	HTMLRoutes   []HTMLRoute
	CSSRoutes    []CSSRoute
	PythonRoutes []PythonRoute
	Metadata     RouteMetadata
}

type RouteMetadata struct {
	TotalRoutes    int
	HTMLCount      int
	CSSCount       int
	PythonCount    int
	AuthRequired   int
	CacheEnabled   int
	LoadOrder      []string
	Dependencies   map[string][]string
	BuildTime      string
}

type AllRoutesBuilder struct {
	templatesDir string
	cssDir       string
	pyHTMXDir    string
	Collection   RouteCollection
}

// NewAllRoutesBuilder creates a unified route builder
func NewAllRoutesBuilder(templatesDir, cssDir, pyHTMXDir string) *AllRoutesBuilder {
	return &AllRoutesBuilder{
		templatesDir: templatesDir,
		cssDir:       cssDir,
		pyHTMXDir:    pyHTMXDir,
		Collection: RouteCollection{
			HTMLRoutes:   make([]HTMLRoute, 0),
			CSSRoutes:    make([]CSSRoute, 0),
			PythonRoutes: make([]PythonRoute, 0),
			Metadata: RouteMetadata{
				Dependencies: make(map[string][]string),
			},
		},
	}
}

// BuildAllRoutes orchestrates building all route types
func (a *AllRoutesBuilder) BuildAllRoutes(htmlFiles, cssFiles, pythonFiles []string) (*RouteCollection, error) {
	log.Printf("=== Building All Routes ===")

	// Step 1: Build CSS routes first (needed for HTML dependencies)
	if err := a.buildCSSRoutes(cssFiles); err != nil {
		return nil, fmt.Errorf("failed to build CSS routes: %w", err)
	}

	// Step 2: Build Python routes (needed for API endpoint mapping)
	if err := a.buildPythonRoutes(pythonFiles); err != nil {
		return nil, fmt.Errorf("failed to build Python routes: %w", err)
	}

	// Step 3: Build HTML routes (can reference CSS and Python routes)
	if err := a.buildHTMLRoutes(htmlFiles); err != nil {
		return nil, fmt.Errorf("failed to build HTML routes: %w", err)
	}

	// Step 4: Cross-reference and validate routes
	if err := a.crossReferenceRoutes(); err != nil {
		return nil, fmt.Errorf("failed to cross-reference routes: %w", err)
	}

	// Step 5: Generate metadata
	a.generateMetadata()

	// Step 6: Log summary
	a.logBuildSummary()

	return &a.Collection, nil
}

func (a *AllRoutesBuilder) buildCSSRoutes(cssFiles []string) error {
	log.Printf("Building CSS routes from %d files...", len(cssFiles))

	cssBuilder := NewCSSRouteBuilder(a.cssDir)
	routes, err := cssBuilder.BuildRoutes(cssFiles)
	if err != nil {
		return err
	}

	a.Collection.CSSRoutes = routes
	log.Printf("Built %d CSS routes", len(routes))
	return nil
}

func (a *AllRoutesBuilder) buildPythonRoutes(pythonFiles []string) error {
	log.Printf("Building Python routes from %d files...", len(pythonFiles))

	pythonBuilder := NewPythonRouteBuilder(a.pyHTMXDir)
	routes, err := pythonBuilder.BuildRoutes(pythonFiles)
	if err != nil {
		return err
	}

	a.Collection.PythonRoutes = routes
	log.Printf("Built %d Python routes", len(routes))
	return nil
}

func (a *AllRoutesBuilder) buildHTMLRoutes(htmlFiles []string) error {
	log.Printf("Building HTML routes from %d files...", len(htmlFiles))

	// Extract CSS file paths for HTML builder
	cssFilePaths := make([]string, len(a.Collection.CSSRoutes))
	for i, route := range a.Collection.CSSRoutes {
		cssFilePaths[i] = route.FilePath
	}

	htmlBuilder := NewHTMLRouteBuilder(a.templatesDir, cssFilePaths)
	routes, err := htmlBuilder.BuildRoutes(htmlFiles)
	if err != nil {
		return err
	}

	a.Collection.HTMLRoutes = routes
	log.Printf("Built %d HTML routes", len(routes))
	return nil
}

func (a *AllRoutesBuilder) crossReferenceRoutes() error {
	log.Printf("Cross-referencing routes and validating dependencies...")

	// Map Python routes for quick lookup
	pythonRouteMap := make(map[string]PythonRoute)
	for _, route := range a.Collection.PythonRoutes {
		pythonRouteMap[route.Route] = route
	}

	// Validate HTML → Python API references
	for i, htmlRoute := range a.Collection.HTMLRoutes {
		deps := a.findHTMLDependencies(htmlRoute, pythonRouteMap)
		a.Collection.Metadata.Dependencies[htmlRoute.Route] = deps

		// Update HTML route with validated dependencies
		a.Collection.HTMLRoutes[i].Metadata["api_dependencies"] = deps
	}

	// Validate CSS dependencies
	for i, cssRoute := range a.Collection.CSSRoutes {
		if len(cssRoute.Dependencies) > 0 {
			a.Collection.Metadata.Dependencies[cssRoute.Route] = cssRoute.Dependencies
		}

		// Validate that dependencies exist
		for _, dep := range cssRoute.Dependencies {
			if !a.cssExists(dep) {
				log.Printf("WARNING: CSS route %s depends on missing CSS: %s", cssRoute.Name, dep)
			}
		}

		a.Collection.CSSRoutes[i] = cssRoute
	}

	return nil
}

func (a *AllRoutesBuilder) findHTMLDependencies(htmlRoute HTMLRoute, pythonRoutes map[string]PythonRoute) []string {
	var dependencies []string

	// TODO: Parse HTML template content to find hx-get, hx-post, etc. attributes
	// For now, we'll use naming conventions

	// Check if there's a matching Python API for this HTML route
	apiPath := "/api/" + htmlRoute.Name
	if _, exists := pythonRoutes[apiPath]; exists {
		dependencies = append(dependencies, apiPath)
	}

	// Check for common API patterns
	commonAPIs := []string{
		"/api/" + htmlRoute.Name + "/load",
		"/api/" + htmlRoute.Name + "/save",
		"/api/" + htmlRoute.Name + "/delete",
		"/api/" + htmlRoute.Name + "/update",
	}

	for _, api := range commonAPIs {
		if _, exists := pythonRoutes[api]; exists {
			dependencies = append(dependencies, api)
		}
	}

	return dependencies
}

func (a *AllRoutesBuilder) cssExists(cssName string) bool {
	for _, route := range a.Collection.CSSRoutes {
		if route.Name == cssName || route.Category == cssName {
			return true
		}
	}
	return false
}

func (a *AllRoutesBuilder) generateMetadata() {
	meta := &a.Collection.Metadata

	// Count totals
	meta.HTMLCount = len(a.Collection.HTMLRoutes)
	meta.CSSCount = len(a.Collection.CSSRoutes)
	meta.PythonCount = len(a.Collection.PythonRoutes)
	meta.TotalRoutes = meta.HTMLCount + meta.CSSCount + meta.PythonCount

	// Count auth-required routes
	for _, route := range a.Collection.HTMLRoutes {
		if route.RequiresAuth {
			meta.AuthRequired++
		}
	}
	for _, route := range a.Collection.PythonRoutes {
		if route.RequiresAuth {
			meta.AuthRequired++
		}
	}

	// Count cache-enabled routes
	for _, route := range a.Collection.PythonRoutes {
		if route.CacheTimeout > 0 {
			meta.CacheEnabled++
		}
	}

	// Generate load order for CSS
	meta.LoadOrder = a.generateCSSLoadOrder()

	// Set build time
	meta.BuildTime = "2025-06-28T12:00:00Z" // TODO: Use actual timestamp
}

func (a *AllRoutesBuilder) generateCSSLoadOrder() []string {
	// CSS routes are already sorted by load order
	loadOrder := make([]string, len(a.Collection.CSSRoutes))
	for i, route := range a.Collection.CSSRoutes {
		loadOrder[i] = route.Route
	}
	return loadOrder
}

func (a *AllRoutesBuilder) logBuildSummary() {
	meta := a.Collection.Metadata

	log.Printf("=== Route Build Summary ===")
	log.Printf("Total Routes: %d", meta.TotalRoutes)
	log.Printf("  - HTML Routes: %d", meta.HTMLCount)
	log.Printf("  - CSS Routes: %d", meta.CSSCount)
	log.Printf("  - Python Routes: %d", meta.PythonCount)
	log.Printf("Authentication Required: %d routes", meta.AuthRequired)
	log.Printf("Cache Enabled: %d routes", meta.CacheEnabled)
	log.Printf("Dependencies Mapped: %d relationships", len(meta.Dependencies))

	// Log route breakdown by category
	a.logRouteBreakdown()
}

func (a *AllRoutesBuilder) logRouteBreakdown() {
	// CSS breakdown by category
	cssCategories := make(map[string]int)
	for _, route := range a.Collection.CSSRoutes {
		cssCategories[route.Category]++
	}

	log.Printf("CSS Routes by Category:")
	for category, count := range cssCategories {
		log.Printf("  - %s: %d", category, count)
	}

	// Python breakdown by method
	pythonMethods := make(map[string]int)
	for _, route := range a.Collection.PythonRoutes {
		pythonMethods[route.Method]++
	}

	log.Printf("Python Routes by Method:")
	for method, count := range pythonMethods {
		log.Printf("  - %s: %d", method, count)
	}

	// HTML breakdown by auth requirement
	authRequired := 0
	publicRoutes := 0
	for _, route := range a.Collection.HTMLRoutes {
		if route.RequiresAuth {
			authRequired++
		} else {
			publicRoutes++
		}
	}

	log.Printf("HTML Routes by Access:")
	log.Printf("  - Public: %d", publicRoutes)
	log.Printf("  - Auth Required: %d", authRequired)
}

// Utility methods for accessing the built routes

// GetRouteCollection returns the complete route collection
func (a *AllRoutesBuilder) GetRouteCollection() *RouteCollection {
	return &a.Collection
}

// GetHTMLRoutes returns all HTML routes
func (a *AllRoutesBuilder) GetHTMLRoutes() []HTMLRoute {
	return a.Collection.HTMLRoutes
}

// GetCSSRoutes returns all CSS routes in load order
func (a *AllRoutesBuilder) GetCSSRoutes() []CSSRoute {
	return a.Collection.CSSRoutes
}

// GetPythonRoutes returns all Python routes
func (a *AllRoutesBuilder) GetPythonRoutes() []PythonRoute {
	return a.Collection.PythonRoutes
}

// GetPublicHTMLRoutes returns HTML routes that don't require auth
func (a *AllRoutesBuilder) GetPublicHTMLRoutes() []HTMLRoute {
	var public []HTMLRoute
	for _, route := range a.Collection.HTMLRoutes {
		if !route.RequiresAuth {
			public = append(public, route)
		}
	}
	return public
}

// GetAPIRoutes returns all Python API routes sorted by path
func (a *AllRoutesBuilder) GetAPIRoutes() []PythonRoute {
	routes := make([]PythonRoute, len(a.Collection.PythonRoutes))
	copy(routes, a.Collection.PythonRoutes)

	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Route < routes[j].Route
	})

	return routes
}

// GenerateRouteMap creates a text representation of all routes
func (a *AllRoutesBuilder) GenerateRouteMap() string {
	var builder strings.Builder

	builder.WriteString("=== HTMLnoJS Route Map ===\n\n")

	// HTML Routes
	builder.WriteString("HTML ROUTES:\n")
	for _, route := range a.Collection.HTMLRoutes {
		auth := ""
		if route.RequiresAuth {
			auth = " [AUTH]"
		}
		builder.WriteString(fmt.Sprintf("  %s %s -> %s%s\n",
			route.Method, route.Route, route.Name, auth))
	}

	// CSS Routes
	builder.WriteString("\nCSS ROUTES (Load Order):\n")
	for i, route := range a.Collection.CSSRoutes {
		builder.WriteString(fmt.Sprintf("  %d. %s -> %s [%s]\n",
			i+1, route.Route, route.Name, route.Category))
	}

	// Python Routes
	builder.WriteString("\nPYTHON API ROUTES:\n")
	for _, route := range a.Collection.PythonRoutes {
		auth := ""
		if route.RequiresAuth {
			auth = " [AUTH]"
		}
		cache := ""
		if route.CacheTimeout > 0 {
			cache = fmt.Sprintf(" [CACHE:%ds]", route.CacheTimeout)
		}
		builder.WriteString(fmt.Sprintf("  %s %s -> %s%s%s\n",
			route.Method, route.Route, route.Function, auth, cache))
	}

	// Summary
	meta := a.Collection.Metadata
	builder.WriteString(fmt.Sprintf("\nSUMMARY: %d total routes (%d HTML, %d CSS, %d Python)\n",
		meta.TotalRoutes, meta.HTMLCount, meta.CSSCount, meta.PythonCount))

	return builder.String()
}