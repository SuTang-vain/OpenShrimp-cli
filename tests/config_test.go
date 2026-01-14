package tests

import (
	"os"
	"path/filepath"
	"testing"

	"ai-manager/internal/config"
)

func TestLoad_DefaultConfig(t *testing.T) {
	// When config file doesn't exist, should return default config
	cfg, err := config.Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("expected no error for non-existent file, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", cfg.Version)
	}
	if cfg.HomeDir != "~/.ai-manager" {
		t.Errorf("expected home_dir ~/.ai-manager, got %s", cfg.HomeDir)
	}
	if len(cfg.Tools) != 3 {
		t.Errorf("expected 3 tools, got %d", len(cfg.Tools))
	}
	if len(cfg.Models) != 3 {
		t.Errorf("expected 3 models, got %d", len(cfg.Models))
	}
}

func TestLoad_CustomConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	customConfig := `
version: "2.0.0"
home_dir: "~/.custom-ai-manager"
tools:
  custom_tool:
    name: "Custom Tool"
    path: "~/.custom"
    enabled: false
models:
  custom-model:
    name: "Custom Model"
    provider: "custom"
    api_endpoint: "https://api.custom.com"
defaults:
  model: "custom-model"
  cleanup_days: 14
retention:
  debug_logs_days: 14
  temp_files_days: 7
  shell_snapshots_days: 60
`
	if err := os.WriteFile(configPath, []byte(customConfig), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Version != "2.0.0" {
		t.Errorf("expected version 2.0.0, got %s", cfg.Version)
	}
	if cfg.HomeDir != "~/.custom-ai-manager" {
		t.Errorf("expected home_dir ~/.custom-ai-manager, got %s", cfg.HomeDir)
	}
	if len(cfg.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(cfg.Tools))
	}
	if cfg.Tools["custom_tool"].Enabled {
		t.Error("expected custom_tool to be disabled")
	}
	if cfg.Defaults.Model != "custom-model" {
		t.Errorf("expected default model custom-model, got %s", cfg.Defaults.Model)
	}
	if cfg.Retention.TempFiles != 7 {
		t.Errorf("expected temp_files 7, got %d", cfg.Retention.TempFiles)
	}
}

func TestSave(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	cfg := &config.Config{
		Version: "1.0.0",
		HomeDir: "~/.test-ai-manager",
		Tools: map[string]config.Tool{
			"test_tool": {
				Name:    "Test Tool",
				Path:    "~/.test",
				Enabled: true,
			},
		},
		Models: map[string]config.Model{
			"test-model": {
				Name:        "Test Model",
				Provider:    "test",
				APIEndpoint: "https://api.test.com",
			},
		},
	}

	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Verify content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	if len(content) == 0 {
		t.Error("config file is empty")
	}

	// Load and verify
	loaded, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}
	if loaded.Version != cfg.Version {
		t.Errorf("version mismatch: expected %s, got %s", cfg.Version, loaded.Version)
	}
}

func TestGetDefaultConfigPath(t *testing.T) {
	path := config.GetDefaultConfigPath()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}
	expected := filepath.Join(home, ".ai-manager", "config.yaml")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestConfig_Tool_Structure(t *testing.T) {
	cfg, err := config.Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check Claude tool
	claude, ok := cfg.Tools["claude"]
	if !ok {
		t.Fatal("claude tool not found")
	}
	if claude.Name != "Claude Code" {
		t.Errorf("expected Claude Code, got %s", claude.Name)
	}
	if claude.Path != "~/.claude" {
		t.Errorf("expected ~/.claude, got %s", claude.Path)
	}
	if !claude.Enabled {
		t.Error("expected claude to be enabled")
	}
	if len(claude.TempPaths) == 0 {
		t.Error("expected temp paths to be defined")
	}
}

func TestConfig_Model_Structure(t *testing.T) {
	cfg, err := config.Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Check Claude Sonnet 4 model
	model, ok := cfg.Models["claude-sonnet-4"]
	if !ok {
		t.Fatal("claude-sonnet-4 model not found")
	}
	if model.Provider != "anthropic" {
		t.Errorf("expected provider anthropic, got %s", model.Provider)
	}
	if model.APIEndpoint != "https://api.anthropic.com" {
		t.Errorf("expected anthropic endpoint, got %s", model.APIEndpoint)
	}
}

func TestConfig_Defaults(t *testing.T) {
	cfg, err := config.Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Defaults.Model != "claude-sonnet-4" {
		t.Errorf("expected default model claude-sonnet-4, got %s", cfg.Defaults.Model)
	}
	if cfg.Defaults.Cleanup != 7 {
		t.Errorf("expected default cleanup 7, got %d", cfg.Defaults.Cleanup)
	}
}

func TestConfig_Retention(t *testing.T) {
	cfg, err := config.Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Retention.DebugLogs != 7 {
		t.Errorf("expected debug_logs 7, got %d", cfg.Retention.DebugLogs)
	}
	if cfg.Retention.TempFiles != 7 {
		t.Errorf("expected temp_files 7, got %d", cfg.Retention.TempFiles)
	}
	if cfg.Retention.ShellSnapshots != 30 {
		t.Errorf("expected shell_snapshots 30, got %d", cfg.Retention.ShellSnapshots)
	}
}
