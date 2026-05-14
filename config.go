package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LFroesch/zap/internal/models"
	"github.com/LFroesch/zap/internal/storage"
)

const (
	registryPathEnv = "ZAP_REGISTRY_PATH"
	demoDataPathEnv = "ZAP_DEMO_DATA_PATH"
)

func resolveRegistryPath() (string, error) {
	if override := os.Getenv(registryPathEnv); override != "" {
		return override, nil
	}

	if xdgHome := os.Getenv("XDG_CONFIG_HOME"); xdgHome != "" {
		return filepath.Join(xdgHome, "zap", "zap-registry.json"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve zap registry path: %w", err)
	}

	return filepath.Join(homeDir, ".config", "zap", "zap-registry.json"), nil
}

func loadConfigs(store *storage.Storage) ([]models.ConfigEntry, error) {
	if primaryExists, err := fileExists(store.GetFilePath()); err != nil {
		return nil, err
	} else if primaryExists {
		return store.Load()
	}

	if demoPath := os.Getenv(demoDataPathEnv); demoPath != "" {
		demoExists, err := fileExists(demoPath)
		if err != nil {
			return nil, err
		}
		if demoExists {
			return storage.New(demoPath).Load()
		}
	}

	return store.Load()
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("stat %s: %w", path, err)
}
