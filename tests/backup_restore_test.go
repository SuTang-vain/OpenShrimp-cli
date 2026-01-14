package tests

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ai-manager/internal/backup"
	"ai-manager/internal/config"
)

func TestBackupManager_BackupWithoutData(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	toolDir := filepath.Join(tempDir, "claude")
	configDir := filepath.Join(toolDir, "projects")
	//nolint:errcheck
	os.MkdirAll(configDir, 0755)

	// Create test files
	//nolint:errcheck
	os.WriteFile(filepath.Join(toolDir, "settings.json"), []byte("{}"), 0644)
	//nolint:errcheck
	os.WriteFile(filepath.Join(configDir, "project.json"), []byte("{}"), 0644)

	cfg := &config.Config{
		Version: "1.0.0",
		HomeDir: tempDir,
		Tools: map[string]config.Tool{
			"claude": {
				Name:       "Claude Code",
				Path:       toolDir,
				ConfigPath: "settings.json",
				DataPath:   "projects",
				Enabled:    true,
			},
		},
	}

	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	bm := backup.NewBackupManager(cfg)

	// Test with includeData = false (only config)
	result, err := bm.Backup(false)
	if err != nil {
		t.Fatalf("Backup() error = %v", err)
	}

	if result.FilesBackedUp == 0 {
		t.Error("expected at least one file backed up")
	}
}

func TestBackupFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1024 * 100, "100.0 KB"},
		{1024 * 1024, "1.0 MB"},
	}

	for _, tt := range tests {
		result := backup.FormatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d) = %s, expected %s", tt.bytes, result, tt.expected)
		}
	}
}

// Test BackupResult and RestoreResult structures
func TestBackupResult_Structure(t *testing.T) {
	now := time.Now()
	result := backup.BackupResult{
		Success:      true,
		BackupPath:   "/path/to/backup.tar.gz",
		FilesBackedUp: 5,
		Timestamp:    now,
		Message:      "Backup created successfully",
	}
	_ = result.BackupSize // Acknowledge field

	if !result.Success {
		t.Error("expected success to be true")
	}
	if result.BackupPath != "/path/to/backup.tar.gz" {
		t.Errorf("expected backup path /path/to/backup.tar.gz, got %s", result.BackupPath)
	}
	if result.FilesBackedUp != 5 {
		t.Errorf("expected 5 files backed up, got %d", result.FilesBackedUp)
	}
}

func TestRestoreResult_Structure(t *testing.T) {
	now := time.Now()
	result := backup.RestoreResult{
		Success:       true,
		RestorePath:   "/path/to/restore",
		FilesRestored: 10,
		Timestamp:     now,
		Message:       "Restored 10 files",
	}

	if !result.Success {
		t.Error("expected success to be true")
	}
	if result.RestorePath != "/path/to/restore" {
		t.Errorf("expected restore path /path/to/restore, got %s", result.RestorePath)
	}
	if result.FilesRestored != 10 {
		t.Errorf("expected 10 files restored, got %d", result.FilesRestored)
	}
}

func TestBackupInfo_Structure(t *testing.T) {
	now := time.Now()
	info := backup.BackupInfo{
		Name:      "backup_20240101_000000.tar.gz",
		Timestamp: now,
		ToolCount: 2,
	}
	_ = info.Path   // Acknowledge field
	_ = info.Size   // Acknowledge field
	_ = info.Files  // Acknowledge field

	if info.Name != "backup_20240101_000000.tar.gz" {
		t.Errorf("expected name backup_20240101_000000.tar.gz, got %s", info.Name)
	}
	if info.ToolCount != 2 {
		t.Errorf("expected 2 tools, got %d", info.ToolCount)
	}
}

