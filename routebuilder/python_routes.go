package routebuilder

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type PythonRoute struct {
	Name           string
	FilePath       string
	Route          string
	Method         string
	Handler        http.HandlerFunc
	Function       string
	Parameters     []string
	ReturnType     string
	RequiresAuth   bool
	RateLimit      int
	CacheTimeout   int
	Documentation  string
	Metadata       map[string]interface{}
}

type PythonRouteBuilder struct {
	pyHTMXDir    string
	routes       []PythonRoute
	htmxServerURL string
	httpClient   *http.Client
}

// NewPythonRouteBuilder creates a new Python HTMX route builder
func NewPythonRouteBuilder(pyHTMXDir string) *PythonRouteBuilder {
	return &PythonRouteBuilder{
		pyHTMXDir:     pyHTMXDir,
		routes:        make([]PythonRoute, 0),
		htmxServerURL: "http://127.0.0.1:8081", // Default HTMX server URL
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
			},
		},
	}
}

// SetHTMXServerURL sets the URL for the HTMX server
func (p *PythonRouteBuilder) SetHTMXServerURL(url string) {
	p.htmxServerURL = strings.TrimSuffix(url, "/")
}

// BuildRoutes discovers and builds Python HTMX routes
func (p *PythonRouteBuilder) BuildRoutes(pythonFiles []string) ([]PythonRoute, error) {
	for _, filePath := range pythonFiles {
		if !strings.HasSuffix(strings.ToLower(filePath), ".py") {
			continue
		}

		routes, err := p.extractRoutesFromFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to extract routes from %s: %w", filePath, err)
		}

		p.routes = append(p.routes, routes...)
	}

	return p.routes, nil
}

func (p *PythonRouteBuilder) extractRoutesFromFile(filePath string) ([]PythonRoute, error) {
	var routes []PythonRoute

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Extract htmx_ functions using regex
	htmxFunctions := p.findHTMXFunctions(string(content))

	// Get relative path for API routing
	relPath, _ := filepath.Rel(p.pyHTMXDir, filePath)
	basePath := strings.TrimSuffix(relPath, ".py")
	basePath = strings.ReplaceAll(basePath, "\\", "/") // Handle Windows paths

	for _, function := range htmxFunctions {
		route := p.buildPythonRoute(filePath, basePath, function)
		routes = append(routes, route)
	}

	return routes, nil
}

func (p *PythonRouteBuilder) findHTMXFunctions(content string) []FunctionInfo {
	var functions []FunctionInfo

	// Regex to match htmx_ functions
	funcRegex := regexp.MustCompile(`def\s+(htmx_\w+)\s*\(([^)]*)\)(?:\s*->\s*([^:]+))?:`)
	matches := funcRegex.FindAllStringSubmatch(content, -1)

	for _, match := range matches {
		functionName := match[1]
		parameters := p.parseParameters(match[2])
		returnType := strings.TrimSpace(match[3])

		// Extract documentation if available
		docstring := p.extractDocstring(content, functionName)

		functions = append(functions, FunctionInfo{
			Name:          functionName,
			Parameters:    parameters,
			ReturnType:    returnType,
			Documentation: docstring,
		})
	}

	return functions
}

func (p *PythonRouteBuilder) parseParameters(paramStr string) []string {
	if paramStr == "" {
		return []string{}
	}

	params := strings.Split(paramStr, ",")
	var cleaned []string

	for _, param := range params {
		param = strings.TrimSpace(param)
		// Remove type hints and default values
		param = strings.Split(param, ":")[0]
		param = strings.Split(param, "=")[0]
		param = strings.TrimSpace(param)
		if param != "" && param != "self" {
			cleaned = append(cleaned, param)
		}
	}

	return cleaned
}

