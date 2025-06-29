package setup

import (
	"embed"
	"fmt"
)

//go:embed readmes/*
var readmeFS embed.FS

// loadREADME loads README content from embedded files
func loadREADME(filename string) (string, error) {
	content, err := readmeFS.ReadFile("readmes/" + filename)
	if err != nil {
		return "", fmt.Errorf("failed to load README %s: %w", filename, err)
	}
	return string(content), nil
}

// GetTemplatesREADME returns the templates directory README content
func GetTemplatesREADME() (string, error) {
	return loadREADME("templates.md")
}

// GetCSSREADME returns the CSS directory README content
func GetCSSREADME() (string, error) {
	return loadREADME("css.md")
}

// GetPyHTMXREADME returns the Python HTMX directory README content
func GetPyHTMXREADME() (string, error) {
	return loadREADME("py_htmx.md")
}