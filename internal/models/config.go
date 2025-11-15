package models

import (
	"path/filepath"
	"strings"
	"time"
)

// ConfigEntry represents a registered file in the registry
type ConfigEntry struct {
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Type        string    `json:"type"`        // json, yaml, toml, ini, txt
	Project     string    `json:"project"`     // project association
	Description string    `json:"description"` // brief description
	LastOpened  time.Time `json:"last_opened,omitempty"`
	Tags        []string  `json:"tags,omitempty"` // flexible tagging
}

// ConfigManager manages the collection of config entries
type ConfigManager struct {
	Configs []ConfigEntry `json:"configs"`
}

// DetectFileType automatically detects file type from extension
func DetectFileType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".toml":
		return "toml"
	case ".ini", ".conf":
		return "ini"
	case ".xml":
		return "xml"
	case ".md", ".markdown":
		return "markdown"
	case ".txt":
		return "txt"
	case ".sh":
		return "shell"
	case ".go":
		return "go"
	case ".py":
		return "python"
	case ".js":
		return "javascript"
	case ".ts":
		return "typescript"
	case ".rb":
		return "ruby"
	default:
		return "txt"
	}
}

// Equals checks if two entries are the same
func (c *ConfigEntry) Equals(other *ConfigEntry) bool {
	return c.Name == other.Name &&
		c.Path == other.Path &&
		c.Project == other.Project
}

// IsValid checks if the config entry has required fields
func (c *ConfigEntry) IsValid() bool {
	return c.Name != "" && c.Path != ""
}
