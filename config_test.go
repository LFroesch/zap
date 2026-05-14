package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LFroesch/zap/internal/storage"
)

func TestResolveRegistryPathPrefersOverride(t *testing.T) {
	t.Setenv(registryPathEnv, "/tmp/zap/custom.json")
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
	t.Setenv("HOME", "/tmp/home")

	path, err := resolveRegistryPath()
	if err != nil {
		t.Fatalf("resolveRegistryPath error = %v", err)
	}
	if path != "/tmp/zap/custom.json" {
		t.Fatalf("resolveRegistryPath = %q, want override", path)
	}
}

func TestResolveRegistryPathUsesXDGHome(t *testing.T) {
	t.Setenv(registryPathEnv, "")
	t.Setenv("XDG_CONFIG_HOME", "/tmp/xdg")
	t.Setenv("HOME", "/tmp/home")

	path, err := resolveRegistryPath()
	if err != nil {
		t.Fatalf("resolveRegistryPath error = %v", err)
	}
	want := filepath.Join("/tmp/xdg", "zap", "zap-registry.json")
	if path != want {
		t.Fatalf("resolveRegistryPath = %q, want %q", path, want)
	}
}

func TestResolveRegistryPathFallsBackToHomeConfig(t *testing.T) {
	t.Setenv(registryPathEnv, "")
	t.Setenv("XDG_CONFIG_HOME", "")

	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	path, err := resolveRegistryPath()
	if err != nil {
		t.Fatalf("resolveRegistryPath error = %v", err)
	}
	want := filepath.Join(homeDir, ".config", "zap", "zap-registry.json")
	if path != want {
		t.Fatalf("resolveRegistryPath = %q, want %q", path, want)
	}
}

func TestLoadConfigsUsesDemoFallbackWhenPrimaryMissing(t *testing.T) {
	tmpDir := t.TempDir()
	primaryPath := filepath.Join(tmpDir, "primary", "zap-registry.json")
	demoPath := filepath.Join(tmpDir, "demo.json")
	t.Setenv(demoDataPathEnv, demoPath)

	if err := os.WriteFile(demoPath, []byte(`{"configs":[{"name":"Demo File","path":"/tmp/demo.txt","type":"txt"}]}`), 0644); err != nil {
		t.Fatalf("write demo config: %v", err)
	}

	configs, err := loadConfigs(storage.New(primaryPath))
	if err != nil {
		t.Fatalf("loadConfigs error = %v", err)
	}
	if len(configs) != 1 || configs[0].Name != "Demo File" {
		t.Fatalf("loadConfigs loaded %+v, want demo fallback entry", configs)
	}
}

func TestLoadConfigsIgnoresDemoFallbackWhenPrimaryExists(t *testing.T) {
	tmpDir := t.TempDir()
	primaryPath := filepath.Join(tmpDir, "primary", "zap-registry.json")
	demoPath := filepath.Join(tmpDir, "demo.json")
	t.Setenv(demoDataPathEnv, demoPath)

	if err := os.MkdirAll(filepath.Dir(primaryPath), 0755); err != nil {
		t.Fatalf("mkdir primary dir: %v", err)
	}
	if err := os.WriteFile(primaryPath, []byte(`{"configs":[{"name":"Primary File","path":"/tmp/primary.txt","type":"txt"}]}`), 0644); err != nil {
		t.Fatalf("write primary config: %v", err)
	}
	if err := os.WriteFile(demoPath, []byte(`{"configs":[{"name":"Demo File","path":"/tmp/demo.txt","type":"txt"}]}`), 0644); err != nil {
		t.Fatalf("write demo config: %v", err)
	}

	configs, err := loadConfigs(storage.New(primaryPath))
	if err != nil {
		t.Fatalf("loadConfigs error = %v", err)
	}
	if len(configs) != 1 || configs[0].Name != "Primary File" {
		t.Fatalf("loadConfigs loaded %+v, want primary entry", configs)
	}
}

func TestLoadConfigsReturnsEmptyWhenNoPrimaryOrFallback(t *testing.T) {
	tmpDir := t.TempDir()
	primaryPath := filepath.Join(tmpDir, "primary", "zap-registry.json")
	t.Setenv(demoDataPathEnv, filepath.Join(tmpDir, "missing-demo.json"))

	configs, err := loadConfigs(storage.New(primaryPath))
	if err != nil {
		t.Fatalf("loadConfigs error = %v", err)
	}
	if len(configs) != 0 {
		t.Fatalf("loadConfigs loaded %+v, want empty configs", configs)
	}
	if _, err := os.Stat(filepath.Dir(primaryPath)); err != nil {
		t.Fatalf("expected primary config dir to be created, stat error = %v", err)
	}
}
