package tests

import (
	"os"
	"path/filepath"
	"testing"

	"ai-manager/internal/config"
	"ai-manager/internal/discovery"
	"ai-manager/internal/models"
)

func TestNewScanner(t *testing.T) {
	cfg := &config.Config{
		Tools: map[string]config.Tool{
			"claude": {Name: "Claude Code"},
		},
	}

	s := discovery.NewScanner(cfg)
	if s == nil {
		t.Fatal("expected non-nil scanner")
	}
}

func TestScanner_Scan_EmptyConfig(t *testing.T) {
	cfg := &config.Config{
		Tools: map[string]config.Tool{},
	}

	s := discovery.NewScanner(cfg)
	result, err := s.Scan()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Empty config uses default tools, so expect 3 tools
	if result.Total != 3 {
		t.Errorf("expected 3 default tools, got %d", result.Total)
	}
	// All default tools are enabled
	if result.Enabled != 3 {
		t.Errorf("expected 3 enabled tools, got %d", result.Enabled)
	}
}

func TestScanner_Scan_SkipsDisabledTools(t *testing.T) {
	cfg := &config.Config{
		Tools: map[string]config.Tool{
			"enabled": {
				Name:    "Enabled Tool",
				Path:    "~/nonexistent-enabled",
				Enabled: true,
			},
			"disabled": {
				Name:    "Disabled Tool",
				Path:    "~/nonexistent-disabled",
				Enabled: false,
			},
		},
	}

	s := discovery.NewScanner(cfg)
	result, err := s.Scan()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Only enabled tools should be in results
	if result.Enabled != 1 {
		t.Errorf("expected 1 enabled tool, got %d", result.Enabled)
	}
	if result.Total != 1 {
		t.Errorf("expected 1 total tool, got %d", result.Total)
	}
}

func TestScanner_Scan_NonExistentPaths(t *testing.T) {
	cfg := &config.Config{
		Tools: map[string]config.Tool{
			"test": {
				Name:    "Test Tool",
				Path:    "~/path/that/does/not/exist/12345",
				Enabled: true,
			},
		},
	}

	s := discovery.NewScanner(cfg)
	result, err := s.Scan()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Tools) != 1 {
		t.Fatalf("expected 1 tool in results, got %d", len(result.Tools))
	}

	tool := result.Tools[0]
	if tool.Found {
		t.Error("expected tool to not be found")
	}
	if tool.Status != models.StatusNotFound {
		t.Errorf("expected status not_found, got %s", tool.Status)
	}
}

func TestScanner_Scan_ExistingPath(t *testing.T) {
	// Create a temporary directory to simulate an existing tool
	tempDir := t.TempDir()

	cfg := &config.Config{
		Tools: map[string]config.Tool{
			"test": {
				Name:       "Test Tool",
				Path:       tempDir,
				Enabled:    true,
				ConfigPath: "config.json",
				DataPath:   "data",
			},
		},
	}

	s := discovery.NewScanner(cfg)
	result, err := s.Scan()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(result.Tools) != 1 {
		t.Fatalf("expected 1 tool in results, got %d", len(result.Tools))
	}

	tool := result.Tools[0]
	if !tool.Found {
		t.Error("expected tool to be found")
	}
	if tool.Path != tempDir {
		t.Errorf("expected path %s, got %s", tempDir, tool.Path)
	}
	if tool.Status != models.StatusWarning {
		t.Errorf("expected status warning (config missing), got %s", tool.Status)
	}
}

func TestScanner_Scan_WithConfigFile(t *testing.T) {
	// Create a temporary directory with a config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.json")
	//nolint:errcheck
	os.WriteFile(configPath, []byte("{}"), 0644)

	cfg := &config.Config{
		Tools: map[string]config.Tool{
			"test": {
				Name:       "Test Tool",
				Path:       tempDir,
				Enabled:    true,
				ConfigPath: configPath,
				DataPath:   "data",
			},
		},
	}

	s := discovery.NewScanner(cfg)
	result, err := s.Scan()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tool := result.Tools[0]
	if tool.Status != models.StatusOK {
		t.Errorf("expected status OK, got %s", tool.Status)
	}
}

func TestScanner_discoverTool(t *testing.T) {
	cfg := &config.Config{}
	s := discovery.NewScanner(cfg)

	// Test with non-existent path
	tool := config.Tool{
		Name:    "Test",
		Path:    "~/path/that/does/not/exist/xyz",
		Enabled: true,
	}

	// Test through Scan method since discoverTool is unexported
	cfg.Tools = map[string]config.Tool{"test": tool}
	result, err := s.Scan()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	info := result.Tools[0]
	if info.Found {
		t.Error("expected found to be false")
	}
	if info.Status != models.StatusNotFound {
		t.Errorf("expected status not_found, got %s", info.Status)
	}
	if info.Name != "Test" {
		t.Errorf("expected name Test, got %s", info.Name)
	}

	// Test with existing path
	existingPath := t.TempDir()
	tool.Path = existingPath
	cfg.Tools = map[string]config.Tool{"test": tool}
	result, err = s.Scan()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	info = result.Tools[0]
	if !info.Found {
		t.Error("expected found to be true")
	}
	if info.Status == models.StatusNotFound {
		t.Error("expected status not not_found for existing path")
	}
}