func (p *PythonRouteBuilder) extractDocstring(content, functionName string) string {
	lines := strings.Split(content, "\n")
	inFunction := false
	inDocstring := false
	var docLines []string

	for _, line := range lines {
		if strings.Contains(line, "def "+functionName) {
			inFunction = true
			continue
		}

		if inFunction {
			trimmed := strings.TrimSpace(line)

			if strings.HasPrefix(trimmed, `"""`) || strings.HasPrefix(trimmed, `'''`) {
				if inDocstring {
					// End of docstring
					break
				}
				inDocstring = true
				// Check if single-line docstring
				if strings.Count(trimmed, `"""`) == 2 || strings.Count(trimmed, `'''`) == 2 {
					docLines = append(docLines, strings.Trim(trimmed, `"'`))
					break
				}
				continue
			}

			if inDocstring {
				docLines = append(docLines, trimmed)
			} else if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
				// Hit actual code without docstring
				break
			}
		}
	}

	return strings.Join(docLines, " ")
}

func (p *PythonRouteBuilder) buildPythonRoute(filePath, basePath string, function FunctionInfo) PythonRoute {
	// Remove htmx_ prefix for route name
	routeName := strings.TrimPrefix(function.Name, "htmx_")

	// Build API route path
	var routePath string
	if basePath == "" || basePath == "." {
		routePath = "/api/" + routeName
	} else {
		routePath = "/api/" + basePath + "/" + routeName
	}

	// Determine HTTP method based on function name patterns
	method := p.determineHTTPMethod(function.Name)

	// Check for special attributes
	requiresAuth := p.checkRequiresAuth(function.Documentation)
	rateLimit := p.extractRateLimit(function.Documentation)
	cacheTimeout := p.extractCacheTimeout(function.Documentation)

	metadata := map[string]interface{}{
		"file":        filePath,
		"base_path":   basePath,
		"parameters":  function.Parameters,
		"return_type": function.ReturnType,
	}

	route := PythonRoute{
		Name:           routeName,
		FilePath:       filePath,
		Route:          routePath,
		Method:         method,
		Handler:        p.createPythonHandler(basePath, function.Name),
		Function:       function.Name,
		Parameters:     function.Parameters,
		ReturnType:     function.ReturnType,
		RequiresAuth:   requiresAuth,
		RateLimit:      rateLimit,
		CacheTimeout:   cacheTimeout,
		Documentation:  function.Documentation,
		Metadata:       metadata,
	}

	return route
}

func (p *PythonRouteBuilder) determineHTTPMethod(functionName string) string {
	name := strings.ToLower(functionName)

	switch {
	case strings.Contains(name, "create") || strings.Contains(name, "add") || strings.Contains(name, "post"):
		return "POST"
	case strings.Contains(name, "update") || strings.Contains(name, "edit") || strings.Contains(name, "put"):
		return "PUT"
	case strings.Contains(name, "delete") || strings.Contains(name, "remove"):
		return "DELETE"
	case strings.Contains(name, "get") || strings.Contains(name, "fetch") || strings.Contains(name, "load"):
		return "GET"
	default:
		return "POST" // Default for HTMX interactions
	}
}

func (p *PythonRouteBuilder) checkRequiresAuth(doc string) bool {
	doc = strings.ToLower(doc)
	return strings.Contains(doc, "@auth") ||
		   strings.Contains(doc, "requires auth") ||
		   strings.Contains(doc, "login required")
}

func (p *PythonRouteBuilder) extractRateLimit(doc string) int {
	// Look for @rate_limit(n) or similar patterns
	rateRegex := regexp.MustCompile(`@rate_limit\((\d+)\)|rate.limit[:\s]+(\d+)`)
	matches := rateRegex.FindStringSubmatch(doc)
	if len(matches) > 1 {
		if matches[1] != "" {
			return parseInt(matches[1])
		}
		if matches[2] != "" {
			return parseInt(matches[2])
		}
	}
	return 0 // No rate limit
}

func (p *PythonRouteBuilder) extractCacheTimeout(doc string) int {
	// Look for @cache(n) or similar patterns
	cacheRegex := regexp.MustCompile(`@cache\((\d+)\)|cache[:\s]+(\d+)`)
	matches := cacheRegex.FindStringSubmatch(doc)
	if len(matches) > 1 {
		if matches[1] != "" {
			return parseInt(matches[1])
		}
		if matches[2] != "" {
			return parseInt(matches[2])
		}
	}
	return 0 // No cache
}

