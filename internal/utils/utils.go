package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// HomeDir returns the user's home directory
func HomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// ExpandPath expands ~ and environment variables in a path
func ExpandPath(path string) string {
	home := HomeDir()
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return os.ExpandEnv(path)
}

// EnsureDir ensures that a directory exists, creating it if necessary
func EnsureDir(path string) error {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// IsSymlink checks if a path is a symbolic link
func IsSymlink(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.Mode()&os.ModeSymlink != 0
}

// ReadSymlink returns the target of a symbolic link
func ReadSymlink(path string) (string, error) {
	return os.Readlink(path)
}

// CreateSymlink creates a symbolic link from link to target
func CreateSymlink(target, link string) error {
	// Remove existing link if it exists
	if _, err := os.Lstat(link); err == nil {
		if err := os.Remove(link); err != nil {
			return fmt.Errorf("failed to remove existing link: %w", err)
		}
	}

	return os.Symlink(target, link)
}

// CheckSymlink checks if a symlink is valid (target exists)
func CheckSymlink(path string) (bool, string, error) {
	if !IsSymlink(path) {
		return false, "", fmt.Errorf("not a symbolic link")
	}

	target, err := ReadSymlink(path)
	if err != nil {
		return false, target, err
	}

	expandedTarget := ExpandPath(target)
	_, err = os.Stat(expandedTarget)
	return err == nil, target, nil
}

// FileSize returns the size of a file or directory
func FileSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// FormatSize formats bytes to human readable format
func FormatSize(bytes int64) string {
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

// CountFiles counts the number of files in a directory
func CountFiles(path string) (int, error) {
	count := 0

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})

	return count, err
}

// IsDirEmpty checks if a directory is empty
func IsDirEmpty(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}
