package discovery

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"ai-manager/internal/config"
	"ai-manager/internal/models"
)

// Scanner scans the system for AI tools
type Scanner struct {
	cfg *config.Config
}

// NewScanner creates a new tool scanner
func NewScanner(cfg *config.Config) *Scanner {
	return &Scanner{cfg: cfg}
}

// Scan discovers all configured AI tools
func (s *Scanner) Scan() (*models.ScanResult, error) {
	result := &models.ScanResult{
		Tools:     make([]models.ToolInfo, 0),
		Timestamp: time.Now(),
	}

	for key, tool := range s.cfg.Tools {
		if !tool.Enabled {
			continue
		}

		info := s.discoverTool(key, tool)
		result.Tools = append(result.Tools, info)
		result.Enabled++
	}

	result.Total = len(result.Tools)
	return result, nil
}

// discoverTool discovers a single tool
func (s *Scanner) discoverTool(key string, tool config.Tool) models.ToolInfo {
	info := models.ToolInfo{
		Name:       tool.Name,
		Enabled:    tool.Enabled,
		ConfigPath: tool.ConfigPath,
		DataPath:   tool.DataPath,
	}

	// Expand paths
	home, _ := os.UserHomeDir()
	toolPath := expandPath(tool.Path, home)
	configPath := expandPath(tool.ConfigPath, home)

	// Check if tool exists
	if _, err := os.Stat(toolPath); os.IsNotExist(err) {
		info.Found = false
		info.Status = models.StatusNotFound
		return info
	}

	info.Found = true
	info.Path = toolPath

	// Check config file
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		info.Status = models.StatusWarning
	} else {
		info.Status = models.StatusOK
	}

	// Calculate disk usage
	if usage, err := models.CalculateDiskUsage(toolPath); err == nil {
		info.DiskUsage = usage
	}

	return info
}

// expandPath expands ~ and environment variables
func expandPath(path string, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return os.ExpandEnv(path)
}
