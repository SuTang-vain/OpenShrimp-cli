package daemon

import (
	"log"
	"os"
	"os/user"

	"ai-manager/internal/config"
)

// Run starts the daemon
func Run() {
	addr := os.Getenv("AIMGR_DAEMON_ADDR")
	if addr == "" {
		addr = ServerAddr
	}

	// Load configuration
	cfg, err := config.Load(config.GetDefaultConfigPath())
	if err != nil {
		log.Printf("Warning: failed to load config: %v", err)
		cfg = &config.Config{}
	}

	// Check if running as current user
	currentUser, _ := user.Current()
	if currentUser != nil {
		log.Printf("Running as user: %s", currentUser.Username)
	}

	// Create and start server
	server := NewServer(cfg)
	if err := server.Start(addr); err != nil {
		log.Fatalf("Failed to start daemon: %v", err)
	}
}
