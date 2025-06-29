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
	pyHTMXDir     string
	routes        []PythonRoute
	fastAPIHost   string
	fastAPIPort   int
	httpClient    *http.Client
}

// NewPythonRouteBuilder creates a new Python HTMX route builder
func NewPythonRouteBuilder(pyHTMXDir string) *PythonRouteBuilder {
	return &PythonRouteBuilder{
		pyHTMXDir:   pyHTMXDir,
		routes:      make([]PythonRoute, 0),
		fastAPIHost: "localhost",
		fastAPIPort: 8081, // Default FastAPI port
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:       10,
				IdleConnTimeout:    30 * time.Second,
				DisableCompression: false,
			},
		},
	}
}

// SetFastAPIServer sets the FastAPI server host and port
func (p *PythonRouteBuilder) SetFastAPIServer(host string, port int) {
	p.fastAPIHost = host
	p.fastAPIPort = port
}

// GetFastAPIURL returns the FastAPI server URL
func (p *PythonRouteBuilder) GetFastAPIURL() string {
	return fmt.Sprintf("http://%s:%d", p.fastAPIHost, p.fastAPIPort)
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

	// Build API route path - this is what the Go server will expose
	var goRoutePath string
	if basePath == "" || basePath == "." {
		goRoutePath = "/api/" + routeName
	} else {
		goRoutePath = "/api/" + basePath + "/" + routeName
	}

	// Determine HTTP method based on function name patterns
	method := p.determineHTTPMethod(function.Name)

	// Check for special attributes
	requiresAuth := p.checkRequiresAuth(function.Documentation)
	rateLimit := p.extractRateLimit(function.Documentation)
	cacheTimeout := p.extractCacheTimeout(function.Documentation)

	metadata := map[string]interface{}{
		"file":         filePath,
		"base_path":    basePath,
		"parameters":   function.Parameters,
		"return_type":  function.ReturnType,
		"fastapi_url":  p.GetFastAPIURL(),
		"fastapi_path": p.buildFastAPIPath(basePath, function.Name),
	}

	route := PythonRoute{
		Name:          routeName,
		FilePath:      filePath,
		Route:         goRoutePath,
		Method:        method,
		Handler:       p.createProxyHandler(basePath, function.Name),
		Function:      function.Name,
		Parameters:    function.Parameters,
		ReturnType:    function.ReturnType,
		RequiresAuth:  requiresAuth,
		RateLimit:     rateLimit,
		CacheTimeout:  cacheTimeout,
		Documentation: function.Documentation,
		Metadata:      metadata,
	}

	return route
}

// buildFastAPIPath creates the path that will be sent to the FastAPI server
func (p *PythonRouteBuilder) buildFastAPIPath(basePath, functionName string) string {
	routeName := strings.TrimPrefix(functionName, "htmx_")

	if basePath == "" || basePath == "." {
		return fmt.Sprintf("/%s", routeName)
	} else {
		return fmt.Sprintf("/%s/%s", basePath, routeName)
	}
}

// createProxyHandler creates an HTTP handler that proxies requests to FastAPI
func (p *PythonRouteBuilder) createProxyHandler(basePath, functionName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Build the FastAPI server URL path
		fastAPIPath := p.buildFastAPIPath(basePath, functionName)
		targetURL := p.GetFastAPIURL() + fastAPIPath

		// Read the request body
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

		// Copy headers from original request (excluding hop-by-hop headers)
		copyHeaders(r.Header, proxyReq.Header)

		// Copy query parameters
		if r.URL.RawQuery != "" {
			proxyReq.URL.RawQuery = r.URL.RawQuery
		}

		// Handle form data for POST requests
		if r.Method == "POST" && strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
			if err := r.ParseForm(); err == nil {
				formData := url.Values(r.PostForm).Encode()
				proxyReq.Body = io.NopCloser(strings.NewReader(formData))
				proxyReq.ContentLength = int64(len(formData))
				proxyReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			}
		}

		// Make the request to the FastAPI server
		resp, err := p.httpClient.Do(proxyReq)
		if err != nil {
			// FastAPI server is not available
			http.Error(w, fmt.Sprintf(`
				<div class="htmx-error" style="color: red; padding: 10px; border: 1px solid red; border-radius: 4px;">
					<strong>Service Unavailable</strong><br>
					The Python handler server is not running on %s<br>
					<small>Error: %v</small>
				</div>
			`, p.GetFastAPIURL(), err), http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		// Copy response headers (excluding hop-by-hop headers)
		copyHeaders(resp.Header, w.Header())

		// Copy status code
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		if _, err := io.Copy(w, resp.Body); err != nil {
			// Log the error but don't send another response since headers are already sent
			fmt.Printf("Error copying response body from FastAPI: %v\n", err)
		}
	}
}

// copyHeaders copies HTTP headers, excluding hop-by-hop headers
func copyHeaders(src, dst http.Header) {
	// Hop-by-hop headers that shouldn't be copied
	hopByHopHeaders := map[string]bool{
		"connection":          true,
		"keep-alive":          true,
		"proxy-authenticate":  true,
		"proxy-authorization": true,
		"te":                  true,
		"trailers":            true,
		"transfer-encoding":   true,
		"upgrade":             true,
	}

	for key, values := range src {
		if !hopByHopHeaders[strings.ToLower(key)] {
			for _, value := range values {
				dst.Add(key, value)
			}
		}
	}
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
	return 0
}

func (p *PythonRouteBuilder) extractCacheTimeout(doc string) int {
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
	return 0
}

// Helper types and functions
type FunctionInfo struct {
	Name          string
	Parameters    []string
	ReturnType    string
	Documentation string
}

func parseInt(s string) int {
	result := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			result = result*10 + int(r-'0')
		}
	}
	return result
}

// Health check for FastAPI server connectivity
func (p *PythonRouteBuilder) CheckFastAPIHealth() error {
	healthURL := p.GetFastAPIURL() + "/health"

	req, err := http.NewRequest("GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("FastAPI server health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FastAPI server returned status %d", resp.StatusCode)
	}

	return nil
}

// Utility methods
func (p *PythonRouteBuilder) GetRoutes() []PythonRoute {
	return p.routes
}

func (p *PythonRouteBuilder) GetRoutesByMethod(method string) []PythonRoute {
	var matches []PythonRoute
	for _, route := range p.routes {
		if route.Method == method {
			matches = append(matches, route)
		}
	}
	return matches
}

func (p *PythonRouteBuilder) GetAuthenticatedRoutes() []PythonRoute {
	var matches []PythonRoute
	for _, route := range p.routes {
		if route.RequiresAuth {
			matches = append(matches, route)
		}
	}
	return matches
}