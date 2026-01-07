package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ai-mgr",
	Short: "AI Tools Manager - Unified management for AI development tools",
	Long: `AI Tools Manager helps you manage AI development tools like
Claude, Gemini, OpenCode, and VSCode in a unified way.

Features:
- Discover AI tools on your system
- Clean up temporary files
- Switch between different AI models
- Manage configurations with backups
- Health checks for your AI tools`,
	SilenceUsage: true,
}

func Run() error {
	// Add subcommands
	rootCmd.AddCommand(
		newScanCmd(),
		newCleanupCmd(),
		newSwitchCmd(),
		newLinkCmd(),
		newCheckCmd(),
		newBackupCmd(),
		newRestoreCmd(),
		newStatsCmd(),
		newVersionCmd(),
	)

	return rootCmd.Execute()
}

// Placeholder commands
func newSwitchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "switch",
		Short: "Switch between AI models",
		Long:  `Switch between different AI model configurations.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Model switching - coming soon!")
		},
	}
}

func newLinkCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "link",
		Short: "Manage symbolic links",
		Long:  `Create or verify symbolic links for AI tool configurations.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Link management - coming soon!")
		},
	}
}

func newBackupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "backup",
		Short: "Backup configurations",
		Long:  `Backup AI tool configurations to a safe location.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Backup - coming soon!")
		},
	}
}

func newRestoreCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "restore",
		Short: "Restore configurations",
		Long:  `Restore AI tool configurations from a backup.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("Restore - coming soon!")
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Long:  `Show the version of AI Tools Manager.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("AI Tools Manager v0.1.0")
		},
	}
}

// GetRootCmd returns the root command for testing purposes
func GetRootCmd() *cobra.Command {
	return rootCmd
}
