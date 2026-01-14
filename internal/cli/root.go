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
		newContextCmd(),
		newSchedulerCmd(),
		newCredentialsCmd(),
	)

	return rootCmd.Execute()
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Long:  `Show the version of AI Tools Manager.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println("AI Tools Manager v0.2.0")
		},
	}
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}
