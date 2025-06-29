package setup

import (
	"path/filepath"
)

type FileSet struct {
	PyHTMXFiles    []string
	TemplateFiles  []string
	CSSFiles       []string
}

// GlobFiles discovers all files in the project directories
func (c *Config) GlobFiles() (*FileSet, error) {
	fs := &FileSet{}

	// Glob all files in py_htmx directory
	pyFiles, err := filepath.Glob(filepath.Join(c.PyHTMXDir, "*"))
	if err != nil {
		return nil, err
	}
	fs.PyHTMXFiles = pyFiles

	// Glob all files in templates directory
	templateFiles, err := filepath.Glob(filepath.Join(c.TemplatesDir, "*"))
	if err != nil {
		return nil, err
	}
	fs.TemplateFiles = templateFiles

	// Glob all files in css directory
	cssFiles, err := filepath.Glob(filepath.Join(c.CSSDir, "*"))
	if err != nil {
		return nil, err
	}
	fs.CSSFiles = cssFiles

	return fs, nil
}