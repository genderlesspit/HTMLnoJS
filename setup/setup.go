package setup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	ProjectDir   string
	PyHTMXDir    string
	CSSDir       string
	TemplatesDir string
}

// Setup creates the required directory structure for HTMLnoJS
// Expects a project directory containing py_htmx/ and templates/ subdirectories
func Setup(projectDir string) (*Config, error) {
	// Validate project directory exists
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("project directory does not exist: %s", projectDir)
	}

	cfg := &Config{
		ProjectDir:   projectDir,
		PyHTMXDir:    filepath.Join(projectDir, "py_htmx"),
		CSSDir:       filepath.Join(projectDir, "css"),
		TemplatesDir: filepath.Join(projectDir, "templates"),
	}

	// Create py_htmx directory
	if err := os.MkdirAll(cfg.PyHTMXDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create py_htmx directory: %w", err)
	}
	log.Printf("Created py_htmx directory: %s", cfg.PyHTMXDir)

	// Create css directory
	if err := os.MkdirAll(cfg.CSSDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create css directory: %w", err)
	}
	log.Printf("Created css directory: %s", cfg.CSSDir)

	// Create templates directory
	if err := os.MkdirAll(cfg.TemplatesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create templates directory: %w", err)
	}
	log.Printf("Created templates directory: %s", cfg.TemplatesDir)

	// Validate structure
	if err := cfg.validate(); err != nil {
		return nil, err
	}

	log.Printf("HTMLnoJS project structure validated in: %s", projectDir)
	return cfg, nil
}

// validate ensures the required directories exist
func (c *Config) validate() error {
	dirs := []string{c.PyHTMXDir, c.CSSDir, c.TemplatesDir}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return fmt.Errorf("required directory missing: %s", dir)
		}
	}

	return nil
}