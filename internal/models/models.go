package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ToolInfo represents discovered AI tool information
type ToolInfo struct {
	Name       string      `json:"name"`
	Path       string      `json:"path"`
	Found      bool        `json:"found"`
	Enabled    bool        `json:"enabled"`
	ConfigPath string      `json:"config_path"`
	DataPath   string      `json:"data_path"`
	DiskUsage  DiskUsage   `json:"disk_usage"`
	LastUsed   time.Time   `json:"last_used"`
	Status     ToolStatus  `json:"status"`
}

// ToolStatus represents the health status of a tool
type ToolStatus string

const (
	StatusOK       ToolStatus = "ok"
	StatusWarning  ToolStatus = "warning"
	StatusError    ToolStatus = "error"
	StatusNotFound ToolStatus = "not_found"
)

// DiskUsage represents disk usage information
type DiskUsage struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	Files     int    `json:"files"`
}

// CleanupResult represents the result of a cleanup operation
type CleanupResult struct {
	Tool      string    `json:"tool"`
	Path      string    `json:"path"`
	FilesDeleted int    `json:"files_deleted"`
	SpaceFreed int64    `json:"space_freed"`
	Duration  time.Duration `json:"duration"`
	Error     error     `json:"error,omitempty"`
}

// ScanResult represents the result of a tool scan
type ScanResult struct {
	Tools     []ToolInfo  `json:"tools"`
	Total     int         `json:"total_found"`
	Enabled   int         `json:"enabled_count"`
	Timestamp time.Time   `json:"timestamp"`
}

// CalculateDiskUsage calculates disk usage for a path
func CalculateDiskUsage(path string) (DiskUsage, error) {
	var size int64
	var count int

	// Expand tilde
	path = expandHome(path)

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			size += info.Size()
			count++
		}
		return nil
	})

	if err != nil {
		return DiskUsage{Path: path}, err
	}

	return DiskUsage{
		Path:      path,
		SizeBytes: size,
		Files:     count,
	}, nil
}

// FormatBytes formats bytes to human readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// expandHome expands tilde to home directory
func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		if home != "" {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}