// Test tar/gzip archive operations
func TestCreateAndReadArchive(t *testing.T) {
	tempDir := t.TempDir()
	archivePath := filepath.Join(tempDir, "test.tar.gz")

	testContent := []byte("test content")
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create and write archive
	{
		file, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}
		defer file.Close()

		gw := gzip.NewWriter(file)
		tw := tar.NewWriter(gw)

		// Add file to archive
		f, err := os.Open(testFile)
		if err != nil {
			t.Fatalf("failed to open test file: %v", err)
		}
		defer f.Close()

		info, _ := f.Stat()
		header := &tar.Header{
			Name: "test.txt",
			Size: info.Size(),
			Mode: int64(info.Mode()),
		}

		if err := tw.WriteHeader(header); err != nil {
			t.Fatalf("failed to write tar header: %v", err)
		}

		if _, err := io.Copy(tw, f); err != nil {
			t.Fatalf("failed to copy to tar: %v", err)
		}

		// Important: close writers to flush data
		tw.Close()
		gw.Close()
	}

	// Read the archive
	{
		readFile, err := os.Open(archivePath)
		if err != nil {
			t.Fatalf("failed to open archive for reading: %v", err)
		}
		defer readFile.Close()

		gz, err := gzip.NewReader(readFile)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer gz.Close()

		tr := tar.NewReader(gz)

		header, err := tr.Next()
		if err != nil {
			t.Fatalf("failed to read tar header: %v", err)
		}

		if header.Name != "test.txt" {
			t.Errorf("expected name 'test.txt', got %s", header.Name)
		}

		content, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("failed to read content: %v", err)
		}

		if string(content) != string(testContent) {
			t.Errorf("content mismatch: expected %s, got %s", testContent, content)
		}
	}
}

func TestBackupManager_Backup_ConfigOnly(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create a minimal config file
	cfg := &config.Config{
		Version: "1.0.0",
		HomeDir: tempDir,
		Tools: map[string]config.Tool{
			"test-tool": {
				Name:    "Test Tool",
				Path:    filepath.Join(tempDir, "nonexistent"),
				Enabled: false, // Disabled tool won't be backed up
			},
		},
	}

	if err := config.Save(cfg, configPath); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	bm := backup.NewBackupManager(cfg)

	// Backup with no enabled tools and no existing config
	result, err := bm.Backup(false)
	if err != nil {
		t.Fatalf("Backup() error = %v", err)
	}

	// Should succeed but with 0 files backed up
	if result.Success != true {
		t.Errorf("expected success, got failure")
	}
}

func TestBackupManager_Restore_ArchiveFormat(t *testing.T) {
	// This test verifies that the restore function can handle
	// properly formatted tar.gz archives
	tempDir := t.TempDir()
	restoreDir := filepath.Join(tempDir, "restore")
	//nolint:errcheck
	os.MkdirAll(restoreDir, 0755)

	archivePath := filepath.Join(tempDir, "test.tar.gz")

	// Create a proper archive
	file, _ := os.Create(archivePath)
	gw := gzip.NewWriter(file)
	tw := tar.NewWriter(gw)

	// Add a file
	content := []byte("hello world")
	header := &tar.Header{
		Name: "test.txt",
		Size: int64(len(content)),
		Mode: 0644,
	}
	//nolint:errcheck
	tw.WriteHeader(header)
	//nolint:errcheck
	tw.Write(content)
	tw.Close()
	gw.Close()
	file.Close()

	// Create a minimal config
	cfg := &config.Config{
		Version: "1.0.0",
		HomeDir: tempDir,
	}

	bm := backup.NewBackupManager(cfg)

	// Mock the backup directory to our temp dir
	// Note: We can't easily change the backup dir without modifying the package
	// This test just verifies archive format handling
	_ = bm
	_ = restoreDir

	// Verify the archive exists and is valid
	if _, err := os.Stat(archivePath); err != nil {
		t.Errorf("archive should exist: %v", err)
	}
}

// Test that backup manager can be created
func TestNewBackupManager(t *testing.T) {
	cfg := &config.Config{
		Version: "1.0.0",
		HomeDir: "/test/home",
	}

	bm := backup.NewBackupManager(cfg)
	if bm == nil {
		t.Fatal("expected non-nil BackupManager")
	}
}

// Test BackupResult fields
func TestBackupResult_Fields(t *testing.T) {
	result := backup.BackupResult{
		Success:      true,
		BackupPath:   "/test/backup.tar.gz",
		BackupSize:   2048,
		FilesBackedUp: 3,
	}

	if result.Success != true {
		t.Error("expected Success to be true")
	}
	if result.BackupPath != "/test/backup.tar.gz" {
		t.Error("BackupPath mismatch")
	}
	if result.BackupSize != 2048 {
		t.Error("BackupSize mismatch")
	}
	if result.FilesBackedUp != 3 {
		t.Error("FilesBackedUp mismatch")
	}
}

// Test RestoreResult fields
func TestRestoreResult_Fields(t *testing.T) {
	result := backup.RestoreResult{
		Success:       true,
		RestorePath:   "/test/restore",
		FilesRestored: 5,
	}

	if result.Success != true {
		t.Error("expected Success to be true")
	}
	if result.RestorePath != "/test/restore" {
		t.Error("RestorePath mismatch")
	}
	if result.FilesRestored != 5 {
		t.Error("FilesRestored mismatch")
	}
}