// createPythonHandler creates an HTTP handler that proxies requests to the HTMX server
func (p *PythonRouteBuilder) createPythonHandler(basePath, functionName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Build the HTMX server URL path
		var htmxPath string
		routeName := strings.TrimPrefix(functionName, "htmx_")

		if basePath == "" || basePath == "." {
			htmxPath = fmt.Sprintf("/%s", routeName)
		} else {
			htmxPath = fmt.Sprintf("/%s/%s", basePath, routeName)
		}

		// Create the full URL to the HTMX server
		targetURL := p.htmxServerURL + htmxPath

		// Create a new request to the HTMX server
		var body io.Reader
		if r.Body != nil {
			bodyBytes, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to read request body: %v", err), http.StatusInternalServerError)
				return
			}
			body = bytes.NewReader(bodyBytes)
		}

		// Create the proxy request
		proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create proxy request: %v", err), http.StatusInternalServerError)
			return
		}

		// Copy headers from original request
		p.copyHeaders(r.Header, proxyReq.Header)

		// Copy query parameters
		if r.URL.RawQuery != "" {
			proxyReq.URL.RawQuery = r.URL.RawQuery
		}

		// Add form data for POST requests
		if r.Method == "POST" && r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			if err := r.ParseForm(); err == nil {
				proxyReq.PostForm = r.PostForm
				// Re-encode form data
				proxyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				formData := url.Values(r.PostForm).Encode()
				proxyReq.Body = io.NopCloser(strings.NewReader(formData))
				proxyReq.ContentLength = int64(len(formData))
			}
		}

		// Make the request to the HTMX server
		resp, err := p.httpClient.Do(proxyReq)
		if err != nil {
			// If HTMX server is not available, return an error message
			http.Error(w, fmt.Sprintf("HTMX server unavailable: %v", err), http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		p.copyHeaders(resp.Header, w.Header())

		// Copy status code
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		if _, err := io.Copy(w, resp.Body); err != nil {
			// Log the error but don't send another response since headers are already sent
			fmt.Printf("Error copying response body: %v\n", err)
		}
	}
}

// copyHeaders copies HTTP headers from source to destination
func (p *PythonRouteBuilder) copyHeaders(src, dst http.Header) {
	for key, values := range src {
		// Skip certain headers that shouldn't be copied
		switch strings.ToLower(key) {
		case "connection", "transfer-encoding", "upgrade":
			continue
		}

		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

// Helper types and functions
type FunctionInfo struct {
	Name          string
	Parameters    []string
	ReturnType    string
	Documentation string
}

func parseInt(s string) int {
	// Simple integer parsing, returns 0 on error
	result := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		}
	}
	return result
}

// GetRoutes returns all built Python routes
func (p *PythonRouteBuilder) GetRoutes() []PythonRoute {
	return p.routes
}

// GetRoutesByMethod returns routes filtered by HTTP method
func (p *PythonRouteBuilder) GetRoutesByMethod(method string) []PythonRoute {
	var matches []PythonRoute
	for _, route := range p.routes {
		if route.Method == method {
			matches = append(matches, route)
		}
	}
	return matches
}

// GetRoutesByFile returns routes from a specific file
func (p *PythonRouteBuilder) GetRoutesByFile(filePath string) []PythonRoute {
	var matches []PythonRoute
	for _, route := range p.routes {
		if route.FilePath == filePath {
			matches = append(matches, route)
		}
	}
	return matches
}

// GetAuthenticatedRoutes returns routes that require authentication
func (p *PythonRouteBuilder) GetAuthenticatedRoutes() []PythonRoute {
	var matches []PythonRoute
	for _, route := range p.routes {
		if route.RequiresAuth {
			matches = append(matches, route)
		}
	}
	return matches
}

// Health check for HTMX server connectivity
func (p *PythonRouteBuilder) CheckHTMXServerHealth() error {
	healthURL := p.htmxServerURL + "/health"

	req, err := http.NewRequest("GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTMX server health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTMX server returned status %d", resp.StatusCode)
	}

	return nil
}