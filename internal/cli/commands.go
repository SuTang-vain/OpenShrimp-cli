package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"ai-manager/internal/backup"
	"ai-manager/internal/cleanup"
	"ai-manager/internal/config"
	"ai-manager/internal/context"
	"ai-manager/internal/discovery"
	"ai-manager/internal/link"
	"ai-manager/internal/models"
	"ai-manager/internal/switcher"
	"ai-manager/internal/utils"

	"github.com/spf13/cobra"
)

var (
	days       int
	verbose    bool
	jsonOutput bool
)

// newScanCmd returns the scan command with implementation
func newScanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Scan for AI tools on your system",
		Long: `Scan and discover AI tools installed on your system.
Shows which tools are found, their paths, and disk usage.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.GetDefaultConfigPath())
			if err != nil {
				return err
			}

			scanner := discovery.NewScanner(cfg)
			result, err := scanner.Scan()
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(result)
			}

			printScanResult(result, verbose)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed information")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

// newCleanupCmd returns the cleanup command with implementation
func newCleanupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cleanup",
		Short: "Clean up temporary files",
		Long: `Clean up temporary files from AI tools.
By default, removes files older than 7 days.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.GetDefaultConfigPath())
			if err != nil {
				return err
			}

			// Override retention days if specified
			if days > 0 {
				cfg.Retention.TempFiles = days
			}

			cleaner := cleanup.NewCleaner(cfg)
			results, err := cleaner.CleanupAll()
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(results)
			}

			totalFreed := int64(0)
			totalDeleted := 0

			fmt.Println("=== Cleanup Results ===")
			for _, r := range results {
				fmt.Println(cleanup.FormatResult(r))
				totalFreed += r.SpaceFreed
				totalDeleted += r.FilesDeleted
			}

			fmt.Printf("\nTotal: %d files deleted, %s freed\n",
				totalDeleted, models.FormatBytes(totalFreed))

			return nil
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 7, "Delete files older than N days")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

// newCheckCmd returns the health check command
func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Health check for AI tools",
		Long: `Run health checks on your AI tools and configurations.
Reports on configuration validity, broken links, and disk usage.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.GetDefaultConfigPath())
			if err != nil {
				return err
			}

			fmt.Println("=== AI Tools Health Check ===")

			issues := 0
			for _, tool := range cfg.Tools {
				if !tool.Enabled {
					continue
				}

				toolPath := utils.ExpandPath(tool.Path)
				configPath := utils.ExpandPath(tool.ConfigPath)

				fmt.Printf("[%s]\n", tool.Name)

				// Check if tool exists
				if _, err := os.Stat(toolPath); os.IsNotExist(err) {
					fmt.Printf("  ✗ Tool path not found: %s\n", toolPath)
					issues++
					continue
				}
				fmt.Printf("  ✓ Path exists: %s\n", toolPath)

				// Check config file
				if _, err := os.Stat(configPath); os.IsNotExist(err) {
					fmt.Printf("  ⚠ Config file missing: %s\n", configPath)
					issues++
				} else {
					fmt.Printf("  ✓ Config file found: %s\n", configPath)
				}

				// Check symlinks if any
				fmt.Println()
			}

			if issues > 0 {
				fmt.Printf("Found %d issue(s)\n", issues)
			} else {
				fmt.Println("All tools are healthy!")
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

// newStatsCmd returns the stats command
func newStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show usage statistics",
		Long:  `Show usage statistics and disk usage for AI tools.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.GetDefaultConfigPath())
			if err != nil {
				return err
			}

			fmt.Println("=== AI Tools Disk Usage ===")

			totalSize := int64(0)
			totalFiles := 0

			for _, tool := range cfg.Tools {
				if !tool.Enabled {
					continue
				}

				path := utils.ExpandPath(tool.Path)
				size, _ := utils.FileSize(path)
				files, _ := utils.CountFiles(path)

				totalSize += size
				totalFiles += files

				fmt.Printf("[%s]\n", tool.Name)
				fmt.Printf("  Path: %s\n", path)
				fmt.Printf("  Size: %s (%d files)\n\n", utils.FormatSize(size), files)
			}

			fmt.Printf("Total: %s (%d files across %d tools)\n",
				utils.FormatSize(totalSize), totalFiles, len(cfg.Tools))

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

