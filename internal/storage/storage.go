package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"configly/internal/models"
)

// Storage handles config file persistence
type Storage struct {
	filePath string
}

// New creates a new Storage instance
func New(filePath string) *Storage {
	return &Storage{filePath: filePath}
}

// Load reads configs from disk
func (s *Storage) Load() ([]models.ConfigEntry, error) {
	var manager models.ConfigManager

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config directory
			if err := os.MkdirAll(filepath.Dir(s.filePath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create config directory: %w", err)
			}
			return []models.ConfigEntry{}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, &manager); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return manager.Configs, nil
}

// Save writes configs to disk atomically
func (s *Storage) Save(configs []models.ConfigEntry) error {
	manager := models.ConfigManager{Configs: configs}

	data, err := json.MarshalIndent(manager, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(s.filePath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Atomic write: write to temp file then rename
	tempFile := s.filePath + ".tmp"
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempFile, s.filePath); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to replace config file: %w", err)
	}

	return nil
}

// SortConfigs sorts configs by project then name
func SortConfigs(configs []models.ConfigEntry) []models.ConfigEntry {
	sorted := make([]models.ConfigEntry, len(configs))
	copy(sorted, configs)

	sort.Slice(sorted, func(i, j int) bool {
		// Handle empty projects by treating them as "General"
		projectI := sorted[i].Project
		if projectI == "" {
			projectI = "General"
		}
		projectJ := sorted[j].Project
		if projectJ == "" {
			projectJ = "General"
		}

		// First sort by project
		if !strings.EqualFold(projectI, projectJ) {
			return strings.ToLower(projectI) < strings.ToLower(projectJ)
		}

		// If projects are the same, sort by name
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})

	return sorted
}

// SortByRecentlyOpened sorts configs by last opened time (most recent first)
func SortByRecentlyOpened(configs []models.ConfigEntry) []models.ConfigEntry {
	sorted := make([]models.ConfigEntry, len(configs))
	copy(sorted, configs)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LastOpened.After(sorted[j].LastOpened)
	})

	return sorted
}

// SortByName sorts configs alphabetically by name
func SortByName(configs []models.ConfigEntry) []models.ConfigEntry {
	sorted := make([]models.ConfigEntry, len(configs))
	copy(sorted, configs)

	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})

	return sorted
}

// SortByType sorts configs by file type, then by name
func SortByType(configs []models.ConfigEntry) []models.ConfigEntry {
	sorted := make([]models.ConfigEntry, len(configs))
	copy(sorted, configs)

	sort.Slice(sorted, func(i, j int) bool {
		// First sort by type
		if !strings.EqualFold(sorted[i].Type, sorted[j].Type) {
			return strings.ToLower(sorted[i].Type) < strings.ToLower(sorted[j].Type)
		}

		// If types are the same, sort by name
		return strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
	})

	return sorted
}

// SortByPath sorts configs alphabetically by full path
func SortByPath(configs []models.ConfigEntry) []models.ConfigEntry {
	sorted := make([]models.ConfigEntry, len(configs))
	copy(sorted, configs)

	sort.Slice(sorted, func(i, j int) bool {
		return strings.ToLower(sorted[i].Path) < strings.ToLower(sorted[j].Path)
	})

	return sorted
}

// FindDuplicates checks if a config with the same path already exists
func FindDuplicates(configs []models.ConfigEntry, path string) *models.ConfigEntry {
	for i := range configs {
		if configs[i].Path == path {
			return &configs[i]
		}
	}
	return nil
}
