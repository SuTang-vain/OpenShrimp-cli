package cleanup

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-manager/internal/config"
	"ai-manager/internal/models"
)

// Cleaner handles cleanup of temporary files
type Cleaner struct {
	cfg *config.Config
}

// NewCleaner creates a new cleanup handler
func NewCleaner(cfg *config.Config) *Cleaner {
	return &Cleaner{cfg: cfg}
}

// CleanupAll runs cleanup for all enabled tools
func (c *Cleaner) CleanupAll() ([]models.CleanupResult, error) {
	results := make([]models.CleanupResult, 0)

	for key, tool := range c.cfg.Tools {
		if !tool.Enabled {
			continue
		}

		result := c.CleanupTool(key, tool)
		results = append(results, result)
	}

	return results, nil
}

// CleanupTool cleans temporary files for a specific tool
func (c *Cleaner) CleanupTool(key string, tool config.Tool) models.CleanupResult {
	result := models.CleanupResult{
		Tool:     tool.Name,
		Duration: time.Duration(0),
	}

	home, _ := os.UserHomeDir()
	basePath := expandPath(tool.Path, home)

	for _, tempPath := range tool.TempPaths {
		fullPath := filepath.Join(basePath, tempPath)

		// Check if path exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			continue
		}

		// Count and delete old files
		deleted, freed := c.cleanPath(fullPath, c.cfg.Retention.TempFiles)
		result.FilesDeleted += deleted
		result.SpaceFreed += freed
		result.Path = fullPath
	}

	return result
}

// cleanPath removes files older than specified days
func (c *Cleaner) cleanPath(path string, days int) (int, int64) {
	var deleted int
	var freed int64

	cutoff := time.Now().AddDate(0, 0, -days)

	filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if !info.IsDir() && info.ModTime().Before(cutoff) {
			size := info.Size()
			if err := os.Remove(filePath); err == nil {
				deleted++
				freed += size
			}
		}

		return nil
	})

	return deleted, freed
}

// CleanupGemini is a simplified cleanup for Gemini specifically
func (c *Cleaner) CleanupGemini() (models.CleanupResult, error) {
	home, _ := os.UserHomeDir()
	geminiPath := filepath.Join(home, ".gemini", "tmp")

	result := models.CleanupResult{
		Tool: "Gemini",
		Path: geminiPath,
	}

	if _, err := os.Stat(geminiPath); os.IsNotExist(err) {
		return result, nil
	}

	deleted, freed := c.cleanPath(geminiPath, c.cfg.Retention.TempFiles)
	result.FilesDeleted = deleted
	result.SpaceFreed = freed
	return result, nil
}

// CleanupClaude is a simplified cleanup for Claude Code specifically
func (c *Cleaner) CleanupClaude() (models.CleanupResult, error) {
	home, _ := os.UserHomeDir()

	result := models.CleanupResult{
		Tool:     "Claude Code",
		Duration: time.Duration(0),
	}

	// Clean debug directory
	debugPath := filepath.Join(home, ".claude", "debug")
	if _, err := os.Stat(debugPath); err == nil {
		deleted, freed := c.cleanPath(debugPath, c.cfg.Retention.DebugLogs)
		result.FilesDeleted += deleted
		result.SpaceFreed += freed
	}

	// Clean shell-snapshots
	snapshotsPath := filepath.Join(home, ".claude", "shell-snapshots")
	if _, err := os.Stat(snapshotsPath); err == nil {
		deleted, freed := c.cleanPath(snapshotsPath, c.cfg.Retention.ShellSnapshots)
		result.FilesDeleted += deleted
		result.SpaceFreed += freed
	}

	return result, nil
}

// expandPath expands ~ to home directory
func expandPath(path string, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return os.ExpandEnv(path)
}

// FormatResult formats a cleanup result for display
func FormatResult(r models.CleanupResult) string {
	if r.Error != nil {
		return "[Error] " + r.Error.Error()
	}
	return "[Done] " + r.Tool + ": " +
		models.FormatBytes(r.SpaceFreed) + " freed, " +
		string(rune('0'+r.FilesDeleted%10)) + " files deleted"
}