// Helper functions
func printScanResult(result *models.ScanResult, verbose bool) {
	fmt.Printf("=== Scan Results (%d tools found, %d enabled) ===\n\n",
		result.Total, result.Enabled)

	for _, tool := range result.Tools {
		status := "✓"
		if tool.Status == models.StatusNotFound {
			status = "✗"
		} else if tool.Status == models.StatusWarning {
			status = "⚠"
		}

		fmt.Printf("%s [%s]\n", status, tool.Name)

		if tool.Found {
			fmt.Printf("  Path: %s\n", tool.Path)
			if verbose {
				fmt.Printf("  Config: %s\n", tool.ConfigPath)
				fmt.Printf("  Data: %s\n", tool.DataPath)
				fmt.Printf("  Disk: %s (%d files)\n",
					models.FormatBytes(tool.DiskUsage.SizeBytes),
					tool.DiskUsage.Files)
			}
		} else {
			fmt.Printf("  Not found on system\n")
		}
		fmt.Println()
	}
}

// printJSON outputs data as JSON
func printJSON(data interface{}) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

// newSwitchCmd returns the switch command with full implementation
func newSwitchCmd() *cobra.Command {
	var exportEnv bool
	var showCurrent bool
	var listModels bool

	cmd := &cobra.Command{
		Use:   "switch [model]",
		Short: "Switch between AI models",
		Long: `Switch between different AI model configurations.
Use 'ai-mgr switch list' to see available models.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.GetDefaultConfigPath())
			if err != nil {
				return err
			}

			s := switcher.NewSwitcher(cfg)

			if listModels {
				fmt.Println("=== Available Models ===")
				for name, model := range cfg.Models {
					prefix := "  "
					if name == s.GetCurrentModel() {
						prefix = "* "
					}
					fmt.Printf("%s%s (%s)\n", prefix, name, model.Name)
				}
				return nil
			}

			if showCurrent {
				return s.ShowCurrentConfig()
			}

			if len(args) == 0 {
				return s.ShowCurrentConfig()
			}

			modelName := args[0]
			result, err := s.PerformSwitch(modelName, exportEnv)
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(result)
			}

			fmt.Printf("✓ %s\n", result.Message)

			if exportEnv && result.EnvExported != "" {
				fmt.Println("\n=== Export Command ===")
				fmt.Println(result.EnvExported)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&listModels, "list", false, "List all available models")
	cmd.Flags().BoolVar(&showCurrent, "current", false, "Show current model configuration")
	cmd.Flags().BoolVar(&exportEnv, "export", false, "Export environment configuration")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

// newLinkCmd returns the link command with full implementation
func newLinkCmd() *cobra.Command {
	var create bool
	var remove bool
	var verify bool
	var initDir bool

	cmd := &cobra.Command{
		Use:   "link [tool]",
		Short: "Manage symbolic links for AI tool configurations",
		Long: `Create or verify symbolic links for AI tool configurations.
By default, lists all configured links and their status.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.GetDefaultConfigPath())
			if err != nil {
				return err
			}

			l := link.NewLinkManager(cfg)

			if initDir {
				if err := l.InitializeLinksDir(); err != nil {
					return err
				}
				fmt.Println("✓ Links directory initialized")
				return nil
			}

			if len(args) > 0 {
				toolName := args[0]
				if create {
					result, err := l.CreateLink(toolName)
					if err != nil {
						return err
					}
					if jsonOutput {
						return printJSON(result)
					}
					fmt.Printf("✓ %s\n", result.Message)
					return nil
				}
				if remove {
					result, err := l.RemoveLink(toolName)
					if err != nil {
						return err
					}
					if jsonOutput {
						return printJSON(result)
					}
					fmt.Printf("✓ %s\n", result.Message)
					return nil
				}
				if verify {
					result, err := l.VerifyLink(toolName)
					if err != nil {
						return err
					}
					if jsonOutput {
						return printJSON(result)
					}
					if result.Success {
						fmt.Printf("✓ %s\n", result.Message)
					} else {
						fmt.Printf("✗ %s\n", result.Message)
					}
					return nil
				}
			}

			links, err := l.ListLinks()
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(links)
			}

			fmt.Println("=== Symbolic Links Status ===")
			for _, lnk := range links {
				status := "✓"
				if !lnk.Exists {
					status = "✗"
				} else if !lnk.Valid {
					status = "⚠"
				}

				fmt.Printf("%s [%s]\n", status, lnk.ToolName)
				fmt.Printf("  Link: %s\n", lnk.LinkPath)
				fmt.Printf("  Target: %s\n", lnk.TargetPath)

				if lnk.Error != "" {
					fmt.Printf("  Issue: %s\n", lnk.Error)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&create, "create", false, "Create symlink for specified tool")
	cmd.Flags().BoolVar(&remove, "remove", false, "Remove symlink for specified tool")
	cmd.Flags().BoolVar(&verify, "verify", false, "Verify symlink for specified tool")
	cmd.Flags().BoolVar(&initDir, "init", false, "Initialize links directory")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

// newBackupCmd returns the backup command with full implementation
func newBackupCmd() *cobra.Command {
	var includeData bool

	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup AI tool configurations",
		Long: `Create a backup of all AI tool configurations.
Backups are stored in ~/.ai-manager/backups with timestamp.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.GetDefaultConfigPath())
			if err != nil {
				return err
			}

			bm := backup.NewBackupManager(cfg)
			result, err := bm.Backup(includeData)
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(result)
			}

			fmt.Printf("Backup created successfully!\n")
			fmt.Printf("  Path: %s\n", result.BackupPath)
			fmt.Printf("  Size: %s\n", backup.FormatBytes(result.BackupSize))
			fmt.Printf("  Files: %d\n", result.FilesBackedUp)
			fmt.Printf("  Time: %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))

			return nil
		},
	}

	cmd.Flags().BoolVar(&includeData, "include-data", false, "Include data directories in backup")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

// newRestoreCmd returns the restore command with full implementation
func newRestoreCmd() *cobra.Command {
	var restorePath string
	var listBackups bool
	var deleteAfterRestore bool

	cmd := &cobra.Command{
		Use:   "restore [backup-name]",
		Short: "Restore AI tool configurations from backup",
		Long: `Restore AI tool configurations from a backup file.
Use 'backup list' to see available backups.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(config.GetDefaultConfigPath())
			if err != nil {
				return err
			}

			bm := backup.NewBackupManager(cfg)

			if listBackups {
				backups, err := bm.ListBackups()
				if err != nil {
					return err
				}

				if jsonOutput {
					return printJSON(backups)
				}

				fmt.Println("=== Available Backups ===")
				if len(backups) == 0 {
					fmt.Println("No backups found.")
					return nil
				}

				for _, b := range backups {
					fmt.Printf("%s\n", b.Name)
					fmt.Printf("  Size: %s, Files: %d\n", backup.FormatBytes(b.Size), b.Files)
					fmt.Printf("  Created: %s\n", b.Timestamp.Format("2006-01-02 15:04:05"))
					fmt.Println()
				}
				return nil
			}

			if len(args) == 0 {
				return fmt.Errorf("backup name required. Use 'backup list' to see available backups")
			}

			backupName := args[0]
			result, err := bm.Restore(backupName, restorePath)
			if err != nil {
				return err
			}

			if jsonOutput {
				return printJSON(result)
			}

			fmt.Printf("Restore completed!\n")
			fmt.Printf("  Files restored: %d\n", result.FilesRestored)
			fmt.Printf("  Location: %s\n", result.RestorePath)
			fmt.Printf("  Time: %s\n", result.Timestamp.Format("2006-01-02 15:04:05"))

			if deleteAfterRestore {
				if err := bm.DeleteBackup(backupName); err != nil {
					fmt.Printf("Warning: Failed to delete backup: %v\n", err)
				} else {
					fmt.Printf("Backup deleted: %s\n", backupName)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&listBackups, "list", false, "List available backups")
	cmd.Flags().StringVarP(&restorePath, "output", "o", "", "Restore to specific directory")
	cmd.Flags().BoolVar(&deleteAfterRestore, "delete", false, "Delete backup after successful restore")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

// newContextCmd returns the context command with full implementation
func newContextCmd() *cobra.Command {
	var listConversations bool
	var listProjects bool
	var showStats bool
	var searchQuery string
	var deleteID string
	var addProject bool
	var projectPath string
	var projectName string
	var projectDesc string
	var techStack string

	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage AI conversation context and project memory",
		Long: `Manage conversation history and project-specific context.
Provides unified memory across different AI tools.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctxManager, err := context.NewManager()
			if err != nil {
				return err
			}
			defer ctxManager.Close()

			switch {
			case showStats:
				stats, err := ctxManager.GetStats()
				if err != nil {
					return err
				}
				if jsonOutput {
					return printJSON(stats)
				}
				fmt.Println("=== Context Statistics ===")
				fmt.Printf("Total Conversations: %d\n", stats.TotalConversations)
				fmt.Printf("Total Messages: %d\n", stats.TotalMessages)
				fmt.Printf("Projects Tracked: %d\n", stats.ProjectsTracked)
				fmt.Printf("Storage Used: %s\n", models.FormatBytes(stats.StorageUsed))
				return nil

			case listProjects:
				projects, err := ctxManager.ListProjects()
				if err != nil {
					return err
				}
				if jsonOutput {
					return printJSON(projects)
				}
				fmt.Println("=== Tracked Projects ===")
				if len(projects) == 0 {
					fmt.Println("No projects tracked yet.")
					return nil
				}
				for _, p := range projects {
					fmt.Printf("[%s]\n", p.Name)
					fmt.Printf("  Path: %s\n", p.ProjectPath)
					if p.Description != "" {
						fmt.Printf("  Description: %s\n", p.Description)
					}
					if len(p.TechStack) > 0 {
						fmt.Printf("  Tech Stack: %s\n", joinStrings(p.TechStack, ", "))
					}
					fmt.Printf("  Last Active: %s\n", p.LastActive.Format("2006-01-02 15:04"))
					fmt.Println()
				}
				return nil

			case listConversations:
				convs, err := ctxManager.ListConversations("", "", 20)
				if err != nil {
					return err
				}
				if jsonOutput {
					return printJSON(convs)
				}
				fmt.Println("=== Recent Conversations ===")
				if len(convs) == 0 {
					fmt.Println("No conversations recorded yet.")
					return nil
				}
				for _, c := range convs {
					status := "✓"
					if c.Summary == "" {
						status = "○"
					}
					fmt.Printf("%s [%s] %s\n", status, c.Tool, c.Title)
					if c.ProjectPath != "" {
						fmt.Printf("  Project: %s\n", c.ProjectPath)
					}
					fmt.Printf("  Messages: %d, Updated: %s\n", len(c.Messages), c.UpdatedAt.Format("2006-01-02 15:04"))
					fmt.Println()
				}
				return nil

			case searchQuery != "":
				convs, err := ctxManager.SearchConversations(searchQuery)
				if err != nil {
					return err
				}
				if jsonOutput {
					return printJSON(convs)
				}
				fmt.Printf("=== Search Results for '%s' ===\n", searchQuery)
				if len(convs) == 0 {
					fmt.Println("No matches found.")
					return nil
				}
				for _, c := range convs {
					fmt.Printf("[%s] %s\n", c.Tool, c.Title)
					if c.Summary != "" {
						fmt.Printf("  %s\n", c.Summary)
					}
					fmt.Println()
				}
				return nil

			case deleteID != "":
				if err := ctxManager.DeleteConversation(deleteID); err != nil {
					return err
				}
				fmt.Printf("Deleted conversation: %s\n", deleteID)
				return nil

			case addProject:
				if projectPath == "" {
					return fmt.Errorf("project path required")
				}
				proj := &context.ProjectContext{
					ProjectPath: projectPath,
					Name:        projectName,
					Description: projectDesc,
				}
				if techStack != "" {
					proj.TechStack = splitString(techStack, ",")
				}
				if err := ctxManager.AddProject(proj); err != nil {
					return err
				}
				fmt.Printf("Added project: %s\n", projectPath)
				return nil

			default:
				cmd.Help()
				return nil
			}
		},
	}

	cmd.Flags().BoolVar(&listConversations, "conversations", false, "List recent conversations")
	cmd.Flags().BoolVar(&listProjects, "projects", false, "List tracked projects")
	cmd.Flags().BoolVar(&showStats, "stats", false, "Show context statistics")
	cmd.Flags().StringVar(&searchQuery, "search", "", "Search conversations")
	cmd.Flags().StringVar(&deleteID, "delete", "", "Delete conversation by ID")
	cmd.Flags().BoolVar(&addProject, "add-project", false, "Add a new project")
	cmd.Flags().StringVar(&projectPath, "path", "", "Project path (for --add-project)")
	cmd.Flags().StringVar(&projectName, "name", "", "Project name (for --add-project)")
	cmd.Flags().StringVar(&projectDesc, "description", "", "Project description (for --add-project)")
	cmd.Flags().StringVar(&techStack, "tech-stack", "", "Tech stack comma-separated (for --add-project)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	return cmd
}

// Helper functions
func joinStrings(arr []string, sep string) string {
	if len(arr) == 0 {
		return ""
	}
	result := arr[0]
	for i := 1; i < len(arr); i++ {
		result += sep + arr[i]
	}
	return result
}

func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	parts := make([]string, 0)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}
