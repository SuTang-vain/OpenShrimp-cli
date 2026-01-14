package tests

import (
	"os"
	"path/filepath"
	"testing"

	"ai-manager/internal/models"
)

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TB"},
	}

	for _, tt := range tests {
		result := models.FormatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d) = %s, expected %s", tt.bytes, result, tt.expected)
		}
	}
}

func TestCalculateDiskUsage(t *testing.T) {
	// Create a temporary directory with some files
	tempDir := t.TempDir()

	// Create test files with known sizes
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")
	subdir := filepath.Join(tempDir, "subdir")
	//nolint:errcheck
	os.MkdirAll(subdir, 0755)
	file3 := filepath.Join(subdir, "file3.txt")

	// Write known sizes (100 bytes each)
	//nolint:errcheck
	os.WriteFile(file1, make([]byte, 100), 0644)
	//nolint:errcheck
	os.WriteFile(file2, make([]byte, 200), 0644)
	//nolint:errcheck
	os.WriteFile(file3, make([]byte, 300), 0644)

	usage, err := models.CalculateDiskUsage(tempDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expectedSize := int64(600) // 100 + 200 + 300
	if usage.SizeBytes != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, usage.SizeBytes)
	}

	// Should count all 3 files
	if usage.Files != 3 {
		t.Errorf("expected 3 files, got %d", usage.Files)
	}

	if usage.Path != tempDir {
		t.Errorf("expected path %s, got %s", tempDir, usage.Path)
	}
}

func TestCalculateDiskUsage_NonExistentPath(t *testing.T) {
	nonExistent := "/path/that/does/not/exist/12345"

	// CalculateDiskUsage returns empty usage for non-existent paths
	// because filepath.Walk skips errors
	usage, err := models.CalculateDiskUsage(nonExistent)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	// Usage should be empty for non-existent paths
	if usage.SizeBytes != 0 {
		t.Errorf("expected 0 size for non-existent path, got %d", usage.SizeBytes)
	}
	if usage.Files != 0 {
		t.Errorf("expected 0 files for non-existent path, got %d", usage.Files)
	}
}

func TestToolInfo_Structure(t *testing.T) {
	info := models.ToolInfo{
		Name:       "Test Tool",
		Path:       "/test/path",
		Found:      true,
		Enabled:    true,
		ConfigPath: "settings.json",
		DataPath:   "data",
		DiskUsage: models.DiskUsage{
			SizeBytes: 1024,
			Files:     5,
		},
		Status: models.StatusOK,
	}

	if info.Name != "Test Tool" {
		t.Errorf("expected name Test Tool, got %s", info.Name)
	}
	if !info.Found {
		t.Error("expected found to be true")
	}
	if info.Status != models.StatusOK {
		t.Errorf("expected status OK, got %s", info.Status)
	}
	if info.DiskUsage.SizeBytes != 1024 {
		t.Errorf("expected size 1024, got %d", info.DiskUsage.SizeBytes)
	}
}

func TestToolStatus_Constants(t *testing.T) {
	// Verify status constants
	if models.StatusOK != "ok" {
		t.Errorf("expected StatusOK to be 'ok', got %s", models.StatusOK)
	}
	if models.StatusWarning != "warning" {
		t.Errorf("expected StatusWarning to be 'warning', got %s", models.StatusWarning)
	}
	if models.StatusError != "error" {
		t.Errorf("expected StatusError to be 'error', got %s", models.StatusError)
	}
	if models.StatusNotFound != "not_found" {
		t.Errorf("expected StatusNotFound to be 'not_found', got %s", models.StatusNotFound)
	}
}

func TestCleanupResult_Structure(t *testing.T) {
	result := models.CleanupResult{
		Tool:        "Test Tool",
		Path:        "/test/path",
		FilesDeleted: 10,
		SpaceFreed:  1024 * 500,
	}

	if result.Tool != "Test Tool" {
		t.Errorf("expected tool Test Tool, got %s", result.Tool)
	}
	if result.FilesDeleted != 10 {
		t.Errorf("expected 10 files deleted, got %d", result.FilesDeleted)
	}
	if result.SpaceFreed != 512000 {
		t.Errorf("expected 512000 space freed, got %d", result.SpaceFreed)
	}
}

func TestScanResult_Structure(t *testing.T) {
	result := models.ScanResult{
		Tools:     []models.ToolInfo{{Name: "Tool1"}, {Name: "Tool2"}},
		Total:     2,
		Enabled:   2,
	}

	if len(result.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(result.Tools))
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
	if result.Enabled != 2 {
		t.Errorf("expected enabled 2, got %d", result.Enabled)
	}
}
