package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-manager/internal/config"
	"ai-manager/internal/models"
)

// BackupResult represents the result of a backup operation
type BackupResult struct {
	Success      bool      `json:"success"`
	BackupPath   string    `json:"backup_path"`
	BackupSize  int64     `json:"backup_size_bytes"`
	FilesBackedUp int     `json:"files_backed_up"`
	Timestamp    time.Time `json:"timestamp"`
	Message      string    `json:"message"`
	Error        error     `json:"error,omitempty"`
}

// RestoreResult represents the result of a restore operation
type RestoreResult struct {
	Success       bool      `json:"success"`
	RestorePath   string    `json:"restore_path"`
	FilesRestored int       `json:"files_restored"`
	Timestamp     time.Time `json:"timestamp"`
	Message       string    `json:"message"`
	Error         error     `json:"error,omitempty"`
}

// BackupInfo represents metadata about a backup
type BackupInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	Size         int64     `json:"size_bytes"`
	Files        int       `json:"file_count"`
	Timestamp    time.Time `json:"timestamp"`
	ToolCount    int       `json:"tool_count"`
}

// BackupManager handles backup and restore operations
type BackupManager struct {
	cfg        *config.Config
	backupDir  string
}

// NewBackupManager creates a new BackupManager
func NewBackupManager(cfg *config.Config) *BackupManager {
	home, _ := os.UserHomeDir()
	backupDir := filepath.Join(home, ".ai-manager", "backups")
	return &BackupManager{
		cfg:       cfg,
		backupDir: backupDir,
	}
}

// Backup creates a backup of all enabled tool configurations
func (bm *BackupManager) Backup(includeData bool) (*BackupResult, error) {
	// Create backup directory if it doesn't exist
	if err := os.MkdirAll(bm.backupDir, 0755); err != nil {
		return &BackupResult{Success: false, Error: err}, err
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	backupName := fmt.Sprintf("backup_%s.tar.gz", timestamp)
	backupPath := filepath.Join(bm.backupDir, backupName)

	// Create tar.gz archive
	file, err := os.Create(backupPath)
	if err != nil {
		return &BackupResult{Success: false, Error: err}, err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	filesBackedUp := 0

	// Backup each enabled tool's configuration
	for toolName, tool := range bm.cfg.Tools {
		if !tool.Enabled {
			continue
		}

		// Backup config file
		if tool.ConfigPath != "" {
			configPath := expandPath(tool.Path, tool.ConfigPath)
			if _, err := os.Stat(configPath); err == nil {
				if err := addFileToTar(tw, configPath, filepath.Join("tools", toolName, "config")); err == nil {
					filesBackedUp++
				}
			}
		}

		// Optionally backup data directory
		if includeData && tool.DataPath != "" {
			dataPath := expandPath(tool.Path, tool.DataPath)
			if _, err := os.Stat(dataPath); err == nil {
				if err := addDirectoryToTar(tw, dataPath, filepath.Join("tools", toolName, "data")); err == nil {
					filesBackedUp++
				}
			}
		}
	}

	// Backup ai-manager config
	configPath := config.GetDefaultConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		//nolint:errcheck
		addFileToTar(tw, configPath, "ai-manager-config.yaml")
		filesBackedUp++
	}

	// Get file size
	info, _ := file.Stat()
	size := info.Size()

	return &BackupResult{
		Success:      true,
		BackupPath:   backupPath,
		BackupSize:   size,
		FilesBackedUp: filesBackedUp,
		Timestamp:    time.Now(),
		Message:      fmt.Sprintf("Backup created successfully: %s", backupName),
	}, nil
}

// Restore restores from a backup file
func (bm *BackupManager) Restore(backupName string, restorePath string) (*RestoreResult, error) {
	backupPath := filepath.Join(bm.backupDir, backupName)

	// Check if backup exists
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return &RestoreResult{Success: false, Error: err}, fmt.Errorf("backup not found: %s", backupName)
	}

	// Create restore directory if specified
	if restorePath != "" {
		if err := os.MkdirAll(restorePath, 0755); err != nil {
			return &RestoreResult{Success: false, Error: err}, err
		}
	} else {
		restorePath = bm.backupDir
	}

	// Open backup file
	file, err := os.Open(backupPath)
	if err != nil {
		return &RestoreResult{Success: false, Error: err}, err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return &RestoreResult{Success: false, Error: err}, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	filesRestored := 0

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // Skip corrupt entries
		}

		targetPath := filepath.Join(restorePath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, 0755); err != nil {
				continue
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				continue
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				continue
			}
			if _, err := io.Copy(outFile, tr); err == nil {
				filesRestored++
			}
			outFile.Close()
		}
	}

	return &RestoreResult{
		Success:       true,
		RestorePath:   restorePath,
		FilesRestored: filesRestored,
		Timestamp:     time.Now(),
		Message:       fmt.Sprintf("Restored %d files to %s", filesRestored, restorePath),
	}, nil
}

// ListBackups returns a list of all available backups
func (bm *BackupManager) ListBackups() ([]BackupInfo, error) {
	entries, err := os.ReadDir(bm.backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []BackupInfo{}, nil
		}
		return nil, err
	}

	var backups []BackupInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}

		info, _ := entry.Info()
		backupPath := filepath.Join(bm.backupDir, entry.Name())

		// Get file count from archive
		fileCount, _ := countFilesInArchive(backupPath)

		backups = append(backups, BackupInfo{
			Name:      entry.Name(),
			Path:      backupPath,
			Size:      info.Size(),
			Files:     fileCount,
			Timestamp: info.ModTime(),
		})
	}

	return backups, nil
}

// DeleteBackup removes a backup file
func (bm *BackupManager) DeleteBackup(backupName string) error {
	backupPath := filepath.Join(bm.backupDir, backupName)
	return os.Remove(backupPath)
}

// Helper functions

func expandPath(base, rel string) string {
	if strings.HasPrefix(base, "~/") {
		home, _ := os.UserHomeDir()
		if home != "" {
			base = filepath.Join(home, base[2:])
		}
	}
	return filepath.Join(base, rel)
}

func addFileToTar(tw *tar.Writer, filePath, arcName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    arcName,
		Size:    info.Size(),
		Mode:    int64(info.Mode()),
		ModTime: info.ModTime(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}

func addDirectoryToTar(tw *tar.Writer, dirPath, arcName string) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		header.Name = filepath.Join(arcName, relPath)

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tw, file)
			return err
		}

		return nil
	})
}

func countFilesInArchive(archivePath string) (int, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	gz, err := gzip.NewReader(file)
	if err != nil {
		return 0, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	count := 0

	for {
		_, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}
		count++
	}

	return count, nil
}

// FormatBytes formats bytes to human readable format (from models)
func FormatBytes(bytes int64) string {
	return models.FormatBytes(bytes)
}
