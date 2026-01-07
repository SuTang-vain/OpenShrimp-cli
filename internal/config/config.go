package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version     string            `yaml:"version"`
	HomeDir     string            `yaml:"home_dir"`
	Tools       map[string]Tool   `yaml:"tools"`
	Models      map[string]Model  `yaml:"models"`
	Defaults    Defaults          `yaml:"defaults"`
	Retention   RetentionPolicy   `yaml:"retention"`
}

type Tool struct {
	Name        string   `yaml:"name"`
	Path        string   `yaml:"path"`
	ConfigPath  string   `yaml:"config_path"`
	DataPath    string   `yaml:"data_path"`
	TempPaths   []string `yaml:"temp_paths"`
	Enabled     bool     `yaml:"enabled"`
}

type Model struct {
	Name        string `yaml:"name"`
	Provider    string `yaml:"provider"`
	APIEndpoint string `yaml:"api_endpoint"`
	ModelID     string `yaml:"model_id"`
	Environment map[string]string `yaml:"environment"`
}

type Defaults struct {
	Model    string `yaml:"model"`
	Cleanup  int    `yaml:"cleanup_days"`
}

type RetentionPolicy struct {
	DebugLogs      int `yaml:"debug_logs_days"`
	TempFiles      int `yaml:"temp_files_days"`
	ShellSnapshots int `yaml:"shell_snapshots_days"`
}

var defaultConfig = &Config{
	Version: "1.0.0",
	HomeDir: "~/.ai-manager",
	Tools: map[string]Tool{
		"claude": {
			Name:       "Claude Code",
			Path:       "~/.claude",
			ConfigPath: "settings.json",
			DataPath:   "projects",
			TempPaths:  []string{"debug", "shell-snapshots"},
			Enabled:    true,
		},
		"gemini": {
			Name:       "Gemini CLI",
			Path:       "~/.gemini",
			ConfigPath: "settings.json",
			DataPath:   "tmp",
			TempPaths:  []string{"tmp"},
			Enabled:    true,
		},
		"opencode": {
			Name:       "OpenCode",
			Path:       "~/.config/opencode",
			ConfigPath: "settings.json",
			DataPath:   "projects",
			TempPaths:  []string{"node_modules", ".cache"},
			Enabled:    true,
		},
	},
	Models: map[string]Model{
		"claude-sonnet-4": {
			Name:        "Claude Sonnet 4",
			Provider:    "anthropic",
			APIEndpoint: "https://api.anthropic.com",
			ModelID:     "claude-sonnet-4-20250514",
		},
		"minimax-m2.1": {
			Name:        "MiniMax M2.1",
			Provider:    "minimax",
			APIEndpoint: "https://api.minimaxi.com/anthropic",
			ModelID:     "miniMax-M2.1-200k",
		},
		"glm-4.7": {
			Name:        "GLM-4.7",
			Provider:    "zhipu",
			APIEndpoint: "https://open.bigmodel.cn/api/anthropic",
			ModelID:     "glm-4.7",
		},
	},
	Defaults: Defaults{
		Model:   "claude-sonnet-4",
		Cleanup: 7,
	},
	Retention: RetentionPolicy{
		DebugLogs:      7,
		TempFiles:      7,
		ShellSnapshots: 30,
	},
}

// Load loads the configuration from the specified path
func Load(configPath string) (*Config, error) {
	// If file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return defaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save saves the configuration to the specified path
func Save(cfg *Config, configPath string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// GetDefaultConfigPath returns the default config path
func GetDefaultConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ai-manager", "config.yaml")
}

// CreateDefaultConfig creates the default config file
func CreateDefaultConfig() error {
	cfg := defaultConfig
	path := GetDefaultConfigPath()
	return Save(cfg, path)
}
