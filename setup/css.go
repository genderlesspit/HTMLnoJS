package setup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type CSSFile struct {
	Name     string
	FilePath string
	Size     int64
	Category string // "global", "component", "utility", "theme"
}

type CSSValidator struct {
	config    *Config
	cssFiles  []CSSFile
	cssCount  int
	totalSize int64
}

// NewCSSValidator creates a new CSS validator
func NewCSSValidator(config *Config) *CSSValidator {
	return &CSSValidator{
		config:   config,
		cssFiles: make([]CSSFile, 0),
	}
}

// GetCSSFiles returns the discovered CSS files
func (c *CSSValidator) GetCSSFiles() []CSSFile {
	return c.cssFiles
}

// GetCSSCount returns the number of CSS files found
func (c *CSSValidator) GetCSSCount() int {
	return c.cssCount
}

// GetTotalSize returns the total size of all CSS files in bytes
func (c *CSSValidator) GetTotalSize() int64 {
	return c.totalSize
}

// ValidateCSSDir ensures css directory contains valid CSS files
// and provisions a README if it doesn't exist
func (c *CSSValidator) ValidateCSSDir() error {
	if err := c.provisionReadme(); err != nil {
		return err
	}

	if err := c.scanCSSFiles(); err != nil {
		return err
	}

	c.logResults()
	return nil
}

func (c *CSSValidator) provisionReadme() error {
	readmePath := filepath.Join(c.config.CSSDir, "README.md")
	if _, err := os.Stat(readmePath); os.IsNotExist(err) {
		content, err := GetCSSREADME()
		if err != nil {
			return fmt.Errorf("failed to load CSS README template: %w", err)
		}

		if err := os.WriteFile(readmePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create css README.md: %w", err)
		}
		log.Printf("Created README.md in css directory")
	}
	return nil
}

func (c *CSSValidator) scanCSSFiles() error {
	return filepath.Walk(c.config.CSSDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".css") {
			return nil
		}

		// Skip README files
		if strings.Contains(strings.ToLower(info.Name()), "readme") {
			return nil
		}

		c.addCSSFile(path, info)
		return nil
	})
}

func (c *CSSValidator) addCSSFile(filePath string, info os.FileInfo) {
	name := strings.TrimSuffix(info.Name(), ".css")
	category := c.categorizeCSS(name)

	cssFile := CSSFile{
		Name:     name,
		FilePath: filePath,
		Size:     info.Size(),
		Category: category,
	}

	c.cssFiles = append(c.cssFiles, cssFile)
	c.cssCount++
	c.totalSize += info.Size()
}

func (c *CSSValidator) categorizeCSS(name string) string {
	name = strings.ToLower(name)

	switch {
	case strings.Contains(name, "main") || strings.Contains(name, "global") || strings.Contains(name, "reset") || strings.Contains(name, "variables"):
		return "global"
	case strings.Contains(name, "theme") || strings.Contains(name, "dark") || strings.Contains(name, "light"):
		return "theme"
	case strings.Contains(name, "util") || strings.Contains(name, "helper"):
		return "utility"
	default:
		return "component"
	}
}

func (c *CSSValidator) logResults() {
	log.Printf("CSS directory validated: %d CSS files found (%.2f KB total)",
		c.cssCount, float64(c.totalSize)/1024)

	if c.cssCount == 0 {
		log.Printf("WARNING: No CSS files found in css directory")
	}

	// Group by category for better reporting
	categories := make(map[string][]CSSFile)
	for _, file := range c.cssFiles {
		categories[file.Category] = append(categories[file.Category], file)
	}

	for category, files := range categories {
		log.Printf("Found %d %s CSS files:", len(files), category)
		for _, file := range files {
			log.Printf("  - %s.css (%.2f KB)", file.Name, float64(file.Size)/1024)
		}
	}
}