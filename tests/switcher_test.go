package tests

import (
	"testing"

	"ai-manager/internal/config"
	"ai-manager/internal/switcher"
)

func TestNewSwitcher(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.Defaults{Model: "claude-sonnet-4"},
		Models: map[string]config.Model{
			"claude-sonnet-4": {
				Name:        "Claude Sonnet 4",
				Provider:    "anthropic",
				APIEndpoint: "https://api.anthropic.com",
			},
		},
	}

	s := switcher.NewSwitcher(cfg)
	if s == nil {
		t.Fatal("expected non-nil switcher")
	}
}

func TestListModels(t *testing.T) {
	cfg := &config.Config{
		Models: map[string]config.Model{
			"model-1": {Name: "Model 1"},
			"model-2": {Name: "Model 2"},
			"model-3": {Name: "Model 3"},
		},
	}

	s := switcher.NewSwitcher(cfg)
	models := s.ListModels()

	if len(models) != 3 {
		t.Errorf("expected 3 models, got %d", len(models))
	}

	// Check that all model names are present
	modelSet := make(map[string]bool)
	for _, m := range models {
		modelSet[m] = true
	}
	for name := range cfg.Models {
		if !modelSet[name] {
			t.Errorf("model %s not in list", name)
		}
	}
}

func TestGetCurrentModel(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.Defaults{Model: "minimax-m2.1"},
	}

	s := switcher.NewSwitcher(cfg)
	current := s.GetCurrentModel()

	if current != "minimax-m2.1" {
		t.Errorf("expected minimax-m2.1, got %s", current)
	}
}

func TestGetModel(t *testing.T) {
	cfg := &config.Config{
		Models: map[string]config.Model{
			"test-model": {
				Name:        "Test Model",
				Provider:    "test",
				APIEndpoint: "https://api.test.com",
				ModelID:     "test-model-id",
			},
		},
	}

	s := switcher.NewSwitcher(cfg)

	// Existing model
	model, ok := s.GetModel("test-model")
	if !ok {
		t.Error("expected to find test-model")
	}
	if model.Name != "Test Model" {
		t.Errorf("expected name Test Model, got %s", model.Name)
	}
	if model.Provider != "test" {
		t.Errorf("expected provider test, got %s", model.Provider)
	}

	// Non-existing model
	_, ok = s.GetModel("nonexistent")
	if ok {
		t.Error("expected not to find nonexistent model")
	}
}

func TestSwitchTo(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.Defaults{Model: "original"},
		Models: map[string]config.Model{
			"original": {Name: "Original"},
			"new":      {Name: "New"},
		},
	}

	s := switcher.NewSwitcher(cfg)

	// Switch to existing model
	if err := s.SwitchTo("new"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if s.GetCurrentModel() != "new" {
		t.Errorf("expected current model to be new, got %s", s.GetCurrentModel())
	}

	// Switch to non-existing model
	if err := s.SwitchTo("nonexistent"); err == nil {
		t.Error("expected error for nonexistent model")
	}
}

func TestShowModelInfo(t *testing.T) {
	cfg := &config.Config{
		Models: map[string]config.Model{
			"test-model": {
				Name:        "Test Model",
				Provider:    "test",
				APIEndpoint: "https://api.test.com",
				ModelID:     "test-id",
				Environment: map[string]string{
					"API_KEY": "secret",
					"DEBUG":   "true",
				},
			},
		},
	}

	s := switcher.NewSwitcher(cfg)

	// Non-existing model
	if err := s.ShowModelInfo("nonexistent"); err == nil {
		t.Error("expected error for nonexistent model")
	}

	// Existing model - just verify it doesn't error
	if err := s.ShowModelInfo("test-model"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestShowCurrentConfig(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.Defaults{Model: "current-model"},
		Models: map[string]config.Model{
			"current-model": {
				Name:        "Current Model",
				Provider:    "current",
				APIEndpoint: "https://api.current.com",
				ModelID:     "current-id",
			},
			"other-model": {
				Name: "Other Model",
			},
		},
	}

	s := switcher.NewSwitcher(cfg)

	if err := s.ShowCurrentConfig(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestExportEnvConfig(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.Defaults{Model: "test-model"},
		Models: map[string]config.Model{
			"test-model": {
				Name:        "Test Model",
				Provider:    "test",
				APIEndpoint: "https://api.test.com",
				ModelID:     "test-id",
				Environment: map[string]string{
					"API_KEY": "secret-value",
					"DEBUG":   "false",
				},
			},
		},
	}

	s := switcher.NewSwitcher(cfg)

	envOutput, err := s.ExportEnvConfig("test-model")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if envOutput == "" {
		t.Error("expected non-empty env output")
	}

	// Check for expected content
	expected := []string{
		"AI_PROVIDER=test",
		"AI_MODEL=test-id",
		"AI_API_ENDPOINT=https://api.test.com",
		"API_KEY=secret-value",
		"DEBUG=false",
	}
	for _, exp := range expected {
		if !contains(envOutput, exp) {
			t.Errorf("expected output to contain %s", exp)
		}
	}

	// Non-existing model
	_, err = s.ExportEnvConfig("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent model")
	}
}

func TestPerformSwitch(t *testing.T) {
	cfg := &config.Config{
		Defaults: config.Defaults{Model: "original"},
		Models: map[string]config.Model{
			"original": {Name: "Original"},
			"new": {
				Name:        "New Model",
				Provider:    "new",
				APIEndpoint: "https://api.new.com",
			},
		},
	}

	s := switcher.NewSwitcher(cfg)

	// Perform switch without export
	result, err := s.PerformSwitch("new", false)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.Message)
	}
	if result.Model != "new" {
		t.Errorf("expected model new, got %s", result.Model)
	}
	if result.Provider != "new" {
		t.Errorf("expected provider new, got %s", result.Provider)
	}
	if result.EnvExported != "" {
		t.Error("expected no env export when exportEnv is false")
	}

	// Switch with export
	cfg.Defaults.Model = "original"
	result, err = s.PerformSwitch("new", true)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.EnvExported == "" {
		t.Error("expected env export when exportEnv is true")
	}

	// Switch to non-existing model
	_, err = s.PerformSwitch("nonexistent", false)
	if err == nil {
		t.Error("expected error for nonexistent model")
	}
}

func TestValidateAPIKey(t *testing.T) {
	// Note: This test verifies the function structure
	// The actual file-based validation depends on test environment

	// Non-existing model file should return false, nil
	exists, err := switcher.ValidateAPIKey("totally-fake-model-xyz")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if exists {
		t.Error("expected false for non-existent API key file")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
